// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import (
	"fmt"
	"math"
)

// MVMUncertaintyResult wraps the MVM output with optional per-element
// uncertainty estimates and ADC saturation metadata.
//
// Uncertainty is computed via first-order error propagation (linear
// approximation):
//
//	sigma_out_i = sqrt( sum_j( (g_ij * sigma_input_j)^2
//	                         + (input_j * sigma_g_ij)^2 ) )
//
// where sigma_g_ij is the per-cell conductance standard deviation
// derived from process variation config, and sigma_input_j is the
// DAC quantization noise (+/-0.5 LSB as uniform RMS).
//
// This is additive to the existing MVM() API and does not modify any
// existing signatures.
type MVMUncertaintyResult struct {
	// Output contains the standard MVM output (same as MVM()).
	Output []float64

	// Uncertainty contains per-element 1-sigma standard deviation of the
	// output. If the array has no variation config (NoiseLevel == 0 and
	// ProcessVariation == nil), all entries will be zero.
	Uncertainty []float64

	// Saturated counts the number of output elements that hit the ADC
	// rail (quantized to 0.0 or 1.0 before ADC rounding). This indicates
	// potential information loss due to ADC dynamic range limits.
	Saturated int
}

// MVMWithUncertainty performs matrix-vector multiplication and estimates
// per-output uncertainty using first-order (linear) error propagation.
//
// The method runs the standard MVM pipeline and then computes an
// analytical uncertainty estimate. It does NOT re-sample noise;
// the uncertainty is deterministic for a given array state.
//
// The three uncertainty sources modeled are:
//   - Device variation: per-cell conductance sigma from NoiseFactor
//   - DAC quantization noise: +/-0.5 LSB modeled as uniform RMS
//   - ADC saturation: flagged per output element
//
// Numerical error (solver discretization) is not included here as it
// applies to the physics engine, not the MVM pipeline itself.
//
// This method acquires a read lock and is safe for concurrent use.
func (a *Array) MVMWithUncertainty(input []float64) (*MVMUncertaintyResult, error) {
	// Run the standard MVM to get the output values.
	output, err := a.MVM(input)
	if err != nil {
		return nil, fmt.Errorf("MVMWithUncertainty: %w", err)
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	rows := a.config.Rows
	cols := len(input)

	// Compute DAC quantization noise sigma.
	// Uniform quantization over N levels produces RMS noise of 0.5 LSB.
	// sigma_dac = (1 / (dacLevels - 1)) * (1 / (2 * sqrt(3)))
	// For uniform distribution over [-0.5*LSB, +0.5*LSB], sigma = LSB / sqrt(12).
	dacSigma := 0.0
	if a.dacLevels > 1 {
		lsb := 1.0 / float64(a.dacLevels-1)
		dacSigma = lsb / math.Sqrt(12.0)
	}

	// Determine per-cell conductance sigma.
	// The NoiseFactor for each cell was drawn as 1 + sigma * N(0,1) at init.
	// The "sigma" used is either ProcessVariation.DeviceSigma or NoiseLevel.
	// For uncertainty propagation we use the configured sigma as the device
	// variation standard deviation (relative to nominal conductance).
	deviceSigma := a.config.NoiseLevel
	if a.config.ProcessVariation != nil {
		deviceSigma = a.config.ProcessVariation.DeviceSigma
	}

	uncertainty := make([]float64, rows)

	// Normalization factor matching mvmCPU: maxCurrent = cols.
	maxCurrent := float64(cols)
	if maxCurrent == 0 {
		maxCurrent = 1.0
	}

	for i := 0; i < rows; i++ {
		var varianceSum float64
		for j := 0; j < cols; j++ {
			gNominal := a.cells[i][j].Conductance

			// DAC input quantized value (reconstruct what mvmCPU used).
			dacIn := a.quantizeDAC(input[j])

			// Term 1: input uncertainty propagated through conductance.
			// d(output_i)/d(input_j) = g_ij / maxCurrent
			// variance contribution = (g_ij / maxCurrent * sigma_input_j)^2
			if dacSigma > 0 {
				term := gNominal * dacSigma / maxCurrent
				varianceSum += term * term
			}

			// Term 2: conductance uncertainty propagated through input.
			// d(output_i)/d(g_ij) = input_j / maxCurrent
			// sigma_g_ij = g_nominal * deviceSigma (relative sigma)
			// variance contribution = (input_j / maxCurrent * g_nominal * deviceSigma)^2
			if deviceSigma > 0 {
				sigmaG := gNominal * deviceSigma
				term := dacIn * sigmaG / maxCurrent
				varianceSum += term * term
			}
		}
		uncertainty[i] = math.Sqrt(varianceSum)
	}

	// Count ADC-saturated outputs.
	// ADC saturation occurs when the pre-quantization normalized current
	// is at or beyond the ADC rail (0.0 or 1.0 after Clamp01). Since the
	// MVM output is already quantized, we detect saturation by checking
	// whether the output equals the minimum or maximum ADC code.
	saturated := 0
	for _, v := range output {
		if v <= 0.0 || v >= 1.0 {
			saturated++
		}
	}

	return &MVMUncertaintyResult{
		Output:      output,
		Uncertainty: uncertainty,
		Saturated:   saturated,
	}, nil
}
