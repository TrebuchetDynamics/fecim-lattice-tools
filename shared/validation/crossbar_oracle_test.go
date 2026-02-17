package validation

import (
	"math"
	"testing"
)

// ── Helpers ──────────────────────────────────────────────────────────────────

func uniformConductances(n, m int, g float64) [][]float64 {
	G := make([][]float64, n)
	for i := range G {
		G[i] = make([]float64, m)
		for j := range G[i] {
			G[i][j] = g
		}
	}
	return G
}

// naiveMVM is the simplest possible behavioural approximation: I[i] = Σ G[i][j]·V[j].
// Used as baseline; the oracle must agree with this exactly when wireR == 0.
func naiveMVM(G [][]float64, V []float64) []float64 {
	N := len(G)
	I := make([]float64, N)
	for i := range G {
		for j, v := range V {
			I[i] += G[i][j] * v
		}
	}
	return I
}

func maxAbsDiff(a, b []float64) float64 {
	d := 0.0
	for i := range a {
		if v := math.Abs(a[i] - b[i]); v > d {
			d = v
		}
	}
	return d
}

// ── Unit tests ───────────────────────────────────────────────────────────────

// TestExactMVM_IdealWires checks that the oracle equals the naïve MVM for wireR=0.
func TestExactMVM_IdealWires(t *testing.T) {
	G := uniformConductances(4, 4, 50e-6)
	V := []float64{1.0, 0.5, 0.25, 0.1}

	res, err := ExactMVM(G, V, 0)
	if err != nil {
		t.Fatalf("ExactMVM error: %v", err)
	}

	expected := naiveMVM(G, V)
	if d := maxAbsDiff(res.RowCurrents, expected); d > 1e-15 {
		t.Errorf("ideal wires: max deviation %.3e A (want < 1e-15)", d)
	}
}

// TestExactMVM_SingleCell verifies Ohm's law for a 1×1 crossbar.
func TestExactMVM_SingleCell(t *testing.T) {
	G := [][]float64{{100e-6}} // 100 µS
	V := []float64{1.0}

	res, err := ExactMVM(G, V, 0)
	if err != nil {
		t.Fatalf("ExactMVM error: %v", err)
	}
	want := 100e-6 // 100 µA
	if math.Abs(res.RowCurrents[0]-want) > 1e-15 {
		t.Errorf("single cell: got %.4e A, want %.4e A", res.RowCurrents[0], want)
	}
}

// TestExactMVM_ZeroInput checks that zero input yields zero output.
func TestExactMVM_ZeroInput(t *testing.T) {
	G := uniformConductances(4, 4, 50e-6)
	V := []float64{0, 0, 0, 0}

	res, err := ExactMVM(G, V, 0)
	if err != nil {
		t.Fatalf("ExactMVM error: %v", err)
	}
	for i, I := range res.RowCurrents {
		if math.Abs(I) > 1e-20 {
			t.Errorf("row %d: got %.3e A, want 0", i, I)
		}
	}
}

// TestExactMVM_WireResistance_ReducesOutputCurrent verifies that adding wire
// resistance reduces the output current compared to the ideal case.
func TestExactMVM_WireResistance_ReducesOutputCurrent(t *testing.T) {
	G := uniformConductances(8, 8, 50e-6)
	V := make([]float64, 8)
	for j := range V {
		V[j] = 1.0
	}

	ideal, err := ExactMVM(G, V, 0)
	if err != nil {
		t.Fatalf("ideal ExactMVM: %v", err)
	}
	// 2.5 Ω matches the default wire resistance in shared/crossbar IRDropSimulator.
	lossy, err := ExactMVM(G, V, 2.5)
	if err != nil {
		t.Fatalf("lossy ExactMVM: %v", err)
	}

	for i := range ideal.RowCurrents {
		if lossy.RowCurrents[i] >= ideal.RowCurrents[i] {
			t.Errorf("row %d: expected wire loss to reduce current, got %.4e >= %.4e",
				i, lossy.RowCurrents[i], ideal.RowCurrents[i])
		}
	}
}

