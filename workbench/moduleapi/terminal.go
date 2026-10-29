package moduleapi

import (
	"context"
	"fmt"
	"maps"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/study"
)

// CancellationOnly distinguishes context cancellation (possibly wrapped/joined
// with itself) from cancellation joined with integrity or other systemic errors.
func CancellationOnly(err error) bool {
	if err == context.Canceled {
		return true
	}
	if err == nil {
		return false
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		if len(causes) == 0 {
			return false
		}
		for _, cause := range causes {
			if !CancellationOnly(cause) {
				return false
			}
		}
		return true
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return CancellationOnly(wrapped.Unwrap())
	}
	return false
}

// CompleteJob projects the authoritative settled outcome, never current project
// configuration. Late pure cancellation with complete records/cancelled=0 is
// valid; results completeness and analysis availability are independent of state.
func CompleteJob(job Job, out study.RunOutcome, err error) Job {
	job.Progress = Progress{Total: job.Progress.Total, Completed: len(out.Summary.Runs)}
	for _, r := range out.Summary.Runs {
		if r.Reused {
			job.Progress.Reused++
		}
		if r.Result.Status == experiment.StatusFailed {
			job.Progress.Failed++
		}
	}
	job.State = "succeeded"
	if err != nil {
		job.State = "failed"
		if CancellationOnly(err) {
			job.State = "cancelled"
			job.Progress.Cancelled = max(0, job.Progress.Total-job.Progress.Completed)
		}
	}
	job.ResultsState = string(out.ResultsState)
	if job.ResultsState == "" {
		job.ResultsState = string(study.ResultsUnavailable)
	}
	job.AnalysisAvailable = false
	job.Counts = nil
	job.ResultsRevision = ""
	if out.AnalysisAvailable {
		page, queryErr := study.QueryResults(out.Analysis, study.ResultsQuery{})
		if queryErr == nil {
			job.AnalysisAvailable = true
			job.Counts = maps.Clone(out.Analysis.Counts)
			job.ResultsRevision = page.Revision
		} else {
			job.State = "failed"
		}
	}
	job.Message = fmt.Sprintf("Execution %s: %d of %d points committed; %d reused, %d failed, %d cancelled. Results %s.", job.State, job.Progress.Completed, job.Progress.Total, job.Progress.Reused, job.Progress.Failed, job.Progress.Cancelled, job.ResultsState)
	if job.AnalysisAvailable {
		job.Message += fmt.Sprintf(" Analysis: %d successful runs, %d failed runs, %d feasible, %d infeasible, %d unusable, %d Pareto.", job.Counts["success"], job.Counts["failed"], job.Counts["feasible"], job.Counts["infeasible"], job.Counts["unusable"], job.Counts["pareto"])
	} else {
		job.Message += " Analysis unavailable; inspect committed runs before reuse."
	}
	return job
}
