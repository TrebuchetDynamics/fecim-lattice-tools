package simulation

import (
	"fmt"
	"sync"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

// CellCoord identifies a cell inside a multi-cell array.
type CellCoord struct {
	Row int
	Col int
}

// CellState captures per-cell dynamic state in a multi-cell simulation.
type CellState struct {
	Voltage       float64
	ElectricField float64
	Polarization  float64
	NormPol       float64
}

// MultiCellArray provides a foundational API for simulating multiple
// ferroelectric cells with independent hysteresis states.
type MultiCellArray struct {
	rows int
	cols int

	material *ferroelectric.HZOMaterial
	models   [][]*ferroelectric.PreisachModel
	states   [][]CellState

	mu sync.RWMutex
}

// NewMultiCellArray creates a multi-cell simulation array where each cell owns
// an independent Preisach model/history.
func NewMultiCellArray(rows, cols int, material *ferroelectric.HZOMaterial) (*MultiCellArray, error) {
	if rows <= 0 || cols <= 0 {
		return nil, fmt.Errorf("invalid array dimensions: %dx%d", rows, cols)
	}
	if material == nil {
		return nil, fmt.Errorf("material cannot be nil")
	}

	arr := &MultiCellArray{
		rows:     rows,
		cols:     cols,
		material: material,
		models:   make([][]*ferroelectric.PreisachModel, rows),
		states:   make([][]CellState, rows),
	}

	for r := 0; r < rows; r++ {
		arr.models[r] = make([]*ferroelectric.PreisachModel, cols)
		arr.states[r] = make([]CellState, cols)
		for c := 0; c < cols; c++ {
			arr.models[r][c] = ferroelectric.NewPreisachModel(material)
		}
	}

	return arr, nil
}

// Size returns array dimensions.
func (m *MultiCellArray) Size() (rows, cols int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rows, m.cols
}

// StepCell updates one cell using an externally provided cell voltage.
func (m *MultiCellArray) StepCell(row, col int, voltage float64) (CellState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.inBounds(row, col) {
		return CellState{}, fmt.Errorf("cell index out of bounds: (%d,%d)", row, col)
	}

	state := m.stepCellLocked(row, col, voltage)
	return state, nil
}

// StepWithVoltageMap updates all cells with the provided voltage matrix.
func (m *MultiCellArray) StepWithVoltageMap(voltageMap [][]float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(voltageMap) != m.rows {
		return fmt.Errorf("voltage map rows mismatch: got %d, want %d", len(voltageMap), m.rows)
	}
	for r := range voltageMap {
		if len(voltageMap[r]) != m.cols {
			return fmt.Errorf("voltage map cols mismatch at row %d: got %d, want %d", r, len(voltageMap[r]), m.cols)
		}
	}

	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.cols; c++ {
			m.stepCellLocked(r, c, voltageMap[r][c])
		}
	}
	return nil
}

// StepWithSelector updates only specified coordinates with the same voltage.
// Useful as a foundation for wordline/bitline selection schemes.
func (m *MultiCellArray) StepWithSelector(cells []CellCoord, voltage float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cell := range cells {
		if !m.inBounds(cell.Row, cell.Col) {
			return fmt.Errorf("cell index out of bounds: (%d,%d)", cell.Row, cell.Col)
		}
		m.stepCellLocked(cell.Row, cell.Col, voltage)
	}
	return nil
}

// GetCellState returns a copy of one cell state.
func (m *MultiCellArray) GetCellState(row, col int) (CellState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.inBounds(row, col) {
		return CellState{}, fmt.Errorf("cell index out of bounds: (%d,%d)", row, col)
	}
	return m.states[row][col], nil
}

// Snapshot returns a deep copy of all cell states.
func (m *MultiCellArray) Snapshot() [][]CellState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([][]CellState, m.rows)
	for r := 0; r < m.rows; r++ {
		out[r] = make([]CellState, m.cols)
		copy(out[r], m.states[r])
	}
	return out
}

// Reset clears all per-cell model states.
func (m *MultiCellArray) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.cols; c++ {
			m.models[r][c].Reset()
			m.states[r][c] = CellState{}
		}
	}
}

func (m *MultiCellArray) stepCellLocked(row, col int, voltage float64) CellState {
	field := voltage / m.material.Thickness
	pol := m.models[row][col].Update(field)
	state := CellState{
		Voltage:       voltage,
		ElectricField: field,
		Polarization:  pol,
		NormPol:       m.models[row][col].NormalizedPolarization(),
	}
	m.states[row][col] = state
	return state
}

func (m *MultiCellArray) inBounds(row, col int) bool {
	return row >= 0 && row < m.rows && col >= 0 && col < m.cols
}
