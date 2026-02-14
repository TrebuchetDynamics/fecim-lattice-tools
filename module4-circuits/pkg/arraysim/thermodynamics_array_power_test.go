package arraysim

import (
	"fmt"
	"math"
	"testing"
)

func TestThermodynamicsArrayPowerBudget_8x8_Compute(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		selector1T1R bool
	}{
		{name: "0T1R", selector1T1R: false},
		{name: "1T1R", selector1T1R: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows, cols := 8, 8
			g := make([][]float64, rows)
			for r := 0; r < rows; r++ {
				g[r] = make([]float64, cols)
				for c := 0; c < cols; c++ {
					g[r][c] = 20e-6 + float64((r+c)%5)*5e-6 // deterministic known matrix
				}
			}

			wl := []float64{0.45, 0.40, 0.35, 0.30, 0.25, 0.20, 0.15, 0.10}
			bl := make([]float64, cols) // grounded compute columns for clean power accounting

			params := SolveParams{
				WLVoltages:  wl,
				BLVoltages:  bl,
				Conductance: g,
				Wire: WireParams{
					RWordLine: 32,
					RBitLine:  41,
				},
				Boundary: BoundaryParams{
					WLDriveResistance: 18,
					BLDriveResistance: 18,
				},
			}
			if tc.selector1T1R {
				params.Selector.Enabled = true
				params.Selector.OnConductance = 1.0 / 5e3 // 5k ohm series selector path
				params.Selector.OffConductance = math.Inf(1)
			}

			tierA := NewTierASolver()
			resA, err := tierA.Solve(params)
			if err != nil {
				t.Fatalf("Tier-A solve failed: %v", err)
			}
			dc, err := NewTierBSolver().SolveDC(params)
			if err != nil {
				t.Fatalf("SolveDC failed: %v", err)
			}

			// Tier-A and DC per-cell outputs should agree.
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					if math.Abs(resA.CellVoltages[r][c]-dc.CellVoltages[r][c]) > 1e-8 {
						t.Fatalf("cell voltage mismatch (%d,%d): tierA=%g dc=%g", r, c, resA.CellVoltages[r][c], dc.CellVoltages[r][c])
					}
					if math.Abs(resA.CellCurrents[r][c]-dc.CellCurrents[r][c]) > 1e-11 {
						t.Fatalf("cell current mismatch (%d,%d): tierA=%g dc=%g", r, c, resA.CellCurrents[r][c], dc.CellCurrents[r][c])
					}
				}
			}

			pCells := 0.0
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					gEff := effectiveCellConductance(params, r, c)
					if gEff <= 0 {
						continue
					}
					i := dc.CellCurrents[r][c]
					rEq := 1.0 / gEff
					p := i * i * rEq
					if p < -1e-21 {
						t.Fatalf("negative cell dissipation at (%d,%d): %g", r, c, p)
					}
					if p < 0 {
						p = 0
					}
					pCells += p
				}
			}

			pWiresAndDrives := wireAndDrivePower(params, dc)
			pTotal := pCells + pWiresAndDrives
			pIn := sourceInputPower(params, dc)
			if pIn <= 0 {
				t.Fatalf("expected positive input power, got %g", pIn)
			}

			relErr := math.Abs(pIn-pTotal) / pIn
			if relErr >= 0.01 {
				t.Fatalf("power budget mismatch: Pin=%g Pcells=%g Pwires+drive=%g Ptotal=%g relErr=%g", pIn, pCells, pWiresAndDrives, pTotal, relErr)
			}
		})
	}
}

func wireAndDrivePower(params SolveParams, dc DCResult) float64 {
	rows := len(dc.WLNodes)
	if rows == 0 {
		return 0
	}
	cols := len(dc.WLNodes[0])
	wire := params.Wire.WithDefaults(params.Geometry.WithDefaults())
	boundary := params.Boundary.WithDefaults(wire)

	total := 0.0

	for r := 0; r < rows; r++ {
		for c := 0; c < cols-1; c++ {
			dv := dc.WLNodes[r][c] - dc.WLNodes[r][c+1]
			total += (dv * dv) / wire.RWordLine
		}
	}
	for c := 0; c < cols; c++ {
		for r := 0; r < rows-1; r++ {
			dv := dc.BLNodes[r][c] - dc.BLNodes[r+1][c]
			total += (dv * dv) / wire.RBitLine
		}
	}

	for r := 0; r < rows; r++ {
		src := 0.0
		if r < len(params.WLVoltages) {
			src = params.WLVoltages[r]
		}
		dv := src - dc.WLNodes[r][0]
		total += (dv * dv) / boundary.WLDriveResistance
	}
	for c := 0; c < cols; c++ {
		src := 0.0
		if c < len(params.BLVoltages) {
			src = params.BLVoltages[c]
		}
		dv := src - dc.BLNodes[0][c]
		total += (dv * dv) / boundary.BLDriveResistance
	}

	return total
}

func sourceInputPower(params SolveParams, dc DCResult) float64 {
	rows := len(dc.WLNodes)
	if rows == 0 {
		return 0
	}
	cols := len(dc.WLNodes[0])
	wire := params.Wire.WithDefaults(params.Geometry.WithDefaults())
	boundary := params.Boundary.WithDefaults(wire)

	pIn := 0.0
	for r := 0; r < rows; r++ {
		vsrc := 0.0
		if r < len(params.WLVoltages) {
			vsrc = params.WLVoltages[r]
		}
		isrc := (vsrc - dc.WLNodes[r][0]) / boundary.WLDriveResistance
		pIn += vsrc * isrc
	}
	for c := 0; c < cols; c++ {
		vsrc := 0.0
		if c < len(params.BLVoltages) {
			vsrc = params.BLVoltages[c]
		}
		isrc := (vsrc - dc.BLNodes[0][c]) / boundary.BLDriveResistance
		pIn += vsrc * isrc
	}
	if math.IsNaN(pIn) || math.IsInf(pIn, 0) {
		panic(fmt.Sprintf("invalid input power: %g", pIn))
	}
	return pIn
}
