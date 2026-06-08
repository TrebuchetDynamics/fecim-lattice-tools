package arraysim

import (
	"math"
	"testing"

	sharedcrossbar "fecim-lattice-tools/shared/crossbar"
)

func TestModule4ArraySimE2ECrossModule2ReadoutBridgeWorkflow(t *testing.T) {
	cases := []struct {
		name    string
		rows    int
		cols    int
		weights [][]float64
		input   []float64
		mode    CouplingMode
	}{
		{name: "square-tier-a", rows: 4, cols: 4, weights: [][]float64{{0.1, 0.3, 0.5, 0.7}, {0.9, 0.2, 0.4, 0.6}, {0.8, 1.0, 0.2, 0.4}, {0.25, 0.5, 0.75, 1.0}}, input: []float64{0.2, 0.4, 0.6, 0.8}, mode: CouplingTierA},
		{name: "wide-tier-b", rows: 3, cols: 5, weights: [][]float64{{0.05, 0.2, 0.4, 0.6, 0.8}, {1.0, 0.7, 0.5, 0.3, 0.1}, {0.15, 0.35, 0.55, 0.75, 0.95}}, input: []float64{0.1, 0.3, 0.5, 0.7, 0.9}, mode: CouplingTierB},
		{name: "tall-ideal", rows: 5, cols: 3, weights: [][]float64{{0.2, 0.4, 0.6}, {0.8, 1.0, 0.1}, {0.3, 0.5, 0.7}, {0.9, 0.2, 0.4}, {0.6, 0.8, 1.0}}, input: []float64{0.25, 0.5, 0.75}, mode: CouplingIdeal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m2, err := sharedcrossbar.NewArray(&sharedcrossbar.Config{Rows: tc.rows, Cols: tc.cols, ADCBits: 8, DACBits: 8, NoiseLevel: 0})
			if err != nil {
				t.Fatalf("module2 NewArray: %v", err)
			}
			defer m2.Destroy()
			if err := m2.ProgramWeightMatrix(tc.weights); err != nil {
				t.Fatalf("module2 ProgramWeightMatrix: %v", err)
			}
			m2Out, err := m2.MVM(tc.input)
			if err != nil {
				t.Fatalf("module2 MVM: %v", err)
			}
			if len(m2Out) != tc.rows {
				t.Fatalf("module2 output len=%d want %d", len(m2Out), tc.rows)
			}

			cfg := ArrayConfig{Rows: tc.rows, Cols: tc.cols, ReadVoltageV: 0.2, CouplingMode: tc.mode, Wire: WireParams{RWordLine: 0.7, RBitLine: 2.4}, Sense: SenseChain{TIA: TIAConfig{Rf: 2e4, Vref: 0, Vmin: 0, Vmax: 1.2}, ADC: ADCConfig{Bits: 6, Vmin: 0, Vmax: 1.2}}}
			params := SolveParams{WLVoltages: wlVectorE2E(tc.rows, 0.2), BLVoltages: make([]float64, tc.cols), Conductance: normalizedWeightsToConductanceE2E(tc.weights), Geometry: DefaultCellGeometry(), Wire: cfg.Wire}
			m4Solve, ok := solveRead(cfg, params)
			if !ok || len(m4Solve.RowCurrents) != tc.rows {
				t.Fatalf("module4 solve invalid ok=%v result=%+v", ok, m4Solve)
			}
			m4Sense := cfg.Sense.ConvertCurrents(m4Solve.RowCurrents)
			m2Sense := cfg.Sense.ConvertCurrents(scaleModule2OutputsToCurrentE2E(m2Out))
			if len(m4Sense) != tc.rows || len(m2Sense) != tc.rows {
				t.Fatalf("sense lengths m4=%d m2=%d want %d", len(m4Sense), len(m2Sense), tc.rows)
			}
			for r := 0; r < tc.rows; r++ {
				if math.IsNaN(m2Out[r]) || math.IsInf(m2Out[r], 0) || math.IsNaN(m4Solve.RowCurrents[r]) || math.IsInf(m4Solve.RowCurrents[r], 0) || m4Sense[r].Code < 0 || m4Sense[r].Code >= 64 || m2Sense[r].Code < 0 || m2Sense[r].Code >= 64 {
					t.Fatalf("row %d bridge invalid m2=%g m2sense=%+v m4=%g m4sense=%+v", r, m2Out[r], m2Sense[r], m4Solve.RowCurrents[r], m4Sense[r])
				}
			}
			margin := ReadMarginAnalysis(cfg, 8)
			if margin.MinMarginV < 0 || len(margin.MarginPerLevel) != 7 {
				t.Fatalf("bridge read margin invalid: %+v", margin)
			}
		})
	}
}

func TestModule4ArraySimE2ECrossModuleBridgeInvalidIsolation(t *testing.T) {
	m2, err := sharedcrossbar.NewArray(&sharedcrossbar.Config{Rows: 2, Cols: 2, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("module2 NewArray: %v", err)
	}
	defer m2.Destroy()
	if err := m2.ProgramWeightMatrix([][]float64{{0.1, 0.2}, {0.3, 0.4}}); err != nil {
		t.Fatalf("module2 ProgramWeightMatrix: %v", err)
	}
	if _, err := m2.MVM([]float64{math.NaN()}); err == nil {
		t.Fatal("module2 NaN input should fail before module4 bridge")
	}
	cfg := ArrayConfig{Rows: 2, Cols: 2, CouplingMode: CouplingTierB, ReadVoltageV: 0.2}
	res := ReadMarginAnalysis(cfg, 0)
	if res.Levels != 0 || len(res.MarginPerLevel) != 0 || res.Reliable {
		t.Fatalf("module4 invalid margin boundary changed: %+v", res)
	}
}

func normalizedWeightsToConductanceE2E(weights [][]float64) [][]float64 {
	out := make([][]float64, len(weights))
	for r := range weights {
		out[r] = make([]float64, len(weights[r]))
		for c, w := range weights[r] {
			if w < 0 {
				w = 0
			}
			if w > 1 {
				w = 1
			}
			out[r][c] = 10e-6 + w*90e-6
		}
	}
	return out
}

func scaleModule2OutputsToCurrentE2E(outputs []float64) []float64 {
	currents := make([]float64, len(outputs))
	for i, out := range outputs {
		currents[i] = out * 50e-6
	}
	return currents
}
