package peripherals

import (
	"fmt"
	"math"
	"testing"
)

func TestPeripheralChainDACCrossbarTIAADC(t *testing.T) {
	// 4x4 programmed crossbar conductance matrix (Siemens)
	conductance := [][]float64{
		{2e-6, 4e-6, 1e-6, 3e-6},
		{5e-6, 2e-6, 4e-6, 1e-6},
		{3e-6, 6e-6, 2e-6, 5e-6},
		{1e-6, 3e-6, 5e-6, 2e-6},
	}

	// Digital input vector to be encoded by the DAC (normalized 0..1)
	input := []float64{0.15, 0.55, 0.85, 0.35}

	dac := DefaultDAC()
	dac.Bits = 8
	dac.VrefLow = 0.0
	dac.VrefHigh = 1.0

	tia := DefaultTIA()
	tia.Gain = 25e3
	tia.OutputOffset = 0.02
	tia.MaxOutputVoltage = 1.0

	// Step 2: Generate DAC codes and reconstructed analog voltages.
	dacCodes := make([]int, len(input))
	dacVoltages := make([]float64, len(input))
	for i, x := range input {
		code := int(math.Round(x * float64(dac.Levels()-1)))
		dacCodes[i] = code
		dacVoltages[i] = dac.Convert(code)
	}

	// Step 3: Run MVM through 4x4 crossbar (I = G * V).
	mvmCurrents := crossbarMVM(conductance, dacVoltages)

	// Reference path for expected result (using ideal input, no DAC quantization).
	idealCurrents := crossbarMVM(conductance, input)

	for _, bits := range []int{4, 6, 8} {
		bits := bits
		t.Run(fmt.Sprintf("adc_bits_%d", bits), func(t *testing.T) {
			adc := DefaultADC()
			adc.Bits = bits
			adc.VrefLow = 0.0
			adc.VrefHigh = 1.0

			for row := 0; row < 4; row++ {
				// Step 4: TIA conversion for actual and expected/ideal currents.
				vTIA := tia.Convert(mvmCurrents[row])
				vIdealTIA := tia.Convert(idealCurrents[row])

				// Step 5: ADC quantization.
				gotCode := adc.Convert(vTIA)
				expectedCode := adc.Convert(vIdealTIA)

				// Step 6: Verify final digital output is within ADC resolution.
				codeDiff := gotCode - expectedCode
				if codeDiff < 0 {
					codeDiff = -codeDiff
				}
				if codeDiff > 1 {
					t.Fatalf("row %d (ADC %db): got code %d, expected %d (|Δ|=%d > 1 LSB); dac_code=%d, dac_v=%.6f V",
						row, bits, gotCode, expectedCode, codeDiff, dacCodes[row], dacVoltages[row])
				}

				// Optional analog-domain sanity: error must stay within one ADC LSB.
				analogErr := math.Abs(vTIA - vIdealTIA)
				if analogErr > adc.Resolution() {
					t.Fatalf("row %d (ADC %db): analog error %.6e V exceeds ADC resolution %.6e V",
						row, bits, analogErr, adc.Resolution())
				}
			}
		})
	}
}

func crossbarMVM(g [][]float64, v []float64) []float64 {
	out := make([]float64, len(g))
	for i := range g {
		var sum float64
		for j := range g[i] {
			sum += g[i][j] * v[j]
		}
		out[i] = sum
	}
	return out
}
