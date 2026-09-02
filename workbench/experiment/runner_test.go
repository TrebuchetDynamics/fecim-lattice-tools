package experiment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
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
	return Result{Status: StatusSuccess, Metrics: []Metric{
		{Name: "energy_pj", Value: value, Unit: "pJ", Model: "test", Assumptions: []string{"test"}, Evidence: EvidenceDerived},
		{Name: "latency_ns", Value: value, Unit: "ns", Model: "test", Assumptions: []string{"test"}, Evidence: EvidenceDerived},
	}}
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

func assertRejectedResult(t *testing.T, bundle project.Bundle, result Result, reason string) {
	t.Helper()
	points, err := Expand(bundle)
	if err != nil {
		t.Fatal(err)
	}
	runID, err := ID(points[0], "test-v1", bundle.Inputs)
	if err != nil {
		t.Fatal(err)
	}

	summary, err := Run(context.Background(), bundle, deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
		return result, nil
	}))
	if !errors.Is(err, ErrInvalidResult) || !strings.Contains(err.Error(), "run "+runID) || !strings.Contains(err.Error(), reason) {
		t.Fatalf("Run error=%v want ErrInvalidResult with run %s and reason %q", err, runID, reason)
	}
	if len(summary.Runs) != 0 {
		t.Fatalf("summary runs=%d want 0", len(summary.Runs))
	}
	if _, statErr := os.Stat(filepath.Join(bundle.Root, "runs")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("runs directory exists after rejected result: %v", statErr)
	}
	assertNoTempRuns(t, bundle.Root)
}

func TestRunRejectsSuccessCarryingFailure(t *testing.T) {
	bundle := runnerBundle(t, 1)
	result := successResult(1)
	result.Failure = &Failure{Kind: "unexpected", Message: "must not persist"}

	assertRejectedResult(t, bundle, result, "success carrying failure")
}

func TestRunRejectsInvalidMetricNames(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Result)
		reason string
	}{
		{name: "blank", mutate: func(result *Result) { result.Metrics[0].Name = " \t" }, reason: "blank metric name"},
		{name: "duplicate", mutate: func(result *Result) { result.Metrics[1].Name = "energy_pj" }, reason: "duplicate metric energy_pj"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := runnerBundle(t, 1)
			result := successResult(1)
			test.mutate(&result)
			assertRejectedResult(t, bundle, result, test.reason)
		})
	}
}

func TestRunRejectsDuplicateOptionalMetricNames(t *testing.T) {
	bundle := runnerBundle(t, 1)
	result := successResult(1)
	optional := Metric{Name: "temperature_c", Value: 25, Unit: "C", Model: "test", Assumptions: []string{"test"}, Evidence: EvidenceDerived}
	result.Metrics = append(result.Metrics, optional, optional)

	assertRejectedResult(t, bundle, result, "duplicate metric temperature_c")
}

func TestRunRejectsNonFiniteMetricValues(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{name: "NaN", value: math.NaN()},
		{name: "positive infinity", value: math.Inf(1)},
		{name: "negative infinity", value: math.Inf(-1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := runnerBundle(t, 1)
			result := successResult(1)
			result.Metrics[0].Value = test.value
			assertRejectedResult(t, bundle, result, "metric energy_pj has non-finite value")
		})
	}
}

func TestRunRejectsMetricMissingUnit(t *testing.T) {
	bundle := runnerBundle(t, 1)
	result := successResult(1)
	result.Metrics[0].Unit = " \t"

	assertRejectedResult(t, bundle, result, "metric energy_pj missing unit")
}

func TestRunRejectsMetricMissingModel(t *testing.T) {
	bundle := runnerBundle(t, 1)
	result := successResult(1)
	result.Metrics[0].Model = " \t"

	assertRejectedResult(t, bundle, result, "metric energy_pj missing model")
}

func TestRunRejectsMetricMissingNonblankAssumption(t *testing.T) {
	tests := []struct {
		name        string
		assumptions []string
	}{
		{name: "absent"},
		{name: "blank only", assumptions: []string{"", " \t"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := runnerBundle(t, 1)
			result := successResult(1)
			result.Metrics[0].Assumptions = test.assumptions
			assertRejectedResult(t, bundle, result, "metric energy_pj missing nonblank assumption")
		})
	}
}

