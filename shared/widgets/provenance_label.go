//go:build legacy_fyne

package widgets

import (
	"image/color"

	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/widgets/display"
)

// ProvenanceLabel is a composite widget that displays a formatted value,
// an inline confidence badge, and an optional source-reference subtitle.
type ProvenanceLabel = display.ProvenanceLabel

// NewProvenanceLabel creates a ProvenanceLabel and initialises its base widget.
func NewProvenanceLabel(value string, prov physics.Provenance, sourceRef string) *ProvenanceLabel {
	return display.NewProvenanceLabel(value, prov, sourceRef)
}

func provenanceColor(p physics.Provenance) color.Color {
	return display.ProvenanceColor(p)
}

func provenanceDisplayName(p physics.Provenance) string {
	return display.ProvenanceDisplayName(p)
}
