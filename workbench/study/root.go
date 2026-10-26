package study

import (
	"context"
	"errors"
	"fmt"
	"os"

	"fecim-lattice-tools/workbench/edaexport"
	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/fecim"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/report"
)

// ValidateProjectRoot validates an existing project through borrowed roots.
// Nil citations skips citation existence checks. Neither handle is closed.
func ValidateProjectRoot(root, citations *os.Root) (project.Bundle, int, error) {
	bundle, err := loadRoot(root, citations)
	if err != nil {
		return project.Bundle{}, 0, err
	}
	points, err := project.Expand(bundle)
	if err != nil {
		return project.Bundle{}, 0, err
	}
	return bundle, len(points), nil
}

// loadRoot also rejects closed citation handles when the project has no keys.
func loadRoot(root, citations *os.Root) (project.Bundle, error) {
	if citations != nil {
		if _, err := citations.Stat("."); err != nil {
			return project.Bundle{}, err
		}
	}
	return project.LoadRoot(root, citations)
}

// RunRoot runs a study using only the supplied roots, never Bundle.Root or
// root.Name() as filesystem authority. CitationDir must be empty; nil citations
// skips citation existence checks. The caller owns both roots and must keep them
// open until this returns, after all started workers have settled.
func RunRoot(ctx context.Context, root, citations *os.Root, opts RunOptions) (RunOutcome, error) {
	return runRoot(ctx, root, citations, opts, fecim.Evaluate, report.GenerateRoot)
}

func runRoot(ctx context.Context, root, citations *os.Root, opts RunOptions, evaluate experiment.Evaluator, writeReport func(*os.Root, project.Bundle, experiment.Analysis) error) (RunOutcome, error) {
	outcome := RunOutcome{ResultsState: ResultsUnavailable}
	if opts.CitationDir != "" {
		return outcome, errors.New("CitationDir is not allowed with supplied roots")
	}
	if opts.Workers < 1 {
		return outcome, errors.New("workers must be at least 1")
	}
	bundle, err := loadRoot(root, citations)
	if err != nil {
		return outcome, err
	}
	outcome.Bundle = bundle
	points, err := experiment.Expand(bundle)
	if err != nil {
		return outcome, err
	}
	// Snapshot only verified active-sweep IDs. Old or corrupt unrelated runs
	// neither contribute results nor prevent collection of this study.
	ids := make([]string, len(points))
	reused := make(map[string]bool, len(points))
	for i, point := range points {
		id, err := experiment.ID(point, fecim.EvaluatorVersion, bundle.Inputs)
		if err != nil {
			return outcome, err
		}
		_, ok, err := experiment.LoadRunRecordRoot(root, id)
		if err != nil {
			// Preserve the runner's nonblocking total snapshot on preflight failure.
			select {
			case opts.Progress <- experiment.ProgressSnapshot{Total: len(points)}:
			default:
			}
			return outcome, err
		}
		ids[i] = id
		reused[id] = ok
	}

	summary, runErr := experiment.RunRoot(ctx, root, bundle, experiment.RunOptions{
		Evaluator:          evaluate,
		EvaluatorVersion:   fecim.EvaluatorVersion,
		Workers:            opts.Workers,
		RepositoryRevision: opts.RepositoryRevision,
		Progress:           opts.Progress,
	})
	// Run has joined its workers. Its delivered records are independently
	// trusted; reload expected IDs to recover commits lost during cancellation.
	collected := make(map[string]experiment.RunRecord, len(summary.Runs))
	for _, record := range summary.Runs {
		if _, expected := reused[record.Manifest.RunID]; expected {
			collected[record.Manifest.RunID] = record
		}
	}
	var reconciliationErr error
	outcome.Summary.Runs = make([]experiment.RunRecord, 0, len(points))
	for _, id := range ids {
		record, ok, err := experiment.LoadRunRecordRoot(root, id)
		if err != nil {
			reconciliationErr = errors.Join(reconciliationErr, err)
		} else if ok {
			record.Reused = reused[id] || collected[id].Reused
			collected[id] = record
		}
		if record, ok := collected[id]; ok {
			record.Reused = record.Reused || reused[id]
			outcome.Summary.Runs = append(outcome.Summary.Runs, record)
		} else if runErr == nil && err == nil {
			reconciliationErr = errors.Join(reconciliationErr, fmt.Errorf("%w: expected run %s is missing after execution", experiment.ErrUnverifiableRun, id))
		}
	}
	if reconciliationErr != nil {
		if len(outcome.Summary.Runs) > 0 {
			outcome.ResultsState = ResultsPartial
		}
		return outcome, errors.Join(runErr, reconciliationErr)
	}
	outcome.ResultsState = ResultsPartial
	if len(outcome.Summary.Runs) == len(points) {
		outcome.ResultsState = ResultsComplete
	}
	outcome.Analysis = experiment.Analyze(bundle, outcome.Summary.Runs)
	outcome.AnalysisAvailable = true
	if runErr != nil {
		return outcome, runErr
	}
	if err := writeReport(root, bundle, outcome.Analysis); err != nil {
		return outcome, err
	}
	return outcome, nil
}

// GenerateReportRoot reads committed runs and writes reports through borrowed
// roots. Nil citations skips citation existence checks; neither root is closed.
func GenerateReportRoot(root, citations *os.Root) (experiment.Analysis, error) {
	bundle, err := loadRoot(root, citations)
	if err != nil {
		return experiment.Analysis{}, err
	}
	runs, err := experiment.LoadRunsRoot(root)
	if err != nil {
		return experiment.Analysis{}, err
	}
	if len(runs) == 0 {
		return experiment.Analysis{}, errors.New("project has no committed runs")
	}
	analysis := experiment.Analyze(bundle, runs)
	if err := report.GenerateRoot(root, bundle, analysis); err != nil {
		return experiment.Analysis{}, fmt.Errorf("generate report: %w", err)
	}
	return analysis, nil
}

// ExportRunRoot exports a verified successful run and returns exports/<id>
// relative to root. The caller retains root ownership on success and failure.
func ExportRunRoot(root *os.Root, id string) (string, error) {
	return edaexport.ExportSelectedRunRoot(root, id)
}
