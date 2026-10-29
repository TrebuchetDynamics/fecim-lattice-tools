//go:build !js

package wasm_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	model "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	comparison "fecim-lattice-tools/workbench/modules/comparison"
)

func TestModuleProductionParityComparison(t *testing.T) {
	if os.Getenv("FECIM_WASM_PARITY") != "1" {
		t.Skip("enable native/Node/Firefox production parity with FECIM_WASM_PARITY=1")
	}
	repo, _ := filepath.Abs("../..")
	dir := t.TempDir()
	old := hysteresis.BuildID
	hysteresis.BuildID = "comparison-production-parity"
	defer func() { hysteresis.BuildID = old }()
	bridge := comparison.NewBridge()
	defer bridge.Close()
	requests := []json.RawMessage{}
	want := [][]byte{}
	wantReplies := [][]byte{}
	revision := uint64(0)
	portRevision := uint64(0)
	send := func(command moduleapi.Command) {
		command.Revision = portRevision
		q := comparison.Request{Identity: comparison.Identity(), Session: "comparison-parity", Request: uint64(len(requests) + 1), Revision: revision, Operation: "action", Input: command}
		data, _ := json.Marshal(q)
		requests = append(requests, data)
		var reply comparison.Reply
		originalReply := bridge.Dispatch(data)
		wantReplies = append(wantReplies, originalReply)
		if err := json.Unmarshal(originalReply, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Error != "" || !reply.Committed || reply.ResultsJSON == "" {
			t.Fatalf("unsuccessful native science: %+v", reply)
		}
		var artifact struct {
			Schema string                 `json:"schema_version"`
			Result model.ScenarioEstimate `json:"result"`
		}
		if err := json.Unmarshal([]byte(reply.ResultsJSON), &artifact); err != nil {
			t.Fatal(err)
		}
		if artifact.Schema != model.ScenarioArtifactVersion || artifact.Result.Energy[0].DailyInferences != 172800 && command.Action != "reset" || artifact.Result.Energy[0].DailyKWh <= 0 || artifact.Result.Run.Outputs["co2_reduction_pct"].UnavailableReason == "" {
			t.Fatal("identical empty, unavailable or unit-invalid science cannot qualify")
		}
		if command.Action == "diff" || command.Action == "sensitivity" {
			found := false
			for _, table := range reply.Result.Tables {
				if table.ID == command.Action {
					found = true
					if command.Action == "diff" && len(table.Rows) != 4 {
						t.Fatal("missing available diff results")
					}
					if command.Action == "sensitivity" && command.Input["output"] == "energy_reduction_pct" && len(table.Rows) != 12 {
						t.Fatal("missing actual perturbations")
					}
				}
			}
			if !found {
				t.Fatal("analysis command returned no analysis")
			}
		}
		revision = reply.Revision
		portRevision = reply.Result.Revision
		want = append(want, []byte(reply.ResultsJSON))
	}
	for _, workload := range model.ScenarioWorkloads() {
		for _, profile := range []string{"conservative", "baseline", "optimistic"} {
			send(moduleapi.Command{Action: "calculate", Input: map[string]string{"workload": workload, "profile": profile, "rate": "2"}})
		}
	}
	send(moduleapi.Command{Action: "diff"})
	send(moduleapi.Command{Action: "sensitivity", Input: map[string]string{"output": "energy_reduction_pct"}})
	send(moduleapi.Command{Action: "sensitivity", Input: map[string]string{"output": "co2_reduction_pct"}})
	send(moduleapi.Command{Action: "reset"})
	data, _ := json.Marshal(requests)
	input := filepath.Join(dir, "requests.json")
	if err := os.WriteFile(input, data, 0600); err != nil {
		t.Fatal(err)
	}
	flags := "-s -w -X fecim-lattice-tools/workbench/hysteresis.BuildID=comparison-production-parity"
	native := command(t, repo, nil, "go", "run", "-ldflags="+flags, "./cmd/fecim-comparison-wasm", input)
	var nativeReplies []json.RawMessage
	if err := json.Unmarshal([]byte(native), &nativeReplies); err != nil {
		t.Fatal(err)
	}
	if len(nativeReplies) != len(want) {
		t.Fatal("native runner lost operations")
	}
	for i, r := range nativeReplies {
		if !bytes.Equal(wantReplies[i], r) {
			t.Fatalf("native runner differs at %d", i)
		}
	}
	wasm := filepath.Join(dir, "comparison.wasm")
	command(t, repo, []string{"GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0"}, "go", "build", "-buildvcs=false", "-trimpath", "-ldflags="+flags, "-o", wasm, "./web/src/modules/comparison/wasm")
	deps := command(t, repo, []string{"GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0"}, "go", "list", "-deps", "./web/src/modules/comparison/wasm")
	for _, forbidden := range []string{"fyne.io/", "net/http", "module6-eda/", "module3-mnist/", "workbench/moduleapi/native"} {
		if strings.Contains(deps, forbidden) {
			t.Fatalf("hosted comparison links %s", forbidden)
		}
	}
	goroot := strings.TrimSpace(command(t, repo, nil, "go", "env", "GOROOT"))
	t.Log(command(t, repo, nil, "node", "validation/wasm/comparison_harness.mjs", wasm, filepath.Join(goroot, "lib/wasm/wasm_exec.js"), input, dir))
	for _, runtime := range []string{"node", "firefox"} {
		for i, expected := range want {
			got, err := os.ReadFile(filepath.Join(dir, runtime+"-"+strconv.Itoa(i)+".json"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(expected, got) {
				t.Fatalf("native/%s operation %d canonical bytes differ: %s", runtime, i, firstDifference(expected, got))
			}
			reply, err := os.ReadFile(filepath.Join(dir, runtime+"-reply-"+strconv.Itoa(i)+".json"))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(wantReplies[i], reply) {
				t.Fatalf("native/%s operation %d original numeric-table reply differs: %s", runtime, i, firstDifference(wantReplies[i], reply))
			}
		}
	}
}
