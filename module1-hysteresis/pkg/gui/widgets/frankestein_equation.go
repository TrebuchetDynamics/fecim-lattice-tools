// Package widgets provides custom GUI widgets for the hysteresis visualization.
package widgets

import (
	"encoding/json"
	"image/color"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const (
	frankesteinEquationSVGPath     = "data/equations/frankestein.svg"
	frankesteinEquationHotspotPath = "data/equations/frankestein.hotspots.json"
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
	if _, err := os.Stat(frankesteinEquationSVGPath); err == nil {
		if widget := newFrankesteinEquationImageWidget(parent, frankesteinEquationSVGPath); widget != nil {
			return widget
		}
	}
	return newFrankesteinEquationTextWidget(parent)
}

func newFrankesteinEquationTextWidget(parent fyne.Window) fyne.CanvasObject {
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

type hotspotDef struct {
	ID      string  `json:"id"`
	Tooltip string  `json:"tooltip"`
	X       float32 `json:"x"`
	Y       float32 `json:"y"`
	W       float32 `json:"w"`
	H       float32 `json:"h"`
}

type hotspotConfig struct {
	BaseWidth  float32      `json:"base_width"`
	BaseHeight float32      `json:"base_height"`
	Hotspots   []hotspotDef `json:"hotspots"`
}

type normalizedHotspotLayout struct {
	hotspots []hotspotDef
}

func (l *normalizedHotspotLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for i, obj := range objects {
		if i >= len(l.hotspots) {
			break
		}
		spot := l.hotspots[i]
		obj.Move(fyne.NewPos(size.Width*spot.X, size.Height*spot.Y))
		obj.Resize(fyne.NewSize(size.Width*spot.W, size.Height*spot.H))
	}
}

func (l *normalizedHotspotLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 0)
}

type Hotspot struct {
	widget.BaseWidget
	parent       fyne.Window
	tooltip      string
	tooltipPopup *widget.PopUp
	tooltipLabel *widget.Label
	debug        bool
}

func NewHotspot(parent fyne.Window, tooltip string, debug bool) *Hotspot {
	h := &Hotspot{
		parent:  parent,
		tooltip: tooltip,
		debug:   debug,
	}
	h.ExtendBaseWidget(h)
	return h
}

func (h *Hotspot) CreateRenderer() fyne.WidgetRenderer {
	fill := color.NRGBA{A: 0}
	stroke := color.NRGBA{A: 0}
	if h.debug {
		fill = color.NRGBA{R: 255, G: 0, B: 0, A: 48}
		stroke = color.NRGBA{R: 255, G: 0, B: 0, A: 120}
	}
	rect := canvas.NewRectangle(fill)
	rect.StrokeColor = stroke
	rect.StrokeWidth = 1
	return widget.NewSimpleRenderer(rect)
}

func (h *Hotspot) MouseIn(e *desktop.MouseEvent) {
	h.showTooltipAt(e.AbsolutePosition)
}

func (h *Hotspot) MouseMoved(e *desktop.MouseEvent) {
	h.showTooltipAt(e.AbsolutePosition)
}

func (h *Hotspot) MouseOut() {
	h.hideTooltip()
}

func (h *Hotspot) Tapped(_ *fyne.PointEvent) {
	if h.tooltipPopup != nil {
		h.hideTooltip()
		return
	}
	if h.parent == nil || h.tooltip == "" {
		return
	}
	app := fyne.CurrentApp()
	if app == nil {
		return
	}
	driver := app.Driver()
	if driver == nil {
		return
	}
	pos := driver.AbsolutePositionForObject(h)
	h.showTooltipAt(pos.Add(fyne.NewPos(8, 8)))
}

func (h *Hotspot) TappedSecondary(_ *fyne.PointEvent) {
	h.Tapped(nil)
}

func (h *Hotspot) showTooltipAt(pos fyne.Position) {
	if h.parent == nil || h.tooltip == "" {
		return
	}
	if h.tooltipLabel == nil {
		h.tooltipLabel = widget.NewLabel(h.tooltip)
		h.tooltipLabel.Wrapping = fyne.TextWrapWord
	} else {
		h.tooltipLabel.SetText(h.tooltip)
	}
	if h.tooltipPopup == nil {
		h.tooltipPopup = widget.NewPopUp(container.NewPadded(h.tooltipLabel), h.parent.Canvas())
	}
	h.tooltipPopup.ShowAtPosition(fyne.NewPos(pos.X+10, pos.Y+10))
}

func (h *Hotspot) hideTooltip() {
	if h.tooltipPopup != nil {
		h.tooltipPopup.Hide()
		h.tooltipPopup = nil
	}
}

func newFrankesteinEquationImageWidget(parent fyne.Window, svgPath string) fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"Frankestein Equation (Module 1)",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	hotspots, minSize := loadFrankesteinHotspots()
	debug := os.Getenv("FECIM_EQUATION_DEBUG") == "1"

	var hotspotWidgets []fyne.CanvasObject
	for _, spot := range hotspots {
		hotspotWidgets = append(hotspotWidgets, NewHotspot(parent, spot.Tooltip, debug))
	}

	image := canvas.NewImageFromFile(svgPath)
	image.FillMode = canvas.ImageFillContain
	if minSize.Width > 0 && minSize.Height > 0 {
		image.SetMinSize(minSize)
	}

	overlay := container.New(&normalizedHotspotLayout{hotspots: hotspots}, hotspotWidgets...)
	stack := container.NewStack(image, overlay)

	caption := widget.NewLabel("Hover or tap any coefficient to see its purpose in Module 1.")
	caption.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		title,
		stack,
		caption,
	)
}

