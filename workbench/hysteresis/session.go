// Package hysteresis owns deterministic, in-memory interactive state over the
// existing Module 1 solvers. It starts no goroutines and performs no persistence.
package hysteresis

import (
	"context"
	"errors"
	"math"
	"sync"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

const (
	MaxSamples    = 100000
	MaxSteps      = 256
	MaxPlotPoints = 2000
	PageSize      = 32
	// The retained ring alone cannot bound hidden Preisach reversal memory/work.
	// Explicit Reset (or changing material) renews this solver lifetime allowance.
	MaxSolverSteps = 100000
	maxTime        = 1e12
)

var (
	ErrStaleRevision = errors.New("stale session revision")
	ErrWorkLimit     = errors.New("solver step limit reached; reset before continuing")
)

type Model string

const (
	Preisach Model = "preisach"
	Landau   Model = "landau-khalatnikov"
)

type Mode string

const (
	Manual   Mode = "manual"
	Sine     Mode = "sine"
	Triangle Mode = "triangle"
)

type View string

const (
	FullHistory View = "full-history"
	LastCycle   View = "last-complete-cycle"
)

// Config is complete, not a partial patch. Config{} selects defaults only in
// NewSession. Quantities use SI: Hz, K, Pa. Seed is explicit execution identity;
// the interactive single-domain solvers have NLS/noise disabled, as in native Go.
type Config struct {
	MaterialID                                string
	Model                                     Model
	Mode                                      Mode
	Frequency, TimeScale, Temperature, Stress float64
	Levels                                    int
	RangeFraction                             float64
	HistoryLimit                              int
	Seed                                      uint64
}

type CommandKind string

const (
	Configure CommandKind = "configure"
	SetField  CommandKind = "set-field"
	SetView   CommandKind = "set-view"
	Run       CommandKind = "run"
	Pause     CommandKind = "pause"
	Cancel    CommandKind = "cancel"
	Reset     CommandKind = "reset"
)

// Command accepts only the payload for its Kind. Configure atomically replaces
// settings; SetField changes manual E in V/m, while Advance alone steps physics.
type Command struct {
	Kind   CommandKind
	Config Config
	Field  float64
	View   View
}

// Budget supplies physical seconds per solver call, independent of playback or
// rendering. Cancellation is checked before every step. A mid-call cancellation
// retains completed steps and their revisions, rather than pretending rollback.
type Budget struct {
	Steps   int
	DT      float64
	Context context.Context
}

type sample struct {
	Index      int
	Time, E, P float64
	Branch     string
}

// Session serializes its operations; every returned snapshot/page is owned by
// the caller and cannot mutate the underlying solver or history.
type Session struct {
	mu                          sync.Mutex
	config                      Config
	material                    Material
	catalogue                   []Material
	preisach                    *ferroelectric.PreisachModel
	lk                          *physics.LKSolver
	revision, configRevision    uint64
	state                       string
	view                        View
	e, p, time                  float64
	level                       int
	history                     []sample
	head, count, dropped, steps int
	lastHistory                 float64
	owner                       *Programmer // active programmer is the only drive/solver writer
	calibrationRevision         uint64      // scientific calibration identity, including change-and-revert
}

func NewSession(c Config) (*Session, error) {
	if c == (Config{}) {
		c = DefaultConfig()
	}
	catalogue := Catalogue()
	m, err := validateConfig(c, catalogue)
	if err != nil {
		return nil, err
	}
	s := &Session{config: c, material: m, catalogue: catalogue, revision: 1, configRevision: 1, state: "paused", view: FullHistory, lastHistory: -1, history: make([]sample, c.HistoryLimit)}
	s.preisach = ferroelectric.NewPreisachModel(&s.material.Parameters)
	s.lk = physics.NewLKSolver()
	s.lk.ConfigureFromMaterial(&s.material.Parameters)
	s.lk.UseNLS = false
	s.lk.EnableNoise = false
	s.lk.Temperature = c.Temperature
	s.lk.Stress = c.Stress
	if c.Model == Landau {
		s.lk.SetState(s.defaultP())
		s.p = s.lk.GetState()
	}
	s.syncLevel()
	return s, nil
}

func validateConfig(c Config, catalogue []Material) (Material, error) {
	var m Material
	for _, v := range catalogue {
		if v.ID == c.MaterialID {
			m = v
			break
		}
	}
	if m.ID == "" {
		return m, errors.New("unsupported material ID")
	}
	if c.Model != Preisach && c.Model != Landau {
		return m, errors.New("unsupported model")
	}
	if c.Mode != Manual && c.Mode != Sine && c.Mode != Triangle {
		return m, errors.New("unsupported core drive mode")
	}
	if !finite(c.Frequency) || c.Frequency <= 0 || !finite(1/c.Frequency) || !finite(2*math.Pi*c.Frequency*maxTime) {
		return m, errors.New("frequency must be positive with representable period and phase")
	}
	if !finite(c.TimeScale) || c.TimeScale < 1e-9 || c.TimeScale > 1 {
		return m, errors.New("playback scale outside [1e-9,1]")
	}
	if !finite(c.Temperature) || c.Temperature < 200 || c.Temperature > 700 || !finite(c.Stress) || c.Stress < 0 || c.Stress > 5e9 {
		return m, errors.New("temperature/stress outside supported controls")
	}
	if c.Model == Preisach && (c.Temperature != 300 || c.Stress != 1e9) {
		return m, errors.New("Preisach conditions are fixed at 300 K and 1 GPa")
	}
	if c.Levels < 2 || c.Levels > 64 || !finite(c.RangeFraction) || c.RangeFraction < 0.5 || c.RangeFraction > 1 {
		return m, errors.New("levels/range outside [2,64] / [0.5,1]")
	}
	if c.HistoryLimit < 1000 || c.HistoryLimit > MaxSamples {
		return m, errors.New("history limit outside [1000,100000]")
	}
	return m, nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (s *Session) defaultP() float64 {
	if p := s.material.Parameters.Pr; p != 0 {
		return -math.Abs(p)
	}
	if p := s.material.Parameters.Ps; p != 0 {
		return -math.Abs(p)
	}
	return -math.Abs(s.lk.PMax)
}
func (s *Session) clearHistory() { s.head = 0; s.count = 0; s.dropped = 0; s.lastHistory = -1 }
func (s *Session) resizeHistory(limit int) {
	keep := min(s.count, limit)
	h := make([]sample, limit)
	for i := 0; i < keep; i++ {
		h[i] = s.at(s.count - keep + i)
	}
	s.dropped += s.count - keep
	s.history = h
	s.head = 0
	s.count = keep
}
func (s *Session) at(i int) sample { return s.history[(s.head+i)%len(s.history)] }

func (s *Session) Apply(c Command) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.owner != nil && c.Kind != SetView {
		return s.snapshot(), ErrDriveOwned
	}
	return s.apply(c)
}

func (s *Session) apply(c Command) (Snapshot, error) {
	fail := func(reason string) (Snapshot, error) { return s.snapshot(), errors.New(reason) }
	if (c.Kind != Configure && c.Config != (Config{})) || (c.Kind != SetField && (c.Field != 0 || math.Signbit(c.Field))) || (c.Kind != SetView && c.View != "") {
		return fail("unexpected command payload")
	}
	switch c.Kind {
	case Configure:
		m, err := validateConfig(c.Config, s.catalogue)
		if err != nil {
			return s.snapshot(), err
		}
		if s.config == c.Config {
			return s.snapshot(), nil
		}
		old := s.config
		if old.MaterialID != c.Config.MaterialID || old.Model != c.Config.Model || old.Levels != c.Config.Levels || old.RangeFraction != c.Config.RangeFraction || old.Temperature != c.Config.Temperature || old.Stress != c.Config.Stress || old.Seed != c.Config.Seed {
			s.calibrationRevision++
		}
		s.config = c.Config
		if old.MaterialID != s.config.MaterialID {
			s.material = m
			s.preisach = ferroelectric.NewPreisachModel(&s.material.Parameters)
			// Native material changes configure the existing LK solver. Some presets
			// omit coefficients and intentionally retain its previous fallback fields.
			s.lk.ConfigureFromMaterial(&s.material.Parameters)
			s.lk.Time = 0
			s.e = 0
			s.p = 0
			s.time = 0
			s.steps = 0
			if s.config.Model == Landau {
				s.lk.SetState(s.defaultP())
				s.p = s.lk.GetState()
			}
		} else if old.Model != s.config.Model {
			if s.config.Model == Landau {
				s.lk.ConfigureFromMaterial(&s.material.Parameters)
				p := s.p
				if !finite(p) || math.Abs(p) < 1e-9 {
					p = s.defaultP()
				}
				s.lk.SetState(p)
				s.p = s.lk.GetState()
				s.lk.Time = 0
				s.lk.UseNLS = false
				s.lk.EnableNoise = false
			} else {
				s.preisach.SetTemperature(300)
				s.preisach.SetStress(1)
				s.preisach.Reset()
				s.p = s.preisach.Update(s.e)
			}
		}
		s.lk.Temperature = s.config.Temperature
		s.lk.Stress = s.config.Stress
		if old.MaterialID != s.config.MaterialID || old.Mode != s.config.Mode || old.Frequency != s.config.Frequency {
			s.time = 0
		}
		if old.MaterialID != s.config.MaterialID || old.Model != s.config.Model || old.Mode != s.config.Mode || old.Frequency != s.config.Frequency || math.Abs(old.Temperature-s.config.Temperature) > 25 {
			s.clearHistory()
		}
		if old.HistoryLimit != s.config.HistoryLimit {
			s.resizeHistory(s.config.HistoryLimit)
		}
		s.configRevision++
		s.syncLevel()
	case SetField:
		if s.config.Mode != Manual {
			return fail("manual field unavailable while automatic waveform owns drive")
		}
		if !finite(c.Field) || math.Abs(c.Field) > s.maxField() {
			return fail("manual field outside material control range")
		}
		s.e = c.Field
	case SetView:
		if c.View != FullHistory && c.View != LastCycle {
			return fail("unsupported history view")
		}
		s.view = c.View
	case Run:
		s.state = "running"
	case Pause:
		s.state = "paused"
	case Cancel:
		s.state = "cancelled"
	case Reset:
		s.e = 0
		s.p = 0
		s.time = 0
		s.steps = 0
		if s.config.Model == Landau {
			s.lk.SetState(s.defaultP())
			s.lk.Time = 0
			s.p = s.lk.GetState()
		} else {
			s.preisach.Reset()
		}
		s.clearHistory()
		s.syncLevel()
	default:
		return fail("unsupported session command")
	}
	s.revision++
	return s.snapshot(), nil
}

// Advance bounds outer solver calls before mutation. Single-domain LK uses its
// existing RK4/at-most-six-Newton-iteration path; no accuracy change is applied.
func (s *Session) Advance(b Budget) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.owner != nil {
		return s.snapshot(), ErrDriveOwned
	}
	if b.Steps < 1 || b.Steps > MaxSteps || !finite(b.DT) || b.DT <= 0 || b.DT > 1 {
		return s.snapshot(), errors.New("budget needs 1–256 steps and physical dt in (0,1] seconds")
	}
	end := s.time + float64(b.Steps)*b.DT
	if !finite(end) || end > maxTime || end <= s.time || s.time+b.DT <= s.time || !finite(2*math.Pi*s.config.Frequency*end) {
		return s.snapshot(), errors.New("unrepresentable time or phase")
	}
	if b.Context != nil {
		if err := b.Context.Err(); err != nil {
			return s.snapshot(), err
		}
	}
	if s.state != "running" {
		return s.snapshot(), nil
	}
	if b.Steps > MaxSolverSteps-s.steps {
		return s.snapshot(), ErrWorkLimit
	}
	for i := 0; i < b.Steps; i++ {
		if b.Context != nil {
			if err := b.Context.Err(); err != nil {
				return s.snapshot(), err
			}
		}
		s.step(b.DT)
		s.revision++
	}
	return s.snapshot(), nil
}

