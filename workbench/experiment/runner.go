package experiment

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"fecim-lattice-tools/workbench/project"
)

type RunOptions struct {
	Evaluator          Evaluator
	EvaluatorVersion   string
	Workers            int
	RepositoryRevision string
	Now                func() time.Time
}

type runJob struct {
	index int
	point DesignPoint
}

type runOutcome struct {
	index  int
	record RunRecord
	err    error
}

func Run(ctx context.Context, bundle project.Bundle, opts RunOptions) (Summary, error) {
	if opts.Evaluator == nil || opts.EvaluatorVersion == "" {
		return Summary{}, errors.New("evaluator and evaluator version are required")
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Workers <= 0 {
		opts.Workers = 1
	}
	points, err := Expand(bundle)
	if err != nil {
		return Summary{}, err
	}

	storage := store{root: bundle.Root}
	records := make([]RunRecord, len(points))
	complete := make([]bool, len(points))
	jobsToRun := make([]runJob, 0, len(points))
	for index, point := range points {
		if err := ctx.Err(); err != nil {
			return orderedSummary(records, complete), err
		}
		id, err := ID(point, opts.EvaluatorVersion, bundle.Inputs)
		if err != nil {
			return orderedSummary(records, complete), err
		}
		point.RunID = id
		if existing, ok, err := storage.load(id); err != nil {
			return orderedSummary(records, complete), err
		} else if ok {
			existing.Reused = true
			records[index] = existing
			complete[index] = true
		} else {
			jobsToRun = append(jobsToRun, runJob{index: index, point: point})
		}
	}
	if len(jobsToRun) == 0 {
		return orderedSummary(records, complete), nil
	}
	if opts.Workers > len(jobsToRun) {
		opts.Workers = len(jobsToRun)
	}

	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan runJob)
	outcomes := make(chan runOutcome, opts.Workers)
	var workers sync.WaitGroup
	workers.Add(opts.Workers)
	for range opts.Workers {
		go func() {
			defer workers.Done()
			for job := range jobs {
				outcome := evaluateJob(workerCtx, storage, bundle, opts, job)
				select {
				case outcomes <- outcome:
				case <-workerCtx.Done():
					if outcome.err != nil {
						select {
						case outcomes <- outcome:
						default:
						}
					}
					return
				}
				if outcome.err != nil {
					return
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, job := range jobsToRun {
			select {
			case jobs <- job:
			case <-workerCtx.Done():
				return
			}
		}
	}()
	go func() {
		workers.Wait()
		close(outcomes)
	}()

	var firstErr error
	for outcome := range outcomes {
		if outcome.err != nil {
			if firstErr == nil {
				firstErr = outcome.err
				cancel()
			}
			continue
		}
		records[outcome.index] = outcome.record
		complete[outcome.index] = true
	}
	summary := orderedSummary(records, complete)
	if err := ctx.Err(); err != nil {
		return summary, err
	}
	if firstErr != nil {
		return summary, firstErr
	}
	return summary, nil
}

func evaluateJob(ctx context.Context, storage store, bundle project.Bundle, opts RunOptions, job runJob) runOutcome {
	if err := ctx.Err(); err != nil {
		return runOutcome{index: job.index, err: err}
	}
	started := opts.Now().UTC()
	result, err := opts.Evaluator(ctx, job.point.Design, job.point.Seed)
	if err != nil {
		return runOutcome{index: job.index, err: fmt.Errorf("run %s: %w", job.point.RunID, err)}
	}
	if result.Status != StatusSuccess && result.Status != StatusFailed {
		return runOutcome{index: job.index, err: fmt.Errorf("run %s returned invalid status %q", job.point.RunID, result.Status)}
	}
	record := RunRecord{
		Manifest: RunManifest{
			SchemaVersion:      2,
			RunID:              job.point.RunID,
			PointIndex:         job.point.Index,
			Seed:               job.point.Seed,
			EvaluatorVersion:   opts.EvaluatorVersion,
			RepositoryRevision: opts.RepositoryRevision,
			Inputs:             bundle.Inputs,
			StartedAt:          started,
			CompletedAt:        opts.Now().UTC(),
			Status:             result.Status,
		},
		Design: job.point.Design,
		Result: result,
	}
	committed, err := storage.commit(record)
	if err != nil {
		return runOutcome{index: job.index, err: err}
	}
	return runOutcome{index: job.index, record: committed}
}

func orderedSummary(records []RunRecord, complete []bool) Summary {
	summary := Summary{Runs: make([]RunRecord, 0, len(records))}
	for index, record := range records {
		if complete[index] {
			summary.Runs = append(summary.Runs, record)
		}
	}
	return summary
}
