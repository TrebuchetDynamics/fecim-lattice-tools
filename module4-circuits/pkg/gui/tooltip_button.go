package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// TooltipButton is a button that shows a lightweight tooltip on hover.
// It uses widget.PopUp to keep the implementation minimal and headless-test friendly.
type TooltipButton struct {
	widget.Button
	tooltipText string
	tooltip     *widget.PopUp
	window      fyne.Window
}

// NewTooltipButton creates a button with optional hover tooltip text.
// If window is nil, tooltips are suppressed.
func NewTooltipButton(label, tooltip string, window fyne.Window, tapped func()) *TooltipButton {
	btn := &TooltipButton{
		tooltipText: tooltip,
		window:      window,
	}
	btn.Text = label
	btn.OnTapped = tapped
	btn.ExtendBaseWidget(btn)
	return btn
}

// MouseIn shows tooltip on hover (desktop only).
func (b *TooltipButton) MouseIn(ev *desktop.MouseEvent) {
	b.Button.MouseIn(ev)
	b.showTooltip()
}

// MouseMoved is required to satisfy desktop.Hoverable.
func (b *TooltipButton) MouseMoved(ev *desktop.MouseEvent) {
	b.Button.MouseMoved(ev)
}

// MouseOut hides the tooltip.
func (b *TooltipButton) MouseOut() {
	b.Button.MouseOut()
	b.hideTooltip()
}

func (b *TooltipButton) showTooltip() {
	if b.tooltipText == "" || b.window == nil {
		return
	}
	if b.tooltip != nil {
		b.tooltip.Hide()
	}
	label := widget.NewLabel(b.tooltipText)
	label.Wrapping = fyne.TextWrapWord
	b.tooltip = widget.NewPopUp(label, b.window.Canvas())

	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(b)
	b.tooltip.ShowAtPosition(fyne.NewPos(pos.X, pos.Y+b.Size().Height+4))
}

func (b *TooltipButton) hideTooltip() {
	if b.tooltip == nil {
		return
	}
	b.tooltip.Hide()
	b.tooltip = nil
}
