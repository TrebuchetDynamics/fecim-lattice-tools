package hysteresis_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"fecim-lattice-tools/workbench/hysteresis"
)

// Catches lost drive ownership, nondeterministic target/phase histories and reset leakage.
func TestProgrammingTrace(t *testing.T) {
	for _, model := range []hysteresis.Model{hysteresis.Preisach, hysteresis.Landau} {
		for _, mode := range []hysteresis.ProgramMode{hysteresis.TargetWrite, hysteresis.WriteRead} {
			for _, domains := range []int{1, 4} {
				if domains > 1 && (model != hysteresis.Landau || mode != hysteresis.WriteRead) {
					continue
				}
				t.Run(string(model)+"/"+string(mode)+string(rune('0'+domains)), func(t *testing.T) {
					c := hysteresis.DefaultConfig()
					c.Mode = hysteresis.Manual
					c.Model = model
					c.Seed = 42
					s, _ := hysteresis.NewSession(c)
					pc := hysteresis.DefaultProgramConfig()
					pc.Mode = mode
					pc.Domains = domains
					p, err := hysteresis.NewProgrammer(s, pc)
					if err != nil {
						t.Fatal(err)
					}
					if _, err = p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 22}); err != nil {
						t.Fatal(err)
					}
					for _, cmd := range []hysteresis.Command{{Kind: hysteresis.SetField, Field: 1}, {Kind: hysteresis.Configure, Config: c}, {Kind: hysteresis.Reset}} {
						if _, err = s.Apply(cmd); !errors.Is(err, hysteresis.ErrDriveOwned) {
							t.Fatalf("external action bypassed programming: %v", err)
						}
					}
					if s.Snapshot().Controls.ManualField.Enabled {
						t.Fatal("manual field advertised while owned")
					}
					if _, err = s.Advance(hysteresis.Budget{Steps: 1, DT: 0.01}); !errors.Is(err, hysteresis.ErrDriveOwned) {
						t.Fatal(err)
					}
					if mode == hysteresis.WriteRead {
						if _, err = p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 8}); err != nil {
							t.Fatal(err)
						}
					}
					first, err := p.Advance(hysteresis.Budget{Steps: 1, DT: 0.01})
					if err != nil {
						t.Fatal(err)
					}
					wantPhase := hysteresis.PrepPhase
					if model == hysteresis.Preisach {
						wantPhase = hysteresis.WritePhase
					}
					wantTarget := 22
					if mode == hysteresis.WriteRead {
						wantTarget = 8
					}
					if first.Phase != wantPhase || first.Core.SolverSteps != 1 || first.Target != wantTarget {
						t.Fatalf("first step: %+v", first)
					}
					if mode == hysteresis.WriteRead && first.QueuedTarget != 0 {
						t.Fatal("queue was not consumed at PREP boundary")
					}
					trace := func() []hysteresis.ProgrammingSnapshot {
						var out []hysteresis.ProgrammingSnapshot
						for i := 0; i < 3000; i++ {
							v, e := p.Advance(hysteresis.Budget{Steps: 1, DT: 0.01})
							if e != nil {
								t.Fatal(e)
							}
							v.Core.Plot = nil
							v.Core.Catalogue = nil
							out = append(out, v)
							if v.Status != hysteresis.ProgramRunning {
								break
							}
						}
						return out
					}
					a := trace()
					if mode == hysteresis.TargetWrite && a[len(a)-1].Status == hysteresis.ProgramRunning {
						t.Fatal("manual job never reached an explicit terminal status")
					}
					if model == hysteresis.Preisach && mode == hysteresis.TargetWrite && !a[len(a)-1].Complete {
						t.Fatal("native converging target was not completed")
					}
					if a[len(a)-1].Core.Count == 0 || a[len(a)-1].Core.P.Presentation.Speakable == "" {
						t.Fatal("no canonical trace")
					}
					if _, err = p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.ResetProgram}); err != nil {
						t.Fatal(err)
					}
					reset := p.Snapshot()
					if reset.Core.Count != 0 || reset.Core.E.SI != 0 || reset.Target != 0 || reset.QueuedTarget != 0 || reset.Phase != hysteresis.IdlePhase {
						t.Fatal("reset leaked write state")
					}
					// Native Preisach Reset preserves physical dynamicP while clearing
					// turning points. Determinism compares identical command histories,
					// not a reset cell with a never-driven constructor.
					s2, _ := hysteresis.NewSession(c)
					p2, _ := hysteresis.NewProgrammer(s2, pc)
					p = p2
					p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 22})
					if mode == hysteresis.WriteRead {
						p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 8})
					}
					p.Advance(hysteresis.Budget{Steps: 1, DT: 0.01})
					b := trace()
					if len(a) != len(b) {
						t.Fatal("identical seeded jobs changed trace length")
					}
					for i := range a {
						a[i].Core.Revision = 0
						b[i].Core.Revision = 0
						if !reflect.DeepEqual(a[i], b[i]) {
							t.Fatalf("identical seeded trace differs at %d", i)
						}
					}
				})
			}
		}
	}
}

