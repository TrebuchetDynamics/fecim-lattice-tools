package crossbar

import (
	"context"
	"encoding/json"
	core "fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/workbench/moduleapi"
	"math"
	"reflect"
	"testing"
	"time"
)

func TestCrossbarProfiles(t *testing.T) {
	ctx := context.Background()
	matrix := [][]float64{{0.2, 0.5, 0.8}, {0.9, 0.4, 0.1}}
	input := []float64{0.5, 1, 0.25}
	for _, profile := range []string{"basic", "enhanced", "parasitic", "fecap"} {
		t.Run(profile, func(t *testing.T) {
			c := DefaultConfig()
			c.Rows = 2
			c.Cols = 3
			c.Profile = profile
			c.Noise = 0.02
			got, err := Evaluate(ctx, c, matrix, input, 32, nil)
			if err != nil {
				t.Fatal(err)
			}
			again, err := Evaluate(ctx, c, matrix, input, 32, nil)
			if err != nil || !reflect.DeepEqual(got, again) {
				t.Fatal("seed is not reproducible", err)
			}
			cfg := &core.Config{Rows: 2, Cols: 3, ADCBits: c.ADCBits, DACBits: c.DACBits, NoiseLevel: c.Noise, QuantizationLevels: c.Levels}
			if profile == "fecap" {
				cfg = core.DefaultFeCAPConfig(3, 2)
				cfg.ADCBits = c.ADCBits
				cfg.DACBits = c.DACBits
				cfg.NoiseLevel = c.Noise
			}
			a, err := core.NewArraySeeded(cfg, c.Seed)
			if err != nil {
				t.Fatal(err)
			}
			defer a.Destroy()
			var want []float64
			switch profile {
			case "fecap":
				if err = a.ProgramCapacitanceMatrix([][]float64{{0.2, 0.9}, {0.5, 0.4}, {0.8, 0.1}}); err != nil {
					t.Fatal(err)
				}
				want, err = a.MVMCharge(input)
				if got.Unit != "C" || got.PhysicalUnit != "F" {
					t.Fatal("charge unit/orientation lost")
				}
			case "parasitic":
				a.ProgramWeightMatrix(matrix)
				s, _ := core.NewParasiticSolver(2, 3, nil)
				s.SetConductances(a.GetConductanceMatrix())
				s.SetParasitics(c.RpRow, c.RpCol)
				r, e := s.SolveMVM(input)
				err = e
				want = r.OutputCurrents
				if got.Solver == nil || !reflect.DeepEqual(got.Solver, r) {
					t.Fatal("solver evidence lost")
				}
			default:
				a.ProgramWeightMatrix(matrix)
				if profile == "basic" {
					want, err = a.MVM(input)
				} else {
					r, e := a.MVMWithNonIdealities(input, &core.MVMOptions{EnableIRDrop: c.IR, EnableSneakPaths: c.Sneak, EnableVariation: true, Temperature: c.Temperature, Architecture: c.Architecture})
					err = e
					want = r.ActualOutput
					if !reflect.DeepEqual(got.IR, r.IRDropAnalysis) || !reflect.DeepEqual(got.Sneak, r.SneakPathAnalysis) {
						t.Fatal("analysis evidence lost")
					}
				}
				if got.Unit != "ratio" {
					t.Fatal("normalized outputs mislabeled as physical current")
				}
			}
			if profile != "fecap" {
				for i := range matrix {
					for j := range matrix[i] {
						if got.Physical[i][j] != a.GetPhysicalConductanceForCell(i, j) {
							t.Fatalf("physical conductance (%d,%d) is not in siemens", i, j)
						}
					}
				}
			} else if reflect.DeepEqual(got.Ideal, got.Actual) {
				t.Fatal("noisy charge copied into ideal reference")
			}
			outputs := 2
			if profile == "parasitic" {
				outputs = 3
			}
			if err != nil || !reflect.DeepEqual(got.Actual, want) || len(got.Actual) != outputs {
				t.Fatalf("%v output=%v want=%v", err, got.Actual, want)
			}
		})
	}
}

