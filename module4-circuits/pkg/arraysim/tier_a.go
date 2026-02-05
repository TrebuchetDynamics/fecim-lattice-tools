// Package arraysim provides approximate array coupling solvers for module4-circuits.
package arraysim

import (
	"errors"
	"math"
)

// TierASolver implements a fast, approximate IR-drop + half-select coupling model.
// It is intended as a lightweight stand-in before DC nodal/MNA is integrated.
type TierASolver struct{}

// NewTierASolver returns a Tier A solver instance.
func NewTierASolver() *TierASolver {
	return &TierASolver{}
}

// Solve computes per-cell voltages and currents using a single-pass IR-drop approximation.
func (t *TierASolver) Solve(params SolveParams) (SolveResult, error) {
	rows := len(params.Conductance)
	if rows == 0 {
		return SolveResult{}, nil
	}

	cols := len(params.BLVoltages)
	if cols == 0 {
		for _, row := range params.Conductance {
			if len(row) > cols {
				cols = len(row)
			}
		}
	}
	if cols == 0 {
		return SolveResult{}, errors.New("arraysim: no columns available")
	}

	geom := params.Geometry.WithDefaults()
	wire := params.Wire.WithDefaults(geom)

	cellVoltages := make([][]float64, rows)
	cellCurrents := make([][]float64, rows)
	idealVoltages := make([][]float64, rows)
	idealCurrents := make([][]float64, rows)
	rowCurrents := make([]float64, rows)
	colCurrents := make([]float64, cols)

	rowActive := func(r int) bool {
		if params.ActiveRows == nil {
			return true
		}
		if r < 0 || r >= len(params.ActiveRows) {
			return false
		}
		return params.ActiveRows[r]
	}

	for r := 0; r < rows; r++ {
		cellVoltages[r] = make([]float64, cols)
		cellCurrents[r] = make([]float64, cols)
		idealVoltages[r] = make([]float64, cols)
		idealCurrents[r] = make([]float64, cols)
		active := rowActive(r)
		for c := 0; c < cols; c++ {
			g := 0.0
			if r < len(params.Conductance) && c < len(params.Conductance[r]) {
				g = params.Conductance[r][c]
			}
			if !active {
				continue
			}
			wl := 0.0
			if r < len(params.WLVoltages) {
				wl = params.WLVoltages[r]
			}
			bl := 0.0
			if c < len(params.BLVoltages) {
				bl = params.BLVoltages[c]
			}
			v := wl - bl
			i := g * v
			idealVoltages[r][c] = v
			idealCurrents[r][c] = i
			rowCurrents[r] += i
			colCurrents[c] += i
		}
	}

	for r := 0; r < rows; r++ {
		if !rowActive(r) {
			continue
		}
		for c := 0; c < cols; c++ {
			vIdeal := idealVoltages[r][c]
			if vIdeal == 0 {
				continue
			}
			g := 0.0
			if r < len(params.Conductance) && c < len(params.Conductance[r]) {
				g = params.Conductance[r][c]
			}
			rowFactor := 0.0
			if cols > 0 {
				rowFactor = (float64(c) + 0.5) / float64(cols)
			}
			colFactor := 0.0
			if rows > 0 {
				colFactor = (float64(r) + 0.5) / float64(rows)
			}
			dropRow := math.Abs(rowCurrents[r]) * wire.RWordLine * rowFactor
			dropCol := math.Abs(colCurrents[c]) * wire.RBitLine * colFactor
			drop := dropRow + dropCol

			desired := math.Abs(vIdeal) - drop
			if desired < 0 {
				desired = 0
			}
			v := math.Copysign(desired, vIdeal)
			i := g * v
			cellVoltages[r][c] = v
			cellCurrents[r][c] = i
		}
	}

	finalRowCurrents := make([]float64, rows)
	finalColCurrents := make([]float64, cols)
	for r := 0; r < rows; r++ {
		if !rowActive(r) {
			continue
		}
		for c := 0; c < cols; c++ {
			i := cellCurrents[r][c]
			finalRowCurrents[r] += i
			finalColCurrents[c] += i
		}
	}

	return SolveResult{
		CellVoltages: cellVoltages,
		CellCurrents: cellCurrents,
		RowCurrents:  finalRowCurrents,
		ColCurrents:  finalColCurrents,
	}, nil
}
