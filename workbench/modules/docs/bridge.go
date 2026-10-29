package docs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
)

func Identity() hysteresis.Identity {
	id := hysteresis.BridgeIdentity()
	id.Protocol = moduleapi.SchemaVersion
	id.Model = "docs-reader/1"
	return id
}
func decodeStrict(data []byte, out any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return errors.New("invalid docs input")
	}
	if d.Decode(new(any)) != io.EOF {
		return errors.New("trailing docs input")
	}
	return nil
}

type Bridge struct {
	port     *Port
	session  string
	sequence uint64
	err      error
}

func NewBridge() *Bridge { p, err := NewPort(); return &Bridge{port: p, err: err} }
func (b *Bridge) Dispatch(data []byte) []byte {
	var req hysteresis.Envelope
	reply := hysteresis.Reply{Identity: Identity()}
	if b.port != nil {
		reply.Revision = b.port.revision
	}
	finish := func(err error) []byte {
		if err != nil {
			reply.Error = err.Error()
		}
		data, _ := json.Marshal(reply)
		return data
	}
	if len(data) > 128<<10 {
		return finish(errors.New("docs envelope limit exceeded"))
	}
	if err := decodeStrict(data, &req); err != nil {
		return finish(err)
	}
	reply.Session = req.Session
	reply.Request = req.Request
	reply.Operation = req.Operation
	if b.err != nil {
		return finish(b.err)
	}
	if req.Identity != Identity() || len(req.Session) == 0 || len(req.Session) > 128 || b.session != "" && b.session != req.Session || req.Request <= b.sequence || req.Revision != b.port.revision {
		return finish(errors.New("identity/session/sequence/revision mismatch"))
	}
	var input map[string]string
	if err := decodeStrict(req.Input, &input); err != nil {
		return finish(err)
	}
	if _, err := b.port.Apply(context.Background(), moduleapi.Command{Action: req.Operation, Revision: req.Revision, Input: input}); err != nil {
		return finish(err)
	}
	b.session = req.Session
	b.sequence = req.Request
	reply.Revision = b.port.revision
	reply.Result, _ = b.port.Output()
	reply.Committed = true
	return finish(nil)
}