// Catches partial/cancelled/capped work being reported as successful.
func TestProgrammingCancellation(t *testing.T) {
	for _, action := range []string{"command", "context", "cap", "calibration-command", "calibration-context", "calibration-cap"} {
		t.Run(action, func(t *testing.T) {
			c := hysteresis.DefaultConfig()
			c.Mode = hysteresis.Manual
			s, _ := hysteresis.NewSession(c)
			pc := hysteresis.DefaultProgramConfig()
			pc.WorkLimit = 2
			pc.CalibrationWorkLimit = 2
			p, err := hysteresis.NewProgrammer(s, pc)
			if err != nil {
				t.Fatal(err)
			}
			kind := hysteresis.WriteTarget
			target := 25
			if len(action) > 11 {
				kind = hysteresis.CalibrateProgram
				target = 0
			}
			p.Apply(hysteresis.ProgramCommand{Kind: kind, Target: target})
			p.Advance(hysteresis.Budget{Steps: 1, DT: 0.01})
			switch action {
			case "command", "calibration-command":
				_, err = p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.CancelProgram})
			case "context", "calibration-context":
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				_, err = p.Advance(hysteresis.Budget{Steps: 1, DT: 0.01, Context: ctx})
				if !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
			case "cap", "calibration-cap":
				p.Advance(hysteresis.Budget{Steps: 1, DT: 0.01})
				_, err = p.Advance(hysteresis.Budget{Steps: 1, DT: 0.01})
				if !errors.Is(err, hysteresis.ErrWorkLimit) {
					t.Fatal(err)
				}
			}
			v := p.Snapshot()
			want := hysteresis.ProgramCancelled
			if action == "cap" || action == "calibration-cap" {
				want = hysteresis.ProgramWorkLimit
			}
			if v.Calibration.Status == hysteresis.CalibrationFresh {
				t.Fatal("partial calibration marked fresh")
			}
			if v.Status != want || v.Complete || v.OwnsDrive {
				t.Fatalf("false completion/ownership: %+v", v)
			}
			if _, err = s.Apply(hysteresis.Command{Kind: hysteresis.SetField, Field: 0}); err != nil {
				t.Fatal("drive not released", err)
			}
		})
	}
}

// Catches stale tables being exported or used by a new write after a settings change.
func TestProgrammingCalibration(t *testing.T) {
	t.Run("full native LK level range completes within admitted work", func(t *testing.T) {
		for _, levels := range []int{30, 64} {
			c := hysteresis.DefaultConfig()
			c.Mode = hysteresis.Manual
			c.Model = hysteresis.Landau
			c.Levels = levels
			s, err := hysteresis.NewSession(c)
			if err != nil {
				t.Fatal(err)
			}
			p, err := hysteresis.NewProgrammer(s, hysteresis.ProgramConfig{})
			if err != nil {
				t.Fatal(err)
			}
			v, err := p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.CalibrateProgram})
			if err != nil {
				t.Fatal(err)
			}
			for v.Status == hysteresis.ProgramRunning {
				v, err = p.Advance(hysteresis.Budget{Steps: 256, DT: 0.01})
				if err != nil {
					t.Fatal(err)
				}
			}
			if v.Calibration.Status != hysteresis.CalibrationFresh || len(v.Calibration.Ascending) != levels || len(v.Calibration.Descending) != levels || !v.Complete {
				t.Fatal("full supported LK calibration could not finish")
			}
		}
	})
	for _, change := range []string{"material", "engine", "levels", "range", "temperature", "stress"} {
		t.Run(change, func(t *testing.T) {
			c := hysteresis.DefaultConfig()
			c.Mode = hysteresis.Manual
			c.Levels = 2
			s, _ := hysteresis.NewSession(c)
			p, err := hysteresis.NewProgrammer(s, hysteresis.ProgramConfig{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = p.ExportCalibration(); err == nil {
				t.Fatal("unavailable calibration exported")
			}
			if _, err = p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.CalibrateProgram}); err != nil {
				t.Fatal(err)
			}
			for i := 0; p.Snapshot().Status == hysteresis.ProgramRunning && i < 10000; i++ {
				if _, err = p.Advance(hysteresis.Budget{Steps: 256, DT: 0.01}); err != nil {
					t.Fatal(err)
				}
			}
			v := p.Snapshot()
			if v.Calibration.Status != hysteresis.CalibrationFresh || !v.Complete || !v.Calibration.CanApply || !v.Calibration.CanExport {
				t.Fatalf("not fresh: %+v", v.Calibration)
			}
			result, err := p.ExportCalibration()
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Ascending) != 2 || len(result.Descending) != 2 || result.Ascending[0].SI != 0 || result.Descending[1].SI != 0 || result.Ascending[1].SI <= 0 || result.Descending[0].SI >= 0 {
				t.Fatal("missing native saturation endpoints")
			}
			result.Ascending[0].SI = 999
			again, _ := p.ExportCalibration()
			if again.Ascending[0].SI == 999 {
				t.Fatal("calibration aliases state")
			}
			base := c
			c.TimeScale = 0.5
			s.Apply(hysteresis.Command{Kind: hysteresis.Configure, Config: c})
			if p.Snapshot().Calibration.Status != hysteresis.CalibrationFresh {
				t.Fatal("playback invalidated calibration")
			}
			c = base
			switch change {
			case "material":
				for _, m := range hysteresis.Catalogue() {
					if m.ID != c.MaterialID {
						c.MaterialID = m.ID
						break
					}
				}
			case "engine":
				c.Model = hysteresis.Landau
			case "levels":
				c.Levels = 64
			case "range":
				c.RangeFraction = 0.5
			case "temperature":
				c.Model = hysteresis.Landau
				c.Temperature = 350
			case "stress":
				c.Model = hysteresis.Landau
				c.Stress = 2e9
			}
			if _, err = s.Apply(hysteresis.Command{Kind: hysteresis.Configure, Config: c}); err != nil {
				t.Fatal(err)
			}
			stale := p.Snapshot().Calibration
			if stale.Status != hysteresis.CalibrationStale || stale.CanApply || stale.CanExport {
				t.Fatalf("stale result eligible: %+v", stale)
			}
			if _, err = p.ExportCalibration(); err == nil {
				t.Fatal("stale result exported")
			}
			s.Apply(hysteresis.Command{Kind: hysteresis.Configure, Config: base})
			if p.Snapshot().Calibration.Status != hysteresis.CalibrationStale {
				t.Fatal("reverting settings resurrected stale calibration")
			}
			p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.CalibrateProgram})
			for p.Snapshot().Status == hysteresis.ProgramRunning {
				if _, err = p.Advance(hysteresis.Budget{Steps: 256, DT: 0.01}); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = p.ExportCalibration(); err != nil {
				t.Fatal("recalibration did not restore freshness", err)
			}
		})
	}
}

