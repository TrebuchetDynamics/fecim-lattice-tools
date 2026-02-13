package arraysim

import (
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

func TestSharedTechnologyNodeUsableInModule4(t *testing.T) {
	n := sharedphysics.TechnologyNodeFromName("65nm")
	g := CellGeometry{
		PitchX:           n.CellPitchX,
		PitchY:           n.CellRowHeight,
		WireWidth:        n.MetalWidth,
		WireThickness:    n.MetalThickness,
		MetalResistivity: n.MetalResistivity,
	}
	if g.PitchX <= 0 || g.PitchY <= 0 || g.WireWidth <= 0 {
		t.Fatalf("shared technology node not usable for module4 geometry: %+v", g)
	}
}
