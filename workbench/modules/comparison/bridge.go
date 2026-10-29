package comparison

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	model "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	"io"
	"sync"
)

func Identity() hysteresis.Identity {
	id := hysteresis.BridgeIdentity()
	id.Protocol = moduleapi.SchemaVersion
	id.Model = model.ScenarioModelVersion
	return id
}

type Request struct {
	Identity  hysteresis.Identity `json:"identity"`
	Session   string              `json:"session"`
	Request   uint64              `json:"request"`
	Revision  uint64              `json:"revision"`
	Operation string              `json:"operation"`
	Input     moduleapi.Command   `json:"input"`
}
type Reply struct {
	Identity    hysteresis.Identity `json:"identity"`
	Session     string              `json:"session"`
	Request     uint64              `json:"request"`
	Revision    uint64              `json:"revision"`
	Operation   string              `json:"operation"`
	Committed   bool                `json:"committed"`
	Error       string              `json:"error,omitempty"`
	Result      moduleapi.Snapshot  `json:"result"`
	ResultsJSON string              `json:"results_json,omitempty"`
}
type Bridge struct {
	mu                 sync.Mutex
	port               *Port
	revision, sequence uint64
	session            string
}

func NewBridge() *Bridge { return &Bridge{port: New()} }
func (b *Bridge) Close() { b.mu.Lock(); defer b.mu.Unlock(); b.port.Close() }
func (b *Bridge) Dispatch(data []byte) []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	r := Reply{Identity: Identity(), Revision: b.revision}
	finish := func(err error) []byte {
		if err != nil {
			r.Error = err.Error()
		}
		out, e := json.Marshal(r)
		if e != nil {
			panic(e)
		}
		return out
	}
	if len(data) > 16<<10 {
		return finish(errors.New("comparison request exceeds 16 KiB budget"))
	}
	var q Request
	check := json.NewDecoder(bytes.NewReader(data))
	if err := uniqueEnvelope(check, 0); err != nil {
		return finish(err)
	}
	var explicit struct {
		Revision *uint64 `json:"revision"`
	}
	if err := json.Unmarshal(data, &explicit); err != nil || explicit.Revision == nil {
		return finish(errors.New("revision must be explicit"))
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&q); err != nil {
		return finish(errors.New("invalid comparison request"))
	}
	if d.Decode(new(any)) != io.EOF {
		return finish(errors.New("trailing comparison request"))
	}
	r.Session = q.Session
	r.Request = q.Request
	r.Operation = q.Operation
	if Identity().Build == "unqualified" || q.Request > 1<<53-1 || q.Revision >= 1<<53-1 || q.Identity != Identity() || len(q.Session) == 0 || len(q.Session) > 128 || b.session != "" && b.session != q.Session || q.Request <= b.sequence || q.Revision != b.revision {
		return finish(errors.New("identity, session, sequence or revision mismatch"))
	}
	b.session = q.Session
	b.sequence = q.Request
	var err error
	switch q.Operation {
	case "snapshot":
		if q.Input.Action != "" || len(q.Input.Input) != 0 || len(q.Input.Resources) != 0 || len(q.Input.Payload) != 0 {
			return finish(errors.New("snapshot accepts no command input"))
		}
		r.Result, err = b.port.Snapshot(context.Background())
	case "action":
		r.Result, err = b.port.Apply(context.Background(), q.Input)
	default:
		err = errors.New("unknown comparison operation")
	}
	if err != nil {
		return finish(err)
	}
	canonical, err := b.port.Canonical("json")
	if err != nil {
		return finish(err)
	}
	b.revision++
	r.Revision = b.revision
	r.Committed = true
	r.ResultsJSON = string(canonical)
	return finish(nil)
}
