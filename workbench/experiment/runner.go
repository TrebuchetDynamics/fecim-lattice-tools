package experiment

import (
	"context"
	"errors"
	"fmt"
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

func Run(ctx context.Context, bundle project.Bundle, opts RunOptions) (Summary, error) {
	if opts.Evaluator == nil || opts.EvaluatorVersion == "" {
		return Summary{}, errors.New("evaluator and evaluator version are required")
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	points, err := Expand(bundle)
	if err != nil {
		return Summary{}, err
	}
	storage := store{root: bundle.Root}
	summary := Summary{Runs: make([]RunRecord, 0, len(points))}
	for _, point := range points {
		if err := ctx.Err(); err != nil {
			return summary, err
		}
		id, err := ID(point, opts.EvaluatorVersion, bundle.Inputs)
		if err != nil {
			return summary, err
		}
		point.RunID = id
		if existing, ok, err := storage.load(id); err != nil {
			return summary, err
		} else if ok {
			existing.Reused = true
			summary.Runs = append(summary.Runs, existing)
			continue
		}
		started := opts.Now().UTC()
		result, err := opts.Evaluator(ctx, point.Design, point.Seed)
		if err != nil {
			return summary, fmt.Errorf("run %s: %w", id, err)
		}
		if result.Status != StatusSuccess && result.Status != StatusFailed {
			return summary, fmt.Errorf("run %s returned invalid status %q", id, result.Status)
		}
		record := RunRecord{
			Manifest: RunManifest{
				SchemaVersion:      1,
				RunID:              id,
				PointIndex:         point.Index,
				Seed:               point.Seed,
				EvaluatorVersion:   opts.EvaluatorVersion,
				RepositoryRevision: opts.RepositoryRevision,
				Inputs:             bundle.Inputs,
				StartedAt:          started,
				CompletedAt:        opts.Now().UTC(),
				Status:             result.Status,
			},
			Design: point.Design,
			Result: result,
		}
		committed, err := storage.commit(record)
		if err != nil {
			return summary, err
		}
		summary.Runs = append(summary.Runs, committed)
	}
	return summary, nil
}
