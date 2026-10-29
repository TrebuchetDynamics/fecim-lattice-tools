package crossbar

import (
	"encoding/json"
	"fecim-lattice-tools/workbench/moduleapi"
	"strings"
	"testing"
	"time"
)

func TestCrossbarWorkerProtocol(t *testing.T) {
	b := NewBridge()
	defer b.Close()
	req := Request{Identity: Identity(), Session: "fixture", Request: 1, Operation: "action", Input: BridgeInput{Command: moduleapi.Command{Action: "configure", Input: map[string]string{"rows": "2", "cols": "3", "profile": "fecap", "matrix": "[[0.2,0.5,0.8],[0.9,0.4,0.1]]", "input": "[0.5,1,0.25]"}}}}
	call := func(op string) Reply {
		req.Operation = op
		data, _ := json.Marshal(req)
		var reply Reply
		if err := json.Unmarshal(b.Dispatch(data), &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Error != "" {
			t.Fatal(reply.Error)
		}
		req.Request++
		req.Revision = reply.Revision
		return reply
	}
	call("action")
	call("start")
	deadline := time.Now().Add(time.Second)
	var reply Reply
	for {
		reply = call("state")
		if reply.Result.Job.State == "succeeded" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Worker science did not settle")
		}
		time.Sleep(time.Millisecond)
	}
	var scientific Result
	if err := json.Unmarshal([]byte(reply.Result.ResultsJSON), &scientific); err != nil || scientific.Unit != "C" || len(scientific.Actual) != 2 || scientific.Actual[0] <= 0 {
		t.Fatal("missing successful canonical charge result", err)
	}
	canonical := reply.Result.ResultsJSON
	req.Input.Command = moduleapi.Command{Action: "program", Input: map[string]string{"row": "1", "col": "2", "weight": "0.2"}}
	changed := call("action")
	if changed.Result.ResultsJSON != canonical || !strings.Contains(strings.Join(changed.Result.Snapshot.Evidence, " "), "STALE") {
		t.Fatal("changed inputs did not label retained results stale")
	}
	data, _ := json.Marshal(req)
	b.Dispatch(data)
	var stale Reply
	json.Unmarshal(b.Dispatch(data), &stale)
	if stale.Error == "" || stale.Committed {
		t.Fatal("replayed request accepted")
	}
}
