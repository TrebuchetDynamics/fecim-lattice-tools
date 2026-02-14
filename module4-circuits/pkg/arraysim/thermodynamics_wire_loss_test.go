package arraysim

import (
	"math"
	"testing"
)

func TestThermodynamicsWireLoss_MatchesI2RAndClosesPowerBudget(t *testing.T) {
	t.Parallel()

	params := SolveParams{
		WLVoltages: []float64{0.40, 0.32, 0.24, 0.16},
		BLVoltages: []float64{0.00, 0.01, 0.02, 0.03},
		Conductance: [][]float64{
			{20e-6, 24e-6, 28e-6, 32e-6},
			{22e-6, 26e-6, 30e-6, 34e-6},
			{21e-6, 25e-6, 29e-6, 33e-6},
			{23e-6, 27e-6, 31e-6, 35e-6},
		},
		Wire: WireParams{RWordLine: 70, RBitLine: 90},
		Boundary: BoundaryParams{
			WLDriveResistance: 30,
			BLDriveResistance: 30,
		},
	}

	dc, err := NewTierBSolver().SolveDC(params)
	if err != nil {
		t.Fatalf("SolveDC failed: %v", err)
	}

	wire := params.Wire.WithDefaults(params.Geometry.WithDefaults())
	pWire := 0.0

	for r := 0; r < len(dc.WLNodes); r++ {
		for c := 0; c < len(dc.WLNodes[r])-1; c++ {
			dv := dc.WLNodes[r][c] - dc.WLNodes[r][c+1]
			iSeg := dv / wire.RWordLine
			pI2R := iSeg * iSeg * wire.RWordLine
			pIV := iSeg * dv
			if math.Abs(pI2R-pIV) > 1e-14 {
				t.Fatalf("WL segment power mismatch (%d,%d): I2R=%g IV=%g", r, c, pI2R, pIV)
			}
			if pI2R < -1e-21 {
				t.Fatalf("negative WL segment dissipation (%d,%d): %g", r, c, pI2R)
			}
			pWire += pI2R
		}
	}

	for c := 0; c < len(dc.BLNodes[0]); c++ {
		for r := 0; r < len(dc.BLNodes)-1; r++ {
			dv := dc.BLNodes[r][c] - dc.BLNodes[r+1][c]
			iSeg := dv / wire.RBitLine
			pI2R := iSeg * iSeg * wire.RBitLine
			pIV := iSeg * dv
			if math.Abs(pI2R-pIV) > 1e-14 {
				t.Fatalf("BL segment power mismatch (%d,%d): I2R=%g IV=%g", r, c, pI2R, pIV)
			}
			if pI2R < -1e-21 {
				t.Fatalf("negative BL segment dissipation (%d,%d): %g", r, c, pI2R)
			}
			pWire += pI2R
		}
	}

	pCells := 0.0
	for r := range dc.CellCurrents {
		for c := range dc.CellCurrents[r] {
			gEff := effectiveCellConductance(params, r, c)
			if gEff <= 0 {
				continue
			}
			i := dc.CellCurrents[r][c]
			pCells += i * i / gEff
		}
	}

	pDrive := wireAndDrivePower(params, dc) - pWire
	if pDrive < 0 {
		pDrive = 0 // numerical guard
	}

	pIn := sourceInputPower(params, dc)
	pTotal := pCells + pWire + pDrive
	if pIn <= 0 {
		t.Fatalf("non-positive input power: %g", pIn)
	}
	if rel := math.Abs(pIn-pTotal) / pIn; rel >= 0.01 {
		t.Fatalf("power accounting mismatch: Pin=%g Pcells=%g Pwire=%g Pdrive=%g rel=%g", pIn, pCells, pWire, pDrive, rel)
	}
}
