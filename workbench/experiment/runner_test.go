package experiment

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/project"
)

func runnerBundle(t *testing.T, points int) project.Bundle {
	t.Helper()
	bundle := fixtureBundle()
	bundle.Root = t.TempDir()
	bundle.Sweep.Parameters = []project.Parameter{{Path: "circuit.adc_bits", Values: make([]float64, points)}}
	for i := 0; i < points; i++ {
		bundle.Sweep.Parameters[0].Values[i] = float64(i + 2)
	}
	bundle.Sweep.MaxPoints = points
	return bundle
}

func successResult(value float64) Result {
	return Result{Status: StatusSuccess, Metrics: []Metric{{Name: "energy_pj", Value: value, Unit: "pJ", Model: "test", Assumptions: []string{"test"}, Evidence: EvidenceDerived}}}
}

func deterministicRunOptions(eval Evaluator) RunOptions {
	return RunOptions{
		Evaluator:          eval,
		EvaluatorVersion:   "test-v1",
		Workers:            1,
		RepositoryRevision: "abc123",
		Now: func() time.Time {
			return time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
		},
	}
}

func TestRunPersistsAndReusesImmutableRecords(t *testing.T) {
	bundle := runnerBundle(t, 2)
	var calls atomic.Int32
	eval := func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		calls.Add(1)
		return successResult(float64(seed)), nil
	}

	first, err := Run(context.Background(), bundle, deterministicRunOptions(eval))
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Runs) != 2 || calls.Load() != 2 {
		t.Fatalf("runs=%d calls=%d", len(first.Runs), calls.Load())
	}
	for _, record := range first.Runs {
		dir := filepath.Join(bundle.Root, "runs", record.Manifest.RunID)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 3 {
			t.Fatalf("%s entries=%d want 3", dir, len(entries))
		}
	}

	second, err := Run(context.Background(), bundle, deterministicRunOptions(eval))
	if err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatalf("evaluator called during reuse: %d", calls.Load())
	}
	for _, record := range second.Runs {
		if !record.Reused {
			t.Fatalf("run %s not marked reused", record.Manifest.RunID)
		}
	}
}

func TestRunCommitsDesignFailureAndContinues(t *testing.T) {
	bundle := runnerBundle(t, 3)
	eval := func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		if seed == bundle.Sweep.Seed+1 {
			return Result{Status: StatusFailed, Failure: &Failure{Kind: "model-domain", Message: "outside domain"}}, nil
		}
		return successResult(float64(seed)), nil
	}
	got, err := Run(context.Background(), bundle, deterministicRunOptions(eval))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Runs) != 3 || got.Runs[1].Result.Status != StatusFailed || got.Runs[2].Result.Status != StatusSuccess {
		t.Fatalf("runs=%+v", got.Runs)
	}
}

func TestRunSystemicErrorStopsWithoutCommittingPoint(t *testing.T) {
	bundle := runnerBundle(t, 3)
	boom := errors.New("storage unavailable")
	eval := func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		if seed == bundle.Sweep.Seed+1 {
			return Result{}, boom
		}
		return successResult(float64(seed)), nil
	}
	got, err := Run(context.Background(), bundle, deterministicRunOptions(eval))
	if !errors.Is(err, boom) {
		t.Fatalf("err=%v want boom", err)
	}
	if len(got.Runs) != 1 {
		t.Fatalf("committed runs=%d want 1", len(got.Runs))
	}
	assertNoTempRuns(t, bundle.Root)
}

func TestRunRejectsCorruptRunPath(t *testing.T) {
	bundle := runnerBundle(t, 1)
	point, err := Expand(bundle)
	if err != nil {
		t.Fatal(err)
	}
	id, err := ID(point[0], "test-v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(bundle.Root, "runs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundle.Root, "runs", id), []byte("corrupt"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = Run(context.Background(), bundle, deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
		return successResult(1), nil
	}))
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("err=%v want corruption", err)
	}
}

func assertNoTempRuns(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "runs"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".tmp-") {
			t.Fatalf("temporary run remained: %s", entry.Name())
		}
	}
}
