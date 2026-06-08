package tabs

import (
	"reflect"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

func TestExportViewerTabIncludesArtifactTree(t *testing.T) {
	app := fyneTest.NewTempApp(t)
	window := app.NewWindow("Export viewer")
	defer window.Close()

	cfg := &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}
	content := MakeExportViewerTab(cfg, window)

	tree := findTree(content)
	if tree == nil {
		t.Fatal("export viewer should include a widget.Tree artifact browser")
	}

	if got, want := tree.ChildUIDs(""), []string{"cells", "data"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("artifact tree root children = %#v, want %#v", got, want)
	}
	if !tree.IsBranch("data") {
		t.Fatal("data should be a branch")
	}

	dataChildren := tree.ChildUIDs("data")
	if !containsTreeID(dataChildren, "data/fecim_crossbar_4x4.v") {
		t.Fatalf("data children = %#v, want generated Verilog artifact", dataChildren)
	}
}

func TestExportViewerArtifactTreeSelectionUpdatesStatus(t *testing.T) {
	app := fyneTest.NewTempApp(t)
	window := app.NewWindow("Export viewer")
	defer window.Close()

	cfg := &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}
	content := MakeExportViewerTab(cfg, window)

	tree := findTree(content)
	if tree == nil {
		t.Fatal("export viewer should include a widget.Tree artifact browser")
	}
	tree.Select("data/fecim_crossbar_4x4.v")

	status := findLabelContaining(content, "Artifact: data/fecim_crossbar_4x4.v")
	if status == nil {
		t.Fatal("selecting artifact tree leaf should update status with selected full-path ID")
	}
}

func findTree(obj fyne.CanvasObject) *widget.Tree {
	switch typed := obj.(type) {
	case *widget.Tree:
		return typed
	case *fyne.Container:
		for _, child := range typed.Objects {
			if found := findTree(child); found != nil {
				return found
			}
		}
	case *container.Scroll:
		return findTree(typed.Content)
	case *container.Split:
		if found := findTree(typed.Leading); found != nil {
			return found
		}
		return findTree(typed.Trailing)
	}
	return nil
}

func findLabelContaining(obj fyne.CanvasObject, text string) *widget.Label {
	switch typed := obj.(type) {
	case *widget.Label:
		if strings.Contains(typed.Text, text) {
			return typed
		}
	case *fyne.Container:
		for _, child := range typed.Objects {
			if found := findLabelContaining(child, text); found != nil {
				return found
			}
		}
	case *container.Scroll:
		return findLabelContaining(typed.Content, text)
	case *container.Split:
		if found := findLabelContaining(typed.Leading, text); found != nil {
			return found
		}
		return findLabelContaining(typed.Trailing, text)
	}
	return nil
}

func containsTreeID(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
