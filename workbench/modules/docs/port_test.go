package docs

import (
	"context"
	"encoding/json"
	"testing"

	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
)

func TestDocsWebPort(t *testing.T) {
	p, err := NewPort()
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	s, err := p.Apply(context.Background(), moduleapi.Command{Action: "open", Input: map[string]string{"id": "glossary"}})
	if err != nil {
		t.Fatal(err)
	}
	if s.Descriptor.ID != "docs" || !s.Descriptor.Available || s.Revision != 1 {
		t.Fatalf("invalid port snapshot %+v", s)
	}
	if _, err = p.Apply(context.Background(), moduleapi.Command{Action: "search", Revision: 0}); err == nil {
		t.Fatal("stale command accepted")
	}
	b := NewBridge()
	wire, _ := json.Marshal(hysteresis.Envelope{Identity: Identity(), Session: "docs-test", Request: 1, Operation: "search", Input: json.RawMessage(`{"query":"zzzznotincorpus","page":"1"}`)})
	var reply hysteresis.Reply
	if err = json.Unmarshal(b.Dispatch(wire), &reply); err != nil || reply.Error != "" || !reply.Committed {
		t.Fatalf("worker reply %s %v", b.Dispatch(nil), err)
	}
	p.Close()
	if _, err = p.Output(); err == nil {
		t.Fatal("closed reader output available")
	}
	data, _ := json.Marshal(reply.Result)
	var out Output
	if err = json.Unmarshal(data, &out); err != nil || out.Page.Total != 0 || out.DocumentCount != 13 {
		t.Fatalf("real Go output: %s %v", data, err)
	}
}