func (s *Session) step(dt float64) {
	prevE := s.e
	s.time += dt
	if s.config.Mode != Manual {
		driveMax := s.material.Parameters.Ec * 2
		if s.config.Model == Landau {
			driveMax = s.material.Parameters.Ec * 2.5
		}
		phase := math.Mod(2*math.Pi*s.config.Frequency*s.time, 2*math.Pi)
		if s.config.Mode == Sine {
			s.e = driveMax * math.Sin(phase)
		} else {
			p := phase / (2 * math.Pi)
			if p < 0.25 {
				s.e = driveMax * (4 * p)
			} else if p < 0.75 {
				s.e = driveMax * (2 - 4*p)
			} else {
				s.e = driveMax * (4*p - 4)
			}
		}
	}
	if s.config.Model == Landau {
		prevP := s.p
		s.p = s.lk.Step(s.e, dt)
		if !finite(s.p) {
			s.p = prevP
			if !finite(s.p) {
				s.p = s.defaultP()
			}
			s.lk.SetState(s.p)
		}
	} else {
		s.p = s.preisach.Update(s.e)
	}
	s.steps++
	s.syncLevel()
	// Match native trail decimation; Page exposes every retained row, not the
	// supplementary plot selection. SolverSteps counts all numerical steps.
	if s.config.Model == Landau && s.count > 0 && s.lastHistory >= 0 && s.time-s.lastHistory < 5e-4 {
		return
	}
	s.lastHistory = s.time
	if s.config.Mode == Manual && s.count > 0 {
		prevE = s.at(s.count - 1).E
	}
	branch := "held"
	if s.e > prevE {
		branch = "ascending"
	} else if s.e < prevE {
		branch = "descending"
	}
	point := sample{Index: s.steps - 1, Time: s.time, E: s.e, P: s.p, Branch: branch}
	if s.count < len(s.history) {
		s.history[(s.head+s.count)%len(s.history)] = point
		s.count++
	} else {
		s.history[s.head] = point
		s.head = (s.head + 1) % len(s.history)
		s.dropped++
	}
}
func (s *Session) syncLevel() {
	norm := 0.0
	if ps := s.material.Parameters.Ps * s.config.RangeFraction; ps != 0 {
		norm = s.p / ps
	}
	norm = math.Max(-1, math.Min(1, norm))
	s.level = int(math.Round((norm + 1) / 2 * float64(s.config.Levels-1)))
}
func (s *Session) maxField() float64 {
	ec := s.material.Parameters.Ec
	if s.config.Model == Preisach {
		ec = s.preisach.GetEffectiveEc()
	}
	return ec * 2.5
}
func (s *Session) Snapshot() Snapshot { s.mu.Lock(); defer s.mu.Unlock(); return s.snapshot() }
