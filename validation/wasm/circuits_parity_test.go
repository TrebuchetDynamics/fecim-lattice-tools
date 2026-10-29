//go:build !js

package wasm_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/modules/circuits"
)

func TestCircuitsHostedDependencyBoundary(t *testing.T) {
	repo, _ := filepath.Abs("../..")
	deps := command(t, repo, []string{"GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0"}, "go", "list", "-deps", "./web/src/modules/circuits/wasm")
	for _, forbidden := range []string{"fyne.io/", "shared/compute", "net/http", "module3-mnist/", "module6-eda/"} {
		if strings.Contains(deps, forbidden) {
			t.Errorf("hosted circuits links %s", forbidden)
		}
	}
}
func TestModuleProductionParityCircuits(t *testing.T) {
	if os.Getenv("FECIM_WASM_PARITY") != "1" {
		t.Skip("enable native/Node/Firefox with FECIM_WASM_PARITY=1")
	}
	repo, _ := filepath.Abs("../..")
	dir := t.TempDir()
	build := "circuits-parity"
	old := hysteresis.BuildID
	hysteresis.BuildID = build
	defer func() { hysteresis.BuildID = old }()
	scenarios := []circuits.Scenario{}
	for _, n := range []int{1, 2, 4, 8} {
		for _, coupling := range []string{"Ideal", "TierA", "TierB"} {
			for _, material := range []string{"FeCIM", "HZO", "BTO", "PZT"} {
				input := make([]int, n)
				active := make([]bool, n)
				mask := make([][]bool, n)
				for i := range input {
					input[i] = 255 - i*23
					active[i] = true
					mask[i] = make([]bool, n)
					for j := range mask[i] {
						mask[i][j] = i == j || i%2 == 0
					}
				}
				if n > 1 {
					input[1] = -128
					active[n-1] = false
				}
				vector, _ := json.Marshal(input)
				rows, _ := json.Marshal(active)
				masks, _ := json.Marshal(mask)
				commands := []moduleapi.Command{{Action: "configure", Input: map[string]string{"rows": fmt.Sprint(n), "cols": fmt.Sprint(n), "material": material, "coupling": coupling, "architecture": "2T1R", "input": string(vector), "active_rows": string(rows), "read_mask": string(masks), "write_mask": string(masks)}}, {Action: "randomize"}, {Action: "select", Input: map[string]string{"row": "0", "col": "0"}}, {Action: "write", Input: map[string]string{"target": "20"}}, {Action: "read"}, {Action: "compute"}, {Action: "undo"}, {Action: "read"}, {Action: "reset"}, {Action: "compute"}}
				scenarios = append(scenarios, circuits.Scenario{Identity: circuits.Identity(), Session: fmt.Sprintf("circuits-%d-%s-%s", n, coupling, material), Commands: commands})
			}
		}
	}
	input := filepath.Join(dir, "input.json")
	data, _ := json.Marshal(scenarios)
	if err := os.WriteFile(input, data, 0600); err != nil {
		t.Fatal(err)
	}
	flags := "-s -w -X fecim-lattice-tools/workbench/hysteresis.BuildID=" + build
	wasm := filepath.Join(dir, "circuits.wasm")
	command(t, repo, []string{"GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0"}, "go", "build", "-buildvcs=false", "-trimpath", "-ldflags="+flags, "-o", wasm, "./web/src/modules/circuits/wasm")
	want := []byte(command(t, repo, nil, "go", "run", "-ldflags="+flags, "./cmd/fecim-circuits-wasm", input))
	var results [][]string
	nativeProfiles := []map[string]any{}
	nativeResults := [][]string{}
	for _, scenario := range scenarios {
		started := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		out, err := circuits.Replay(ctx, scenario)
		cancel()
		if err != nil {
			t.Fatal(err)
		}
		var memory runtime.MemStats
		runtime.ReadMemStats(&memory)
		if time.Since(started) > 15*time.Second || memory.HeapAlloc > 128<<20 {
			t.Fatal("native per-profile CPU/observed Go-heap cap exceeded")
		}
		nativeProfiles = append(nativeProfiles, map[string]any{"session": scenario.Session, "ms": time.Since(started).Milliseconds(), "observedGoHeapBytes": memory.HeapAlloc})
		nativeResults = append(nativeResults, out)
	}
	nativeBytes, _ := json.Marshal(nativeResults)
	if !bytes.Equal(want, nativeBytes) {
		t.Fatal("native CLI and measured native protocol bytes differ")
	}
	receipt, _ := json.Marshal(nativeProfiles)
	t.Log("native per-profile resources (observed Go heap, not peak RSS): " + string(receipt))
	if err := json.Unmarshal(want, &results); err != nil || len(results) != len(scenarios) {
		t.Fatal("native scenarios missing", err)
	}
	for _, scenario := range results {
		if len(scenario) != 5 {
			t.Fatalf("expected WRITE READ COMPUTE undo READ reset COMPUTE: %d results", len(scenario))
		}
		for _, raw := range scenario {
			var r circuits.Result
			if err := json.Unmarshal([]byte(raw), &r); err != nil || len(r.Device.CellCurrents) != r.Config.Rows {
				t.Fatal("missing actual device results", err)
			}
		}
	}
	goroot := strings.TrimSpace(command(t, repo, nil, "go", "env", "GOROOT"))
	t.Log(command(t, repo, nil, "node", "validation/wasm/circuits_harness.mjs", wasm, filepath.Join(goroot, "lib/wasm/wasm_exec.js"), input, dir))
	for _, runtime := range []string{"node", "firefox"} {
		got, err := os.ReadFile(filepath.Join(dir, runtime+"-circuits.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(want, got) {
			i := 0
			for i < len(want) && i < len(got) && want[i] == got[i] {
				i++
			}
			start := i - 100
			if start < 0 {
				start = 0
			}
			end := i + 100
			if end > len(want) {
				end = len(want)
			}
			t.Fatalf("native/%s original scientific bytes differ: %s near %q versus %q", runtime, firstDifference(want, got), want[start:end], got[start:end])
		}
	}
}
