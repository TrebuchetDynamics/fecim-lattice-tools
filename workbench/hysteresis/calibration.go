package hysteresis

import (
	"errors"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
)

// MaxCalibrationSteps bounds native LK calibration separately from the live
// session allowance. Up to 126 target writes each need at most ~1e6 native
// steps (100 pulse durations / (pulse duration / 10000)); reserve rounding room.
// This admits all 2–64 level sweeps, not a browser responsiveness claim.
const MaxCalibrationSteps = 128000000

type CalibrationStatus string

const (
	CalibrationUnavailable CalibrationStatus = "unavailable"
	CalibrationFresh       CalibrationStatus = "fresh"
	CalibrationStale       CalibrationStatus = "stale"
)

// Fresh means a completed table matches current settings, not that every target
// converged or that a device was measured. FallbackTargets explicitly counts the
// native LK ±Ec fallback estimates. No partial table is exposed as a result.
type CalibrationSnapshot struct {
	Status                CalibrationStatus
	Reason, Method        string
	ConfigRevision        uint64
	Config                Config
	CanApply, CanExport   bool
	FallbackTargets       int
	Ascending, Descending []Quantity
}
type calibrationState struct {
	job                      *controller.LevelCalibration
	manager                  *algo.CalibrationManager
	revision, configRevision uint64
	config                   Config
	stale                    bool
	fallbacks                int
	method                   string
}

func (p *Programmer) calibrationSnapshot() CalibrationSnapshot {
	s := p.session
	c := &p.calibration
	v := CalibrationSnapshot{Status: CalibrationUnavailable, Reason: "No completed in-memory level calibration", ConfigRevision: c.configRevision, Config: c.config, Method: c.method, FallbackTargets: c.fallbacks}
	if c.manager == nil {
		return v
	}
	v.Status = CalibrationFresh
	v.Reason = "Completed simulation calibration; not device measurements"
	if c.stale || c.revision != s.calibrationRevision {
		v.Status = CalibrationStale
		v.Reason = "Settings changed or recalibration interrupted; explicitly recalibrate before applying or exporting"
	}
	v.CanApply = v.Status == CalibrationFresh
	v.CanExport = v.CanApply
	v.Ascending = make([]Quantity, len(c.manager.CalibrationUp))
	v.Descending = make([]Quantity, len(c.manager.CalibrationDown))
	for i := range v.Ascending {
		v.Ascending[i] = field(c.manager.CalibrationUp[i])
		v.Descending[i] = field(c.manager.CalibrationDown[i])
	}
	return v
}

// ExportCalibration returns only a fresh, owned in-memory table. It does no
// serialization or persistence; a future consented local adapter owns file I/O.
func (p *Programmer) ExportCalibration() (CalibrationSnapshot, error) {
	s := p.session
	s.mu.Lock()
	defer s.mu.Unlock()
	v := p.calibrationSnapshot()
	if !v.CanExport {
		return CalibrationSnapshot{}, errors.New(v.Reason)
	}
	return v, nil
}
func (p *Programmer) calibrationManager() *algo.CalibrationManager {
	if p.calibration.manager != nil && !p.calibration.stale && p.calibration.revision == p.session.calibrationRevision {
		return p.calibration.manager
	}
	return nil
}
func (p *Programmer) startCalibration() error {
	s := p.session
	job, err := controller.NewLevelCalibration(s.material.Parameters, s.config.Model == Landau, s.config.Levels, s.config.RangeFraction, s.config.Temperature, s.config.Stress)
	if err != nil {
		return err
	}
	p.calibration.job = job
	p.calibration.stale = true
	p.phase = IdlePhase
	p.phaseTime = 0
	p.target = 0
	p.queued = 0
	p.ctrl = nil
	p.work = 0
	p.status = ProgramRunning
	p.reason = "Calibrating in memory with native scratch solvers"
	s.owner = p
	return nil
}
func (p *Programmer) advanceCalibration() {
	c := &p.calibration
	s := p.session
	c.job.Step()
	if !c.job.Done() {
		return
	}
	c.manager, _ = c.job.Result()
	c.fallbacks = c.job.Fallbacks()
	c.revision = s.calibrationRevision
	c.configRevision = s.configRevision
	c.config = s.config
	c.stale = false
	c.method = "native Preisach zero-field binary search"
	if s.config.Model == Landau {
		c.method = "native single-domain LK write-controller sweep (±Ec fallback estimates counted)"
	}
	p.finish(ProgramComplete, "Calibration completed in memory; no files saved")
}
func (p *Programmer) learnCalibration() {
	c := p.calibrationManager()
	if c == nil {
		return
	}
	idx := p.target - 1
	ec := p.session.material.Parameters.Ec
	if p.target > p.session.config.Levels/2 {
		c.UpdateCalibrationUp(idx, 0, ec)
		c.CalibrationUp[idx] = p.ctrl.CurrentField
	} else {
		c.UpdateCalibrationDown(idx, 0, ec)
		c.CalibrationDown[idx] = p.ctrl.CurrentField
	}
}
