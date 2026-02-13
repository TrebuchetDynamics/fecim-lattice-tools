package ferroelectric

import sharedphysics "fecim-lattice-tools/shared/physics"

// LevelBins defines discrete polarization bins with a guard band for read margin.
// Re-exported from shared/physics for backward compatibility.
type LevelBins = sharedphysics.LevelBins

// NewLevelBins creates a LevelBins helper with sane bounds.
// Re-exported from shared/physics for backward compatibility.
func NewLevelBins(ps float64, numLevels int, rangeFrac float64, guardFrac float64) LevelBins {
	return sharedphysics.NewLevelBins(ps, numLevels, rangeFrac, guardFrac)
}