// Catches invalid requests mutating a live job and manual clicks spuriously
// programming the already selected level (native clicks are a no-op).
func TestProgrammingInputs(t *testing.T) {
	c := hysteresis.DefaultConfig()
	c.Mode = hysteresis.Manual
	s, _ := hysteresis.NewSession(c)
	p, _ := hysteresis.NewProgrammer(s, hysteresis.ProgramConfig{})
	before := s.Snapshot()
	v, err := p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: before.Level + 1})
	if err != nil || v.Status != hysteresis.ProgramIdle || !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("current-level click should not acquire drive")
	}
	for _, cmd := range []hysteresis.ProgramCommand{{Kind: hysteresis.WriteTarget, Target: 0}, {Kind: hysteresis.WriteTarget, Target: 65}, {Kind: hysteresis.CancelProgram, Target: 2}, {Kind: "unknown"}} {
		before := p.Snapshot()
		if _, err = p.Apply(cmd); err == nil || !reflect.DeepEqual(before, p.Snapshot()) {
			t.Fatal("invalid command mutated state")
		}
	}
	p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 25})
	other, _ := hysteresis.NewProgrammer(s, hysteresis.ProgramConfig{})
	if _, err = other.Apply(hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 4}); !errors.Is(err, hysteresis.ErrDriveOwned) {
		t.Fatal("second programmer stole drive", err)
	}
	for _, b := range []hysteresis.Budget{{Steps: 257, DT: 0.01}, {Steps: 1, DT: 0}, {Steps: 1, DT: 2}} {
		before := p.Snapshot()
		if _, err = p.Advance(b); err == nil || !reflect.DeepEqual(before, p.Snapshot()) {
			t.Fatal("invalid budget mutated work")
		}
	}
	for _, modify := range []func(*hysteresis.ProgramConfig){func(c *hysteresis.ProgramConfig) { c.Method = "dcc" }, func(c *hysteresis.ProgramConfig) { c.Domains = 129 }, func(c *hysteresis.ProgramConfig) { c.WorkLimit = 0 }} {
		pc := hysteresis.DefaultProgramConfig()
		modify(&pc)
		if _, err = hysteresis.NewProgrammer(s, pc); err == nil {
			t.Fatal("invalid config accepted")
		}
	}
}

func TestTimeResolvedUnavailable(t *testing.T) {
	s, _ := hysteresis.NewSession(hysteresis.Config{})
	s.Apply(hysteresis.Command{Kind: hysteresis.Run})
	s.Advance(hysteresis.Budget{Steps: 2, DT: 0.01})
	pc := hysteresis.DefaultProgramConfig()
	pc.Mode = hysteresis.TimeResolved
	p, err := hysteresis.NewProgrammer(s, pc)
	if err != nil {
		t.Fatal(err)
	}
	before := s.Snapshot()
	v, err := p.Apply(hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 5})
	if err != nil {
		t.Fatal(err)
	}
	v, err = p.Advance(hysteresis.Budget{Steps: 10, DT: 0.01})
	if err != nil {
		t.Fatal(err)
	}
	if v.Status != hysteresis.LegacyProtocolUnavailable || v.Reason == "" || v.Complete || v.OwnsDrive || len(v.Core.Plot) != 0 || !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("stub fabricated switching work")
	}
}