func TestRunRejectsInvalidMetricEvidence(t *testing.T) {
	for _, evidence := range []Evidence{"", "expert-opinion"} {
		t.Run(string(evidence), func(t *testing.T) {
			bundle := runnerBundle(t, 1)
			result := successResult(1)
			result.Metrics[0].Evidence = evidence
			assertRejectedResult(t, bundle, result, fmt.Sprintf("metric energy_pj has invalid evidence %q", evidence))
		})
	}
}

func TestRunRejectsMissingObjectiveMetric(t *testing.T) {
	bundle := runnerBundle(t, 1)
	result := successResult(1)
	result.Metrics = result.Metrics[:1]

	assertRejectedResult(t, bundle, result, "missing objective metric latency_ns")
}

func TestRunRejectsMissingConstraintMetric(t *testing.T) {
	bundle := runnerBundle(t, 1)
	bundle.Project.Constraints = []project.Constraint{{Metric: "area_um2", Operator: "<=", Value: 5000, Unit: "um2"}}

	assertRejectedResult(t, bundle, successResult(1), "missing constraint metric area_um2")
}

func TestRunRejectsConstraintMetricUnitMismatch(t *testing.T) {
	bundle := runnerBundle(t, 1)
	bundle.Project.Constraints = []project.Constraint{{Metric: "area_um2", Operator: "<=", Value: 5000, Unit: "um2"}}
	result := successResult(1)
	result.Metrics = append(result.Metrics, Metric{Name: "area_um2", Value: 100, Unit: "mm2", Model: "test", Assumptions: []string{"test"}, Evidence: EvidenceDerived})

	assertRejectedResult(t, bundle, result, "constraint metric area_um2 unit mm2 does not match um2")
}

func TestRunCommitsStructurallyValidResult(t *testing.T) {
	bundle := runnerBundle(t, 1)
	result := successResult(1)
	result.Metrics[0].Evidence = EvidenceLiterature
	result.Metrics[1].Evidence = EvidenceExperiment
	result.Metrics = append(result.Metrics, Metric{Name: "extra", Value: 3, Unit: "ratio", Model: "test", Assumptions: []string{"test"}, Evidence: EvidenceDefault})

	summary, err := Run(context.Background(), bundle, deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
		return result, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Runs) != 1 {
		t.Fatalf("summary runs=%d want 1", len(summary.Runs))
	}
	if _, err := os.Stat(filepath.Join(bundle.Root, "runs", summary.Runs[0].Manifest.RunID)); err != nil {
		t.Fatalf("committed run directory: %v", err)
	}
}

func TestRunCommitsConstraintViolationForInfeasibleAnalysis(t *testing.T) {
	bundle := runnerBundle(t, 1)
	bundle.Project.Constraints = []project.Constraint{{Metric: "area_um2", Operator: "<=", Value: 50, Unit: "um2"}}
	result := successResult(1)
	result.Metrics = append(result.Metrics, Metric{Name: "area_um2", Value: 100, Unit: "um2", Model: "test", Assumptions: []string{"test"}, Evidence: EvidenceDerived})

	summary, err := Run(context.Background(), bundle, deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
		return result, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	analysis := Analyze(bundle, summary.Runs)
	if len(summary.Runs) != 1 || len(analysis.Runs) != 1 || analysis.Runs[0].Feasible || analysis.Counts["infeasible"] != 1 {
		t.Fatalf("summary=%+v analysis=%+v want committed infeasible run", summary, analysis)
	}
}

func TestRunInvalidResultStopsSchedulingAndPreservesCompletedRuns(t *testing.T) {
	bundle := runnerBundle(t, 3)
	points, err := Expand(bundle)
	if err != nil {
		t.Fatal(err)
	}
	rejectedRunID, err := ID(points[1], "test-v1", bundle.Inputs)
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	eval := func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		calls.Add(1)
		result := successResult(float64(seed))
		if seed == bundle.Sweep.Seed+1 {
			result.Metrics = result.Metrics[:1]
		}
		return result, nil
	}

	summary, err := Run(context.Background(), bundle, deterministicRunOptions(eval))
	if !errors.Is(err, ErrInvalidResult) || !strings.Contains(err.Error(), "run "+rejectedRunID) || !strings.Contains(err.Error(), "missing objective metric latency_ns") {
		t.Fatalf("Run error=%v want rejected run context", err)
	}
	if len(summary.Runs) != 1 || calls.Load() != 2 {
		t.Fatalf("summary runs=%d evaluator calls=%d want 1 and 2", len(summary.Runs), calls.Load())
	}
	if _, statErr := os.Stat(filepath.Join(bundle.Root, "runs", rejectedRunID)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("rejected run directory exists: %v", statErr)
	}
	assertNoTempRuns(t, bundle.Root)
}

