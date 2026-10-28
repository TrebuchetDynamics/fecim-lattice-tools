package hysteresis

import (
	"errors"
	"math"
	"math/rand/v2"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

var ErrDriveOwned = errors.New("programming owns session drive; cancel or reset through Programmer")

type ProgramMode string

const (
	TargetWrite  ProgramMode = "target-write"
	WriteRead    ProgramMode = "write-read"
	TimeResolved ProgramMode = "time-resolved"
)

type ProgramStatus string

const (
	ProgramIdle               ProgramStatus = "idle"
	ProgramRunning            ProgramStatus = "running"
	ProgramComplete           ProgramStatus = "complete"
	ProgramFailed             ProgramStatus = "failed"
	ProgramCancelled          ProgramStatus = "cancelled"
	ProgramWorkLimit          ProgramStatus = "work_limit"
	LegacyProtocolUnavailable ProgramStatus = "legacy_protocol_unavailable"
)

type ProgramPhase string

const (
	IdlePhase    ProgramPhase = "idle"
	PrepPhase    ProgramPhase = "prep"
	WritePhase   ProgramPhase = "write"
	DisplayPhase ProgramPhase = "display"
)

type ISPPMethod string

const (
	LinearISPP      ISPPMethod = "linear"
	LogarithmicISPP ISPPMethod = "logarithmic"
)

// ProgramConfig controls orchestration, never physics equations. Domains > 1
// selects native polydomain LK only in WriteRead. Seed comes from Session.Config.
// WorkLimit bounds the entire job; PulseLimit uses the native total+pulse guard.
// Config{} selects defaults only in NewProgrammer; other configs are complete.
type ProgramConfig struct {
	Mode                           ProgramMode
	Method                         ISPPMethod
	Domains, WorkLimit, PulseLimit int
	CalibrationWorkLimit           int
	GuardFraction                  float64
}

func DefaultProgramConfig() ProgramConfig {
	return ProgramConfig{Mode: TargetWrite, Method: LogarithmicISPP, Domains: 1, WorkLimit: MaxSolverSteps, PulseLimit: 30, GuardFraction: 0.15, CalibrationWorkLimit: MaxCalibrationSteps}
}

type ProgramCommandKind string

const (
	WriteTarget      ProgramCommandKind = "write-target"
	CancelProgram    ProgramCommandKind = "cancel"
	ResetProgram     ProgramCommandKind = "reset"
	CalibrateProgram ProgramCommandKind = "calibrate"
)

// Targets are native 1-based levels; Core.Level remains zero-based.
type ProgramCommand struct {
	Kind   ProgramCommandKind
	Target int
}

type ControllerSnapshot struct {
	Phase                                                                                    controller.WriteState
	PhaseText                                                                                string
	Field, LowerBound, UpperBound                                                            Quantity
	Pulses, TotalPulses, GuardPulses, Overshoots, ResetDirection, LastVerifyLevel, LastError int
}
type ProgrammingSnapshot struct {
	Core                                              Snapshot
	Config                                            ProgramConfig
	Status                                            ProgramStatus
	Reason                                            string
	Phase                                             ProgramPhase
	Target, QueuedTarget, ReadLevel, LastBranch, Work int
	OwnsDrive, Complete                               bool
	Controller                                        ControllerSnapshot
	Calibration                                       CalibrationSnapshot
}

// Programmer shares Session's mutex, drive and persistent solvers. It starts no
// goroutines, uses only a private seeded RNG, and never writes calibration files.
type Programmer struct {
	session                                                 *Session
	config                                                  ProgramConfig
	status                                                  ProgramStatus
	reason                                                  string
	phase                                                   ProgramPhase
	phaseTime                                               float64
	target, queued, readLevel, startLevel, lastBranch, work int
	ctrl                                                    *controller.WriteController
	rng                                                     *rand.Rand
	ensemble                                                bool
	skipPrep                                                bool
	nextCycle                                               bool
	forceReset                                              bool
	stepping                                                bool // release ownership only after the final admitted solver call
	rngSeed                                                 uint64
	calibration                                             calibrationState
}

func NewProgrammer(s *Session, c ProgramConfig) (*Programmer, error) {
	if s == nil {
		return nil, errors.New("session is required")
	}
	if c == (ProgramConfig{}) {
		c = DefaultProgramConfig()
	}
	if c.Mode != TargetWrite && c.Mode != WriteRead && c.Mode != TimeResolved {
		return nil, errors.New("unsupported programming mode")
	}
	if c.Method != LinearISPP && c.Method != LogarithmicISPP {
		return nil, errors.New("unsupported ISPP method; DCC is not implemented")
	}
	if c.Domains < 1 || c.Domains > 128 || (c.Domains > 1 && c.Mode != WriteRead) || c.WorkLimit < 1 || c.WorkLimit > MaxSolverSteps || c.CalibrationWorkLimit < 1 || c.CalibrationWorkLimit > MaxCalibrationSteps || c.PulseLimit < 1 || c.PulseLimit > 1000 || !finite(c.GuardFraction) || c.GuardFraction < 0 || c.GuardFraction >= 0.5 {
		return nil, errors.New("programming configuration outside supported bounds")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.Domains > 1 && s.config.Model != Landau {
		return nil, errors.New("polydomain programming requires LK")
	}
	p := &Programmer{session: s, config: c, status: ProgramIdle, phase: IdlePhase}
	p.resetRNG()
	if c.Mode == TimeResolved {
		p.status = LegacyProtocolUnavailable
		p.reason = "Time-Resolved Switching is an unavailable legacy GUI stub; no switching protocol or numerical trace is implemented"
	}
	return p, nil
}
func (p *Programmer) resetRNG() {
	p.rngSeed = p.session.config.Seed
	p.rng = rand.New(rand.NewPCG(p.rngSeed, 0))
}
func (p *Programmer) Snapshot() ProgrammingSnapshot {
	s := p.session
	s.mu.Lock()
	defer s.mu.Unlock()
	return p.snapshot()
}
func (p *Programmer) snapshot() ProgrammingSnapshot {
	s := p.session
	v := ProgrammingSnapshot{Core: s.snapshot(), Config: p.config, Status: p.status, Reason: p.reason, Phase: p.phase, Target: p.target, QueuedTarget: p.queued, ReadLevel: p.readLevel, LastBranch: p.lastBranch, Work: p.work, OwnsDrive: s.owner == p, Complete: p.status == ProgramComplete, Calibration: p.calibrationSnapshot()}
	if p.config.Mode == TimeResolved {
		v.Core.Plot = nil
	} // existing session rows are not switching-protocol samples
	if p.ctrl != nil {
		w := p.ctrl
		v.Controller = ControllerSnapshot{Phase: w.State, PhaseText: w.State.String(), Field: field(w.CurrentField), LowerBound: field(w.VMin), UpperBound: field(w.VMax), Pulses: w.PulseCount, TotalPulses: w.TotalPulses, GuardPulses: w.GuardPulseCount, Overshoots: w.OvershootCount, ResetDirection: w.ResetDirection(), LastVerifyLevel: w.LastVerifyLevel, LastError: w.LastError}
	}
	return v
}
func (p *Programmer) Apply(c ProgramCommand) (ProgrammingSnapshot, error) {
	s := p.session
	s.mu.Lock()
	defer s.mu.Unlock()
	fail := func(e string) (ProgrammingSnapshot, error) { return p.snapshot(), errors.New(e) }
	if c.Kind != WriteTarget && c.Target != 0 {
		return fail("unexpected programming command payload")
	}
	if c.Kind != WriteTarget && c.Kind != CancelProgram && c.Kind != ResetProgram && c.Kind != CalibrateProgram {
		return fail("unsupported programming command")
	}
	if s.owner != nil && s.owner != p {
		return p.snapshot(), ErrDriveOwned
	}
	if p.config.Mode == TimeResolved {
		return p.snapshot(), nil
	}
	switch c.Kind {
	case WriteTarget:
		if c.Target < 1 || c.Target > s.config.Levels {
			return fail("target outside native 1-based level range")
		}
		if s.owner == p {
			if p.calibration.job != nil {
				return fail("calibration owns programming workflow")
			}
			if p.config.Mode != WriteRead {
				return fail("manual target write already active")
			}
			p.queued = c.Target
			s.revision++
			return p.snapshot(), nil
		}
		if p.config.Mode == TargetWrite && c.Target == s.level+1 {
			return p.snapshot(), nil
		}
		if s.config.Mode != Manual {
			return fail("select core Manual mode before programming")
		}
		if p.config.Domains > 1 && s.config.Model != Landau {
			return fail("polydomain programming requires LK")
		}
		p.work = 0
		if p.rngSeed != s.config.Seed {
			p.resetRNG()
		}
		p.forceReset = false
		p.nextCycle = false
		p.queued = 0
		p.reason = ""
		p.status = ProgramRunning
		p.readLevel = 0
		s.owner = p
		s.state = "running"
		if p.config.Domains > 1 {
			s.lk.EnableEnsemble(p.config.Domains, &s.material.Parameters, s.config.Seed)
			s.lk.SetState(s.p)
			s.p = s.lk.GetState()
			p.ensemble = true
		}
		p.beginTarget(c.Target, true)
	case CancelProgram:
		p.finish(ProgramCancelled, "Cancelled; retained steps are not a complete result")
	case ResetProgram:
		p.release()
		if _, err := s.apply(Command{Kind: Reset}); err != nil {
			return p.snapshot(), err
		}
		p.phase = IdlePhase
		p.status = ProgramIdle
		p.reason = ""
		p.target = 0
		p.queued = 0
		p.readLevel = 0
		p.startLevel = 0
		p.lastBranch = 0
		p.forceReset = false
		p.work = 0
		p.phaseTime = 0
		p.nextCycle = false
		p.ctrl = nil
		p.calibration.job = nil
		p.resetRNG()
	case CalibrateProgram:
		if s.owner == p {
			return fail("cancel active work before calibration")
		}
		if err := p.startCalibration(); err != nil {
			return p.snapshot(), err
		}
	}
	s.revision++
	return p.snapshot(), nil
}
func (p *Programmer) beginTarget(target int, first bool) {
	s := p.session
	p.target = target
	p.startLevel = s.level + 1
	p.phaseTime = 0
	ec := s.maxField() / 2.5
	p.ctrl = controller.NewWriteController(s.config.Levels, ec, 2.5*ec, p.calibrationManager())
	p.ctrl.StepMode = string(p.config.Method)
	p.ctrl.PulseDuration = 0.4 / s.config.Frequency
	prep := p.needsPrep(target, first)
	p.forceReset = false
	p.phase = PrepPhase
	p.skipPrep = !prep
	if !first && !prep {
		p.phase = WritePhase
		p.ctrl.Start(target, false)
	}
	if prep && p.config.Mode == WriteRead {
		s.preisach.Reset()
	}
}
func (p *Programmer) needsPrep(target int, first bool) bool {
	s := p.session
	if p.config.Mode == TargetWrite {
		return s.config.Model == Landau
	}
	force := target <= 3 || target >= s.config.Levels-2
	nextBranch := direction(float64(target - (s.level + 1)))
	return force || p.forceReset || (s.config.Model == Landau && (first || (p.lastBranch != 0 && p.lastBranch != nextBranch)))
}
func direction(v float64) int {
	if v > 0 {
		return 1
	}
	if v < 0 {
		return -1
	}
	return 0
}
func ramp(current, target, step float64) float64 {
	if math.Abs(target-current) < step {
		return target
	}
	return current + math.Copysign(step, target-current)
}
func (p *Programmer) release() {
	s := p.session
	if s.owner != p {
		return
	}
	if p.ensemble {
		s.lk.EnableEnsemble(1, &s.material.Parameters, 0)
		s.lk.SetState(s.p)
		s.p = s.lk.GetState()
		p.ensemble = false
	}
	s.owner = nil
	s.state = "paused"
}
func (p *Programmer) finish(status ProgramStatus, reason string) {
	p.status = status
	p.reason = reason
	p.calibration.job = nil
	if !p.stepping {
		p.release()
	}
	p.session.revision++
}

// Advance accepts the same bounded physical timestep contract as Session. Job
// cancellation/caps are terminal, not success; already completed work is retained.
func (p *Programmer) Advance(b Budget) (ProgrammingSnapshot, error) {
	s := p.session
	s.mu.Lock()
	defer s.mu.Unlock()
	if b.Steps < 1 || b.Steps > MaxSteps || !finite(b.DT) || b.DT <= 0 || b.DT > 1 {
		return p.snapshot(), errors.New("budget needs 1–256 steps and physical dt in (0,1] seconds")
	}
	if p.status != ProgramRunning {
		return p.snapshot(), nil
	}
	end := s.time + float64(b.Steps)*b.DT
	if !finite(end) || end > maxTime || s.time+b.DT <= s.time || !finite(2*math.Pi*s.config.Frequency*end) {
		return p.snapshot(), errors.New("unrepresentable time or phase")
	}
	for i := 0; i < b.Steps && p.status == ProgramRunning; i++ {
		if b.Context != nil {
			if err := b.Context.Err(); err != nil {
				p.finish(ProgramCancelled, err.Error())
				return p.snapshot(), err
			}
		}
		limit := p.config.WorkLimit
		if p.calibration.job != nil {
			limit = p.config.CalibrationWorkLimit
		}
		if p.work >= limit || (p.calibration.job == nil && s.steps >= MaxSolverSteps) {
			p.finish(ProgramWorkLimit, ErrWorkLimit.Error())
			return p.snapshot(), ErrWorkLimit
		}
		if p.calibration.job != nil {
			p.advanceCalibration()
		} else {
			p.step(b.DT)
		}
		p.work++
		s.revision++
	}
	return p.snapshot(), nil
}
func (p *Programmer) step(dt float64) {
	s := p.session
	p.stepping = true
	defer func() {
		p.stepping = false
		if p.status != ProgramRunning {
			p.release()
		}
	}()
	if p.nextCycle {
		target := p.queued
		p.queued = 0
		p.nextCycle = false
		p.beginTarget(target, false)
	}
	// Native level clicks queue until the next PREP or WRITE boundary. A
	// click during PREP can therefore replace the target at WRITE entry.
	if p.queued > 0 && p.phaseTime == 0 && (p.phase == PrepPhase || p.phase == WritePhase) {
		p.target = p.queued
		p.queued = 0
		p.startLevel = s.level + 1
		p.ctrl.ResetState()
		force := p.target <= 3 || p.target >= s.config.Levels-2
		if p.phase == WritePhase && !force {
			p.ctrl.Start(p.target, false)
		} else {
			p.phase = PrepPhase
			p.skipPrep = s.config.Model == Preisach && !force
			if !p.skipPrep {
				s.preisach.Reset()
			}
		}
	}
	p.phaseTime += dt
	duration := 1 / s.config.Frequency
	emax := s.maxField()
	rate := 4 * emax * s.config.Frequency
	if p.config.Mode == WriteRead {
		rate = 3 * s.material.Parameters.Ec * 2 * s.config.Frequency
	}
	switch p.phase {
	case PrepPhase:
		if p.skipPrep {
			p.phase = WritePhase
			p.phaseTime = 0
			p.ctrl.Start(p.target, false)
			break
		}
		targetE := emax
		if p.config.Mode == WriteRead {
			emax = s.material.Parameters.Ec * 2
			targetE = emax
		}
		if p.target > s.config.Levels/2 {
			targetE = -targetE
		}
		s.e = ramp(s.e, targetE, rate*dt)
		ref := s.material.Parameters.Pr
		if ref == 0 {
			ref = s.material.Parameters.Ps
		}
		threshold := 0.2 * ref
		hold := 0.3
		saturated := math.Abs(s.p) >= 0.5*s.material.Parameters.Ps && direction(s.p) == direction(targetE)
		if p.config.Mode == WriteRead {
			hold = 0.25
			if p.target <= 3 || p.target >= s.config.Levels-2 {
				threshold = 0.5 * s.material.Parameters.Ps
			}
		}
		biased := ref <= 0 || (math.Abs(s.p) >= threshold && direction(s.p) == direction(targetE))
		if p.phaseTime > duration*hold && math.Abs(s.e-targetE) < 0.01*emax && biased {
			p.phase = WritePhase
			p.phaseTime = 0
			p.ctrl.Start(p.target, p.config.Mode == TargetWrite || saturated)
		}
	case WritePhase:
		level, guard := s.level+1, 0
		if p.config.GuardFraction > 0 {
			ps := s.material.Parameters.Ps
			if ps == 0 {
				ps = s.material.Parameters.Pr
			}
			bins := ferroelectric.NewLevelBins(ps, s.config.Levels, s.config.RangeFraction, p.config.GuardFraction)
			l, inError, delta := bins.LevelForP(s.p)
			level = l
			if inError && l == p.target {
				guard = direction(delta)
			}
		}
		e, done := p.ctrl.Update(dt, s.e, level, guard)
		s.e = ramp(s.e, e, rate*1.5*dt)
		if !done && p.ctrl.TotalPulses+p.ctrl.PulseCount >= p.config.PulseLimit {
			p.finish(ProgramWorkLimit, "ISPP pulse limit reached; target is not complete")
		} else if done {
			if p.ctrl.State != controller.StateSuccess {
				p.finish(ProgramFailed, "ISPP controller did not converge")
			} else {
				p.phase = DisplayPhase
				p.phaseTime = 0
				if branch := direction(float64(p.target - p.startLevel)); branch != 0 {
					p.lastBranch = branch
				}
				p.forceReset = p.ctrl.MaxOvershootDelta > 3
				p.learnCalibration()
			}
		}
	case DisplayPhase:
		s.e = ramp(s.e, 0, rate*0.4*dt)
		if math.Abs(s.e) < 0.01*(emax/2.5) {
			p.readLevel = s.level + 1
		}
		hold := 0.4
		if p.config.Mode == WriteRead {
			hold = 0.6
		}
		if p.phaseTime > duration*hold && math.Abs(s.e) < 0.01*emax {
			if p.config.Mode == TargetWrite {
				p.finish(ProgramComplete, "")
			} else {
				target := p.queued
				if target == 0 {
					target = p.nextTarget()
				}
				p.queued = target
				p.nextCycle = true
				p.phase = WritePhase
				if p.needsPrep(target, false) {
					p.phase = PrepPhase
				}
				p.phaseTime = 0
			}
		}
	}
	s.step(dt)
}
func (p *Programmer) nextTarget() int {
	s := p.session
	current := s.level + 1
	lo, hi := 2, s.config.Levels-1
	if lo > hi {
		return 1 + (current % s.config.Levels)
	}
	dir := p.lastBranch
	if dir == 0 {
		dir = 1
		if current > s.config.Levels/2 {
			dir = -1
		}
	}
	if dir > 0 && current+1 <= hi {
		return current + 1 + p.rng.IntN(hi-current)
	}
	if lo <= current-1 {
		return lo + p.rng.IntN(current-lo)
	}
	if current+1 <= hi {
		return current + 1 + p.rng.IntN(hi-current)
	}
	return p.target
}
