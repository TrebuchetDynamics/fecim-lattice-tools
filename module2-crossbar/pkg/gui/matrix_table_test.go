package gui

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestMatrixTableModelProvidesStableCellAndHeaderText(t *testing.T) {
	model := NewMatrixTableModel([][]float64{
		{0.125, 1.5},
		{2.25, 3.0},
	})

	if rows, cols := model.Dimensions(); rows != 2 || cols != 2 {
		t.Fatalf("Dimensions = %dx%d, want 2x2", rows, cols)
	}
	if got := model.CellText(0, 0); got != "0.125" {
		t.Fatalf("CellText(0,0) = %q, want 0.125", got)
	}
	if got := model.CellText(1, 1); got != "3.000" {
		t.Fatalf("CellText(1,1) = %q, want 3.000", got)
	}
	if got := model.ColumnHeader(1); got != "C2" {
		t.Fatalf("ColumnHeader(1) = %q, want C2", got)
	}
	if got := model.RowHeader(1); got != "R2" {
		t.Fatalf("RowHeader(1) = %q, want R2", got)
	}
}

func TestNewMatrixTableUsesVirtualizedFyneCallbacks(t *testing.T) {
	fyneTest.NewTempApp(t)

	model := NewMatrixTableModel([][]float64{{0.1, 0.2}, {0.3, 0.4}})
	table := NewMatrixTable(model)
	if table == nil {
		t.Fatal("NewMatrixTable returned nil")
	}

	if rows, cols := table.Length(); rows != 2 || cols != 2 {
		t.Fatalf("table.Length = %dx%d, want 2x2", rows, cols)
	}
	if !table.ShowHeaderRow || !table.ShowHeaderColumn {
		t.Fatal("matrix table should show row and column headers")
	}

	cell := table.CreateCell()
	table.UpdateCell(widget.TableCellID{Row: 1, Col: 0}, cell)
	if got := cell.(*widget.Label).Text; got != "0.300" {
		t.Fatalf("updated cell text = %q, want 0.300", got)
	}

	header := table.CreateHeader()
	table.UpdateHeader(widget.TableCellID{Row: -1, Col: 1}, header)
	if got := header.(*widget.Label).Text; got != "C2" {
		t.Fatalf("column header text = %q, want C2", got)
	}
	table.UpdateHeader(widget.TableCellID{Row: 1, Col: -1}, header)
	if got := header.(*widget.Label).Text; got != "R2" {
		t.Fatalf("row header text = %q, want R2", got)
	}
}
