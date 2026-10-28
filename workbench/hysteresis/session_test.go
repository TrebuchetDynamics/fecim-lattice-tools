package hysteresis_test

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"

	"fecim-lattice-tools/workbench/hysteresis"
)

func TestSession(t *testing.T) {
	newSession := func(c hysteresis.Config) *hysteresis.Session {
		t.Helper()
		s, err := hysteresis.NewSession(c)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	apply := func(s *hysteresis.Session, c hysteresis.Command) hysteresis.Snapshot {
		t.Helper()
		v, err := s.Apply(c)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	advance := func(s *hysteresis.Session, n int) hysteresis.Snapshot {
		t.Helper()
		v, err := s.Advance(hysteresis.Budget{Steps: n, DT: 0.01})
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	t.Run("manual reversals retain memory until reset", func(t *testing.T) {
		c := hysteresis.DefaultConfig()
		c.Mode = hysteresis.Manual
		s := newSession(c)
		v := s.Snapshot()
		if v.State != "paused" || !v.Controls.ManualField.Enabled || v.Controls.Temperature.Enabled || v.Evidence != "simulation_only" {
			t.Fatal("initial state/metadata")
		}
		apply(s, hysteresis.Command{Kind: hysteresis.Run})
		var p [2]float64
		for i, e := range []float64{1e8, 0, -1e8, 0} {
			apply(s, hysteresis.Command{Kind: hysteresis.SetField, Field: e})
			v = advance(s, 1)
			if i == 1 {
				p[0] = v.P.SI
			}
			if i == 3 {
				p[1] = v.P.SI
			}
		}
		if p[0] == p[1] {
			t.Fatal("lost Preisach reversal memory at zero")
		}
		v = apply(s, hysteresis.Command{Kind: hysteresis.Pause})
		if !reflect.DeepEqual(v, advance(s, 1)) {
			t.Fatal("pause advanced science")
		}
		v = apply(s, hysteresis.Command{Kind: hysteresis.SetView, View: hysteresis.LastCycle})
		if v.Count != 4 || v.P.SI != p[1] {
			t.Fatal("view changed science")
		}
		v = apply(s, hysteresis.Command{Kind: hysteresis.Reset})
		if v.E.SI != 0 || v.P.SI != 0 || v.Time != 0 || v.Count != 0 {
			t.Fatal("reset incomplete")
		}
		apply(s, hysteresis.Command{Kind: hysteresis.Run})
		for _, e := range []float64{1e8, 0} {
			apply(s, hysteresis.Command{Kind: hysteresis.SetField, Field: e})
			v = advance(s, 1)
		}
		if v.P.SI != p[0] {
			t.Fatal("reset trajectory differs")
		}
		v = apply(s, hysteresis.Command{Kind: hysteresis.Cancel})
		if v.State != "cancelled" || !reflect.DeepEqual(v, advance(s, 1)) {
			t.Fatal("cancel advanced science")
		}
	})
	t.Run("invalid settings commands and budgets are atomic", func(t *testing.T) {
		s := newSession(hysteresis.Config{})
		apply(s, hysteresis.Command{Kind: hysteresis.Run})
		advance(s, 1)
		before := s.Snapshot()
		for _, mutate := range []func(*hysteresis.Config){
			func(c *hysteresis.Config) { c.MaterialID = "missing" }, func(c *hysteresis.Config) { c.Model = "unknown" }, func(c *hysteresis.Config) { c.Mode = "square" },
			func(c *hysteresis.Config) { c.Frequency = math.NaN() }, func(c *hysteresis.Config) { c.Frequency = math.MaxFloat64 }, func(c *hysteresis.Config) { c.Frequency = 0 }, func(c *hysteresis.Config) { c.Frequency = math.SmallestNonzeroFloat64 },
			func(c *hysteresis.Config) { c.TimeScale = 1e-10 }, func(c *hysteresis.Config) { c.Temperature = math.Inf(1) }, func(c *hysteresis.Config) { c.Temperature = 350 }, func(c *hysteresis.Config) { c.Stress = -1 },
			func(c *hysteresis.Config) { c.Levels = 65 }, func(c *hysteresis.Config) { c.RangeFraction = 0.4 }, func(c *hysteresis.Config) { c.HistoryLimit = math.MaxInt },
		} {
			c := before.Config
			mutate(&c)
			if _, err := s.Apply(hysteresis.Command{Kind: hysteresis.Configure, Config: c}); err == nil {
				t.Fatalf("accepted invalid config %+v", c)
			}
			if _, err := hysteresis.NewSession(c); err == nil {
				t.Fatal("constructor accepted invalid config")
			}
			if !reflect.DeepEqual(before, s.Snapshot()) {
				t.Fatal("invalid config mutated session")
			}
		}
		for _, c := range []hysteresis.Command{{Kind: "unknown"}, {Kind: hysteresis.SetField, Field: 0}, {Kind: hysteresis.SetField, Field: math.NaN()}, {Kind: hysteresis.SetView, View: "missing"}, {Kind: hysteresis.Reset, Field: 1}, {Kind: hysteresis.Pause, Config: before.Config}} {
			if _, err := s.Apply(c); err == nil {
				t.Fatalf("accepted invalid command %+v", c)
			}
		}
		for _, b := range []hysteresis.Budget{{Steps: 257, DT: 0.01}, {Steps: -1, DT: 1}, {Steps: 1, DT: 0}, {Steps: 1, DT: math.Inf(1)}, {Steps: 1, DT: math.MaxFloat64}} {
			if _, err := s.Advance(b); err == nil {
				t.Fatalf("accepted invalid budget %+v", b)
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := s.Advance(hysteresis.Budget{Steps: 1, DT: 0.01, Context: ctx}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, s.Snapshot()) {
			t.Fatal("rejected work mutated session")
		}
		c := before.Config
		c.Mode = hysteresis.Manual
		apply(s, hysteresis.Command{Kind: hysteresis.Configure, Config: c})
		before = s.Snapshot()
		for _, e := range []float64{math.NaN(), math.Inf(-1), math.MaxFloat64, before.Controls.ManualField.Max * 2} {
			if _, err := s.Apply(hysteresis.Command{Kind: hysteresis.SetField, Field: e}); err == nil {
				t.Fatal("invalid manual field accepted")
			}
		}
		if !reflect.DeepEqual(before, s.Snapshot()) {
			t.Fatal("invalid manual field mutated session")
		}
	})
	t.Run("revision bound pages cover ring and plot is bounded", func(t *testing.T) {
		c := hysteresis.DefaultConfig()
		c.HistoryLimit = 3000
		s := newSession(c)
		apply(s, hysteresis.Command{Kind: hysteresis.Run})
		for i := 0; i < 13; i++ {
			advance(s, 256)
		}
		v := s.Snapshot()
		if v.Count != 3000 || v.Dropped != 328 || len(v.Plot) != 2000 {
			t.Fatal("retention/plot cap")
		}
		c.HistoryLimit = 1000
		v = apply(s, hysteresis.Command{Kind: hysteresis.Configure, Config: c})
		n := 0
		for offset := 0; offset < v.Count; offset += 32 {
			p, err := s.Page(hysteresis.PageQuery{Revision: v.Revision, Offset: offset})
			if err != nil {
				t.Fatal(err)
			}
			if len(p.Rows) > 32 || p.Total != 1000 || p.Position == "" {
				t.Fatal("page metadata")
			}
			for _, r := range p.Rows {
				if r.Index != 2328+n || r.E.Presentation.Speakable == "" || r.Branch == "" {
					t.Fatal("row lost identity/context")
				}
				n++
			}
		}
		if n != 1000 || v.Dropped != 2328 {
			t.Fatal("missing retained rows")
		}
		p, _ := s.Page(hysteresis.PageQuery{Revision: v.Revision})
		p.Rows[0].P.SI = 999
		again, _ := s.Page(hysteresis.PageQuery{Revision: v.Revision})
		if again.Rows[0].P.SI == 999 {
			t.Fatal("page aliases session")
		}
		v.Catalogue[0].Parameters.Ec = 0
		if s.Snapshot().Catalogue[0].Parameters.Ec == 0 {
			t.Fatal("catalogue aliases session")
		}
		for _, offset := range []int{-1, 1001, math.MaxInt} {
			if _, err := s.Page(hysteresis.PageQuery{Revision: v.Revision, Offset: offset}); err == nil {
				t.Fatal("invalid offset accepted")
			}
		}
		if a := testing.AllocsPerRun(2, func() { _, _ = s.Page(hysteresis.PageQuery{Revision: v.Revision}) }); a > 1000 {
			t.Fatalf("unbounded paging allocations: %g", a)
		}
		advance(s, 1)
		if _, err := s.Page(hysteresis.PageQuery{Revision: v.Revision}); !errors.Is(err, hysteresis.ErrStaleRevision) {
			t.Fatal(err)
		}
	})
	t.Run("settings invalidation preserves engine memory", func(t *testing.T) {
		c := hysteresis.DefaultConfig()
		c.Mode = hysteresis.Manual
		s := newSession(c)
		apply(s, hysteresis.Command{Kind: hysteresis.Run})
		apply(s, hysteresis.Command{Kind: hysteresis.SetField, Field: 1e8})
		v := advance(s, 1)
		c.TimeScale = 0.1
		next := apply(s, hysteresis.Command{Kind: hysteresis.Configure, Config: c})
		if next.Count != v.Count || next.P != v.P || next.Time != v.Time {
			t.Fatal("playback reset physics")
		}
		c.Frequency = 1
		next = apply(s, hysteresis.Command{Kind: hysteresis.Configure, Config: c})
		if next.Count != 0 || next.Time != 0 || next.P != v.P || next.ConfigRevision == v.ConfigRevision {
			t.Fatal("frequency invalidation")
		}
		c.Model = hysteresis.Landau
		next = apply(s, hysteresis.Command{Kind: hysteresis.Configure, Config: c})
		if next.P != v.P || !next.Controls.Temperature.Enabled {
			t.Fatal("engine carry/control")
		}
		c.Temperature = 350
		c.Stress = 2e9
		apply(s, hysteresis.Command{Kind: hysteresis.Configure, Config: c})
		c.Model = hysteresis.Preisach
		c.Temperature = 300
		c.Stress = 1e9
		apply(s, hysteresis.Command{Kind: hysteresis.Configure, Config: c})
	})
	t.Run("solver lifetime work is bounded until reset", func(t *testing.T) {
		s := newSession(hysteresis.Config{})
		apply(s, hysteresis.Command{Kind: hysteresis.Run})
		for i := 0; i < hysteresis.MaxSolverSteps; {
			n := min(256, hysteresis.MaxSolverSteps-i)
			advance(s, n)
			i += n
		}
		before := s.Snapshot()
		if _, err := s.Advance(hysteresis.Budget{Steps: 1, DT: 0.01}); !errors.Is(err, hysteresis.ErrWorkLimit) {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, s.Snapshot()) {
			t.Fatal("work limit changed state")
		}
		apply(s, hysteresis.Command{Kind: hysteresis.Reset})
		advance(s, 1)
	})
}
