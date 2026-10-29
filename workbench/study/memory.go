package study

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/fecim"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/report"
)

// RunMemory executes a validated bundle (from project.LoadBytes or LoadRoot)
// without opening files or using Bundle.Root/CitationDir as authority.
// Admission is <=256 points and <=32 rows
// and columns, including every swept coordinate, checked before Expand. Workers
// must be positive; this bounded implementation evaluates serially in sweep
// order. Records have scientific identities but no disk artifact/record hashes.
// ResultsJSON uses the original report.WriteJSON serializer, also on cancellation.
func RunMemory(ctx context.Context, bundle project.Bundle, opts RunOptions) (RunOutcome, error) {
	return runMemory(ctx, bundle, opts, fecim.Evaluate)
}

func runMemory(ctx context.Context, bundle project.Bundle, opts RunOptions, evaluate experiment.Evaluator) (RunOutcome, error) {
	out := RunOutcome{ResultsState: ResultsUnavailable}
	if opts.CitationDir != "" {
		return out, errors.New("CitationDir is not allowed for memory studies")
	}
	if opts.Workers < 1 {
		return out, errors.New("workers must be at least 1")
	}
	if err := preflightMemory(bundle); err != nil {
		return out, err
	}
	points, err := experiment.Expand(bundle)
	if err != nil {
		return out, err
	}
	out.Bundle = bundle
	out.Summary.Runs = make([]experiment.RunRecord, 0, len(points))
	progress := experiment.ProgressSnapshot{Total: len(points)}
	send := func() {
		select {
		case opts.Progress <- progress:
		default:
		}
	}
	send()
	finish := func(runErr error) (RunOutcome, error) {
		out.ResultsState = ResultsPartial
		if len(out.Summary.Runs) == len(points) {
			out.ResultsState = ResultsComplete
		}
		out.Analysis = experiment.Analyze(bundle, out.Summary.Runs)
		out.AnalysisAvailable = true
		var data bytes.Buffer
		serializationErr := report.WriteJSON(&data, bundle, out.Analysis)
		if serializationErr == nil {
			out.ResultsJSON = data.Bytes()
		}
		if ctx.Err() != nil {
			progress.Cancelled = progress.Total - progress.Completed
		}
		send()
		return out, errors.Join(runErr, serializationErr)
	}
	for _, point := range points {
		if opts.PointYield != nil {
			if err := opts.PointYield(ctx); err != nil {
				return finish(err)
			}
		}
		if err := ctx.Err(); err != nil {
			return finish(err)
		}
		id, err := experiment.ID(point, fecim.EvaluatorVersion, bundle.Inputs)
		if err != nil {
			return finish(err)
		}
		started := time.Now().UTC()
		result, err := evaluate(ctx, point.Design, point.Seed)
		if err != nil {
			return finish(err)
		}
		// The production evaluator has a fixed metric schema; guard its numeric
		// output before Analyze/JSON, as the native runner also does.
		if result.Status != experiment.StatusSuccess && result.Status != experiment.StatusFailed {
			return finish(experiment.ErrInvalidResult)
		}
		if result.Status == experiment.StatusSuccess {
			metrics := make(map[string]experiment.Metric, len(result.Metrics))
			for _, metric := range result.Metrics {
				if math.IsNaN(metric.Value) || math.IsInf(metric.Value, 0) {
					return finish(experiment.ErrInvalidResult)
				}
				metrics[metric.Name] = metric
			}
			for _, objective := range bundle.Project.Objectives {
				if _, ok := metrics[objective.Metric]; !ok {
					return finish(fmt.Errorf("%w: missing objective metric %s", experiment.ErrInvalidResult, objective.Metric))
				}
			}
			for _, constraint := range bundle.Project.Constraints {
				metric, ok := metrics[constraint.Metric]
				if !ok || metric.Unit != constraint.Unit {
					return finish(fmt.Errorf("%w: missing or incompatible constraint metric %s", experiment.ErrInvalidResult, constraint.Metric))
				}
			}
		}
		record := experiment.RunRecord{Manifest: experiment.RunManifest{
			SchemaVersion: 2, RunID: id, PointIndex: point.Index, Seed: point.Seed,
			ModelVersion: bundle.Project.ModelVersion, EvaluatorVersion: fecim.EvaluatorVersion,
			RepositoryRevision: opts.RepositoryRevision, Inputs: bundle.Inputs,
			StartedAt: started, CompletedAt: time.Now().UTC(), Status: result.Status,
		}, Design: point.Design, Result: result}
		// A completed result is retained even if cancellation arrives during Evaluate.
		out.Summary.Runs = append(out.Summary.Runs, record)
		progress.Completed++
		if result.Status == experiment.StatusFailed {
			progress.Failed++
		}
		send()
	}
	return finish(ctx.Err())
}

func preflightMemory(bundle project.Bundle) error {
	const maxPoints = 256
	const maxDimension = 32
	if bundle.Project.SchemaVersion != 1 || bundle.Design.SchemaVersion != 1 || bundle.Sweep.SchemaVersion != 1 {
		return errors.New("all schema_version values must be 1")
	}
	if err := project.ValidateDesign(bundle.Design); err != nil {
		return err
	}
	if bundle.Design.Array.Rows > maxDimension || bundle.Design.Array.Cols > maxDimension {
		return errors.New("memory arrays must not exceed 32 by 32")
	}
	if len(bundle.Sweep.Parameters) == 0 || len(bundle.Sweep.Parameters) > 11 {
		return errors.New("memory sweep requires 1 to 11 parameters")
	}
	limit := bundle.Sweep.MaxPoints
	if limit <= 0 {
		limit = 10000
	}
	if limit > 100000 {
		return errors.New("max_points must not exceed 100000")
	}
	if limit > maxPoints {
		limit = maxPoints
	}
	integerLimit := math.Ldexp(1, strconv.IntSize-1)
	points := 1
	for _, p := range bundle.Sweep.Parameters {
		count := len(p.Values)
		if r := p.Range; r != nil {
			if count != 0 || r.Count < 1 || math.IsNaN(r.Start) || math.IsInf(r.Start, 0) || math.IsNaN(r.Stop) || math.IsInf(r.Stop, 0) || math.IsInf(r.Stop-r.Start, 0) {
				return errors.New("invalid sweep range")
			}
			count = r.Count
		}
		if count < 1 || count > limit || points > limit/count {
			return fmt.Errorf("memory sweep exceeds %d points", limit)
		}
		points *= count
		check := func(value float64) error {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return errors.New("sweep values must be finite")
			}
			switch p.Path {
			case "array.rows", "array.cols":
				if value < 1 || value > maxDimension {
					return errors.New("memory array coordinates must be within 1 to 32")
				}
				fallthrough
			case "device.conductance_levels", "circuit.adc_bits", "circuit.dac_bits":
				if value < -integerLimit || value >= integerLimit {
					return errors.New("integer sweep coordinate overflows")
				}
			}
			return nil
		}
		for _, value := range p.Values {
			if err := check(value); err != nil {
				return err
			}
		}
		if r := p.Range; r != nil {
			if err := check(r.Start); err != nil {
				return err
			}
			if r.Count > 1 {
				if err := check(r.Stop); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