func snapshotRunFiles(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("unexpected directory in cached run: %s", entry.Name())
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		files[entry.Name()] = data
	}
	return files
}

func TestRunRejectsCachedSuccessMissingNewObjective(t *testing.T) {
	bundle := runnerBundle(t, 1)
	initial, err := Run(context.Background(), bundle, deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
		return successResult(1), nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(initial.Runs) != 1 || initial.Runs[0].Manifest.SchemaVersion != 2 {
		t.Fatalf("initial cache=%+v want one schema-2 run", initial)
	}
	runID := initial.Runs[0].Manifest.RunID
	runDir := filepath.Join(bundle.Root, "runs", runID)
	before := snapshotRunFiles(t, runDir)
	bundle.Project.Objectives = append(bundle.Project.Objectives, project.Objective{Metric: "throughput_ops", Direction: project.Maximize})
	var calls atomic.Int32

	summary, err := Run(context.Background(), bundle, deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
		calls.Add(1)
		return successResult(2), nil
	}))
	if !errors.Is(err, ErrInvalidResult) || !strings.Contains(err.Error(), "run "+runID) || !strings.Contains(err.Error(), "missing objective metric throughput_ops") {
		t.Fatalf("Run error=%v want ErrInvalidResult with cached run %s and missing objective reason", err, runID)
	}
	if len(summary.Runs) != 0 {
		t.Fatalf("summary runs=%d want no reused record", len(summary.Runs))
	}
	if calls.Load() != 0 {
		t.Fatalf("evaluator calls=%d want 0 during invalid reuse check", calls.Load())
	}
	if after := snapshotRunFiles(t, runDir); !reflect.DeepEqual(after, before) {
		t.Fatalf("cached artifacts changed:\nbefore=%v\nafter=%v", before, after)
	}
	assertNoTempRuns(t, bundle.Root)
}

func TestRunRejectsCachedSuccessWithChangedConstraintUnit(t *testing.T) {
	bundle := runnerBundle(t, 1)
	bundle.Project.Constraints = []project.Constraint{{Metric: "energy_pj", Operator: "<=", Value: 10, Unit: "pJ"}}
	initial, err := Run(context.Background(), bundle, deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
		return successResult(1), nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(initial.Runs) != 1 || initial.Runs[0].Manifest.SchemaVersion != 2 {
		t.Fatalf("initial cache=%+v want one schema-2 run", initial)
	}
	runID := initial.Runs[0].Manifest.RunID
	runDir := filepath.Join(bundle.Root, "runs", runID)
	before := snapshotRunFiles(t, runDir)
	bundle.Project.Constraints[0].Unit = "fJ"
	var calls atomic.Int32

	summary, err := Run(context.Background(), bundle, deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
		calls.Add(1)
		return successResult(2), nil
	}))
	if !errors.Is(err, ErrInvalidResult) || !strings.Contains(err.Error(), "run "+runID) || !strings.Contains(err.Error(), "constraint metric energy_pj unit pJ does not match fJ") {
		t.Fatalf("Run error=%v want ErrInvalidResult with cached run %s and unit mismatch reason", err, runID)
	}
	if len(summary.Runs) != 0 {
		t.Fatalf("summary runs=%d want no reused record", len(summary.Runs))
	}
	if calls.Load() != 0 {
		t.Fatalf("evaluator calls=%d want 0 during invalid reuse check", calls.Load())
	}
	if after := snapshotRunFiles(t, runDir); !reflect.DeepEqual(after, before) {
		t.Fatalf("cached artifacts changed:\nbefore=%v\nafter=%v", before, after)
	}
	assertNoTempRuns(t, bundle.Root)
}

