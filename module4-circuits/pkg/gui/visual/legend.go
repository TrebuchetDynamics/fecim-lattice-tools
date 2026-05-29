//go:build legacy_fyne

package visual

// VCellLegendSpec describes the signed cell-voltage legend shown by the unified view.
type VCellLegendSpec struct {
	Title    string
	Min      float64
	Max      float64
	Ticks    []float64
	TickText []string
	SignText string
}

// NewVCellLegendSpec returns the canonical symmetric legend for signed cell voltage.
func NewVCellLegendSpec(maxAbs float64) VCellLegendSpec {
	if maxAbs <= 0 {
		maxAbs = 1.0
	}
	return VCellLegendSpec{
		Title:    "Cell Voltage (V)",
		Min:      -maxAbs,
		Max:      maxAbs,
		Ticks:    []float64{-maxAbs, 0, maxAbs},
		TickText: []string{"-Vmax", "0", "+Vmax"},
		SignText: "+ = BL>WL", // negative implies WL>BL
	}
}
