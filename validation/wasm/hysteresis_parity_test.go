//go:build !js

package wasm_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/hysteresis"
)

func TestHysteresisProductionParity(t *testing.T) { productionGate(t, "parity") }
func TestHysteresisWorkerLifecycle(t *testing.T)  { productionGate(t, "lifecycle") }
func TestHysteresisWorkerBudgets(t *testing.T)    { productionGate(t, "budgets") }

// A byte-equality gate must not certify identical incomplete/failed science.
func TestHysteresisRejectsIncompleteParityTranscript(t *testing.T) {
	old := hysteresis.BuildID
	hysteresis.BuildID = "test-pinned"
	defer func() { hysteresis.BuildID = old }()
	identity, _ := json.Marshal(hysteresis.BridgeIdentity())
	b := hysteresis.NewBridge()
	var replies []json.RawMessage
	for _, q := range productionRequests(t, identity) {
		replies = append(replies, b.Dispatch(q))
	}
	data, _ := json.Marshal(replies)
	if err := validateProductionTranscript(data); err != nil {
		t.Fatal(err)
	}
	// Fault injection changes only the real completed write's terminal evidence.
	for i, r := range replies {
		var v struct {
			Operation string                         `json:"operation"`
			Result    hysteresis.ProgrammingSnapshot `json:"result"`
		}
		json.Unmarshal(r, &v)
		if v.Operation == "program.advance" && v.Result.Target == 1 && v.Result.Complete {
			replies[i] = bytes.Replace(r, []byte(`"Complete":true`), []byte(`"Complete":false`), 1)
			break
		}
	}
	data, _ = json.Marshal(replies)
	if err := validateProductionTranscript(data); err == nil {
		t.Fatal("identical incomplete programming qualified as parity")
	}
}