// TestExactMVM_WireResistance_SmallError checks that 2.5 Ω/segment wire resistance
// causes less than 30% error vs ideal for an 8×8 array at G=50µS.
// (Validates that the oracle reports physically plausible IR-drop magnitude.)
func TestExactMVM_WireResistance_SmallError(t *testing.T) {
	G := uniformConductances(8, 8, 50e-6)
	V := make([]float64, 8)
	for j := range V {
		V[j] = 1.0
	}

	ideal, _ := ExactMVM(G, V, 0)
	lossy, err := ExactMVM(G, V, 2.5)
	if err != nil {
		t.Fatalf("lossy ExactMVM: %v", err)
	}

	for i, I0 := range ideal.RowCurrents {
		relErr := math.Abs(lossy.RowCurrents[i]-I0) / I0
		if relErr > 0.30 {
			t.Errorf("row %d: relative error %.1f%% > 30%%", i, relErr*100)
		}
	}
}

// TestExactMVM_KCLConservation verifies that the sum of all cell currents equals
// the sum of all row output currents (current conservation).
func TestExactMVM_KCLConservation(t *testing.T) {
	G := uniformConductances(6, 6, 30e-6)
	V := []float64{1.0, 0.8, 0.6, 0.4, 0.2, 0.1}

	for _, wireR := range []float64{0, 1.0, 2.5} {
		res, err := ExactMVM(G, V, wireR)
		if err != nil {
			t.Fatalf("wireR=%.1f: ExactMVM error: %v", wireR, err)
		}

		var sumCell, sumRow float64
		for i := range res.CellCurrents {
			for j := range res.CellCurrents[i] {
				sumCell += res.CellCurrents[i][j]
			}
		}
		for _, I := range res.RowCurrents {
			sumRow += I
		}

		// For ideal wires, sumCell == sumRow exactly.
		// For lossy wires, some current is dissipated in row wires between
		// the cells and the sensing node, so sumCell >= sumRow.
		if wireR == 0 {
			if math.Abs(sumCell-sumRow) > 1e-12 {
				t.Errorf("wireR=0: KCL violation: cell sum %.6e != row sum %.6e", sumCell, sumRow)
			}
		} else {
			if sumCell < sumRow-1e-12 {
				t.Errorf("wireR=%.1f: more current at sense node (%.6e) than through cells (%.6e)",
					wireR, sumRow, sumCell)
			}
		}
	}
}

// TestExactMVM_IdealVsNaive_RandomMatrix validates oracle against naïve MVM for
// a variety of array sizes with wireR=0.
func TestExactMVM_IdealVsNaive_RandomMatrix(t *testing.T) {
	cases := []struct{ n, m int }{{2, 2}, {4, 8}, {8, 4}, {8, 8}, {16, 16}}

	for _, tc := range cases {
		G := make([][]float64, tc.n)
		for i := range G {
			G[i] = make([]float64, tc.m)
			for j := range G[i] {
				// Simple deterministic conductance pattern
				G[i][j] = (float64(i*tc.m+j+1) * 10e-6) // 10–(n·m)×10 µS
			}
		}
		V := make([]float64, tc.m)
		for j := range V {
			V[j] = float64(j+1) * 0.1 // 0.1 – m×0.1 V
		}

		res, err := ExactMVM(G, V, 0)
		if err != nil {
			t.Fatalf("%dx%d: %v", tc.n, tc.m, err)
		}
		naive := naiveMVM(G, V)
		if d := maxAbsDiff(res.RowCurrents, naive); d > 1e-13 {
			t.Errorf("%dx%d: max deviation %.3e A (want < 1e-13)", tc.n, tc.m, d)
		}
	}
}

// TestExactMVM_DimensionValidation checks error handling for bad inputs.
func TestExactMVM_DimensionValidation(t *testing.T) {
	_, err := ExactMVM(nil, nil, 0)
	if err == nil {
		t.Error("expected error for nil conductances")
	}

	G := [][]float64{{1e-6, 2e-6}}
	_, err = ExactMVM(G, []float64{1.0}, 0) // wrong number of inputs
	if err == nil {
		t.Error("expected error for input-column mismatch")
	}
}

// TestGaussianEliminate_Identity verifies the solver on a known system.
func TestGaussianEliminate_Identity(t *testing.T) {
	A := [][]float64{
		{2, 1},
		{1, 3},
	}
	b := []float64{5, 10}
	// Expected: x[0] = 1, x[1] = 3  (2·1 + 1·3 = 5, 1·1 + 3·3 = 10)

	x, err := gaussianEliminate(A, b, 2)
	if err != nil {
		t.Fatalf("gaussianEliminate: %v", err)
	}
	if math.Abs(x[0]-1) > 1e-12 || math.Abs(x[1]-3) > 1e-12 {
		t.Errorf("solution: got [%.6f, %.6f], want [1, 3]", x[0], x[1])
	}
}
