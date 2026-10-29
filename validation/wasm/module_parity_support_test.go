//go:build !js

package wasm_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/research"
	"fecim-lattice-tools/workbench/study"
)

func TestModuleHostedDependencyBoundary(t *testing.T) {
	repo, _ := filepath.Abs("../..")
	deps := command(t, repo, []string{"GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0"}, "go", "list", "-deps", "./web/src/modules/research/wasm")
	for _, forbidden := range []string{"module6-eda/", "workbench/edaexport", "module3-mnist/", "fyne.io/", "net/http"} {
		if strings.Contains(deps, forbidden) {
			t.Errorf("hosted science links forbidden dependency %s", forbidden)
		}
	}
}

// qualifyModule uses real Go-generated envelopes and original scientific bytes.
// Feature owners can reuse this helper with their own production entry/fixture;
// it is deliberately not another registry of feature-specific tests.
func qualifyModule(t *testing.T, repo, entry, nativeEntry, fixture, build string, want []byte) {
	t.Helper()
	dir := t.TempDir()
	wasm := filepath.Join(dir, "module.wasm")
	flags := "-s -w -X fecim-lattice-tools/workbench/hysteresis.BuildID=" + build
	command(t, repo, []string{"GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0"}, "go", "build", "-buildvcs=false", "-trimpath", "-ldflags="+flags, "-o", wasm, entry)
	goroot := strings.TrimSpace(command(t, repo, nil, "go", "env", "GOROOT"))
	input := filepath.Join(dir, "input.json")
	if err := os.WriteFile(input, []byte(fixture), 0600); err != nil {
		t.Fatal(err)
	}
	native := []byte(command(t, repo, nil, "go", "run", "-ldflags="+flags, nativeEntry, input))
	if !bytes.Equal(want, native) {
		t.Fatal("native production runner differs from persisted original report")
	}
	output := command(t, repo, nil, "node", "validation/wasm/module_harness.mjs", wasm, filepath.Join(goroot, "lib/wasm/wasm_exec.js"), input, dir)
	t.Log(output)
	for _, runtime := range []string{"node", "firefox"} {
		got, err := os.ReadFile(filepath.Join(dir, runtime+"-results.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateExampleReport(got); err != nil {
			t.Fatalf("%s did not execute successful finite science: %v", runtime, err)
		}
		if !bytes.Equal(want, got) {
			t.Fatalf("native/%s original bytes differ: %s", runtime, firstDifference(want, got))
		}
	}
}
func TestModuleResearchNativeWASM(t *testing.T) {
	if os.Getenv("FECIM_WASM_PARITY") != "1" {
		t.Skip("enable required module production parity with FECIM_WASM_PARITY=1")
	}
	repo, _ := filepath.Abs("../..")
	nativeDir := t.TempDir()
	for name, data := range research.ExampleDocuments() {
		if err := os.WriteFile(filepath.Join(nativeDir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	native, err := study.Run(ctx, nativeDir, study.RunOptions{Workers: 1, RepositoryRevision: "module-parity"})
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(nativeDir, "reports/results.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = validateExampleReport(want); err != nil {
		t.Fatal(err)
	}
	// Use the production bridge natively too; transport progress/timestamps are
	// intentionally not canonical scientific serialization.
	old := hysteresis.BuildID
	hysteresis.BuildID = "module-parity"
	defer func() { hysteresis.BuildID = old }()
	bridge := research.New()
	defer bridge.Close()
	base := moduleapi.ResearchRequest{Identity: research.Identity(), Session: "module-parity", Request: 1, Operation: "example", Input: moduleapi.ResearchInput{RepositoryRevision: "module-parity"}}
	fixture, _ := json.Marshal(base)
	dispatch := func(op string) moduleapi.ResearchReply {
		base.Operation = op
		data, _ := json.Marshal(base)
		var r moduleapi.ResearchReply
		if err := json.Unmarshal(bridge.Dispatch(data), &r); err != nil {
			t.Fatal(err)
		}
		if r.Error != "" {
			t.Fatal(r.Error)
		}
		base.Request++
		base.Revision = r.Revision
		return r
	}
	dispatch("example")
	dispatch("start")
	for {
		r := dispatch("state")
		if r.Result.Job.State == "finished" {
			if !bytes.Equal(want, []byte(r.Result.ResultsJSON)) {
				t.Fatal("native production bridge differs")
			}
			break
		}
		if ctx.Err() != nil {
			t.Fatal("bridge timeout")
		}
		time.Sleep(time.Millisecond)
	}
	if len(native.Summary.Runs) != 8 {
		t.Fatal("fixture no longer has eight runs")
	}
	qualifyModule(t, repo, "./web/src/modules/research/wasm", "./validation/wasm/testdata/module_runner", string(fixture), "module-parity", want)
}
