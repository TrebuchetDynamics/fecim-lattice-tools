//go:build legacy_fyne

// Package widgets provides reusable UI components.
package widgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	widgetsa11y "fecim-lattice-tools/shared/widgets/accessibility"
)

// AccessibilityMode represents the current accessibility setting.
type AccessibilityMode = widgetsa11y.AccessibilityMode

const (
	// AccessibilityNormal is the standard mode.
	AccessibilityNormal = widgetsa11y.AccessibilityNormal
	// AccessibilityHighContrast uses higher contrast colors.
	AccessibilityHighContrast = widgetsa11y.AccessibilityHighContrast
	// AccessibilityLargeText uses larger text sizes.
	AccessibilityLargeText = widgetsa11y.AccessibilityLargeText
)

// HighContrastColors provides WCAG AAA compliant colors (7:1 contrast ratio).
var HighContrastColors = widgetsa11y.HighContrastColors

// FocusIndicator wraps a widget with a visible focus indicator.
type FocusIndicator = widgetsa11y.FocusIndicator

// NewFocusIndicator creates a focus indicator wrapper.
func NewFocusIndicator(content fyne.CanvasObject) *FocusIndicator {
	return widgetsa11y.NewFocusIndicator(content)
}

// AccessibleButton creates a button with enhanced accessibility.
func AccessibleButton(label, accessibleName string, icon fyne.Resource, onTap func()) *widget.Button {
	return widgetsa11y.AccessibleButton(label, accessibleName, icon, onTap)
}

// KeyboardNavigationHelp creates a help dialog showing keyboard shortcuts.
func KeyboardNavigationHelp() fyne.CanvasObject { return widgetsa11y.KeyboardNavigationHelp() }

// ShowKeyboardHelp displays the keyboard navigation help dialog.
func ShowKeyboardHelp(parent fyne.Window) { widgetsa11y.ShowKeyboardHelp(parent) }

// CreateAccessibilityMenu creates a menu with accessibility options.
func CreateAccessibilityMenu(parent fyne.Window) *fyne.MenuItem {
	return widgetsa11y.CreateAccessibilityMenu(parent)
}

// SkipToContent creates a skip link for keyboard users.
func SkipToContent(target fyne.Focusable) *widget.Button { return widgetsa11y.SkipToContent(target) }

// ContrastChecker verifies WCAG contrast compliance.
type ContrastChecker = widgetsa11y.ContrastChecker

// Announce sends an accessibility announcement to assistive technology bridges.
func Announce(message string) { widgetsa11y.Announce(message) }

// LastAnnouncement returns the most recent accessibility announcement.
func LastAnnouncement() string { return widgetsa11y.LastAnnouncement() }

// SetAccessibleLabel stores an accessibility label for a canvas object.
func SetAccessibleLabel(obj fyne.CanvasObject, label string) {
	widgetsa11y.SetAccessibleLabel(obj, label)
}

// GetAccessibleLabel returns the stored accessibility label for obj.
func GetAccessibleLabel(obj fyne.CanvasObject) (string, bool) {
	return widgetsa11y.GetAccessibleLabel(obj)
}

func resetAccessibilityStateForTest() { widgetsa11y.ResetStateForTest() }

// HighContrastTheme returns a high contrast theme variant.
type HighContrastTheme = widgetsa11y.HighContrastTheme

// NewHighContrastTheme creates a high contrast theme wrapper.
func NewHighContrastTheme(base fyne.Theme) *HighContrastTheme {
	return widgetsa11y.NewHighContrastTheme(base)
}
