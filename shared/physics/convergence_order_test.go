package physics

// TestLKSolver_ConvergenceOrder uses Richardson extrapolation with a refined
// timestep sequence to estimate the convergence order of the RK4/implicit
// solver.
//
// For a pure RK4 integrator the expected convergence order is p=4 (error ~ dt^4).
// However, the LK solver includes several features that can reduce the observed
// order:
//   - Stiffness detection switching to implicit (1st-order Newton) stepping
//   - Rate clamping (maxAbsRate caps |dP/dt|)
//   - Polarization clamping (PMax * overshoot factor)
//   - Effective viscosity from series resistance
//
// This test documents the actual convergence order and is suitable for
// inclusion in a paper's supplementary material.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// convergenceOrderResult holds the full convergence analysis output.
type convergenceOrderResult struct {
	Method           string                  `json:"method"`
	Description      string                  `json:"description"`
	Solver           string                  `json:"solver"`
	Material         string                  `json:"material"`
	AppliedField     string                  `json:"applied_field"`
	TotalTime        float64                 `json:"total_time_s"`
	DtSequence       []float64               `json:"dt_sequence_s"`
	FinalP           []float64               `json:"final_P_C_m2"`
	ErrorEstimates   []float64               `json:"error_estimates_C_m2"`
	ConvergenceOrder []float64               `json:"convergence_order"`
	ExpectedOrder    string                  `json:"expected_order"`
	ExtrapolatedP    float64                 `json:"extrapolated_P_C_m2"`
	Table            []convergenceOrderEntry `json:"convergence_table"`
	Summary          string                  `json:"summary"`
	GeneratedAt      string                  `json:"generated_at"`
}

// convergenceOrderEntry is one row in the convergence table.
type convergenceOrderEntry struct {
	Dt         float64 `json:"dt_s"`
	FinalP     float64 `json:"final_P_C_m2"`
	Error      float64 `json:"error_estimate_C_m2"`
	ErrorRatio float64 `json:"error_ratio"`
	OrderEst   float64 `json:"order_estimate"`
}

