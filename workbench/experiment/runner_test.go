package experiment

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
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

func TestRunWorkerDeterminism(t *testing.T) {
	one := runnerBundle(t, 4)
	four := one
	four.Root = t.TempDir()
	eval := func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		return successResult(float64(seed)), nil
	}
	oneOpts := deterministicRunOptions(eval)
	oneOpts.Workers = 1
	fourOpts := deterministicRunOptions(eval)
	fourOpts.Workers = 4
	first, err := Run(context.Background(), one, oneOpts)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Run(context.Background(), four, fourOpts)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Runs, second.Runs) {
		t.Fatalf("worker count changed ordered results:\n1=%+v\n4=%+v", first.Runs, second.Runs)
	}
}

func TestRunBoundsConcurrentEvaluatorCalls(t *testing.T) {
	bundle := runnerBundle(t, 4)
	started := make(chan struct{}, 4)
	release := make(chan struct{}, 4)
	eval := func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		started <- struct{}{}
		<-release
		return successResult(float64(seed)), nil
	}
	opts := deterministicRunOptions(eval)
	opts.Workers = 2
	done := make(chan error, 1)
	go func() {
		_, err := Run(context.Background(), bundle, opts)
		done <- err
	}()
	<-started
	select {
	case <-started:
	case <-time.After(100 * time.Millisecond):
		// Drain a sequential implementation before failing RED so no goroutine leaks.
		for i := 0; i < 4; i++ {
			release <- struct{}{}
		}
		<-done
		t.Fatal("runner did not start two workers")
	}
	select {
	case <-started:
		t.Fatal("runner exceeded worker limit")
	case <-time.After(25 * time.Millisecond):
	}
	for i := 0; i < 2; i++ {
		release <- struct{}{}
	}
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(500 * time.Millisecond):
			t.Fatal("runner did not schedule remaining work")
		}
		release <- struct{}{}
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestRunCancellationPreservesAndReusesCommittedRuns(t *testing.T) {
	bundle := runnerBundle(t, 4)
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	eval := func(ctx context.Context, _ project.Design, seed int64) (Result, error) {
		if calls.Add(1) == 3 {
			cancel()
			return Result{}, ctx.Err()
		}
		return successResult(float64(seed)), nil
	}
	opts := deterministicRunOptions(eval)
	opts.Workers = 1
	partial, err := Run(ctx, bundle, opts)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v want context.Canceled", err)
	}
	if len(partial.Runs) != 2 {
		t.Fatalf("partial runs=%d want 2", len(partial.Runs))
	}

	var resumedCalls atomic.Int32
	resumeOpts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		resumedCalls.Add(1)
		return successResult(float64(seed)), nil
	})
	resumeOpts.Workers = 2
	resumed, err := Run(context.Background(), bundle, resumeOpts)
	if err != nil {
		t.Fatal(err)
	}
	if len(resumed.Runs) != 4 || resumedCalls.Load() != 2 || !resumed.Runs[0].Reused || !resumed.Runs[1].Reused {
		t.Fatalf("resumed=%+v calls=%d", resumed.Runs, resumedCalls.Load())
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
