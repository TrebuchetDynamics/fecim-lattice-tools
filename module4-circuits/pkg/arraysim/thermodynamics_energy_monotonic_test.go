package arraysim

import "testing"

func TestThermodynamicsEnergyMonotonic_WriteISPPPulses(t *testing.T) {
	t.Parallel()

	// 4x4 array with a target WRITE location; emulate ISPP by increasing pulse amplitude.
	rows, cols := 4, 4
	targetR, targetC := 1, 2
	g := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		g[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			g[r][c] = 30e-6
		}
	}

	params := SolveParams{
		Conductance: g,
		Wire:        WireParams{RWordLine: 40, RBitLine: 40},
		Boundary: BoundaryParams{
			WLDriveResistance: 20,
			BLDriveResistance: 20,
		},
	}

	pulseWidth := 100e-9 // 100 ns
	pulseVs := []float64{1.6, 1.7, 1.8, 1.9, 2.0, 2.1}

	cumEnergy := 0.0
	prev := 0.0

	for k, vp := range pulseVs {
		wl := make([]float64, rows)
		bl := make([]float64, cols)
		for r := range wl {
			wl[r] = vp / 2
		}
		for c := range bl {
			bl[c] = vp / 2
		}
		wl[targetR] = vp
		bl[targetC] = 0
		params.WLVoltages = wl
		params.BLVoltages = bl

		res, err := NewTierASolver().Solve(params)
		if err != nil {
			t.Fatalf("pulse %d solve failed: %v", k, err)
		}

		// Instantaneous power from target cell + source-connected wire/drive network.
		iTarget := res.CellCurrents[targetR][targetC]
		gEff := effectiveCellConductance(params, targetR, targetC)
		if gEff <= 0 {
			t.Fatalf("pulse %d invalid target conductance", k)
		}
		pTarget := (iTarget * iTarget) / gEff

		dc, err := NewTierBSolver().SolveDC(params)
		if err != nil {
			t.Fatalf("pulse %d solveDC failed: %v", k, err)
		}
		pNet := wireAndDrivePower(params, dc)
		pInst := pTarget + pNet
		if pInst < 0 {
			t.Fatalf("pulse %d negative instantaneous power: %g", k, pInst)
		}

		cumEnergy += pInst * pulseWidth
		if cumEnergy+1e-21 < prev {
			t.Fatalf("cumulative energy decreased at pulse %d: prev=%g now=%g", k, prev, cumEnergy)
		}
		prev = cumEnergy

		// Emulate successful ISPP step by increasing target conductance slightly.
		params.Conductance[targetR][targetC] *= 1.03
	}

	if cumEnergy <= 0 {
		t.Fatalf("final cumulative energy must be >0, got %g", cumEnergy)
	}
}
