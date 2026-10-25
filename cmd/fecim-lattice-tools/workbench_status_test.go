package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/study"
)

func TestWorkbenchTerminalOutcome(t *testing.T) {
	complete := study.RunOutcome{
		ResultsState: study.ResultsComplete, AnalysisAvailable: true,
		Summary: experiment.Summary{Runs: []experiment.RunRecord{
			{Result: experiment.Result{Status: experiment.StatusSuccess}, Reused: true},
			{Result: experiment.Result{Status: experiment.StatusSuccess}},
			{Result: experiment.Result{Status: experiment.StatusFailed}},
		}},
		Analysis: experiment.Analysis{Counts: map[string]int{"feasible": 2, "pareto": 1, "failed": 1}},
	}
	partial := complete
	partial.ResultsState = study.ResultsPartial
	integrity := partial
	integrity.AnalysisAvailable = false // Stale counts must not leak through.
	empty := study.RunOutcome{ResultsState: study.ResultsPartial, AnalysisAvailable: true,
		Analysis: experiment.Analysis{Counts: map[string]int{"feasible": 0, "pareto": 0, "failed": 0}}}
	missing := partial
	missing.Analysis = experiment.Analysis{}
	unknown := complete
	unknown.ResultsState = "future-state"
	reportErr := errors.New("report write failed")
	cancelErr := fmt.Errorf("study stopped: %w", context.Canceled)

	for _, tc := range []struct {
		name    string
		outcome study.RunOutcome
		err     error
		state   string
		counts  string
		partial bool
	}{
		{"complete", complete, nil, "complete", "total=3 reused=1 failed=1 feasible=2 pareto=1", false},
		{"report failure retains complete results", complete, reportErr, "failed", "total=3 reused=1 failed=1 feasible=2 pareto=1", false},
		{"wrapped cancellation", partial, cancelErr, "cancelled", "total=3 reused=1 failed=1 feasible=2 pareto=1", true},
		{"systemic failure retains partial results", partial, reportErr, "failed", "total=3 reused=1 failed=1 feasible=2 pareto=1", true},
		{"mixed joined error is failure", partial, fmt.Errorf("operation: %w", errors.Join(cancelErr, reportErr)), "failed", "total=3 reused=1 failed=1 feasible=2 pareto=1", true},
		{"joined cancellations", partial, errors.Join(cancelErr, context.Canceled), "cancelled", "total=3 reused=1 failed=1 feasible=2 pareto=1", true},
		{"complete cancellation", complete, cancelErr, "cancelled", "total=3 reused=1 failed=1 feasible=2 pareto=1", false},
		{"valid empty partial analysis", empty, cancelErr, "cancelled", "total=0 reused=0 failed=0 feasible=0 pareto=0", true},
		{"integrity invalid analysis", integrity, reportErr, "failed", "total=3 reused=1 failed=1 feasible=unavailable pareto=unavailable", true},
		{"missing analysis keys", missing, reportErr, "failed", "total=3 reused=1 failed=1 feasible=unavailable pareto=unavailable", true},
		{"zero outcome", study.RunOutcome{}, reportErr, "failed", "total=unavailable reused=unavailable failed=unavailable feasible=unavailable pareto=unavailable", false},
		{"unknown state", unknown, reportErr, "failed", "total=unavailable reused=unavailable failed=unavailable feasible=unavailable pareto=unavailable", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, details := formatWorkbenchOutcome(tc.outcome, tc.err)
			if !strings.HasPrefix(status, "Simulation "+tc.state) || !strings.Contains(status, tc.counts) || !strings.Contains(details, tc.counts) {
				t.Fatalf("status=%q details=%q; want state=%s counts=%s", status, details, tc.state, tc.counts)
			}
			if strings.Contains(strings.ToLower(status+details), "partial") != tc.partial {
				t.Fatalf("partial qualification: status=%q details=%q", status, details)
			}
			if tc.err != nil && !strings.Contains(details, tc.err.Error()) {
				t.Fatalf("details lost operation error: %q", details)
			}
		})
	}
}
