package crossbar

import (
	core "fecim-lattice-tools/shared/crossbar"
	"fmt"
	"math"
)

type mvmWorkflow struct{ state CrossbarState }

func newMVMWorkflow(state CrossbarState) mvmWorkflow { return mvmWorkflow{state: state} }

// The legacy viewmodel uses the explicitly normalized enhanced profile.
// It does not publish physical microamperes or a plain, unnormalized dot product.
func (w mvmWorkflow) compute() (CrossbarState, error) {
	s := w.state
	if len(s.Conductances) != s.Rows || len(s.InputVector) != s.Cols {
		return s, fmt.Errorf("crossbar: exact matrix/vector shape required")
	}
	for _, row := range s.Conductances {
		if len(row) != s.Cols {
			return s, fmt.Errorf("crossbar: ragged matrix")
		}
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				return s, fmt.Errorf("crossbar: weights must be finite normalized values")
			}
		}
	}
	for _, v := range s.InputVector {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return s, fmt.Errorf("crossbar: inputs must be finite values in [0,1]")
		}
	}
	a, err := core.NewArray(&core.Config{Rows: s.Rows, Cols: s.Cols, ADCBits: 16, DACBits: 16})
	if err != nil {
		return s, err
	}
	defer a.Destroy()
	if err = a.ProgramWeightMatrix(s.Conductances); err != nil {
		return s, err
	}
	result, err := a.MVMWithNonIdealities(s.InputVector, &core.MVMOptions{EnableIRDrop: s.EnableIRDrop, EnableSneakPaths: s.SneakPaths, Temperature: 300, Architecture: "0T1R"})
	if err != nil {
		return s, err
	}
	s.OutputVector = result.ActualOutput
	s.IRDrop = 0
	if result.IRDropAnalysis != nil {
		s.IRDrop = result.IRDropAnalysis.MaxIRDrop
	}
	return s, nil
}
