package simulation

import (
	"fmt"
	"sync"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

type CellCoord struct{ Row, Col int }

const maxMultiCellCount = 1_000_000

type CellState struct {
	Voltage       float64
	ElectricField float64
	Polarization  float64
	NormPol       float64
}

type MultiCellArray struct {
	rows, cols int
	material   *ferroelectric.HZOMaterial
	models     [][]*ferroelectric.PreisachModel
	states     [][]CellState
	mu         sync.RWMutex
}

func NewMultiCellArray(rows, cols int, material *ferroelectric.HZOMaterial) (*MultiCellArray, error) {
	if !isValidArrayDimensions(rows, cols) {
		return nil, fmt.Errorf("invalid array dimensions: %dx%d", rows, cols)
	}
	materialSnapshot := snapshotMaterial(material)
	if materialSnapshot == nil {
		return nil, fmt.Errorf("material cannot be nil")
	}
	if !isValidMaterialThickness(materialSnapshot.Thickness) {
		return nil, fmt.Errorf("material thickness must be finite and > 0 m: got %.3e m", materialSnapshot.Thickness)
	}
	firstModel := ferroelectric.NewPreisachModel(materialSnapshot)
	if firstModel == nil {
		return nil, fmt.Errorf("material cannot initialize Preisach model: %s", materialSnapshot.Name)
	}
	m := &MultiCellArray{rows: rows, cols: cols, material: materialSnapshot, models: make([][]*ferroelectric.PreisachModel, rows), states: make([][]CellState, rows)}
	for r := 0; r < rows; r++ {
		m.models[r] = make([]*ferroelectric.PreisachModel, cols)
		m.states[r] = make([]CellState, cols)
		for c := 0; c < cols; c++ {
			if r == 0 && c == 0 {
				m.models[r][c] = firstModel
				continue
			}
			m.models[r][c] = ferroelectric.NewPreisachModel(materialSnapshot)
		}
	}
	return m, nil
}

func isValidArrayDimensions(rows, cols int) bool {
	return rows > 0 && cols > 0 && rows <= maxMultiCellCount && cols <= maxMultiCellCount && cols <= maxMultiCellCount/rows
}

func (m *MultiCellArray) Size() (int, int) { m.mu.RLock(); defer m.mu.RUnlock(); return m.rows, m.cols }

func (m *MultiCellArray) StepCell(row, col int, voltage float64) (CellState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.inBounds(row, col) {
		return CellState{}, fmt.Errorf("cell index out of bounds: (%d,%d)", row, col)
	}
	if err := validateAppliedVoltage(voltage, m.material.Thickness); err != nil {
		return CellState{}, err
	}
	return m.stepCellLocked(row, col, voltage), nil
}

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
		for c := range voltageMap[r] {
			if err := validateAppliedVoltage(voltageMap[r][c], m.material.Thickness); err != nil {
				return fmt.Errorf("voltage map cell (%d,%d): %w", r, c, err)
			}
		}
	}
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.cols; c++ {
			m.stepCellLocked(r, c, voltageMap[r][c])
		}
	}
	return nil
}

func (m *MultiCellArray) StepWithSelector(cells []CellCoord, voltage float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cell := range cells {
		if !m.inBounds(cell.Row, cell.Col) {
			return fmt.Errorf("cell index out of bounds: (%d,%d)", cell.Row, cell.Col)
		}
	}
	if err := validateAppliedVoltage(voltage, m.material.Thickness); err != nil {
		return err
	}
	for _, cell := range cells {
		m.stepCellLocked(cell.Row, cell.Col, voltage)
	}
	return nil
}

func (m *MultiCellArray) GetCellState(row, col int) (CellState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.inBounds(row, col) {
		return CellState{}, fmt.Errorf("cell index out of bounds: (%d,%d)", row, col)
	}
	return m.states[row][col], nil
}

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
	var field float64
	if m.material.Thickness > 0 {
		field = voltage / m.material.Thickness
	}
	p := m.models[row][col].Update(field)
	s := CellState{Voltage: voltage, ElectricField: field, Polarization: p, NormPol: m.models[row][col].NormalizedPolarization()}
	m.states[row][col] = s
	return s
}

func validateAppliedVoltage(voltage, thickness float64) error {
	if !isFinite(voltage) {
		return fmt.Errorf("voltage must be finite: got %.3g V", voltage)
	}
	if !isRepresentableField(voltage, thickness) {
		return fmt.Errorf("voltage %.3g V overflows electric field for thickness %.3e m", voltage, thickness)
	}
	return nil
}

func (m *MultiCellArray) inBounds(row, col int) bool {
	return row >= 0 && row < m.rows && col >= 0 && col < m.cols
}
