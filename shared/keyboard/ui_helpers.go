package keyboard

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// SelectNextTab selects the next tab in an AppTabs control.
func SelectNextTab(tabs *container.AppTabs) bool {
	if tabs == nil || len(tabs.Items) == 0 {
		return false
	}
	items := tabs.Items
	currentIdx := 0
	for i, item := range items {
		if item == tabs.Selected() {
			currentIdx = i
			break
		}
	}
	nextIdx := (currentIdx + 1) % len(items)
	tabs.Select(items[nextIdx])
	return true
}

// SelectPrevTab selects the previous tab in an AppTabs control.
func SelectPrevTab(tabs *container.AppTabs) bool {
	if tabs == nil || len(tabs.Items) == 0 {
		return false
	}
	items := tabs.Items
	currentIdx := 0
	for i, item := range items {
		if item == tabs.Selected() {
			currentIdx = i
			break
		}
	}
	prevIdx := currentIdx - 1
	if prevIdx < 0 {
		prevIdx = len(items) - 1
	}
	tabs.Select(items[prevIdx])
	return true
}

// SelectNextOption selects the next option in a widget.Select.
func SelectNextOption(selector *widget.Select) bool {
	if selector == nil || len(selector.Options) == 0 {
		return false
	}
	currentIdx := 0
	for i, opt := range selector.Options {
		if opt == selector.Selected {
			currentIdx = i
			break
		}
	}
	nextIdx := (currentIdx + 1) % len(selector.Options)
	selector.SetSelected(selector.Options[nextIdx])
	return true
}

// SelectPrevOption selects the previous option in a widget.Select.
func SelectPrevOption(selector *widget.Select) bool {
	if selector == nil || len(selector.Options) == 0 {
		return false
	}
	currentIdx := 0
	for i, opt := range selector.Options {
		if opt == selector.Selected {
			currentIdx = i
			break
		}
	}
	prevIdx := currentIdx - 1
	if prevIdx < 0 {
		prevIdx = len(selector.Options) - 1
	}
	selector.SetSelected(selector.Options[prevIdx])
	return true
}

// ShowHelpTextDialog displays raw help text in a standard scrollable dialog.
func ShowHelpTextDialog(w fyne.Window, title, helpText string, width, height float32) {
	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	helpContent := container.NewVScroll(helpLabel)
	helpContent.SetMinSize(fyne.NewSize(width, height))

	helpDialog := dialog.NewCustom(title, "Close", helpContent, w)
	helpDialog.Show()
}
