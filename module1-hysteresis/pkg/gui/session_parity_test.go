package gui

import (
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/workbench/hysteresis"
)

// The oracle is native UpdatePhysics, not a second adapter or a fresh major
// loop for each field. No window, goroutine, calibration, persistence or logger
// sampling is started, and no app is registered in the global preset manager.
func TestWorkbenchSessionCharacterization(t *testing.T) {
	InitTestLogger()
	nativeMaterials := ferroelectric.AllMaterials()
	preferred, _ := defaultMaterialSelection(nativeMaterials)
	defaults := hysteresis.DefaultConfig()
	t.Run("material transitions preserve native solver fallback configuration", func(t *testing.T) {
		// ConfigureFromMaterial intentionally retains fallback fields omitted by
		// incomplete presets. Test transitions as well as fresh constructors.
		s, err := hysteresis.NewSession(defaults)
		if err != nil {
			t.Fatal(err)
		}
		a := &App{material: preferred, preisach: ferroelectric.NewPreisachModel(preferred), lkSolver: physics.NewLKSolver(), physicsEngine: PhysicsPreisach, numLevels: physics.DefaultLevels, maxHistory: 50000, lastHistorySample: -1, frequency: 0.5, waveform: WaveformSine, autoMode: true}
		a.lkSolver.ConfigureFromMaterial(preferred)
		a.lkSolver.UseNLS, a.lkSolver.EnableNoise = false, false
		c := defaults
		for _, next := range hysteresis.Catalogue() {
			if next.ID == c.MaterialID {
				continue
			}
			a.material = &next.Parameters
			a.preisach = ferroelectric.NewPreisachModel(a.material)
			a.lkSolver.ConfigureFromMaterial(a.material)
			a.lkSolver.Time = 0
			a.electricField, a.polarization, a.simTime = 0, 0, 0
			a.numLevels = min(64, max(2, a.material.GetNumLevels()))
			a.wrdRangeFrac = rangeFracForMaterial(a.material)
			a.resetHistoryLocked()
			if a.useLKSolver() {
				a.lkSolver.SetState(a.lkDefaultPolarization())
				a.polarization = a.lkSolver.GetState()
			}
			a.syncDiscreteLevelLocked()
			if next.DefaultLevels != a.numLevels || next.DefaultRangeFraction != a.wrdRangeFrac {
				t.Fatalf("Go catalogue material-selection defaults differ for %s", next.ID)
			}
			c.MaterialID, c.Levels, c.RangeFraction = next.ID, next.DefaultLevels, next.DefaultRangeFraction
			if _, err = s.Apply(hysteresis.Command{Kind: hysteresis.Configure, Config: c}); err != nil {
				t.Fatal(err)
			}
			a.setPhysicsEngine(PhysicsLandau)
			c.Model = hysteresis.Landau
			if _, err = s.Apply(hysteresis.Command{Kind: hysteresis.Configure, Config: c}); err != nil {
				t.Fatal(err)
			}
			if _, err = s.Apply(hysteresis.Command{Kind: hysteresis.Run}); err != nil {
				t.Fatal(err)
			}
			a.UpdatePhysics(0.125)
			v, err := s.Advance(hysteresis.Budget{Steps: 1, DT: 0.125})
			if err != nil {
				t.Fatal(err)
			}
			if v.E.SI != a.electricField || v.P.SI != a.polarization || v.Time != a.simTime || v.Level != a.discreteLevel || v.Count != a.historyLengthLocked() {
				t.Fatalf("material transition %s differs from native", next.ID)
			}
		}
	})
	for _, material := range hysteresis.Catalogue() {
		if material.ID == defaults.MaterialID && material.Parameters.Name != preferred.Name {
			t.Fatal("default catalogue drift")
		}
		for _, model := range []hysteresis.Model{hysteresis.Preisach, hysteresis.Landau} {
			for _, mode := range []hysteresis.Mode{hysteresis.Manual, hysteresis.Sine, hysteresis.Triangle} {
				t.Run(material.ID+"/"+string(model)+"/"+string(mode), func(t *testing.T) {
					mat := material.Parameters
					a := &App{material: &mat, preisach: ferroelectric.NewPreisachModel(&mat), lkSolver: physics.NewLKSolver(), physicsEngine: PhysicsPreisach, numLevels: physics.DefaultLevels, maxHistory: 50000, lastHistorySample: -1, frequency: 0.5, timeScale: 1, wrdRangeFrac: rangeFracForMaterial(&mat), autoMode: true, waveform: WaveformSine}
					a.lkSolver.ConfigureFromMaterial(&mat)
					a.lkSolver.UseNLS = false
					a.lkSolver.EnableNoise = false
					if model == hysteresis.Landau {
						a.setPhysicsEngine(PhysicsLandau)
					}
					c := defaults
					c.MaterialID = material.ID
					c.Model = model
					c.Mode = mode
					c.RangeFraction = a.wrdRangeFrac
					s, err := hysteresis.NewSession(c)
					if err != nil {
						t.Fatal(err)
					}
					if mode == hysteresis.Manual {
						a.waveform = WaveformManual
						a.autoMode = false
					}
					if mode == hysteresis.Triangle {
						a.waveform = WaveformTriangle
					}
					apply := func(cmd hysteresis.Command) hysteresis.Snapshot {
						t.Helper()
						v, err := s.Apply(cmd)
						if err != nil {
							t.Fatal(err)
						}
						return v
					}
					check := func(v hysteresis.Snapshot) {
						t.Helper()
						if v.E.SI != a.electricField || v.P.SI != a.polarization || v.Time != a.simTime || v.Level != a.discreteLevel || v.Count != a.historyLengthLocked() {
							t.Fatalf("E/P/time/level/history: session %g/%g/%g/%d/%d native %g/%g/%g/%d/%d", v.E.SI, v.P.SI, v.Time, v.Level, v.Count, a.electricField, a.polarization, a.simTime, a.discreteLevel, a.historyLengthLocked())
						}
						p, err := s.Page(hysteresis.PageQuery{Revision: v.Revision})
						if err != nil {
							t.Fatal(err)
						}
						for i, r := range p.Rows {
							e, p := a.historyAtLocked(i)
							if r.E.SI != e || r.P.SI != p {
								t.Fatalf("history row %d", i)
							}
						}
					}
					apply(hysteresis.Command{Kind: hysteresis.Run})
					for pass := 0; pass < 2; pass++ {
						for i, ratio := range []float64{0, 1.5, 0.4, 1, -0.5, -1.5, 0.2, 0} {
							if mode == hysteresis.Manual {
								a.electricField = ratio * mat.Ec
								apply(hysteresis.Command{Kind: hysteresis.SetField, Field: a.electricField})
							}
							dt := 0.125
							if i < 2 {
								dt = 1e-4
							}
							a.UpdatePhysics(dt)
							v, err := s.Advance(hysteresis.Budget{Steps: 1, DT: dt})
							if err != nil {
								t.Fatal(err)
							}
							check(v)
						}
						a.resetAllStateLocked()
						check(apply(hysteresis.Command{Kind: hysteresis.Reset}))
					}
					for i := 0; i < 64; i++ {
						a.UpdatePhysics(0.125)
					}
					if _, err := s.Advance(hysteresis.Budget{Steps: 64, DT: 0.125}); err != nil {
						t.Fatal(err)
					}
					v := apply(hysteresis.Command{Kind: hysteresis.SetView, View: hysteresis.LastCycle})
					es, ps := a.historySnapshotLocked()
					es, ps = lastCompleteCycle(es, ps)
					p, err := s.Page(hysteresis.PageQuery{Revision: v.Revision})
					if err != nil {
						t.Fatal(err)
					}
					if p.Total != len(es) {
						t.Fatal("cycle selection length")
					}
					for i, r := range p.Rows {
						if r.E.SI != es[i] || r.P.SI != ps[i] {
							t.Fatal("cycle selection differs")
						}
					}
					apply(hysteresis.Command{Kind: hysteresis.SetView, View: hysteresis.FullHistory})
					a.timeScale = 0.1
					c.TimeScale = 0.1
					check(apply(hysteresis.Command{Kind: hysteresis.Configure, Config: c}))
					a.resizeHistoryLocked(1000)
					c.HistoryLimit = 1000
					check(apply(hysteresis.Command{Kind: hysteresis.Configure, Config: c}))
					for _, next := range []hysteresis.Model{hysteresis.Landau, hysteresis.Preisach} {
						engine := PhysicsPreisach
						if next == hysteresis.Landau {
							engine = PhysicsLandau
						}
						a.setPhysicsEngine(engine)
						c.Model = next
						check(apply(hysteresis.Command{Kind: hysteresis.Configure, Config: c}))
					}
					// Conditions are applied by Go, not merely echoed as metadata.
					a.setPhysicsEngine(PhysicsLandau)
					c.Model = hysteresis.Landau
					check(apply(hysteresis.Command{Kind: hysteresis.Configure, Config: c}))
					a.lkSolver.Temperature, a.lkSolver.Stress = 350, 2e9
					a.resetHistoryLocked() // native temperature delta >25 K
					c.Temperature, c.Stress = 350, 2e9
					check(apply(hysteresis.Command{Kind: hysteresis.Configure, Config: c}))
					a.UpdatePhysics(0.125)
					v, err = s.Advance(hysteresis.Budget{Steps: 1, DT: 0.125})
					if err != nil {
						t.Fatal(err)
					}
					check(v)
				})
			}
		}
	}
}
