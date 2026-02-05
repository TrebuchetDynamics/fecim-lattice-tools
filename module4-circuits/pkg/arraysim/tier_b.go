package arraysim

import "fmt"

// DCEngine extends Engine with access to the full DC nodal solution (internal line node voltages).
//
// Tier A provides a fast approximation; Tier B is intended to be a true DC nodal solve.
//
// This is scaffolding: Tier B currently delegates to a dense reference solve for small arrays.
//
// TODO(tier-b): Replace the dense reference solver with a scalable sparse/iterative solver.
// TODO(tier-b): Add support for more realistic boundary conditions and selector devices.
// TODO(tier-b): Validate against SPICE golden vectors.
type DCEngine interface {
	Engine
	SolveDC(params SolveParams) (DCResult, error)
}

// DCResult contains the standard per-cell outputs plus the internal line node voltages.
//
// WLNodes and BLNodes are sized [rows][cols] and correspond to the wordline/bitline
// node at each cell intersection.
type DCResult struct {
	SolveResult
	WLNodes [][]float64
	BLNodes [][]float64
}

// TierBSolver is a placeholder Tier B DC solver.
//
// It is intentionally conservative and only supports a dense reference solve for small arrays.
// It is never selected by default.
type TierBSolver struct {
	// MaxDenseSize caps the dense reference solve (in number of cells).
	// If <= 0, a conservative default is used.
	MaxDenseSize int
}

// NewTierBSolver returns a Tier B solver instance.
func NewTierBSolver() *TierBSolver {
	return &TierBSolver{MaxDenseSize: 16}
}

func (t *TierBSolver) maxSize() int {
	if t == nil || t.MaxDenseSize <= 0 {
		return 16
	}
	return t.MaxDenseSize
}

// Solve satisfies Engine by returning only the per-cell outputs.
func (t *TierBSolver) Solve(params SolveParams) (SolveResult, error) {
	res, err := t.SolveDC(params)
	if err != nil {
		return SolveResult{}, err
	}
	return res.SolveResult, nil
}

// SolveDC returns the full DC solution including internal line node voltages.
func (t *TierBSolver) SolveDC(params SolveParams) (DCResult, error) {
	rows := len(params.Conductance)
	cols := len(params.BLVoltages)
	if cols == 0 {
		for _, row := range params.Conductance {
			if len(row) > cols {
				cols = len(row)
			}
		}
	}
	if rows == 0 || cols == 0 {
		return DCResult{SolveResult: SolveResult{}}, nil
	}

	if rows*cols > t.maxSize() {
		return DCResult{}, fmt.Errorf("arraysim: tier-b dense reference solver supports up to %d cells, got %d", t.maxSize(), rows*cols)
	}

	return referenceSolveDense(params)
}
