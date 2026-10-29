package crossbar

import (
	"bytes"
	"context"
	"encoding/json"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	"fmt"
	"io"
	"sync"
)

func Identity() hysteresis.Identity {
	id := hysteresis.BridgeIdentity()
	id.Protocol = moduleapi.SchemaVersion
	id.Model = ModelVersion
	return id
}

type BridgeInput struct {
	Command moduleapi.Command `json:"command"`
	Job     string            `json:"job,omitempty"`
}
type Request struct {
	Identity  hysteresis.Identity `json:"identity"`
	Session   string              `json:"session"`
	Request   uint64              `json:"request"`
	Revision  uint64              `json:"revision"`
	Operation string              `json:"operation"`
	Input     BridgeInput         `json:"input"`
}
type BridgeResult struct {
	Snapshot    moduleapi.Snapshot `json:"snapshot"`
	Job         moduleapi.Job      `json:"job"`
	ResultsJSON string             `json:"results_json,omitempty"`
}
type Reply struct {
	Identity  hysteresis.Identity `json:"identity"`
	Session   string              `json:"session"`
	Request   uint64              `json:"request"`
	Revision  uint64              `json:"revision"`
	Operation string              `json:"operation"`
	Committed bool                `json:"committed"`
	Error     string              `json:"error,omitempty"`
	Result    BridgeResult        `json:"result"`
}
type Bridge struct {
	mu                 sync.Mutex
	port               *Port
	session            string
	sequence, revision uint64
	job                string
	closed             bool
}

func NewBridge() *Bridge { return &Bridge{port: New(32)} }
func (b *Bridge) Close() { b.mu.Lock(); defer b.mu.Unlock(); b.closed = true; b.port.Close() }
func (b *Bridge) Dispatch(data []byte) []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	reply := Reply{Identity: Identity(), Revision: b.revision}
	finish := func(err error) []byte {
		if err != nil {
			reply.Error = err.Error()
		}
		out, _ := json.Marshal(reply)
		return out
	}
	if len(data) > 2<<20 {
		return finish(fmt.Errorf("crossbar request exceeds 2 MiB"))
	}
	var req Request
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return finish(fmt.Errorf("invalid crossbar request"))
	}
	if decoder.Decode(new(any)) != io.EOF {
		return finish(fmt.Errorf("trailing crossbar request"))
	}
	reply.Session, reply.Request, reply.Operation = req.Session, req.Request, req.Operation
	if b.closed || req.Identity != Identity() || req.Session == "" || len(req.Session) > 128 || b.session != "" && b.session != req.Session || req.Request <= b.sequence || req.Revision != b.revision {
		return finish(fmt.Errorf("identity, session, sequence or revision mismatch"))
	}
	b.session, b.sequence = req.Session, req.Request
	ctx := context.Background()
	snapshot, err := b.port.Snapshot(ctx)
	if err != nil {
		return finish(err)
	}
	command := req.Input.Command
	command.Revision = snapshot.Revision
	switch req.Operation {
	case "snapshot", "state":
	case "action":
		snapshot, err = b.port.Apply(ctx, command)
	case "start":
		command.Action = "run"
		var job moduleapi.Job
		job, err = b.port.Start(ctx, command)
		if err == nil {
			b.job = job.ID
		}
	case "cancel":
		_, err = b.port.Cancel(ctx, b.job)
	default:
		err = fmt.Errorf("unknown crossbar operation")
	}
	if err != nil {
		return finish(err)
	}
	snapshot, err = b.port.Snapshot(ctx)
	if err != nil {
		return finish(err)
	}
	job := moduleapi.Job{State: "idle", ResultsState: "unavailable"}
	if b.job != "" {
		job, err = b.port.Job(ctx, b.job)
		if err != nil {
			return finish(err)
		}
	}
	b.revision++
	reply.Revision = b.revision
	reply.Committed = true
	reply.Result = BridgeResult{Snapshot: snapshot, Job: job, ResultsJSON: b.port.Canonical()}
	return finish(nil)
}
