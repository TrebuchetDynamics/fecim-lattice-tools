package hysteresis_test

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/hysteresis"
)

func TestHysteresisBridgeTerminalWorkLimit(t *testing.T) {
	old := hysteresis.BuildID
	hysteresis.BuildID = "test-pinned"
	defer func() { hysteresis.BuildID = old }()
	b := hysteresis.NewBridge()
	revision := uint64(0)
	id := uint64(0)
	send := func(op string, input any) hysteresis.Reply {
		id++
		payload, _ := json.Marshal(input)
		data, _ := json.Marshal(hysteresis.Envelope{Identity: hysteresis.BridgeIdentity(), Session: "s", Request: id, Revision: revision, Operation: op, Input: payload})
		var r hysteresis.Reply
		if err := json.Unmarshal(b.Dispatch(data), &r); err != nil {
			t.Fatal(err)
		}
		revision = r.Revision
		return r
	}
	c := hysteresis.DefaultConfig()
	c.Mode = hysteresis.Manual
	p := hysteresis.DefaultProgramConfig()
	p.WorkLimit = 1
	for _, r := range []hysteresis.Reply{send("session.new", c), send("program.new", p), send("program.apply", hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 22}), send("program.advance", map[string]any{"Steps": 1, "DT": 0.01})} {
		if r.Error != "" {
			t.Fatal(r.Error)
		}
	}
	before := revision
	r := send("program.advance", map[string]any{"Steps": 1, "DT": 0.01})
	data, _ := json.Marshal(r.Result)
	var terminal hysteresis.ProgrammingSnapshot
	json.Unmarshal(data, &terminal)
	if r.Error == "" || r.Revision != before+1 || terminal.Status != hysteresis.ProgramWorkLimit || terminal.OwnsDrive || terminal.Complete {
		t.Fatalf("terminal mutation hidden or mislabeled: error=%q revision=%d status=%s owns=%v complete=%v", r.Error, r.Revision, terminal.Status, terminal.OwnsDrive, terminal.Complete)
	}
}

// The public byte seam must reject malformed/adversarial input before changing
// a valid live solver, its revision, or the request replay barrier.
func TestHysteresisBridgeRejectsBeforeMutation(t *testing.T) {
	old := hysteresis.BuildID
	hysteresis.BuildID = "test-pinned"
	defer func() { hysteresis.BuildID = old }()
	b := hysteresis.NewBridge()
	makeRequest := func(op string, input any) []byte {
		q := hysteresis.Envelope{Identity: hysteresis.BridgeIdentity(), Session: "s", Request: 2, Revision: 1, Operation: op}
		q.Input, _ = json.Marshal(input)
		data, _ := json.Marshal(q)
		return data
	}
	first := makeRequest("session.new", hysteresis.Config{})
	first = bytes.Replace(first, []byte(`"request":2,"revision":1`), []byte(`"request":1,"revision":0`), 1)
	var r hysteresis.Reply
	for _, bad := range [][]byte{bytes.Replace(first, []byte(`,"revision":0`), nil, 1), bytes.Replace(first, []byte(`"revision":0`), []byte(`"revision":null`), 1)} {
		if err := json.Unmarshal(b.Dispatch(bad), &r); err != nil || r.Error == "" || r.Revision != 0 {
			t.Fatalf("initial revision must be explicit integer zero: %+v %v", r, err)
		}
	}
	r = hysteresis.Reply{}
	if err := json.Unmarshal(b.Dispatch(first), &r); err != nil || r.Error != "" || r.Revision != 1 {
		t.Fatalf("new session: %+v %v", r, err)
	}
	valid := makeRequest("core.apply", hysteresis.Command{Kind: hysteresis.Run})
	c := hysteresis.DefaultConfig()
	c.HistoryLimit = hysteresis.MaxSamples + 1
	unsafeSeed := hysteresis.DefaultConfig()
	unsafeSeed.Seed = 9007199254740992
	cases := map[string][]byte{
		"oversize":          bytes.Repeat([]byte(" "), hysteresis.MaxEnvelopeBytes+1),
		"version":           bytes.Replace(valid, []byte("fecim-hysteresis/1"), []byte("fecim-hysteresis/0"), 1),
		"build":             bytes.Replace(valid, []byte("test-pinned"), []byte("another-build"), 1),
		"session":           bytes.Replace(valid, []byte(`"session":"s"`), []byte(`"session":"other"`), 1),
		"replay":            bytes.Replace(valid, []byte(`"request":2`), []byte(`"request":1`), 1),
		"stale":             bytes.Replace(valid, []byte(`"revision":1`), []byte(`"revision":0`), 1),
		"unsafe-integer":    bytes.Replace(valid, []byte(`"request":2`), []byte(`"request":9007199254740992`), 1),
		"unsupported":       makeRequest("science.fake", nil),
		"unsafe-seed":       makeRequest("core.apply", hysteresis.Command{Kind: hysteresis.Configure, Config: unsafeSeed}),
		"allocation-cap":    makeRequest("core.apply", hysteresis.Command{Kind: hysteresis.Configure, Config: c}),
		"work-cap":          makeRequest("core.advance", map[string]any{"Steps": 257, "DT": 0.01}),
		"nonfinite":         makeRequest("core.advance", map[string]any{"Steps": 1, "DT": "NaN"}),
		"infinity":          bytes.Replace(makeRequest("core.advance", map[string]any{"Steps": 1, "DT": 1}), []byte(`"DT":1`), []byte(`"DT":1e999`), 1),
		"unknown":           append(append([]byte{}, valid[:len(valid)-1]...), []byte(`,"surprise":1}`)...),
		"duplicate":         bytes.Replace(valid, []byte(`"revision":1`), []byte(`"revision":0,"revision":1`), 1),
		"duplicate-payload": makeRequest("core.apply", json.RawMessage(`{"Kind":"reset","kind":"run"}`)),
		"trailing":          append(append([]byte{}, valid...), []byte(` {}`)...),
	}
	t.Run("reference-preallocation", func(t *testing.T) {
		data := makeRequest("analysis.new", json.RawMessage(`{"Kind":"pund","References":[`+strings.Repeat("null,", 199999)+`null]}`))
		if len(data) > hysteresis.MaxEnvelopeBytes {
			t.Fatal("fixture must fit envelope cap")
		}
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		reply := b.Dispatch(data)
		runtime.ReadMemStats(&after)
		var got hysteresis.Reply
		if err := json.Unmarshal(reply, &got); err != nil || !strings.Contains(got.Error, "32 reference") || got.Revision != 1 {
			t.Fatalf("collection cap did not reject before mutation: %s %v", reply, err)
		}
		if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 8<<20 {
			t.Fatalf("known reference cap rejected only after allocating %d bytes", allocated)
		}
	})
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			var got hysteresis.Reply
			if err := json.Unmarshal(b.Dispatch(data), &got); err != nil || got.Error == "" || got.Revision != 1 {
				t.Fatalf("must reject without state mutation: %+v %v", got, err)
			}
		})
	}
	// Same request ID is still usable after all rejected traffic. The solver must
	// still have zero samples and its initial paused state before the accepted Run.
	var snapshot struct {
		Error  string              `json:"error"`
		Result hysteresis.Snapshot `json:"result"`
	}
	if err := json.Unmarshal(b.Dispatch(makeRequest("core.snapshot", nil)), &snapshot); err != nil || snapshot.Error != "" || snapshot.Result.Count != 0 || snapshot.Result.State != "paused" {
		t.Fatalf("rejections changed session: %+v %v", snapshot, err)
	}
}
