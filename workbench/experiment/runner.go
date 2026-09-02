package experiment

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"fecim-lattice-tools/workbench/project"
)

var ErrInvalidResult = errors.New("invalid evaluator result")

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
			if existing.Result.Status == StatusSuccess {
				if err := validateSuccessResult(bundle.Project, existing.Result); err != nil {
					return orderedSummary(records, complete), fmt.Errorf("run %s: %w", point.RunID, err)
				}
			}
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
	if result.Status == StatusSuccess {
		if err := validateSuccessResult(bundle.Project, result); err != nil {
			return runOutcome{index: job.index, err: fmt.Errorf("run %s: %w", job.point.RunID, err)}
		}
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

func validateSuccessResult(config project.Project, result Result) error {
	if result.Failure != nil {
		return fmt.Errorf("%w: success carrying failure", ErrInvalidResult)
	}
	metrics := make(map[string]Metric, len(result.Metrics))
	for _, metric := range result.Metrics {
		if strings.TrimSpace(metric.Name) == "" {
			return fmt.Errorf("%w: blank metric name", ErrInvalidResult)
		}
		if _, exists := metrics[metric.Name]; exists {
			return fmt.Errorf("%w: duplicate metric %s", ErrInvalidResult, metric.Name)
		}
		metrics[metric.Name] = metric
	}
	for _, metric := range result.Metrics {
		if math.IsNaN(metric.Value) || math.IsInf(metric.Value, 0) {
			return fmt.Errorf("%w: metric %s has non-finite value", ErrInvalidResult, metric.Name)
		}
	}
	for _, metric := range result.Metrics {
		if strings.TrimSpace(metric.Unit) == "" {
			return fmt.Errorf("%w: metric %s missing unit", ErrInvalidResult, metric.Name)
		}
		if strings.TrimSpace(metric.Model) == "" {
			return fmt.Errorf("%w: metric %s missing model", ErrInvalidResult, metric.Name)
		}
		hasAssumption := false
		for _, assumption := range metric.Assumptions {
			if strings.TrimSpace(assumption) != "" {
				hasAssumption = true
				break
			}
		}
		if !hasAssumption {
			return fmt.Errorf("%w: metric %s missing nonblank assumption", ErrInvalidResult, metric.Name)
		}
		switch metric.Evidence {
		case EvidenceLiterature, EvidenceExperiment, EvidenceDefault, EvidenceDerived:
		default:
			return fmt.Errorf("%w: metric %s has invalid evidence %q", ErrInvalidResult, metric.Name, metric.Evidence)
		}
	}
	for _, objective := range config.Objectives {
		if _, exists := metrics[objective.Metric]; !exists {
			return fmt.Errorf("%w: missing objective metric %s", ErrInvalidResult, objective.Metric)
		}
	}
	for _, constraint := range config.Constraints {
		metric, exists := metrics[constraint.Metric]
		if !exists {
			return fmt.Errorf("%w: missing constraint metric %s", ErrInvalidResult, constraint.Metric)
		}
		if metric.Unit != constraint.Unit {
			return fmt.Errorf("%w: constraint metric %s unit %s does not match %s", ErrInvalidResult, constraint.Metric, metric.Unit, constraint.Unit)
		}
	}
	return nil
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
