// Package arraysim provides approximate array coupling solvers for module4-circuits.
package arraysim

// TierASolver implements the default coupled READ-path solver.
//
// Tier A now solves the full WL/BL resistive network (wire resistance + IR drop)
// using the dense DC nodal reference implementation, then returns per-cell
// voltages/currents for the read chain.
//
// This intentionally avoids per-cell ideal fallbacks inside the Tier-A path.
type TierASolver struct{}

// NewTierASolver returns a Tier A solver instance.
func NewTierASolver() *TierASolver {
	return &TierASolver{}
}

// Solve computes per-cell voltages and currents from a coupled resistive-network
// solve (full-array nodal model).
func (t *TierASolver) Solve(params SolveParams) (SolveResult, error) {
	res, err := referenceSolveDense(params)
	if err != nil {
		return SolveResult{}, err
	}
	return res.SolveResult, nil
}
