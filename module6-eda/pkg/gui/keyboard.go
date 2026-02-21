// Package gui provides Fyne-based GUI components for EDA suite.
package gui

import (
	"fecim-lattice-tools/shared/keyboard"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Custom actions specific to EDA module.
const (
	actionJumpView1 keyboard.Action = "eda_view_1"
	actionJumpView2 keyboard.Action = "eda_view_2"
	actionJumpView3 keyboard.Action = "eda_view_3"
	actionJumpView4 keyboard.Action = "eda_view_4"
)

// SetupKeyboard initializes keyboard shortcuts for the EDA window.
func SetupKeyboard(w fyne.Window, viewSelector *widget.Select) {
	km := keyboard.NewManager(w)

	// Standard handlers
	km.SetHandlers(map[keyboard.Action]func(){
		keyboard.ActionSave: func() {
			showInfoDialog(w, "Save", "Save: Feature coming soon")
		},
		keyboard.ActionExport: func() {
			showInfoDialog(w, "Export", "Export: Use the Export buttons in each tab")
		},
		keyboard.ActionReset: func() {
			if viewSelector != nil && len(viewSelector.Options) > 0 {
				viewSelector.SetSelected(viewSelector.Options[0])
			}
		},
		keyboard.ActionPauseResume: func() {
			if viewSelector != nil {
				keyboard.SelectNextOption(viewSelector)
			}
		},
		keyboard.ActionNavigateLeft: func() {
			if viewSelector != nil {
				keyboard.SelectPrevOption(viewSelector)
			}
		},
		keyboard.ActionNavigateRight: func() {
			if viewSelector != nil {
				keyboard.SelectNextOption(viewSelector)
			}
		},
		keyboard.ActionHelp: func() {
			ShowKeyboardHelp(w)
		},
	})

	// M6-specific compound shortcuts
	km.AddCustomShortcut(actionJumpView1, fyne.Key1, 0, "Go to Builder & Validation")
	km.SetHandler(actionJumpView1, func() {
		if viewSelector != nil && len(viewSelector.Options) > 0 {
			viewSelector.SetSelected(viewSelector.Options[0])
		}
	})

	km.AddCustomShortcut(actionJumpView2, fyne.Key2, 0, "Go to Layout Visualizer")
	km.SetHandler(actionJumpView2, func() {
		if viewSelector != nil && len(viewSelector.Options) > 1 {
			viewSelector.SetSelected(viewSelector.Options[1])
		}
	})

	km.AddCustomShortcut(actionJumpView3, fyne.Key3, 0, "Go to Learn")
	km.SetHandler(actionJumpView3, func() {
		if viewSelector != nil && len(viewSelector.Options) > 2 {
			viewSelector.SetSelected(viewSelector.Options[2])
		}
	})

	km.AddCustomShortcut(actionJumpView4, fyne.Key4, 0, "Go to Flow Scripts")
	km.SetHandler(actionJumpView4, func() {
		if viewSelector != nil && len(viewSelector.Options) > 3 {
			viewSelector.SetSelected(viewSelector.Options[3])
		}
	})

	// NOTE: Ctrl+Shift+G/V/E (Generate All / Validate All / Export Package) are
	// registered directly by MakeBuilderValidationTab via window.Canvas().AddShortcut
	// using fyne.KeyModifierShortcutDefault (Cmd on macOS, Ctrl on Linux/Windows).
	// Do NOT add them here — duplicate Canvas.AddShortcut registrations overwrite
	// the working handlers, breaking the shortcuts on Linux/Windows.

	km.Register()
}

// showInfoDialog shows a simple info dialog (kept as helper used by handlers).
func showInfoDialog(w fyne.Window, title, message string) {
	fyne.Do(func() {
		dialog.ShowInformation(title, message, w)
	})
}

// ShowKeyboardHelp displays a dialog with all keyboard shortcuts
func ShowKeyboardHelp(w fyne.Window) {
	helpText := keyboard.FormatHelpMetadata(keyboard.HelpMetadata{
		Sections: []keyboard.ShortcutSection{
			{Title: "Navigation", Shortcuts: []keyboard.ShortcutMetadata{
				{Key: "←/→", Description: "Switch views"},
				{Key: "Space", Description: "Cycle through views"},
				{Key: "1", Description: "Go to Builder & Validation"},
				{Key: "2", Description: "Go to Layout Visualizer"},
				{Key: "3", Description: "Go to Learn"},
				{Key: "4", Description: "Go to Flow Scripts"},
			}},
			{Title: "Data", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Ctrl+S", Description: "Save (coming soon)"}, {Key: "Ctrl+E", Description: "Export (use tab buttons)"}, {Key: "Ctrl+R", Description: "Reset to first view"}}},
			{Title: "Builder Actions", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Ctrl+Shift+G", Description: "Generate All"}, {Key: "Ctrl+Shift+V", Description: "Validate All"}, {Key: "Ctrl+Shift+E", Description: "Export Package"}}},
			{Title: "Help", Shortcuts: []keyboard.ShortcutMetadata{{Key: "/ or ?", Description: "Show this help dialog"}}},
		},
		Tips: []string{"Use number keys (1-4) to quickly jump to any view", "Each tab has its own export functionality", "Generated files are educational examples only"},
	})

	keyboard.ShowHelpTextDialog(w, "Keyboard Shortcuts", helpText, 360, 320)
}
