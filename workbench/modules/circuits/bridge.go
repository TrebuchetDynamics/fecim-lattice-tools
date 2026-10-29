package circuits

import (
	"bytes"
	"context"
	"encoding/json"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	"fmt"
	"io"
	"runtime"
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
	Resources   RuntimeResources   `json:"resources"`
}

// RuntimeResources are measured protocol evidence, excluded from scientific bytes.
// HeapAlloc is Go live allocated heap at a response, not process RSS or peak memory.
type RuntimeResources struct {
	HeapAllocBytes  uint64 `json:"heap_alloc_bytes"`
	GoReservedBytes uint64 `json:"go_reserved_bytes"`
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

func NewBridge() *Bridge { return &Bridge{port: New(8)} }
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
		return finish(fmt.Errorf("circuits request exceeds 2 MiB"))
	}
	var req Request
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return finish(fmt.Errorf("invalid circuits request"))
	}
	if decoder.Decode(new(any)) != io.EOF {
		return finish(fmt.Errorf("trailing circuits request"))
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
		if command.Action == "write" || command.Action == "read" || command.Action == "compute" {
			return finish(fmt.Errorf("scientific operations require cancellable start"))
		}
		snapshot, err = b.port.Apply(ctx, command)
	case "start":
		var job moduleapi.Job
		job, err = b.port.Start(ctx, command)
		if err == nil {
			b.job = job.ID
		}
	case "cancel":
		_, err = b.port.Cancel(ctx, b.job)
	default:
		err = fmt.Errorf("unknown circuits operation")
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
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	reply.Result = BridgeResult{Snapshot: snapshot, Job: job, ResultsJSON: b.port.Canonical(), Resources: RuntimeResources{memory.HeapAlloc, memory.Sys}}
	return finish(nil)
}