func loadFrankesteinHotspots() ([]hotspotDef, fyne.Size) {
	defaultHotspots, defaultSize := defaultFrankesteinHotspots()
	data, err := os.ReadFile(frankesteinEquationHotspotPath)
	if err != nil {
		return defaultHotspots, defaultSize
	}

	var cfg hotspotConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("failed to parse hotspots file: %v", err)
		return defaultHotspots, defaultSize
	}

	hotspots := defaultHotspots
	if len(cfg.Hotspots) > 0 {
		hotspots = cfg.Hotspots
	}

	size := defaultSize
	if cfg.BaseWidth > 0 && cfg.BaseHeight > 0 {
		size = fyne.NewSize(cfg.BaseWidth, cfg.BaseHeight)
	}

	return hotspots, size
}

func defaultFrankesteinHotspots() ([]hotspotDef, fyne.Size) {
	return []hotspotDef{
		{
			ID:      "rho_eff_main",
			Tooltip: "Effective viscosity: intrinsic damping plus series-resistance RC delay.",
			X:       0.05, Y: 0.12, W: 0.12, H: 0.12,
		},
		{
			ID:      "e_applied",
			Tooltip: "Applied electric field drive term (external voltage across the film).",
			X:       0.34, Y: 0.12, W: 0.16, H: 0.12,
		},
		{
			ID:      "k_dep",
			Tooltip: "Depolarization factor: models interfacial layer; slants the loop for analog states.",
			X:       0.55, Y: 0.12, W: 0.10, H: 0.12,
		},
		{
			ID:      "alpha",
			Tooltip: "Dynamic stiffness: temperature + stress dependent curvature of energy wells.",
			X:       0.22, Y: 0.30, W: 0.08, H: 0.12,
		},
		{
			ID:      "beta",
			Tooltip: "First-order nonlinearity: negative for HZO to create the switching barrier.",
			X:       0.40, Y: 0.30, W: 0.08, H: 0.12,
		},
		{
			ID:      "gamma",
			Tooltip: "Sixth-order stabilizer: keeps energy bounded at large polarization.",
			X:       0.58, Y: 0.30, W: 0.08, H: 0.12,
		},
		{
			ID:      "noise",
			Tooltip: "Stochastic noise term (optional): captures thermal variability.",
			X:       0.24, Y: 0.48, W: 0.10, H: 0.12,
		},
		{
			ID:      "rho_eff_def",
			Tooltip: "Effective viscosity definition used in the headless hysteresis path.",
			X:       0.06, Y: 0.64, W: 0.12, H: 0.12,
		},
		{
			ID:      "rho",
			Tooltip: "Intrinsic viscosity / damping coefficient.",
			X:       0.30, Y: 0.64, W: 0.06, H: 0.12,
		},
		{
			ID:      "r_series",
			Tooltip: "Series resistance: absorbs RC delay into viscosity.",
			X:       0.44, Y: 0.64, W: 0.12, H: 0.12,
		},
	}, fyne.NewSize(1200, 320)
}
