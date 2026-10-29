//go:build !js

package wasm_test

import (
	"bytes"
	"encoding/json"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	crossbar "fecim-lattice-tools/workbench/modules/crossbar"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCrossbarHostedDependencyBoundary(t *testing.T) {
	repo, _ := filepath.Abs("../..")
	deps := command(t, repo, []string{"GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0"}, "go", "list", "-deps", "./web/src/modules/crossbar/wasm")
	for _, forbidden := range []string{"fyne.io/", "shared/compute", "net/http", "module3-mnist/", "module6-eda/"} {
		if strings.Contains(deps, forbidden) {
			t.Errorf("hosted CPU bundle links %s", forbidden)
		}
	}
}

func TestModuleProductionParityCrossbar(t *testing.T) {
	if os.Getenv("FECIM_WASM_PARITY") != "1" {
		t.Skip("enable real native/Node/Firefox parity with FECIM_WASM_PARITY=1")
	}
	repo, _ := filepath.Abs("../..")
	dir := t.TempDir()
	build := "crossbar-parity"
	old := hysteresis.BuildID
	hysteresis.BuildID = build
	defer func() { hysteresis.BuildID = old }()
	requests := []crossbar.Request{}
	for _, profile := range []string{"basic", "enhanced", "parasitic", "fecap"} {
		requests = append(requests, crossbar.Request{Identity: crossbar.Identity(), Session: "crossbar-parity-" + profile, Request: 1, Operation: "action", Input: crossbar.BridgeInput{Command: moduleapi.Command{Action: "configure", Input: map[string]string{"rows": "2", "cols": "3", "profile": profile, "noise": "0.02", "sneak": "true", "matrix": "[[0.2,0.5,0.8],[0.9,0.4,0.1]]", "input": "[0.5,1,0.25]"}}}})
	}
	for _, profile := range []string{"basic", "enhanced", "parasitic", "fecap"} {
		requests = append(requests, crossbar.Request{Identity: crossbar.Identity(), Session: "crossbar-max-" + profile, Request: 1, Operation: "action", Input: crossbar.BridgeInput{Command: moduleapi.Command{Action: "configure", Input: map[string]string{"rows": "32", "cols": "32", "profile": profile, "noise": "0.02", "sneak": "true"}}}})
	}
	input := filepath.Join(dir, "input.json")
	data, _ := json.Marshal(requests)
	if err := os.WriteFile(input, data, 0600); err != nil {
		t.Fatal(err)
	}
	flags := "-s -w -X fecim-lattice-tools/workbench/hysteresis.BuildID=" + build
	wasm := filepath.Join(dir, "crossbar.wasm")
	command(t, repo, []string{"GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0"}, "go", "build", "-buildvcs=false", "-trimpath", "-ldflags="+flags, "-o", wasm, "./web/src/modules/crossbar/wasm")
	want := []byte(command(t, repo, nil, "go", "run", "-ldflags="+flags, "./cmd/fecim-crossbar-wasm", input))
	var successful []string
	if err := json.Unmarshal(want, &successful); err != nil || len(successful) != len(requests) {
		t.Fatal("native did not produce all admitted scientific results", err)
	}
	for i, raw := range successful {
		var result crossbar.Result
		if err := json.Unmarshal([]byte(raw), &result); err != nil || len(result.Actual) < 2 || result.Actual[0] <= 0 || len(result.Physical) != result.Config.Rows {
			t.Fatalf("profile %d lacks successful cell/vector science: %v", i, err)
		}
	}
	goroot := strings.TrimSpace(command(t, repo, nil, "go", "env", "GOROOT"))
	t.Log(command(t, repo, nil, "node", "validation/wasm/crossbar_harness.mjs", wasm, filepath.Join(goroot, "lib/wasm/wasm_exec.js"), input, dir))
	for _, runtime := range []string{"node", "firefox"} {
		got, err := os.ReadFile(filepath.Join(dir, runtime+"-crossbar.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(want, got) {
			t.Fatalf("native/%s scientific bytes differ: %s", runtime, firstDifference(want, got))
		}
	}
}