func TestCrossbarStackAndProgrammingLevels(t *testing.T) {
	// An opt-in programming grid must not change legacy serialized defaults.
	legacy, _ := json.Marshal(core.Config{})
	var fields map[string]any
	json.Unmarshal(legacy, &fields)
	for _, key := range []string{"QuantizationLevels", "quantization_levels"} {
		if _, ok := fields[key]; ok {
			t.Fatal("new grid field changed legacy default JSON bytes")
		}
	}
	ctx := context.Background()
	p := New(32)
	defer p.Close()
	snap, err := p.Apply(ctx, moduleapi.Command{Action: "configure", Input: map[string]string{"rows": "2", "cols": "3", "levels": "16", "matrix": "[[0.5,0,1],[0.5,0,1]]"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range snap.Tables {
		if table.ID == "matrix" && *table.Rows[0][2].Value != 8.0/15.0 {
			t.Fatal("programmable conductance levels not used by actual core")
		}
	}
	apply := func(layer string) moduleapi.Snapshot {
		t.Helper()
		var err error
		snap, err = p.Apply(ctx, moduleapi.Command{Action: "stack", Revision: snap.Revision, Input: map[string]string{"layers": "512", "layer": layer}})
		if err != nil {
			t.Fatal(err)
		}
		return snap
	}
	first := apply("0")
	last := apply("511")
	for _, tc := range []struct {
		s        moduleapi.Snapshot
		min, max float64
	}{{first, 0, 0.15}, {last, 0.85, 1}} {
		found := false
		for _, table := range tc.s.Tables {
			if table.ID == "stack-cells" {
				found = true
				if len(table.Rows) != 6 {
					t.Fatal("stack cells not accessible")
				}
				for _, row := range table.Rows {
					v := *row[2].Value
					if v < tc.min || v > tc.max {
						t.Fatal("layer identity/gradient lost")
					}
				}
			}
		}
		if !found {
			t.Fatal("missing selected stack layer")
		}
	}
	again := apply("0")
	if !reflect.DeepEqual(first.Tables, again.Tables) {
		t.Fatal("seeded stack layer changed on reselection")
	}
	before := again
	if _, err = p.Apply(ctx, moduleapi.Command{Action: "stack", Revision: again.Revision, Input: map[string]string{"layers": "513", "layer": "0"}}); err == nil {
		t.Fatal("unbounded stack admitted")
	}
	after, _ := p.Snapshot(ctx)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("invalid stack action changed state")
	}
}

func TestCrossbarPresentationAndNonconvergence(t *testing.T) {
	ctx := context.Background()
	p := New(32)
	defer p.Close()
	snap, err := p.Apply(ctx, moduleapi.Command{Action: "configure", Input: map[string]string{"rows": "2", "cols": "3", "profile": "fecap", "noise": "0.02"}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := Evaluate(ctx, p.cfg, p.matrix, p.input, 32, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range snap.Tables {
		if table.ID == "matrix" {
			for _, row := range table.Rows {
				i, j := int(*row[0].Value), int(*row[1].Value)
				if *row[3].Value != result.Physical[i][j] {
					t.Fatalf("cell (%d,%d) presentation differs from actual core capacitance", i, j)
				}
			}
		}
	}
	c := DefaultConfig()
	c.Rows = 32
	c.Cols = 32
	c.Profile = "parasitic"
	c.RpRow = 0.01
	c.RpCol = 0.01
	failed, err := Evaluate(ctx, c, SeedMatrix(c), SeedInput(c), 32, nil)
	if err != nil || failed.Solver == nil || failed.Solver.Converged || failed.Solver.Iterations != 100 || len(failed.Solver.DeviceVoltages) != 32 {
		t.Fatalf("nonconvergence lost actual core diagnostics: %v %+v", err, failed.Solver)
	}
}

func TestCrossbarPortTransactions(t *testing.T) {
	p := New(32)
	defer p.Close()
	ctx := context.Background()
	before, _ := p.Snapshot(ctx)
	for _, input := range []map[string]string{{"rows": "33"}, {"input": "[1]"}, {"temperature": "NaN"}, {"noise": "Inf"}, {"matrix": "[[0.5]]"}, {"profile": "webgpu"}} {
		if _, err := p.Apply(ctx, moduleapi.Command{Action: "configure", Revision: before.Revision, Input: input}); err == nil {
			t.Fatalf("accepted invalid input %v", input)
		}
		after, _ := p.Snapshot(ctx)
		if !reflect.DeepEqual(before, after) {
			t.Fatal("failed action mutated snapshot")
		}
	}
	c := DefaultConfig()
	c.Rows = 129
	if _, err := Evaluate(ctx, c, nil, nil, 128, nil); err == nil {
		t.Fatal("local cap bypassed")
	}
	c = DefaultConfig()
	c.Noise = math.NaN()
	if _, err := Evaluate(ctx, c, nil, nil, 32, nil); err == nil {
		t.Fatal("nonfinite direct config accepted")
	}
	c = DefaultConfig()
	cancelCtx, cancel := context.WithCancel(ctx)
	if _, err := Evaluate(cancelCtx, c, SeedMatrix(c), SeedInput(c), 32, func(context.Context) error { cancel(); return cancelCtx.Err() }); err == nil {
		t.Fatal("yield cancellation ignored")
	}
	job, err := p.Start(ctx, moduleapi.Command{Action: "run", Revision: before.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Cancel(ctx, job.ID); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for {
		job, err = p.Job(ctx, job.ID)
		if err != nil {
			t.Fatal(err)
		}
		if job.State == "cancelled" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("cancel did not settle: %+v", job)
		}
		time.Sleep(time.Millisecond)
	}
	after, _ := p.Snapshot(ctx)
	data, _ := json.Marshal(after)
	if len(data) == 0 || job.Progress.Cancelled != 1 {
		t.Fatal("missing terminal counts")
	}
}
