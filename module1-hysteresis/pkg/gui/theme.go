// Package gui provides a Fyne-based graphical user interface for the hysteresis demo.
package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Exported colors - accessible from widgets subpackage
var (
	ColorPrimary    = color.RGBA{0, 212, 255, 255}   // Cyan
	ColorSecondary  = color.RGBA{255, 107, 107, 255} // Coral red
	ColorAccent     = color.RGBA{78, 205, 196, 255}  // Teal
	ColorWarning    = color.RGBA{255, 230, 109, 255} // Yellow
	ColorBackground = color.RGBA{0, 50, 100, 255}    // FeCIM blue #003264
	ColorGrid       = color.RGBA{0, 70, 130, 128}    // Grid lines (lighter blue)
	ColorAxis       = color.RGBA{150, 180, 200, 255} // Axis lines
	ColorPositive   = color.RGBA{255, 100, 100, 255} // Positive polarization
	ColorNegative   = color.RGBA{100, 150, 255, 255} // Negative polarization
)

// Legacy unexported aliases for backward compatibility within gui package
var (
	colorPrimary    = ColorPrimary
	colorSecondary  = ColorSecondary
	colorAccent     = ColorAccent
	colorWarning    = ColorWarning
	colorBackground = ColorBackground
	colorGrid       = ColorGrid
	colorAxis       = ColorAxis
	colorPositive   = ColorPositive
	colorNegative   = ColorNegative
)

// ============================================================
// Custom Theme
// ============================================================

type feCIMTheme struct{}

func (t *feCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ColorBackground // FeCIM blue #003264
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return ColorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255} // Slightly lighter blue
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255} // Darker blue for inputs
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255} // Separator lines
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *feCIMTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *feCIMTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *feCIMTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// ============================================================
// Fixed Width Layout
// ============================================================

// fixedWidthLayout is a custom layout that enforces a fixed width
type fixedWidthLayout struct {
	width float32
}

func (l *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minH := float32(0)
	for _, o := range objects {
		if o.Visible() {
			minH = fyne.Max(minH, o.MinSize().Height)
		}
	}
	return fyne.NewSize(l.width, minH)
}

func (l *fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(l.width, size.Height))
		o.Move(fyne.NewPos(0, 0))
	}
}

// fixedMinWidthLayout enforces a minimum width but allows expansion
type fixedMinWidthLayout struct {
	minWidth float32
}

func (l *fixedMinWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minH := float32(0)
	for _, o := range objects {
		if o.Visible() {
			minH = fyne.Max(minH, o.MinSize().Height)
		}
	}
	return fyne.NewSize(l.minWidth, minH)
}

func (l *fixedMinWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		// Only use the given width, but respect child's MinSize for height
		// This prevents vertical stretching
		childMinSize := o.MinSize()
		o.Resize(fyne.NewSize(size.Width, childMinSize.Height))
		o.Move(fyne.NewPos(0, 0))
	}
}