func productionGate(t *testing.T, mode string) {
	t.Helper()
	if os.Getenv("FECIM_WASM_PARITY") != "1" {
		t.Skip("enable production three-runtime gate with FECIM_WASM_PARITY=1")
	}
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if parent := os.Getenv("FECIM_WASM_ARTIFACT_DIR"); parent != "" {
		if err := os.MkdirAll(parent, 0700); err != nil {
			t.Fatal(err)
		}
		dir, err = os.MkdirTemp(parent, "hysteresis-"+mode+"-")
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("production %s artifacts: %s", mode, dir)
	// The identity hashes tracked Go source/dependency bytes plus the production
	// driver and exact toolchain shim, rather than trusting a mutable git label.
	goroot := strings.TrimSpace(command(t, repo, nil, "go", "env", "GOROOT"))
	shim := filepath.Join(goroot, "lib/wasm/wasm_exec.js")
	names := strings.Fields(command(t, repo, nil, "git", "ls-files", "--cached", "--others", "--exclude-standard", "*.go", "go.mod", "go.sum"))
	names = append(names, "config/constants.yaml", "config/materials.yaml", "web/src/worker/hysteresis-worker.mjs", "wasm_exec.js", "workbench/moduleapi/openapi.json", "workbench/moduleapi/research/example/project.yaml", "workbench/moduleapi/research/example/design.yaml", "workbench/moduleapi/research/example/sweep.yaml", "web/scripts/build-models.mjs", "web/scripts/generate-api.mjs")
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		p := name
		if name == "wasm_exec.js" {
			p = shim
		} else {
			p = filepath.Join(repo, p)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		hash.Write([]byte(name))
		hash.Write(data)
	}
	build := hex.EncodeToString(hash.Sum(nil))
	flags := "-s -w -X fecim-lattice-tools/workbench/hysteresis.BuildID=" + build
	native := filepath.Join(dir, "hysteresis-native")
	command(t, repo, nil, "go", "build", "-buildvcs=false", "-trimpath", "-ldflags="+flags, "-o", native, "./cmd/fecim-hysteresis-wasm")
	wasm := filepath.Join(dir, "hysteresis.wasm")
	command(t, repo, []string{"GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0"}, "go", "build", "-buildvcs=false", "-trimpath", "-ldflags="+flags, "-o", wasm, "./cmd/fecim-hysteresis-wasm")
	identity := json.RawMessage(strings.TrimSpace(command(t, repo, nil, native, "identity")))
	requests := productionRequests(t, identity)
	input := filepath.Join(dir, "input.json")
	data, _ := json.Marshal(requests)
	if err := os.WriteFile(input, data, 0600); err != nil {
		t.Fatal(err)
	}
	nativeJSON := command(t, repo, nil, native, input)
	if err := os.WriteFile(filepath.Join(dir, "native-input.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	rawStrings := func(trace []json.RawMessage) []string {
		out := make([]string, len(trace))
		for i, q := range trace {
			out[i] = string(q)
		}
		return out
	}
	fixture, _ := json.Marshal(map[string]any{"parity": rawStrings(requests), "cold": rawStrings(coldRequests(t, identity))})
	if err := os.WriteFile(input, fixture, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "native-results.json"), []byte(nativeJSON), 0600); err != nil {
		t.Fatal(err)
	}
	// Identical failures/empty results must not qualify as scientific parity.
	var replies []struct {
		Error  string          `json:"error"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal([]byte(nativeJSON), &replies); err != nil {
		t.Fatal(err)
	}
	if len(replies) != len(requests) {
		t.Fatal("incomplete native transcript")
	}
	for i, r := range replies {
		if r.Error != "" || len(r.Result) == 0 || bytes.Equal(r.Result, []byte("null")) {
			t.Fatalf("native operation %d failed: %s", i, nativeJSON)
		}
	}
	if err := validateProductionTranscript([]byte(nativeJSON)); err != nil {
		t.Fatal(err)
	}
	output := command(t, repo, nil, "node", "validation/wasm/hysteresis_harness.mjs", mode, wasm, shim, input, dir)
	t.Log(output)
	for _, runtime := range []string{"node", "firefox"} {
		actual, err := os.ReadFile(filepath.Join(dir, runtime+"-results.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal([]byte(nativeJSON), actual) {
			t.Errorf("native vs %s: %s", runtime, firstDifference([]byte(nativeJSON), actual))
		}
	}
}

func validateProductionTranscript(data []byte) error {
	var replies []struct {
		Operation string          `json:"operation"`
		Error     string          `json:"error"`
		Result    json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &replies); err != nil {
		return err
	}
	coverage := map[string]bool{"completed programming": false, "fresh calibration": false, "stale calibration": false, "unavailable protocol": false, "independent PUND": false, "maximum FORC": false, "unverified reference": false, "negative field": false, "positive field": false, "cancelled": false, "LK": false, "Preisach": false, "sine": false, "triangle": false}
	for _, reply := range replies {
		if reply.Error != "" || len(reply.Result) == 0 || bytes.Equal(reply.Result, []byte("null")) {
			return fmt.Errorf("unsuccessful production operation %s: %s", reply.Operation, reply.Error)
		}
		switch reply.Operation {
		case "core.page":
			var p hysteresis.Page
			if err := json.Unmarshal(reply.Result, &p); err != nil {
				return err
			}
			for _, r := range p.Rows {
				coverage["negative field"] = coverage["negative field"] || r.E.SI < 0
				coverage["positive field"] = coverage["positive field"] || r.E.SI > 0
			}
		case "core.advance":
			var s hysteresis.Snapshot
			if err := json.Unmarshal(reply.Result, &s); err != nil {
				return err
			}
			if s.SolverSteps > 0 {
				if s.Config.Model == hysteresis.Landau {
					coverage["LK"] = true
				} else if s.Config.Model == hysteresis.Preisach {
					coverage["Preisach"] = true
				}
				if s.Config.Mode == hysteresis.Sine {
					coverage["sine"] = true
				}
				if s.Config.Mode == hysteresis.Triangle {
					coverage["triangle"] = true
				}
			}
		case "program.new", "program.snapshot", "program.advance":
			var p hysteresis.ProgrammingSnapshot
			if err := json.Unmarshal(reply.Result, &p); err != nil {
				return err
			}
			if p.Status == hysteresis.ProgramFailed || p.Status == hysteresis.ProgramWorkLimit {
				return fmt.Errorf("programming failed: %s", p.Reason)
			}
			if reply.Operation == "program.advance" && p.Target == 1 && p.Status == hysteresis.ProgramComplete && p.Complete && p.Work > 0 {
				coverage["completed programming"] = true
			}
			if p.Calibration.Status == hysteresis.CalibrationFresh && p.Calibration.CanExport && len(p.Calibration.Ascending) == 2 && len(p.Calibration.Descending) == 2 {
				coverage["fresh calibration"] = true
			}
			if p.Calibration.Status == hysteresis.CalibrationStale && !p.Calibration.CanExport && !p.Calibration.CanApply {
				coverage["stale calibration"] = true
			}
			if p.Status == hysteresis.LegacyProtocolUnavailable && p.Reason != "" && len(p.Core.Plot) == 0 {
				coverage["unavailable protocol"] = true
			}
		case "analysis.result":
			var a hysteresis.AnalysisResult
			if err := json.Unmarshal(reply.Result, &a); err != nil {
				return err
			}
			if a.ExperimentalValidation || a.Evidence != "simulation_only" {
				return fmt.Errorf("diagnostic evidence mislabeled")
			}
			if a.PUND != nil && a.Diagnostic.Profile == "legacy-pund-demo" && a.Diagnostic.Model == hysteresis.Preisach && a.Context.Config.Model == hysteresis.Landau {
				valid := true
				for _, trace := range a.PUND.Traces {
					valid = valid && len(trace) > 0
				}
				coverage["independent PUND"] = valid
			}
			if a.FORC != nil && a.Diagnostic.Profile == "gui-forc" && a.Diagnostic.Model == hysteresis.Preisach && len(a.FORC.Sweep.Curves) == 64 && a.FORC.Matrix.Rows == 64 && a.FORC.Matrix.Cols == 256 && len(a.FORC.Matrix.Density) == 64 {
				valid := true
				for _, row := range a.FORC.Matrix.Density {
					valid = valid && len(row) == 256
				}
				coverage["maximum FORC"] = valid
			}
			for _, ref := range a.References {
				if ref.Evidence == "imported_unverified" && !ref.ExperimentalValidation && len(ref.Points) > 0 {
					coverage["unverified reference"] = true
				}
			}
		case "job.cancel":
			var p struct {
				State  string
				Status string
			}
			if err := json.Unmarshal(reply.Result, &p); err != nil {
				return err
			}
			coverage["cancelled"] = coverage["cancelled"] || p.State == "cancelled" || p.Status == "cancelled"
		}
	}
	var missing []string
	for name, seen := range coverage {
		if !seen {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return fmt.Errorf("incomplete production science: %s", strings.Join(missing, ", "))
	}
	return nil
}

func coldRequests(t *testing.T, identity json.RawMessage) []json.RawMessage {
	t.Helper()
	c := hysteresis.DefaultConfig()
	inputs := []struct {
		op string
		in any
	}{
		{"session.new", hysteresis.Config{}},
		{"core.apply", hysteresis.Command{Kind: hysteresis.Run}},
		{"core.advance", map[string]any{"Steps": 64, "DT": 1 / (c.Frequency * 64)}},
		{"core.apply", hysteresis.Command{Kind: hysteresis.Pause}},
		{"core.page", map[string]any{"Offset": 0}},
	}
	var trace []json.RawMessage
	for i, q := range inputs {
		b, err := json.Marshal(map[string]any{"identity": identity, "session": "qualification", "request": i + 1, "revision": i, "operation": q.op, "input": q.in})
		if err != nil {
			t.Fatal(err)
		}
		trace = append(trace, b)
	}
	return trace
}

func productionRequests(t *testing.T, identity json.RawMessage) []json.RawMessage {
	t.Helper()
	var out []json.RawMessage
	revision := 0
	add := func(op string, input any) {
		b, err := json.Marshal(map[string]any{"identity": identity, "session": "qualification", "request": len(out) + 1, "revision": revision, "operation": op, "input": input})
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, b)
		revision++
	}
	c := hysteresis.DefaultConfig()
	c.Mode = hysteresis.Manual
	c.Seed = 42
	add("session.new", c)
	add("core.apply", hysteresis.Command{Kind: hysteresis.Run})
	for _, e := range []float64{math.Copysign(0, -1), 1.5e8, 0, -1.5e8, 0, 8e7, -8e7} {
		add("core.apply", hysteresis.Command{Kind: hysteresis.SetField, Field: e})
		add("core.advance", map[string]any{"Steps": 1, "DT": 1e-9})
	}
	add("core.apply", hysteresis.Command{Kind: hysteresis.Pause})
	add("core.page", map[string]any{"Offset": 0})
	add("analysis.new", map[string]any{"Kind": "pund"})
	add("analysis.advance", map[string]any{"Steps": 1})
	add("analysis.result", nil)
	add("analysis.page", map[string]any{"Table": "pund-traces", "Offset": 0})
	add("analysis.new", map[string]any{"Kind": "forc", "FORC": hysteresis.FORCParameters{EmaxVm: 1, Reversals: 3, Resolution: 3}})
	for range 7 {
		add("analysis.advance", map[string]any{"Steps": 1})
	}
	add("analysis.result", nil)
	add("analysis.page", map[string]any{"Table": "forc-density", "Offset": 0})
	p := hysteresis.DefaultProgramConfig()
	p.Mode = hysteresis.TimeResolved
	add("program.new", p)
	add("program.snapshot", nil)
	p.Mode = hysteresis.TargetWrite
	add("program.new", p)
	add("program.apply", hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 1})
	add("program.advance", map[string]any{"Steps": 256, "DT": 0.1}) // complete a real target write
	add("program.snapshot", nil)
	// Complete native Preisach calibration, then invalidate by settings change.
	c.Levels = 2
	add("core.apply", hysteresis.Command{Kind: hysteresis.Configure, Config: c})
	add("program.apply", hysteresis.ProgramCommand{Kind: hysteresis.CalibrateProgram})
	for range 8 {
		add("program.advance", map[string]any{"Steps": 256, "DT": 0.01})
	}
	c.Levels = 64
	c.HistoryLimit = hysteresis.MaxSamples
	add("core.apply", hysteresis.Command{Kind: hysteresis.Configure, Config: c})
	add("program.snapshot", nil)
	add("program.apply", hysteresis.ProgramCommand{Kind: hysteresis.CalibrateProgram})
	add("program.advance", map[string]any{"Steps": 256, "DT": 1})
	add("job.cancel", nil)
	// Both loop engines and waveform modes, including their signed complete
	// tables. These commands do not retune physics for cross-runtime comparison.
	for _, model := range []hysteresis.Model{hysteresis.Preisach, hysteresis.Landau} {
		for _, mode := range []hysteresis.Mode{hysteresis.Sine, hysteresis.Triangle} {
			c.Model = model
			c.Mode = mode
			add("core.apply", hysteresis.Command{Kind: hysteresis.Configure, Config: c})
			add("core.apply", hysteresis.Command{Kind: hysteresis.Reset})
			add("core.apply", hysteresis.Command{Kind: hysteresis.Run})
			add("core.advance", map[string]any{"Steps": 256, "DT": 1 / (c.Frequency * 256)})
			add("core.apply", hysteresis.Command{Kind: hysteresis.Pause})
			add("core.page", map[string]any{"Offset": 128})
		}
	}
	// Maximum admitted domain count and both methods share the same seeded LK
	// engine. Measure one bounded batch and cancellation, not a convergence claim.
	c.Mode = hysteresis.Manual
	add("core.apply", hysteresis.Command{Kind: hysteresis.Configure, Config: c})
	p.Mode = hysteresis.WriteRead
	p.Domains = 128
	p.Method = hysteresis.LinearISPP
	add("program.new", p)
	add("program.apply", hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 22})
	add("program.advance", map[string]any{"Steps": 256, "DT": 0.01})
	add("job.cancel", nil)
	p.Domains = 1
	add("program.new", p)
	add("program.apply", hysteresis.ProgramCommand{Kind: hysteresis.CalibrateProgram})
	add("program.advance", map[string]any{"Steps": 256, "DT": 0.01})
	add("job.cancel", nil)
	// Actual diagnostic identity must remain independent of selected LK context.
	add("analysis.new", map[string]any{"Kind": "pund", "References": []hysteresis.ReferenceInput{{Format: "csv", Source: "synthetic cited example", Data: []byte("E_MV_cm,P_uC_cm2\n-1,-20\n1,20\n")}}})
	add("analysis.advance", map[string]any{"Steps": 1})
	add("analysis.result", nil)
	add("core.snapshot", nil)
	add("analysis.new", map[string]any{"Kind": "forc", "FORC": hysteresis.FORCParameters{EmaxVm: 1, Reversals: 64, Resolution: 256}})
	for range 129 {
		add("analysis.advance", map[string]any{"Steps": 1})
	}
	add("analysis.result", nil)
	add("analysis.page", map[string]any{"Table": "forc-density", "Offset": 16352})
	add("analysis.new", map[string]any{"Kind": "forc", "FORC": hysteresis.DefaultFORCParameters()})
	add("analysis.advance", map[string]any{"Steps": 1})
	add("job.cancel", nil)
	return out
}
