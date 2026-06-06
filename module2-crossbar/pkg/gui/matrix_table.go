package gui

import (
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// MatrixTableModel adapts dense numeric matrices to Fyne's virtualized
// widget.Table callback shape.
type MatrixTableModel struct {
	mu   sync.RWMutex
	data [][]float64
	rows int
	cols int
}

// NewMatrixTableModel copies matrix data into a stable model for virtualized
// table rendering. Rows may be ragged; missing cells render as empty strings.
func NewMatrixTableModel(data [][]float64) *MatrixTableModel {
	model := &MatrixTableModel{}
	model.SetData(data)
	return model
}

// SetData replaces the matrix snapshot used by virtualized table callbacks.
func (m *MatrixTableModel) SetData(data [][]float64) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rows = len(data)
	m.cols = 0
	m.data = make([][]float64, len(data))
	for row, values := range data {
		if len(values) > m.cols {
			m.cols = len(values)
		}
		m.data[row] = append([]float64(nil), values...)
	}
}

// Dimensions returns the matrix row and column counts.
func (m *MatrixTableModel) Dimensions() (rows, cols int) {
	if m == nil {
		return 0, 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rows, m.cols
}

// CellText returns formatted matrix cell text, or an empty string for missing
// ragged cells/out-of-range coordinates.
func (m *MatrixTableModel) CellText(row, col int) string {
	if m == nil || row < 0 || col < 0 {
		return ""
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if row >= len(m.data) || col >= len(m.data[row]) {
		return ""
	}
	return fmt.Sprintf("%.3f", m.data[row][col])
}

// ColumnHeader returns a one-based column label.
func (m *MatrixTableModel) ColumnHeader(col int) string {
	if col < 0 {
		return ""
	}
	return fmt.Sprintf("C%d", col+1)
}

// RowHeader returns a one-based row label.
func (m *MatrixTableModel) RowHeader(row int) string {
	if row < 0 {
		return ""
	}
	return fmt.Sprintf("R%d", row+1)
}

// NewMatrixTable creates a virtualized Fyne table for matrix data.
func NewMatrixTable(model *MatrixTableModel) *widget.Table {
	if model == nil {
		model = NewMatrixTableModel(nil)
	}

	table := widget.NewTableWithHeaders(
		model.Dimensions,
		func() fyne.CanvasObject {
			label := widget.NewLabel("0.000")
			label.Alignment = fyne.TextAlignTrailing
			label.TextStyle = fyne.TextStyle{Monospace: true}
			return label
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cell.(*widget.Label).SetText(model.CellText(id.Row, id.Col))
		},
	)
	table.ShowHeaderRow = true
	table.ShowHeaderColumn = true
	table.CreateHeader = func() fyne.CanvasObject {
		label := widget.NewLabel("")
		label.Alignment = fyne.TextAlignCenter
		label.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
		return label
	}
	table.UpdateHeader = func(id widget.TableCellID, cell fyne.CanvasObject) {
		label := cell.(*widget.Label)
		switch {
		case id.Row < 0 && id.Col >= 0:
			label.SetText(model.ColumnHeader(id.Col))
		case id.Col < 0 && id.Row >= 0:
			label.SetText(model.RowHeader(id.Row))
		default:
			label.SetText("")
		}
	}
	return table
}
