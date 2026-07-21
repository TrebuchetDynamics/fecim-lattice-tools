package experiment

import (
	"fmt"

	"fecim-lattice-tools/workbench/project"
)

type ConstraintOutcome struct {
	Constraint project.Constraint `json:"constraint"`
	Passed     bool               `json:"passed"`
	Actual     float64            `json:"actual"`
}

type AnalyzedRun struct {
	Run            RunRecord           `json:"run"`
	Feasible       bool                `json:"feasible"`
	Pareto         bool                `json:"pareto"`
	UnusableReason string              `json:"unusable_reason,omitempty"`
	Constraints    []ConstraintOutcome `json:"constraints"`
}

type Analysis struct {
	Runs   []AnalyzedRun  `json:"runs"`
	Counts map[string]int `json:"counts"`
}

func Analyze(bundle project.Bundle, runs []RunRecord) Analysis {
	analysis := Analysis{
		Runs: make([]AnalyzedRun, len(runs)),
		Counts: map[string]int{
			"success": 0, "failed": 0, "feasible": 0,
			"infeasible": 0, "unusable": 0, "pareto": 0,
		},
	}
	lookups := make([]map[string]Metric, len(runs))
	for index, run := range runs {
		analyzed := AnalyzedRun{Run: run, Constraints: []ConstraintOutcome{}}
		analysis.Runs[index] = analyzed
		if run.Result.Status == StatusFailed {
			analysis.Counts["failed"]++
			continue
		}
		if run.Result.Status != StatusSuccess {
			analysis.Runs[index].UnusableReason = fmt.Sprintf("unknown run status %q", run.Result.Status)
			analysis.Counts["unusable"]++
			continue
		}
		analysis.Counts["success"]++
		metrics, duplicate := metricLookup(run.Result.Metrics)
		lookups[index] = metrics
		if duplicate != "" {
			analysis.Runs[index].UnusableReason = "duplicate metric " + duplicate
			analysis.Counts["unusable"]++
			continue
		}
		if missing := missingObjectives(bundle.Project.Objectives, metrics); missing != "" {
			analysis.Runs[index].UnusableReason = "missing objective metric " + missing
			analysis.Counts["unusable"]++
			continue
		}
		usable := true
		feasible := true
		outcomes := make([]ConstraintOutcome, 0, len(bundle.Project.Constraints))
		for _, constraint := range bundle.Project.Constraints {
			metric, ok := metrics[constraint.Metric]
			if !ok {
				analysis.Runs[index].UnusableReason = "missing constraint metric " + constraint.Metric
				usable = false
				break
			}
			if metric.Unit != constraint.Unit {
				analysis.Runs[index].UnusableReason = fmt.Sprintf("constraint metric %s unit %s does not match %s", constraint.Metric, metric.Unit, constraint.Unit)
				usable = false
				break
			}
			passed := compare(metric.Value, constraint.Operator, constraint.Value)
			outcomes = append(outcomes, ConstraintOutcome{Constraint: constraint, Passed: passed, Actual: metric.Value})
			feasible = feasible && passed
		}
		analysis.Runs[index].Constraints = outcomes
		if !usable {
			analysis.Counts["unusable"]++
			continue
		}
		analysis.Runs[index].Feasible = feasible
		if feasible {
			analysis.Counts["feasible"]++
		} else {
			analysis.Counts["infeasible"]++
		}
	}

	for i := range analysis.Runs {
		if !analysis.Runs[i].Feasible {
			continue
		}
		dominated := false
		for j := range analysis.Runs {
			if i != j && analysis.Runs[j].Feasible && dominates(lookups[j], lookups[i], bundle.Project.Objectives) {
				dominated = true
				break
			}
		}
		if !dominated {
			analysis.Runs[i].Pareto = true
			analysis.Counts["pareto"]++
		}
	}
	return analysis
}

func metricLookup(metrics []Metric) (map[string]Metric, string) {
	lookup := make(map[string]Metric, len(metrics))
	for _, metric := range metrics {
		if _, exists := lookup[metric.Name]; exists {
			return lookup, metric.Name
		}
		lookup[metric.Name] = metric
	}
	return lookup, ""
}

func missingObjectives(objectives []project.Objective, metrics map[string]Metric) string {
	for _, objective := range objectives {
		if _, ok := metrics[objective.Metric]; !ok {
			return objective.Metric
		}
	}
	return ""
}

func compare(actual float64, operator string, expected float64) bool {
	switch operator {
	case "<":
		return actual < expected
	case "<=":
		return actual <= expected
	case ">":
		return actual > expected
	case ">=":
		return actual >= expected
	case "==":
		return actual == expected
	default:
		return false
	}
}

func dominates(a, b map[string]Metric, objectives []project.Objective) bool {
	strictlyBetter := false
	for _, objective := range objectives {
		av := a[objective.Metric].Value
		bv := b[objective.Metric].Value
		switch objective.Direction {
		case project.Minimize:
			if av > bv {
				return false
			}
			strictlyBetter = strictlyBetter || av < bv
		case project.Maximize:
			if av < bv {
				return false
			}
			strictlyBetter = strictlyBetter || av > bv
		}
	}
	return strictlyBetter
}