func TestRunStillReusesCachedDesignFailure(t *testing.T) {
	bundle := runnerBundle(t, 1)
	failed := Result{Status: StatusFailed, Failure: &Failure{Kind: "model-domain", Message: "outside domain"}}
	initial, err := Run(context.Background(), bundle, deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
		return failed, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32

	reused, err := Run(context.Background(), bundle, deterministicRunOptions(func(context.Context, project.Design, int64) (Result, error) {
		calls.Add(1)
		return successResult(1), nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(initial.Runs) != 1 || len(reused.Runs) != 1 || reused.Runs[0].Result.Status != StatusFailed || !reused.Runs[0].Reused {
		t.Fatalf("initial=%+v reused=%+v want reused design failure", initial, reused)
	}
	if calls.Load() != 0 {
		t.Fatalf("evaluator calls=%d want 0", calls.Load())
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
		if record.Manifest.SchemaVersion != 2 {
			t.Fatalf("schema version=%d want 2", record.Manifest.SchemaVersion)
		}
		if len(record.Manifest.ArtifactSHA256) != 2 || record.Manifest.ArtifactSHA256["resolved-design.json"] == "" || record.Manifest.ArtifactSHA256["result.json"] == "" {
			t.Fatalf("artifact digests=%v want exactly resolved design and result", record.Manifest.ArtifactSHA256)
		}
		if record.Manifest.RecordSHA256 == "" {
			t.Fatal("record digest is empty")
		}
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

func TestRunRejectsOversizedArtifactsBeforeCreatingRunStorage(t *testing.T) {
	large := strings.Repeat("x", int(maxRunArtifactBytes))
	tests := []struct {
		name     string
		artifact string
		prepare  func(*project.Bundle, *RunOptions)
	}{
		{
			name:     "design",
			artifact: "resolved-design.json",
			prepare: func(bundle *project.Bundle, _ *RunOptions) {
				bundle.Design.Device.Material = large
			},
		},
		{
			name:     "result",
			artifact: "result.json",
			prepare: func(_ *project.Bundle, opts *RunOptions) {
				opts.Evaluator = func(_ context.Context, _ project.Design, _ int64) (Result, error) {
					result := successResult(1)
					result.Warnings = []string{large}
					return result, nil
				}
			},
		},
		{
			name:     "manifest",
			artifact: "manifest.json",
			prepare: func(_ *project.Bundle, opts *RunOptions) {
				opts.RepositoryRevision = large
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := runnerBundle(t, 1)
			opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
				return successResult(float64(seed)), nil
			})
			test.prepare(&bundle, &opts)

			_, err := Run(context.Background(), bundle, opts)
			if err == nil || !strings.Contains(err.Error(), test.artifact) || !strings.Contains(err.Error(), "exceeds 16 MiB") {
				t.Fatalf("Run error=%v want %s 16 MiB error", err, test.artifact)
			}
			if _, statErr := os.Stat(filepath.Join(bundle.Root, "runs")); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("runs directory exists after rejected commit: %v", statErr)
			}
		})
	}
}

func rewriteRunJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsTamperedCachedResult(t *testing.T) {
	bundle := runnerBundle(t, 1)
	opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		return successResult(float64(seed)), nil
	})
	first, err := Run(context.Background(), bundle, opts)
	if err != nil {
		t.Fatal(err)
	}

	record := first.Runs[0]
	record.Result.Metrics[0].Value = 999
	resultPath := filepath.Join(bundle.Root, "runs", record.Manifest.RunID, "result.json")
	rewriteRunJSON(t, resultPath, record.Result)

	if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) {
		t.Fatalf("Run error=%v want ErrUnverifiableRun", err)
	}
}

func TestRunRejectsCachedDesignThatDoesNotMatchRunID(t *testing.T) {
	bundle := runnerBundle(t, 1)
	opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		return successResult(float64(seed)), nil
	})
	first, err := Run(context.Background(), bundle, opts)
	if err != nil {
		t.Fatal(err)
	}

	record := first.Runs[0]
	record.Design.Array.Rows++
	designBytes, err := canonicalJSON(record.Design)
	if err != nil {
		t.Fatal(err)
	}
	resultBytes, err := canonicalJSON(record.Result)
	if err != nil {
		t.Fatal(err)
	}
	record.Manifest.ArtifactSHA256["resolved-design.json"] = sha256Hex(designBytes)
	record.Manifest.RecordSHA256, err = recordSHA256(record.Manifest, designBytes, resultBytes)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(bundle.Root, "runs", record.Manifest.RunID)
	if err := os.WriteFile(filepath.Join(dir, "resolved-design.json"), designBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	rewriteRunJSON(t, filepath.Join(dir, "manifest.json"), record.Manifest)

	if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) || !strings.Contains(err.Error(), "identity mismatch") {
		t.Fatalf("Run error=%v want ErrUnverifiableRun identity mismatch", err)
	}
}