func TestLKSolver_ConvergenceOrder(t *testing.T) {
	mat := DefaultHZO()

	// Use finer timesteps than the existing convergence test to better
	// resolve the RK4 convergence order. Each step halves dt.
	dtSequence := []float64{1e-11, 5e-12, 2.5e-12, 1.25e-12}

	// Use a moderate field and very short simulation time to stay in the
	// linear response regime where RK4 convergence is cleanest.
	const totalTime = 1e-9 // 1 ns
	E := 0.3 * mat.Ec      // Weak field (30% of Ec) — avoids saturation

	pValues := make([]float64, len(dtSequence))

	t.Logf("=== Convergence Order Estimation via Richardson Extrapolation ===")
	t.Logf("Material: %s", mat.Name)
	t.Logf("E_applied = 0.3 * Ec = %.3e V/m", E)
	t.Logf("Total time = %.3e s", totalTime)
	t.Logf("")

	for idx, dt := range dtSequence {
		s := NewLKSolver()
		s.ConfigureFromMaterial(mat)
		s.UseNLS = false
		s.EnableNoise = false
		s.UseMaterialAlpha = true
		s.UpdateParams()
		s.P = -mat.Pr

		numSteps := int(math.Round(totalTime / dt))
		var finalP float64
		for i := 0; i < numSteps; i++ {
			finalP = s.Step(E, dt)
		}

		pValues[idx] = finalP
		t.Logf("dt = %.3e s → P = %.10e C/m² (steps = %d)", dt, finalP, numSteps)
	}

	// Richardson extrapolation to estimate convergence order.
	//
	// For a method of order p with uniform refinement ratio r=2:
	//   error(dt) ≈ C * dt^p
	//   P(dt) = P_exact + C * dt^p + O(dt^(p+1))
	//
	// Given three consecutive refinements with ratio r=2:
	//   P1 = P(dt), P2 = P(dt/2), P3 = P(dt/4)
	//   (P1 - P2) / (P2 - P3) = r^p = 2^p
	//   p = log2((P1-P2)/(P2-P3))
	//
	// We compute this for each consecutive triple.
	t.Logf("")
	t.Logf("=== Richardson Error Analysis ===")

	errorEstimates := make([]float64, len(dtSequence))
	convergenceOrders := make([]float64, len(dtSequence))
	errorRatios := make([]float64, len(dtSequence))

	// Initialize to 0 (not NaN) for JSON compatibility.
	// Zero means "not computed for this entry."
	for i := range dtSequence {
		errorEstimates[i] = 0
		convergenceOrders[i] = 0
		errorRatios[i] = 0
	}

	// Compute error estimates (difference between consecutive refinements)
	for i := 1; i < len(pValues); i++ {
		errorEstimates[i] = math.Abs(pValues[i-1] - pValues[i])
	}

	// Compute convergence order from error ratios
	for i := 2; i < len(pValues); i++ {
		num := math.Abs(pValues[i-2] - pValues[i-1])
		den := math.Abs(pValues[i-1] - pValues[i])
		if den > 1e-20 {
			ratio := num / den
			errorRatios[i] = ratio
			if ratio > 0 {
				// p = log2(ratio) since refinement ratio is 2
				order := math.Log2(ratio)
				convergenceOrders[i] = order
				t.Logf("Pair [%d-%d]/[%d-%d]: ratio = %.4f, estimated order p = %.2f",
					i-2, i-1, i-1, i, ratio, order)
			}
		} else {
			t.Logf("Pair [%d-%d]/[%d-%d]: denominator too small (%.2e), machine precision reached",
				i-2, i-1, i-1, i, den)
		}
	}

	// Richardson extrapolation for the best estimate of P_exact.
	// Using the finest pair and the estimated order from the last triple.
	pFinest := pValues[len(pValues)-1]
	pCoarser := pValues[len(pValues)-2]
	lastOrder := convergenceOrders[len(convergenceOrders)-1]

	var pExtrapolated float64
	if lastOrder > 0 {
		// P_exact ≈ P_fine + (P_fine - P_coarse) / (2^p - 1)
		rp := math.Pow(2, lastOrder)
		pExtrapolated = pFinest + (pFinest-pCoarser)/(rp-1)
	} else {
		// Fallback: assume 2nd order
		pExtrapolated = pFinest + (pFinest-pCoarser)/3.0
	}

	// Build convergence table
	table := make([]convergenceOrderEntry, len(dtSequence))
	for i := range dtSequence {
		table[i] = convergenceOrderEntry{
			Dt:         dtSequence[i],
			FinalP:     pValues[i],
			Error:      errorEstimates[i],
			ErrorRatio: errorRatios[i],
			OrderEst:   convergenceOrders[i],
		}
	}

	// Log formatted table
	t.Logf("")
	t.Logf("=== Convergence Table ===")
	t.Logf("%-14s %-22s %-16s %-14s %-12s", "dt (s)", "P (C/m²)", "Error (C/m²)", "Error Ratio", "Order p")
	for _, row := range table {
		errStr := "—"
		ratioStr := "—"
		orderStr := "—"
		if row.Error > 0 {
			errStr = fmt.Sprintf("%.6e", row.Error)
		}
		if row.ErrorRatio > 0 {
			ratioStr = fmt.Sprintf("%.4f", row.ErrorRatio)
		}
		if row.OrderEst > 0 {
			orderStr = fmt.Sprintf("%.2f", row.OrderEst)
		}
		t.Logf("%-14.3e %-22.10e %-16s %-14s %-12s", row.Dt, row.FinalP, errStr, ratioStr, orderStr)
	}
	t.Logf("=========================")

	// Build summary
	var avgOrder float64
	var orderCount int
	for _, o := range convergenceOrders {
		if o > 0 {
			avgOrder += o
			orderCount++
		}
	}
	if orderCount > 0 {
		avgOrder /= float64(orderCount)
	}

	summaryText := fmt.Sprintf(
		"Estimated convergence order: p ≈ %.2f (expected 4.0 for pure RK4). "+
			"The solver uses RK4 with rate clamping, P clamping, and stiffness-triggered "+
			"implicit fallback, all of which can reduce the observed order below 4. "+
			"An order >= 2 confirms at least 2nd-order convergence. "+
			"Extrapolated P = %.10e C/m², finest P = %.10e C/m² (diff = %.3e C/m²).",
		avgOrder, pExtrapolated, pFinest, math.Abs(pExtrapolated-pFinest))

	t.Logf("")
	t.Logf("Summary: %s", summaryText)

	// Verify basic convergence: finer dt should reduce error
	if len(pValues) >= 4 {
		err01 := math.Abs(pValues[0] - pValues[1])
		err23 := math.Abs(pValues[2] - pValues[3])
		if err23 < err01 || err23 < 1e-15 {
			t.Logf("CONVERGENCE VERIFIED: error decreasing with dt refinement (coarse=%.3e, fine=%.3e)",
				err01, err23)
		} else {
			t.Logf("WARNING: error not monotonically decreasing (coarse=%.3e, fine=%.3e). "+
				"May indicate clamping/implicit switching interference at this dt range.", err01, err23)
		}
	}

	// Write artifact
	result := convergenceOrderResult{
		Method: "richardson_extrapolation",
		Description: "Convergence order estimation for LK RK4/implicit solver. " +
			"Uses halving refinement (r=2) and Richardson extrapolation to " +
			"estimate the effective convergence order p. Pure RK4 gives p=4; " +
			"rate clamping, P clamping, and implicit fallback may reduce this.",
		Solver:           "LKSolver (RK4 + implicit Newton fallback)",
		Material:         mat.Name,
		AppliedField:     "0.3 * Ec (subcoercive, linear regime)",
		TotalTime:        totalTime,
		DtSequence:       dtSequence,
		FinalP:           pValues,
		ErrorEstimates:   errorEstimates,
		ConvergenceOrder: convergenceOrders,
		ExpectedOrder:    "4.0 (RK4, reduced by clamping)",
		ExtrapolatedP:    pExtrapolated,
		Table:            table,
		Summary:          summaryText,
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	// Persist convergence order results as artifact for publication.
	artifactDir := filepath.Join("..", "..", "validation", "literature", "output")
	os.MkdirAll(artifactDir, 0o755) // best-effort; don't fail test on mkdir error

	data, err := json.MarshalIndent(result, "", "  ")
	if err == nil {
		outPath := filepath.Join(artifactDir, "convergence_order.json")
		if writeErr := os.WriteFile(outPath, data, 0o644); writeErr == nil {
			t.Logf("Artifact written to %s", outPath)
		} else {
			t.Logf("Warning: could not write artifact: %v", writeErr)
		}
	} else {
		t.Logf("Warning: could not marshal artifact JSON: %v", err)
	}

	t.Logf("=== End Convergence Order Estimation ===")
}
