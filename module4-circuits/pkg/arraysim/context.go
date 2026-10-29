package arraysim

import "context"

// SolveContext keeps the legacy Engine contract intact and adds cancellation to
// the iterative solver. Never expose a partially solved network on cancellation.
func SolveContext(ctx context.Context, engine Engine, p SolveParams) (SolveResult, error) {
	if err := ctx.Err(); err != nil {
		return SolveResult{}, err
	}
	if b, ok := engine.(*TierBSolver); ok {
		copy := *b
		copy.context = ctx
		engine = &copy
	}
	result, err := engine.Solve(p)
	if err != nil {
		return SolveResult{}, err
	}
	if err = ctx.Err(); err != nil {
		return SolveResult{}, err
	}
	return result, nil
}

func DeviceSense(s DeviceSettings, current, gain, vref float64) (float64, int, bool) {
	return calculation(context.Background(), s, 0, 0).convertSenseLocked(current, gain, vref)
}
