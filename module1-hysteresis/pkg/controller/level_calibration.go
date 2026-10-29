package controller

import (
	"errors"
	"math"

	"fecim-lattice-tools/module1-hysteresis/pkg/algo"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

// LevelCalibration is the native GUI's level-table sweep, made resumable. Each
// Step performs at most one solver update. It owns scratch solvers; neither the
// interactive solver nor any file/cache is touched. The caller owns work caps
// and cancellation. LK retains native single-domain calibration, even when the
// interactive write/read solver uses an ensemble.
type LevelCalibration struct {
	material                                   physics.HZOMaterial
	lk                                         *physics.LKSolver
	preisach                                   *ferroelectric.PreisachModel
	manager                                    *algo.CalibrationManager
	ec, effPs, pulse, dt, elapsed, e           float64
	target, stage, iteration, sample, bestDiff int
	descending, done                           bool
	low, high, best, mid                       float64
	writer                                     *WriteController
	fallbacks                                  int
}

func NewLevelCalibration(mat physics.HZOMaterial, landau bool, levels int, fraction, temp, stress float64) (*LevelCalibration, error) {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	if levels < 2 || levels > 64 || !finite(fraction) || fraction < 0.5 || fraction > 1 || !finite(temp) || temp < 200 || temp > 700 || !finite(stress) || stress < 0 || stress > 5e9 || !finite(mat.Ec) || mat.Ec <= 0 || !finite(mat.Ps) || mat.Ps < 0 || !finite(mat.Pr) || mat.Pr < 0 {
		return nil, errors.New("invalid level calibration inputs")
	}
	j := &LevelCalibration{material: mat, manager: algo.NewCalibrationManager(levels), target: 1, ec: mat.Ec, effPs: mat.Ps * fraction}
	if j.effPs <= 0 {
		j.effPs = mat.Pr
	}
	if j.effPs <= 0 {
		j.effPs = 1
	}
	if landau {
		j.lk = physics.NewLKSolver()
		j.lk.ConfigureFromMaterial(&j.material)
		j.lk.Temperature = temp
		j.lk.Stress = stress
		j.lk.EnableNoise = false
		j.lk.UseNLS = false
		if !j.lk.UseMaterialAlpha {
			j.lk.UpdateParams()
		}
		j.pulse = mat.Tau
		if j.pulse <= 0 {
			j.pulse = 10e-9
		}
		j.dt = math.Min(1e-4, j.pulse/10000)
		if j.dt <= 0 {
			j.dt = 1e-12
		}
	} else {
		j.preisach = ferroelectric.NewPreisachModel(&j.material)
		j.preisach.SetTemperature(temp)
		j.ec = j.preisach.GetEffectiveEc()
		if j.ec == 0 {
			j.ec = mat.Ec
		}
	}
	for i := 0; i < levels; i++ {
		j.manager.CalibUpLow[i] = j.ec * 0.3
		j.manager.CalibUpHigh[i] = j.ec * 2
		j.manager.CalibDownLow[i] = -j.ec * 2
		j.manager.CalibDownHigh[i] = -j.ec * 0.3
	}
	return j, nil
}
func (j *LevelCalibration) Done() bool           { return j.done }
func (j *LevelCalibration) Fallbacks() int       { return j.fallbacks }
func (j *LevelCalibration) EffectiveEc() float64 { return j.ec }

// Result is available only after all targets, search iterations and table repair
// finish. The returned manager is caller-owned; the completed job no longer uses it.
func (j *LevelCalibration) Result() (*algo.CalibrationManager, error) {
	if !j.done {
		return nil, errors.New("calibration is incomplete")
	}
	return j.manager, nil
}
func (j *LevelCalibration) level(p float64) int {
	return int(math.Round((math.Max(-1, math.Min(1, p/j.effPs)) + 1) / 2 * float64(j.manager.NumLevels-1)))
}
func (j *LevelCalibration) nextTarget(e float64) {
	if j.descending {
		j.manager.CalibrationDown[j.target] = e
		j.target--
	} else {
		j.manager.CalibrationUp[j.target] = e
		j.target++
	}
	if !j.descending && j.target == j.manager.NumLevels {
		j.descending = true
		j.target = j.manager.NumLevels - 2
	}
	if j.descending && j.target < 0 {
		j.manager.RepairLevelTables(j.ec)
		j.done = true
	}
	j.stage = 0
	j.iteration = 0
	j.writer = nil
}

// Step preserves the GUI's 50-update saturation, zero-field read, 15-iteration
// Preisach binary search and LK controller timing/fallback policy, not the
// distinct headless CLI sweep and not a uniform field-spacing estimate.
func (j *LevelCalibration) Step() {
	if j.done {
		return
	}
	if j.lk != nil {
		j.stepLK()
		return
	}
	if j.stage == 0 {
		ratio := float64(j.target) / float64(j.manager.NumLevels-1)
		j.low = j.ec * 0.3
		j.high = j.ec * 2
		j.best = j.ec * (0.8 + ratio*1.2)
		if j.descending {
			ratio = float64(j.manager.NumLevels-1-j.target) / float64(j.manager.NumLevels-1)
			j.low = -j.ec * 2
			j.high = -j.ec * 0.3
			j.best = -j.ec * (0.8 + ratio*1.2)
		}
		j.bestDiff = j.manager.NumLevels
		j.stage = 1
	}
	if j.stage == 1 {
		j.mid = (j.low + j.high) / 2
		j.preisach.Reset()
		j.sample = 0
		j.stage = 2
	}
	sat := -2 * j.ec
	if j.descending {
		sat = -sat
	}
	switch {
	case j.sample < 50:
		j.preisach.Update(sat)
	case j.sample == 50:
		j.preisach.Update(0)
	case j.sample == 51:
		j.preisach.Update(j.mid)
	default:
		level := j.level(j.preisach.Update(0))
		diff := level - j.target
		if absCalibration(diff) < absCalibration(j.bestDiff) {
			j.bestDiff = diff
			j.best = j.mid
		}
		j.iteration++
		if diff == 0 || j.iteration == 15 {
			j.nextTarget(j.best)
			return
		}
		if level < j.target {
			j.low = j.mid
		} else {
			j.high = j.mid
		}
		j.stage = 1
	}
	j.sample++
}
func absCalibration(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
func (j *LevelCalibration) stepLK() {
	if j.writer == nil {
		j.writer = NewWriteController(j.manager.NumLevels, j.ec, j.ec*2.5, nil)
		j.writer.PulseDuration = j.pulse
		j.writer.Start(j.target+1, true)
		p := -math.Abs(j.effPs)
		if j.descending {
			p = -p
		}
		j.lk.SetState(p)
		j.lk.Time = 0
		j.e = 0
		j.elapsed = 0
	}
	e, done := j.writer.Update(j.dt, j.e, j.level(j.lk.GetState())+1, 0)
	j.e = e
	j.lk.Step(e, j.dt)
	j.elapsed += j.dt
	if done || j.elapsed >= j.pulse*float64(j.writer.MaxRetries+100) {
		e = j.writer.CurrentField
		if !done || j.writer.State != StateSuccess {
			j.fallbacks++
			e = j.ec
			if j.descending {
				e = -e
			}
		}
		j.nextTarget(e)
	}
}
