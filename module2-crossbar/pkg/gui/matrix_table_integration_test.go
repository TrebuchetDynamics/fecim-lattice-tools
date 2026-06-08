package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestEnhancedLayoutIncludesVirtualizedConductanceMatrixTable(t *testing.T) {
	fyneTest.NewTempApp(t)

	app, err := NewCrossbarApp()
	if err != nil {
		t.Fatalf("NewCrossbarApp: %v", err)
	}
	_ = app.createEnhancedMainLayout()

	matrixTab := findTabByText(app.tabs, "Matrix Table")
	if matrixTab == nil {
		t.Fatal("enhanced layout should include a Matrix Table tab")
	}

	table := findMatrixTable(matrixTab.Content)
	if table == nil {
		t.Fatal("Matrix Table tab should contain a virtualized widget.Table")
	}
	rows, cols := table.Length()
	if rows != app.config.Rows || cols != app.config.Cols {
		t.Fatalf("matrix table dimensions = %dx%d, want %dx%d", rows, cols, app.config.Rows, app.config.Cols)
	}
	if !table.ShowHeaderRow || !table.ShowHeaderColumn {
		t.Fatal("matrix table should include row and column headers")
	}
}

func TestConductanceDisplayRefreshUpdatesMatrixTable(t *testing.T) {
	fyneTest.NewTempApp(t)

	app, err := NewCrossbarApp()
	if err != nil {
		t.Fatalf("NewCrossbarApp: %v", err)
	}
	_ = app.createEnhancedMainLayout()

	matrixTab := findTabByText(app.tabs, "Matrix Table")
	if matrixTab == nil {
		t.Fatal("enhanced layout should include a Matrix Table tab")
	}
	table := findMatrixTable(matrixTab.Content)
	if table == nil {
		t.Fatal("Matrix Table tab should contain a virtualized widget.Table")
	}

	if err := app.array.ProgramWeight(0, 0, 1.0); err != nil {
		t.Fatalf("ProgramWeight: %v", err)
	}
	app.updateConductanceDisplay()

	waitForTableCellText(t, table, 0, 0, "1.000")
}

func findTabByText(tabs *container.AppTabs, text string) *container.TabItem {
	if tabs == nil {
		return nil
	}
	for _, item := range tabs.Items {
		if item.Text == text {
			return item
		}
	}
	return nil
}

func findMatrixTable(obj any) *widget.Table {
	switch o := obj.(type) {
	case *widget.Table:
		return o
	case *fyne.Container:
		for _, child := range o.Objects {
			if table := findMatrixTable(child); table != nil {
				return table
			}
		}
	case *container.Scroll:
		return findMatrixTable(o.Content)
	}
	return nil
}

func waitForTableCellText(t *testing.T, table *widget.Table, row, col int, want string) {
	t.Helper()
	deadline := time.Now().Add(500 * time.Millisecond)
	cell := table.CreateCell()
	for time.Now().Before(deadline) {
		table.UpdateCell(widget.TableCellID{Row: row, Col: col}, cell)
		if got := cell.(*widget.Label).Text; got == want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	table.UpdateCell(widget.TableCellID{Row: row, Col: col}, cell)
	t.Fatalf("table cell [%d,%d] text = %q, want %q", row, col, cell.(*widget.Label).Text, want)
}
