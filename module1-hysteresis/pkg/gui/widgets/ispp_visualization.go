package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
)

// ISPPVisualization is a placeholder widget (implementation WIP)
type ISPPVisualization struct {
	widget.BaseWidget
	stats *physics.WriteVerifyStats
}

func NewISPPVisualization() *ISPPVisualization {
	return &ISPPVisualization{
		stats: physics.NewWriteVerifyStats(),
	}
}

func (v *ISPPVisualization) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabel("ISPP Visualization (Coming Soon)")
	return widget.NewSimpleRenderer(label)
}

func (v *ISPPVisualization) AddPulse(voltage float64) {
	// Placeholder
}

func (v *ISPPVisualization) AddConductance(conductance float64) {
	// Placeholder
}

func (v *ISPPVisualization) SetTarget(target float64) {
	// Placeholder
}

func (v *ISPPVisualization) Refresh() {
	v.BaseWidget.Refresh()
}

func (v *ISPPVisualization) MinSize() fyne.Size {
	return fyne.NewSize(300, 100)
}

func (v *ISPPVisualization) GetStats() *physics.WriteVerifyStats {
	return v.stats
}

func (v *ISPPVisualization) SetAnimationState(_ float64, _ float64, _ int, _ float64, _ bool) {
	// Placeholder
}
