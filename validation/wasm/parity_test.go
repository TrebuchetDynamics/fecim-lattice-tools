//go:build !js

package wasm_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/study"
)

// This gate catches any numerical, identity, ordering or serialization difference
// between the real native persisted study and the Go/WASM compute/report path.
// Missing tools are errors when explicitly enabled, never a passing skip.
func TestExampleResultsJSONNativeWASM(t *testing.T) {
	if os.Getenv("FECIM_WASM_PARITY") != "1" {
		t.Skip("enable the required three-runtime gate with FECIM_WASM_PARITY=1; see README.md")
	}
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	artifacts := t.TempDir()
	if parent := os.Getenv("FECIM_WASM_ARTIFACT_DIR"); parent != "" {
		if err := os.MkdirAll(parent, 0o700); err != nil {
			t.Fatal(err)
		}
		artifacts, err = os.MkdirTemp(parent, "parity-")
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("artifacts: %s", artifacts)
	root := filepath.Join(artifacts, "native-project")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"project.yaml", "design.yaml", "sweep.yaml"} {
		data, err := os.ReadFile(filepath.Join(repo, "examples/research-design-workbench", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	revision := strings.TrimSpace(command(t, repo, nil, "git", "rev-parse", "HEAD"))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	outcome, err := study.Run(ctx, root, study.RunOptions{Workers: 1, CitationDir: filepath.Join(repo, "citations/papers"), RepositoryRevision: revision})
	if err != nil {
		t.Fatal(err)
	}
	if len(outcome.Summary.Runs) != 8 {
		t.Fatalf("example expanded to %d runs, want 8", len(outcome.Summary.Runs))
	}
	native, err := os.ReadFile(filepath.Join(root, "reports/results.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateExampleReport(native); err != nil {
		t.Fatalf("native example no longer exercises numerical results: %v", err)
	}
	if err := os.WriteFile(filepath.Join(artifacts, "native-results.json"), native, 0o600); err != nil {
		t.Fatal(err)
	}
	// Only transport changes: the bundle is resolved by the existing Go loader.
	// The WASM harness intentionally has no browser filesystem or YAML fork.
	bundle := outcome.Bundle
	bundle.Root = ""
	input, err := json.Marshal(struct {
		Bundle   project.Bundle `json:"bundle"`
		Revision string         `json:"revision"`
	}{bundle, revision})
	if err != nil {
		t.Fatal(err)
	}
	inputPath := filepath.Join(artifacts, "input.json")
	if err := os.WriteFile(inputPath, input, 0o600); err != nil {
		t.Fatal(err)
	}
	wasm := filepath.Join(artifacts, "study.wasm")
	command(t, repo, []string{"GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0"}, "go", "build", "-trimpath", "-ldflags=-s -w", "-o", wasm, "./validation/wasm/testdata/runner")
	goroot := strings.TrimSpace(command(t, repo, nil, "go", "env", "GOROOT"))
	output := command(t, repo, nil, "node", "validation/wasm/harness.mjs", wasm, filepath.Join(goroot, "lib/wasm/wasm_exec.js"), inputPath, artifacts)
	t.Log(output)
	for _, runtime := range []string{"node", "firefox"} {
		actual, err := os.ReadFile(filepath.Join(artifacts, runtime+"-results.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(native, actual) {
			t.Errorf("native vs %s results.json differ: first field %s; artifacts in %s (set FECIM_WASM_ARTIFACT_DIR to retain after test cleanup)", runtime, firstDifference(native, actual), artifacts)
		}
	}
}

func command(t *testing.T, dir string, extra []string, name string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	if err := protectProcessTree(cmd); err != nil {
		t.Fatal(err)
	}
	cmd.Dir = dir
	cmd.Env = os.Environ()
	for _, value := range extra {
		key := strings.SplitN(value, "=", 2)[0] + "="
		filtered := cmd.Env[:0]
		for _, entry := range cmd.Env {
			if !strings.HasPrefix(entry, key) {
				filtered = append(filtered, entry)
			}
		}
		cmd.Env = append(filtered, value)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
	return string(output)
}

func validateExampleReport(data []byte) error {
	var parsed struct {
		Runs []struct {
			Result experiment.Result `json:"result"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	if len(parsed.Runs) != 8 {
		return fmt.Errorf("want 8 numerical runs, got %d", len(parsed.Runs))
	}
	for i, run := range parsed.Runs {
		if run.Result.Status != experiment.StatusSuccess || run.Result.Failure != nil {
			return fmt.Errorf("run %d is not successful", i)
		}
		want := map[string]bool{"area_um2": false, "energy_pj": false, "latency_ns": false, "tia_snr_db": false}
		if len(run.Result.Metrics) != len(want) {
			return fmt.Errorf("run %d lacks its four numerical metrics", i)
		}
		for _, metric := range run.Result.Metrics {
			seen, known := want[metric.Name]
			if !known || seen || math.IsNaN(metric.Value) || math.IsInf(metric.Value, 0) {
				return fmt.Errorf("run %d has invalid numerical metric %q", i, metric.Name)
			}
			want[metric.Name] = true
		}
	}
	return nil
}

func TestExampleRequiresNumericalResults(t *testing.T) {
	fixture := func() []byte {
		var runs []map[string]any
		for i := 0; i < 8; i++ {
			runs = append(runs, map[string]any{"result": map[string]any{"status": "success", "metrics": []map[string]any{
				{"name": "area_um2", "value": 1}, {"name": "energy_pj", "value": 2},
				{"name": "latency_ns", "value": 3}, {"name": "tia_snr_db", "value": 4},
			}}})
		}
		data, _ := json.Marshal(map[string]any{"runs": runs})
		return data
	}
	valid := fixture()
	if err := validateExampleReport(valid); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{
		bytes.ReplaceAll(valid, []byte(`"success"`), []byte(`"failed"`)),
		bytes.ReplaceAll(valid, []byte(`"area_um2"`), []byte(`"missing_metric"`)),
		[]byte(`{"runs":[]}`),
	} {
		if err := validateExampleReport(bad); err == nil {
			t.Error("numerically empty/failed/missing-metric example passed")
		}
	}
}

func TestFirstDifference(t *testing.T) {
	for _, tc := range []struct{ a, b, want string }{
		{`{"runs":[{"value":1}]}`, `{"runs":[{"value":1.0000000001}]}`, "/runs/0/value"},
		{`{"counts":{"failed":0}}`, `{"counts":{"failed":1}}`, "/counts/failed"},
		{`{"runs":[]}`, `{"runs":[{}]}`, "/runs/length"},
		{`{"x":null}`, `{}`, "/x"},
		{`{"x":1}`, `[]`, "<root>"},
		{"{\"x\":1}\n", `{ "x":1}`, "<serialization>"},
	} {
		if got := firstDifference([]byte(tc.a), []byte(tc.b)); got != tc.want {
			t.Errorf("firstDifference(%s,%s)=%s, want %s", tc.a, tc.b, got, tc.want)
		}
	}
}

// Diagnostic only: byte equality above is authoritative. UseNumber preserves
// exact numeric lexemes and never rounds a mismatch away for a pretty message.
func firstDifference(a, b []byte) string {
	decode := func(data []byte) (any, error) {
		var value any
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		err := decoder.Decode(&value)
		return value, err
	}
	left, err := decode(a)
	if err != nil {
		return "<invalid native JSON>"
	}
	right, err := decode(b)
	if err != nil {
		return "<invalid runtime JSON>"
	}
	var walk func(any, any, string) string
	walk = func(a, b any, path string) string {
		if reflect.TypeOf(a) != reflect.TypeOf(b) {
			if path == "" {
				return "<root>"
			}
			return path
		}
		switch a := a.(type) {
		case map[string]any:
			b := b.(map[string]any)
			keys := make(map[string]bool)
			for k := range a {
				keys[k] = true
			}
			for k := range b {
				keys[k] = true
			}
			ordered := make([]string, 0, len(keys))
			for k := range keys {
				ordered = append(ordered, k)
			}
			sort.Strings(ordered)
			for _, k := range ordered {
				av, aok := a[k]
				bv, bok := b[k]
				next := path + "/" + strings.ReplaceAll(strings.ReplaceAll(k, "~", "~0"), "/", "~1")
				if aok != bok {
					return next
				}
				if d := walk(av, bv, next); d != "" {
					return d
				}
			}
		case []any:
			b := b.([]any)
			if len(a) != len(b) {
				return path + "/length"
			}
			for i := range a {
				if d := walk(a[i], b[i], fmt.Sprintf("%s/%d", path, i)); d != "" {
					return d
				}
			}
		default:
			if !reflect.DeepEqual(a, b) {
				return path
			}
		}
		return ""
	}
	if path := walk(left, right, ""); path != "" {
		return path
	}
	return "<serialization>"
}