func TestRunRejectsNonCanonicalCachedManifest(t *testing.T) {
	transforms := []struct {
		name      string
		transform func(*testing.T, []byte) []byte
	}{
		{
			name: "whitespace",
			transform: func(_ *testing.T, data []byte) []byte {
				return append([]byte(" \n"), data...)
			},
		},
		{
			name: "key-order",
			transform: func(t *testing.T, data []byte) []byte {
				var manifest map[string]any
				if err := json.Unmarshal(data, &manifest); err != nil {
					t.Fatal(err)
				}
				reordered, err := canonicalJSON(manifest)
				if err != nil {
					t.Fatal(err)
				}
				return reordered
			},
		},
		{
			name: "duplicate-provenance-key",
			transform: func(t *testing.T, data []byte) []byte {
				trimmed := bytes.TrimSpace(data)
				if len(trimmed) == 0 || trimmed[len(trimmed)-1] != '}' {
					t.Fatalf("unexpected manifest bytes: %q", data)
				}
				return append(append(trimmed[:len(trimmed)-1], []byte(",\n  \"repository_revision\": \"abc123\"\n}")...), '\n')
			},
		},
	}
	for _, transform := range transforms {
		t.Run(transform.name, func(t *testing.T) {
			bundle := runnerBundle(t, 1)
			opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
				return successResult(float64(seed)), nil
			})
			first, err := Run(context.Background(), bundle, opts)
			if err != nil {
				t.Fatal(err)
			}
			manifestPath := filepath.Join(bundle.Root, "runs", first.Runs[0].Manifest.RunID, "manifest.json")
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(manifestPath, transform.transform(t, data), 0o644); err != nil {
				t.Fatal(err)
			}

			if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) || !strings.Contains(err.Error(), "manifest.json is not canonical") {
				t.Fatalf("Run error=%v want ErrUnverifiableRun canonical-manifest error", err)
			}
		})
	}
}

func TestRunRejectsTamperedCachedManifestProvenance(t *testing.T) {
	bundle := runnerBundle(t, 1)
	opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		return successResult(float64(seed)), nil
	})
	first, err := Run(context.Background(), bundle, opts)
	if err != nil {
		t.Fatal(err)
	}

	record := first.Runs[0]
	record.Manifest.RepositoryRevision = "tampered-revision"
	manifestPath := filepath.Join(bundle.Root, "runs", record.Manifest.RunID, "manifest.json")
	rewriteRunJSON(t, manifestPath, record.Manifest)
	if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) || !strings.Contains(err.Error(), "record digest mismatch") {
		t.Fatalf("Run error=%v want ErrUnverifiableRun record digest mismatch", err)
	}
}

func TestRunRejectsLegacySchemaOneCache(t *testing.T) {
	bundle := runnerBundle(t, 1)
	opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		return successResult(float64(seed)), nil
	})
	first, err := Run(context.Background(), bundle, opts)
	if err != nil {
		t.Fatal(err)
	}

	record := first.Runs[0]
	record.Manifest.SchemaVersion = 1
	record.Manifest.ArtifactSHA256 = nil
	record.Manifest.RecordSHA256 = ""
	dir := filepath.Join(bundle.Root, "runs", record.Manifest.RunID)
	manifestPath := filepath.Join(dir, "manifest.json")
	rewriteRunJSON(t, manifestPath, record.Manifest)
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	before = append([]byte(" \n"), before...)
	if err := os.WriteFile(manifestPath, before, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = Run(context.Background(), bundle, opts)
	if !errors.Is(err, ErrUnverifiableRun) || !strings.Contains(err.Error(), "archive or remove runs/"+record.Manifest.RunID+"/") {
		t.Fatalf("Run error=%v want legacy-schema archive/removal guidance", err)
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("legacy manifest was modified")
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("legacy run directory changed: info=%v err=%v", info, err)
	}
}

func TestRunRejectsMalformedCachedArtifacts(t *testing.T) {
	corruptions := map[string]func([]byte) []byte{
		"truncated": func([]byte) []byte { return []byte("{") },
		"unknown-field": func(data []byte) []byte {
			return bytes.Replace(data, []byte("{"), []byte(`{"unexpected":true,`), 1)
		},
		"multiple-values": func(data []byte) []byte { return append(data, []byte("{}\n")...) },
	}
	for _, artifact := range []string{"manifest.json", "resolved-design.json", "result.json"} {
		for corruption, corrupt := range corruptions {
			t.Run(artifact+"/"+corruption, func(t *testing.T) {
				bundle := runnerBundle(t, 1)
				opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
					return successResult(float64(seed)), nil
				})
				first, err := Run(context.Background(), bundle, opts)
				if err != nil {
					t.Fatal(err)
				}

				path := filepath.Join(bundle.Root, "runs", first.Runs[0].Manifest.RunID, artifact)
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, corrupt(data), 0o644); err != nil {
					t.Fatal(err)
				}
				if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) {
					t.Fatalf("Run error=%v want ErrUnverifiableRun", err)
				}
			})
		}
	}
}

