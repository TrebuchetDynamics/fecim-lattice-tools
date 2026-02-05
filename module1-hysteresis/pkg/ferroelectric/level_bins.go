package ferroelectric

import "math"

// LevelBins defines discrete polarization bins with a guard band for read margin.
// RangeFrac scales the usable polarization range (1.0 = full Ps range).
// GuardFrac is the fraction of each bin width considered unsafe near the edges.
type LevelBins struct {
	NumLevels int
	Ps        float64
	RangeFrac float64
	GuardFrac float64
}

// NewLevelBins creates a LevelBins helper with sane bounds.
func NewLevelBins(ps float64, numLevels int, rangeFrac float64, guardFrac float64) LevelBins {
	if numLevels < 2 {
		numLevels = 2
	}
	if rangeFrac <= 0 || rangeFrac > 1 {
		rangeFrac = 1
	}
	if guardFrac < 0 {
		guardFrac = 0
	}
	if guardFrac > 0.5 {
		guardFrac = 0.5
	}
	return LevelBins{
		NumLevels: numLevels,
		Ps:        ps,
		RangeFrac: rangeFrac,
		GuardFrac: guardFrac,
	}
}

// EffectivePs returns the target saturation used for spacing levels.
func (b LevelBins) EffectivePs() float64 {
	if b.Ps == 0 {
		return 0
	}
	rangeFrac := b.RangeFrac
	if rangeFrac <= 0 || rangeFrac > 1 {
		rangeFrac = 1
	}
	return b.Ps * rangeFrac
}

// Step returns the center-to-center spacing between adjacent levels.
func (b LevelBins) Step() float64 {
	if b.NumLevels <= 1 || b.Ps == 0 {
		return 0
	}
	effectivePs := b.EffectivePs()
	if effectivePs == 0 {
		return 0
	}
	return 2 * effectivePs / float64(b.NumLevels-1)
}

// LevelForP maps polarization to a 1-based level index, reports guard-band status,
// and returns the signed delta from the bin center.
func (b LevelBins) LevelForP(P float64) (level int, inError bool, delta float64) {
	effectivePs := b.EffectivePs()
	if b.Ps == 0 || effectivePs == 0 {
		return 1, true, P
	}

	if P > effectivePs {
		P = effectivePs
	} else if P < -effectivePs {
		P = -effectivePs
	}

	step := b.Step()
	if step == 0 {
		return 1, true, P
	}

	norm := (P/effectivePs + 1) / 2
	idx := int(math.Round(norm * float64(b.NumLevels-1)))
	if idx < 0 {
		idx = 0
	} else if idx > b.NumLevels-1 {
		idx = b.NumLevels - 1
	}

	center := -effectivePs + float64(idx)*step
	low := center - 0.5*step
	high := center + 0.5*step
	guard := b.GuardFrac * step

	if guard > 0 {
		distToEdge := math.Min(P-low, high-P)
		inError = distToEdge < guard
	}

	return idx + 1, inError, P - center
}
