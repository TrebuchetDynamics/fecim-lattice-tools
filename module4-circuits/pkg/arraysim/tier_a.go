// Package arraysim provides approximate array coupling solvers for module4-circuits.
package arraysim

import (
	"errors"
	"math"
)

// TierASolver implements a fast, approximate IR-drop + half-select coupling model.
//
// It is intended as a lightweight stand-in before DC nodal/MNA is integrated.
// Compared to CouplingIdeal, Tier A includes a simple line IR-drop approximation.
type TierASolver struct{}

// NewTierASolver returns a Tier A solver instance.
func NewTierASolver() *TierASolver {
	return &TierASolver{}
}

// Solve computes per-cell voltages and currents using a lightweight fixed-point
// iteration to make the IR-drop coupling self-consistent.
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

	// Precompute ideal voltages and seed the current estimate with the ideal currents.
	for r := 0; r < rows; r++ {
		cellVoltages[r] = make([]float64, cols)
		cellCurrents[r] = make([]float64, cols)
		idealVoltages[r] = make([]float64, cols)

		active := rowActive(r)
		wl := 0.0
		if r < len(params.WLVoltages) {
			wl = params.WLVoltages[r]
		}
		for c := 0; c < cols; c++ {
			if !active {
				continue
			}
			g := 0.0
			if r < len(params.Conductance) && c < len(params.Conductance[r]) {
				g = params.Conductance[r][c]
			}
			if g == 0 {
				continue
			}
			bl := 0.0
			if c < len(params.BLVoltages) {
				bl = params.BLVoltages[c]
			}
			v := wl - bl
			idealVoltages[r][c] = v
			i := g * v
			rowCurrents[r] += i
			colCurrents[c] += i
		}
	}

	// Fixed-point iteration (typically converges in a handful of iterations).
	const (
		minIter = 2
		maxIter = 5
		relTol  = 1e-6
		absTol  = 1e-12
	)

	newRowCurrents := make([]float64, rows)
	newColCurrents := make([]float64, cols)

	for iter := 0; iter < maxIter; iter++ {
		for i := range newRowCurrents {
			newRowCurrents[i] = 0
		}
		for i := range newColCurrents {
			newColCurrents[i] = 0
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
				if g == 0 {
					continue
				}

				rowFactor := (float64(c) + 0.5) / float64(cols)
				colFactor := (float64(r) + 0.5) / float64(rows)
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
				newRowCurrents[r] += i
				newColCurrents[c] += i
			}
		}

		maxErr := 0.0
		maxRef := 0.0
		for r := 0; r < rows; r++ {
			if !rowActive(r) {
				continue
			}
			err := math.Abs(newRowCurrents[r] - rowCurrents[r])
			ref := math.Max(math.Abs(newRowCurrents[r]), math.Abs(rowCurrents[r]))
			if err > maxErr {
				maxErr = err
			}
			if ref > maxRef {
				maxRef = ref
			}
		}
		for c := 0; c < cols; c++ {
			err := math.Abs(newColCurrents[c] - colCurrents[c])
			ref := math.Max(math.Abs(newColCurrents[c]), math.Abs(colCurrents[c]))
			if err > maxErr {
				maxErr = err
			}
			if ref > maxRef {
				maxRef = ref
			}
		}

		tol := absTol + relTol*maxRef
		converged := maxErr <= tol

		copy(rowCurrents, newRowCurrents)
		copy(colCurrents, newColCurrents)

		if iter+1 >= minIter && converged {
			break
		}
	}

	return SolveResult{
		CellVoltages: cellVoltages,
		CellCurrents: cellCurrents,
		RowCurrents:  rowCurrents,
		ColCurrents:  colCurrents,
	}, nil
}
