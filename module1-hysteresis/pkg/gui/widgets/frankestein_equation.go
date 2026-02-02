// Package widgets provides custom GUI widgets for the hysteresis visualization.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// TermChip is a small hoverable label that shows a tooltip for a coefficient.
type TermChip struct {
	widget.BaseWidget
	parent       fyne.Window
	tooltip      string
	label        *widget.Label
	tooltipPopup *widget.PopUp
	tooltipLabel *widget.Label
}

// NewTermChip creates a new term chip with hover tooltip text.
func NewTermChip(parent fyne.Window, text, tooltip string) *TermChip {
	t := &TermChip{
		parent:  parent,
		tooltip: tooltip,
	}
	t.label = widget.NewLabel(text)
	t.label.TextStyle = fyne.TextStyle{Monospace: true}
	t.ExtendBaseWidget(t)
	return t
}

// CreateRenderer implements fyne.Widget.
func (t *TermChip) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.label)
}

// MouseIn shows the tooltip on hover.
func (t *TermChip) MouseIn(e *desktop.MouseEvent) {
	t.showTooltip(e)
}

// MouseMoved keeps the tooltip near the cursor.
func (t *TermChip) MouseMoved(e *desktop.MouseEvent) {
	t.showTooltip(e)
}

// MouseOut hides the tooltip.
func (t *TermChip) MouseOut() {
	t.hideTooltip()
}

func (t *TermChip) showTooltip(e *desktop.MouseEvent) {
	if t.parent == nil || t.tooltip == "" {
		return
	}
	if t.tooltipLabel == nil {
		t.tooltipLabel = widget.NewLabel(t.tooltip)
		t.tooltipLabel.Wrapping = fyne.TextWrapWord
	} else {
		t.tooltipLabel.SetText(t.tooltip)
	}
	if t.tooltipPopup == nil {
		t.tooltipPopup = widget.NewPopUp(container.NewPadded(t.tooltipLabel), t.parent.Canvas())
	}
	pos := fyne.NewPos(e.AbsolutePosition.X+10, e.AbsolutePosition.Y+10)
	t.tooltipPopup.ShowAtPosition(pos)
}

func (t *TermChip) hideTooltip() {
	if t.tooltipPopup != nil {
		t.tooltipPopup.Hide()
		t.tooltipPopup = nil
	}
}

func mathLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.TextStyle = fyne.TextStyle{Monospace: true}
	return label
}

// NewFrankesteinEquationWidget builds the equation display with tooltips.
func NewFrankesteinEquationWidget(parent fyne.Window) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"Frankestein Equation (Module 1)",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	line1 := container.NewHBox(
		NewTermChip(parent, "\\rho_{eff}", "Effective viscosity: intrinsic damping plus series-resistance RC delay."),
		mathLabel(" dP/dt = "),
		NewTermChip(parent, "E_{applied}", "Applied electric field drive term (external voltage across the film)."),
		mathLabel(" - "),
		NewTermChip(parent, "k_{dep}", "Depolarization factor: models interfacial layer; slants the loop for analog states."),
		mathLabel(" P - ("),
	)

	line2 := container.NewHBox(
		NewTermChip(parent, "2\\alpha", "Dynamic stiffness: temperature + stress dependent curvature of energy wells."),
		mathLabel(" P + "),
		NewTermChip(parent, "4\\beta", "First-order nonlinearity: negative for HZO to create the switching barrier."),
		mathLabel(" P^3 + "),
		NewTermChip(parent, "6\\gamma", "Sixth-order stabilizer: keeps energy bounded at large polarization."),
		mathLabel(" P^5)"),
	)

	line3 := container.NewHBox(
		mathLabel("+ "),
		NewTermChip(parent, "\\xi(t)", "Stochastic noise term (optional): captures thermal variability."),
	)

	line4 := container.NewHBox(
		NewTermChip(parent, "\\rho_{eff}", "Effective viscosity definition used in the headless hysteresis path."),
		mathLabel(" = "),
		NewTermChip(parent, "\\rho", "Intrinsic viscosity / damping coefficient."),
		mathLabel(" + ("),
		NewTermChip(parent, "R_{series}", "Series resistance: absorbs RC delay into viscosity."),
		mathLabel(" A) / d"),
	)

	caption := widget.NewLabel("Hover any coefficient to see its purpose in Module 1.")
	caption.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		title,
		line1,
		line2,
		line3,
		line4,
		caption,
	)
}
