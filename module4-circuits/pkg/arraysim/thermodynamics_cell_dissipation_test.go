package arraysim

import (
	"math"
	"testing"
)

func TestThermodynamicsCellDissipation_4x4_NoNegativePower(t *testing.T) {
	t.Parallel()

	params := SolveParams{
		WLVoltages: []float64{0.30, 0.25, 0.20, 0.15},
		BLVoltages: []float64{0.00, 0.02, 0.03, 0.04},
		Conductance: [][]float64{
			{30e-6, 40e-6, 50e-6, 60e-6},
			{35e-6, 45e-6, 55e-6, 65e-6},
			{32e-6, 42e-6, 52e-6, 62e-6},
			{38e-6, 48e-6, 58e-6, 68e-6},
		},
		Wire: WireParams{RWordLine: 25, RBitLine: 25},
		Boundary: BoundaryParams{
			WLDriveResistance: 15,
			BLDriveResistance: 15,
		},
	}

	res, err := NewTierASolver().Solve(params)
	if err != nil {
		t.Fatalf("Tier-A solve failed: %v", err)
	}

	for r := 0; r < len(res.CellCurrents); r++ {
		for c := 0; c < len(res.CellCurrents[r]); c++ {
			gEff := effectiveCellConductance(params, r, c)
			if gEff <= 0 {
				continue
			}
			i := res.CellCurrents[r][c]
			v := res.CellVoltages[r][c]
			rEq := 1.0 / gEff

			pI2R := i * i * rEq
			pVI := v * i

			if pI2R < -1e-21 {
				t.Fatalf("negative I^2R at (%d,%d): %g", r, c, pI2R)
			}
			if pVI < -1e-21 {
				t.Fatalf("negative V*I at (%d,%d): %g", r, c, pVI)
			}

			if math.Abs(pI2R-pVI) > 1e-12 {
				t.Fatalf("power cross-check mismatch at (%d,%d): I^2R=%g V*I=%g", r, c, pI2R, pVI)
			}
		}
	}
}