func TestRunRejectsOversizedCachedArtifacts(t *testing.T) {
	for _, artifact := range []string{"manifest.json", "resolved-design.json", "result.json"} {
		t.Run(artifact, func(t *testing.T) {
			bundle := runnerBundle(t, 1)
			opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
				return successResult(float64(seed)), nil
			})
			first, err := Run(context.Background(), bundle, opts)
			if err != nil {
				t.Fatal(err)
			}

			path := filepath.Join(bundle.Root, "runs", first.Runs[0].Manifest.RunID, artifact)
			if err := os.Truncate(path, maxRunArtifactBytes+1); err != nil {
				t.Fatal(err)
			}
			if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) || !strings.Contains(err.Error(), "exceeds 16 MiB") {
				t.Fatalf("Run error=%v want ErrUnverifiableRun oversized-artifact error", err)
			}
		})
	}
}

func TestRunRejectsMissingCachedArtifacts(t *testing.T) {
	for _, artifact := range []string{"manifest.json", "resolved-design.json", "result.json"} {
		t.Run(artifact, func(t *testing.T) {
			bundle := runnerBundle(t, 1)
			opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
				return successResult(float64(seed)), nil
			})
			first, err := Run(context.Background(), bundle, opts)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(bundle.Root, "runs", first.Runs[0].Manifest.RunID, artifact)
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) {
				t.Fatalf("Run error=%v want ErrUnverifiableRun", err)
			}
		})
	}
}

func TestRunRejectsCachedStatusInconsistency(t *testing.T) {
	bundle := runnerBundle(t, 1)
	opts := deterministicRunOptions(func(_ context.Context, _ project.Design, seed int64) (Result, error) {
		return successResult(float64(seed)), nil
	})
	first, err := Run(context.Background(), bundle, opts)
	if err != nil {
		t.Fatal(err)
	}

	record := first.Runs[0]
	record.Result.Status = StatusFailed
	resultBytes, err := canonicalJSON(record.Result)
	if err != nil {
		t.Fatal(err)
	}
	designBytes, err := canonicalJSON(record.Design)
	if err != nil {
		t.Fatal(err)
	}
	record.Manifest.ArtifactSHA256["result.json"] = sha256Hex(resultBytes)
	record.Manifest.RecordSHA256, err = recordSHA256(record.Manifest, designBytes, resultBytes)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(bundle.Root, "runs", record.Manifest.RunID)
	if err := os.WriteFile(filepath.Join(dir, "result.json"), resultBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	rewriteRunJSON(t, filepath.Join(dir, "manifest.json"), record.Manifest)

	if _, err := Run(context.Background(), bundle, opts); !errors.Is(err, ErrUnverifiableRun) || !strings.Contains(err.Error(), "inconsistent status") {
		t.Fatalf("Run error=%v want ErrUnverifiableRun inconsistent status", err)
	}
}

func TestStoreCommitPropagatesUnverifiableRenameWinner(t *testing.T) {
	bundle := runnerBundle(t, 1)
	points, err := Expand(bundle)
	if err != nil {
		t.Fatal(err)
	}
	id, err := ID(points[0], "test-v1", bundle.Inputs)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	record := RunRecord{
		Manifest: RunManifest{
			SchemaVersion:    2,
			RunID:            id,
			PointIndex:       points[0].Index,
			Seed:             points[0].Seed,
			EvaluatorVersion: "test-v1",
			Inputs:           bundle.Inputs,
			StartedAt:        now,
			CompletedAt:      now,
			Status:           StatusSuccess,
		},
		Design: points[0].Design,
		Result: successResult(1),
	}
	renameErr := errors.New("rename lost")
	storage := store{
		root: bundle.Root,
		rename: func(_, finalDir string) error {
			if err := os.Mkdir(finalDir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(finalDir, "manifest.json"), []byte("{"), 0o644); err != nil {
				t.Fatal(err)
			}
			return renameErr
		},
	}

	_, err = storage.commit(record)
	if !errors.Is(err, ErrUnverifiableRun) {
		t.Fatalf("commit error=%v want ErrUnverifiableRun", err)
	}
	if errors.Is(err, renameErr) {
		t.Fatalf("commit returned raw rename error instead of corrupt winner error: %v", err)
	}
	assertNoTempRuns(t, bundle.Root)
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
