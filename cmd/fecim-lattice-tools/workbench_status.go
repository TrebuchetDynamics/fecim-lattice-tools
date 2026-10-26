package main

import (
	"context"
	"errors"
	"fmt"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/study"
)

func formatWorkbenchOutcome(outcome study.RunOutcome, err error) (status, details string) {
	status = "Simulation complete"
	if err != nil {
		status = "Simulation failed"
		if workbenchCancellationOnly(err) {
			status = "Simulation cancelled"
		}
	}

	counts := "total=unavailable reused=unavailable failed=unavailable feasible=unavailable pareto=unavailable"
	qualification := "Results unavailable"
	if outcome.ResultsState == study.ResultsComplete || outcome.ResultsState == study.ResultsPartial {
		reused, failed := 0, 0
		for _, run := range outcome.Summary.Runs {
			if run.Reused {
				reused++
			}
			if run.Result.Status == experiment.StatusFailed {
				failed++
			}
		}
		feasible, hasFeasible := outcome.Analysis.Counts["feasible"]
		pareto, hasPareto := outcome.Analysis.Counts["pareto"]
		analysis := "feasible=unavailable pareto=unavailable"
		if outcome.AnalysisAvailable && hasFeasible && hasPareto {
			analysis = fmt.Sprintf("feasible=%d pareto=%d", feasible, pareto)
		}
		counts = fmt.Sprintf("total=%d reused=%d failed=%d %s", len(outcome.Summary.Runs), reused, failed, analysis)
		qualification = "Complete current-sweep results"
		if outcome.ResultsState == study.ResultsPartial {
			qualification = "Partial results (trusted current-sweep records only)"
			status += " · partial"
		}
	}
	status += " · " + counts
	details = qualification + ": " + counts + "\n"
	if err != nil {
		details += err.Error() + "\n"
	}
	return status, details
}

// A joined error is cancellation only when every leaf is cancellation. An
// errors.Is check on the joined root would hide simultaneous execution failures.
func workbenchCancellationOnly(err error) bool {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		children := joined.Unwrap()
		if len(children) == 0 {
			return false
		}
		for _, child := range children {
			if !workbenchCancellationOnly(child) {
				return false
			}
		}
		return true
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return workbenchCancellationOnly(wrapped.Unwrap())
	}
	return errors.Is(err, context.Canceled)
}
