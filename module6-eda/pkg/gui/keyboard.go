// Package gui provides Fyne-based GUI components for EDA suite.
package gui

import (
	"fecim-lattice-tools/shared/keyboard"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// SetupKeyboard initializes keyboard shortcuts for the EDA window.
func SetupKeyboard(w fyne.Window, viewSelector *widget.Select) {
	// Register Ctrl+S for save
	ctrlS := &desktop.CustomShortcut{
		KeyName:  fyne.KeyS,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlS, func(_ fyne.Shortcut) {
		showInfoDialog(w, "Save", "Save: Feature coming soon")
	})

	// Register Ctrl+E for export
	ctrlE := &desktop.CustomShortcut{
		KeyName:  fyne.KeyE,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlE, func(_ fyne.Shortcut) {
		showInfoDialog(w, "Export", "Export: Use the Export buttons in each tab")
	})

	// Register Ctrl+R for reset
	ctrlR := &desktop.CustomShortcut{
		KeyName:  fyne.KeyR,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlR, func(_ fyne.Shortcut) {
		if viewSelector != nil && len(viewSelector.Options) > 0 {
			viewSelector.SetSelected(viewSelector.Options[0])
		}
	})

	// Handle non-modifier keys
	w.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		switch ke.Name {
		case fyne.KeySpace:
			// Toggle between views
			if viewSelector != nil {
				cycleViewSelector(viewSelector)
			}

		case fyne.Key1:
			if viewSelector != nil && len(viewSelector.Options) > 0 {
				viewSelector.SetSelected(viewSelector.Options[0])
			}

		case fyne.Key2:
			if viewSelector != nil && len(viewSelector.Options) > 1 {
				viewSelector.SetSelected(viewSelector.Options[1])
			}

		case fyne.KeySlash:
			ShowKeyboardHelp(w)

		case fyne.KeyLeft:
			if viewSelector != nil {
				prevView(viewSelector)
			}

		case fyne.KeyRight:
			if viewSelector != nil {
				nextView(viewSelector)
			}
		}
	})
}

// cycleViewSelector cycles to the next view
func cycleViewSelector(selector *widget.Select) {
	keyboard.SelectNextOption(selector)
}

// nextView switches to the next view
func nextView(selector *widget.Select) {
	cycleViewSelector(selector)
}

// prevView switches to the previous view
func prevView(selector *widget.Select) {
	keyboard.SelectPrevOption(selector)
}

// showInfoDialog shows a simple info dialog
func showInfoDialog(w fyne.Window, title, message string) {
	dialog.ShowInformation(title, message, w)
}

// ShowKeyboardHelp displays a dialog with all keyboard shortcuts
func ShowKeyboardHelp(w fyne.Window) {
	helpText := keyboard.FormatHelpMetadata(keyboard.HelpMetadata{
		Sections: []keyboard.ShortcutSection{
			{Title: "Navigation", Shortcuts: []keyboard.ShortcutMetadata{{Key: "←/→", Description: "Switch views"}, {Key: "Space", Description: "Cycle through views"}, {Key: "1", Description: "Go to Builder & Validation"}, {Key: "2", Description: "Go to Learn"}}},
			{Title: "Data", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Ctrl+S", Description: "Save (coming soon)"}, {Key: "Ctrl+E", Description: "Export (use tab buttons)"}, {Key: "Ctrl+R", Description: "Reset to first view"}}},
			{Title: "Builder Actions", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Ctrl+Shift+G", Description: "Generate All"}, {Key: "Ctrl+Shift+V", Description: "Validate All"}, {Key: "Ctrl+Shift+E", Description: "Export Package"}}},
			{Title: "Help", Shortcuts: []keyboard.ShortcutMetadata{{Key: "/ or ?", Description: "Show this help dialog"}}},
		},
		Tips: []string{"Use number keys (1-2) to quickly jump to views", "Each tab has its own export functionality", "Generated files are educational examples only"},
	})

	keyboard.ShowHelpTextDialog(w, "Keyboard Shortcuts", helpText, 360, 320)
}
