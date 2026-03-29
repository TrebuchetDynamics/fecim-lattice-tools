package physics

import "math"

// AgingEngine models coupled wake-up, fatigue, and retention degradation.
//
// The wakeup model captures the non-monotonic Pr trajectory observed in HZO:
// Pr initially increases during early cycling (tetragonal -> orthorhombic
// phase conversion) before declining at high cycle counts (orthorhombic ->
// monoclinic fatigue).
//
// References:
//   - Pesic et al., Advanced Functional Materials 26, 2016
//   - Zhou et al., ACS Applied Materials & Interfaces 12, 2020
//   - PMC11789571 (2025): 90-degree UCDW -> 180-degree DW transition
type AgingEngine struct {
	PrFresh            float64
	WakeupBoost        float64 // fractional max increase by wake-up
	WakeupTauCycles    float64 // characteristic wake-up cycle count
	FatigueStartCycle  float64 // cycle count where fatigue begins
	FatigueRate        float64 // fatigue strength coefficient
	RetentionTauSec    float64 // retention decay time constant (seconds)
	CycleHistory       []int
	LastCycle          int
	LastPolarizationPr float64
}

func NewAgingEngine(prFresh float64) *AgingEngine {
	return &AgingEngine{
		PrFresh:           prFresh,
		WakeupBoost:       0.18,
		WakeupTauCycles:   250,
		FatigueStartCycle: 1_000,
		FatigueRate:       0.18,
		RetentionTauSec:   3600 * 24 * 30,
		CycleHistory:      make([]int, 0, 64),
	}
}

// ApplyCycle updates aging state for a target cycle and retention hold time.
//
// The wakeup factor models the non-monotonic Pr behavior: Pr rises during
// early cycling (saturating at WakeupTauCycles) then degrades when the
// cycle count exceeds FatigueStartCycle. The wakeup contribution is not
// artificially capped; instead, the exponential saturation naturally limits
// the wakeup gain.
//
// The fatigue factor uses a stretched-power-law model that engages only
// after FatigueStartCycle, producing a gradual decline consistent with
// orthorhombic -> monoclinic phase conversion at high cycle counts.
func (a *AgingEngine) ApplyCycle(cycle int, holdTimeSec float64) float64 {
	if cycle < 1 {
		cycle = 1
	}

	// Wakeup: exponential saturation toward full boost. The wakeup gain
	// naturally saturates via the exponential, no hard cap needed.
	wakeup := 1 + a.WakeupBoost*(1-math.Exp(-float64(cycle)/a.WakeupTauCycles))

	// Fatigue: gradual stretched-power-law decay after onset.
	fatigue := 1.0
	if float64(cycle) > a.FatigueStartCycle {
		denom := 1_000_000 - a.FatigueStartCycle
		if denom <= 0 {
			denom = 1
		}
		x := (float64(cycle) - a.FatigueStartCycle) / denom
		if x > 1 {
			x = 1
		}
		fatigue = math.Exp(-a.FatigueRate * x)
	}

	// Retention: exponential decay with hold time.
	retention := 1.0
	if holdTimeSec > 0 {
		retention = math.Exp(-holdTimeSec / a.RetentionTauSec)
	}

	pr := a.PrFresh * wakeup * fatigue * retention
	a.LastCycle = cycle
	a.LastPolarizationPr = pr
	a.CycleHistory = append(a.CycleHistory, cycle)
	return pr
}
