package hysteresis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"runtime"
	"strings"
	"sync"
)

// BuildID is supplied by the asset builder (-X) from source, dependency and
// matching wasm_exec.js bytes. Unpinned binaries are deliberately not admitted.
var BuildID = "unqualified"

const MaxEnvelopeBytes = 1 << 20
const MaxBridgeSteps = MaxSteps // core/program outer steps; analysis yields per unit
const maxJSInteger = 1<<53 - 1

type Identity struct {
	Protocol string `json:"protocol"`
	Model    string `json:"model"`
	Build    string `json:"build"`
	Go       string `json:"go"`
}

func BridgeIdentity() Identity {
	return Identity{"fecim-hysteresis/1", "hysteresis-producers/1", BuildID, runtime.Version()}
}

type Operation struct {
	Name    string `json:"name"`
	Mutates bool   `json:"mutates"`
}

// Operations is the sole registry. Browsers render these capabilities; they do
// not invent scientific defaults or maintain a second operation list.
func Operations() []Operation {
	return []Operation{
		{"session.new", true}, {"core.snapshot", false}, {"core.apply", true}, {"core.advance", true}, {"core.page", false},
		{"program.new", true}, {"program.apply", true}, {"program.advance", true}, {"program.snapshot", false},
		{"analysis.new", true}, {"analysis.advance", true}, {"analysis.page", false}, {"analysis.result", false}, {"job.cancel", true},
	}
}

type Envelope struct {
	Identity  Identity        `json:"identity"`
	Session   string          `json:"session"`
	Request   uint64          `json:"request"`
	Revision  uint64          `json:"revision"`
	Operation string          `json:"operation"`
	Input     json.RawMessage `json:"input"`
}
type Reply struct {
	Identity  Identity `json:"identity"`
	Session   string   `json:"session"`
	Request   uint64   `json:"request"`
	Revision  uint64   `json:"revision"`
	Operation string   `json:"operation"`
	Result    any      `json:"result,omitempty"`
	Error     string   `json:"error,omitempty"`
	Committed bool     `json:"committed"`
}
type Bridge struct {
	mu                    sync.Mutex
	session               *Session
	programmer            *Programmer
	analysis              *AnalysisJob
	analysisRunning       bool
	id                    string
	revision, lastRequest uint64
}

func NewBridge() *Bridge { return &Bridge{} }

// encoding/json otherwise accepts duplicate (including case-folded) fields,
// allowing ambiguous commands. Bound nesting before allocating typed inputs.
func uniqueJSON(d *json.Decoder, depth, arrayLimit int) error {
	if depth > 64 {
		return errors.New("JSON nesting exceeds 64")
	}
	token, err := d.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delimiter == '{' {
		keys := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok {
				return errors.New("invalid JSON key")
			}
			name = strings.ToLower(name)
			if keys[name] {
				return errors.New("duplicate JSON field")
			}
			keys[name] = true
			limit := -1
			if name == "references" {
				limit = 32
			} // NewAnalysis branch cap, before typed slice growth
			if err := uniqueJSON(d, depth+1, limit); err != nil {
				return err
			}
		}
	} else if delimiter == '[' {
		count := 0
		for d.More() {
			if arrayLimit >= 0 && count >= arrayLimit {
				return errors.New("at most 32 reference branches (preallocation cap)")
			}
			count++
			if err := uniqueJSON(d, depth+1, -1); err != nil {
				return err
			}
		}
	} else {
		return errors.New("unexpected JSON delimiter")
	}
	_, err = d.Token()
	return err
}
func decode(data []byte, into any) error {
	check := json.NewDecoder(bytes.NewReader(data))
	check.UseNumber()
	if err := uniqueJSON(check, 0, -1); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(into); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("expected exactly one JSON value")
	}
	return nil
}

// Dispatch checks the byte cap before decoding any payload, then checks build,
// session, monotonic request and transport revision before producer allocation.
// Replies contain Go's original scientific serialization, never JS numbers.
func (b *Bridge) Dispatch(data []byte) []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	r := Reply{Identity: BridgeIdentity(), Revision: b.revision}
	var q Envelope
	fail := func(err error) []byte { r.Error = err.Error(); out, _ := json.Marshal(r); return out }
	if len(data) > MaxEnvelopeBytes {
		return fail(errors.New("request envelope exceeds 1 MiB"))
	}
	wire := struct {
		*Envelope
		Revision *uint64 `json:"revision"`
	}{Envelope: &q}
	if err := decode(data, &wire); err != nil {
		return fail(err)
	}
	if wire.Revision == nil {
		return fail(errors.New("revision must be an explicit integer"))
	}
	q.Revision = *wire.Revision
	r.Session = q.Session
	r.Request = q.Request
	r.Operation = q.Operation
	if BuildID == "unqualified" || q.Identity != BridgeIdentity() {
		return fail(errors.New("protocol/model/build/runtime identity mismatch"))
	}
	if len(q.Session) < 1 || len(q.Session) > 128 || q.Request == 0 || q.Request > maxJSInteger || q.Request <= b.lastRequest || q.Revision != b.revision || q.Revision >= maxJSInteger {
		return fail(errors.New("invalid session/request or stale revision"))
	}
	if b.session != nil && q.Session != b.id {
		return fail(errors.New("session mismatch"))
	}
	known := false
	for _, op := range Operations() {
		if op.Name == q.Operation {
			known = true
			break
		}
	}
	if !known {
		return fail(errors.New("unsupported operation"))
	}
	if b.session == nil && q.Operation != "session.new" {
		return fail(errors.New("session.new required"))
	}
	if q.Operation == "session.new" && b.session != nil {
		return fail(errors.New("terminate and recreate Worker to replace session"))
	}
	var beforeCore, beforeAnalysis uint64
	if b.session != nil {
		beforeCore = b.session.revision
	}
	if b.analysis != nil {
		beforeAnalysis = b.analysis.result.Revision
	}
	result, err := b.dispatch(q)
	if err != nil {
		changed := (b.session != nil && b.session.revision != beforeCore) || (b.analysis != nil && b.analysis.result.Revision != beforeAnalysis)
		if changed {
			b.lastRequest = q.Request
			b.revision++
			r.Revision = b.revision
			r.Result = result
			r.Committed = true
		}
		return fail(err)
	}
	r.Committed = true
	b.id = q.Session
	b.lastRequest = q.Request
	b.revision++
	r.Revision = b.revision
	r.Result = result
	out, err := json.Marshal(r)
	if err != nil {
		return fail(err)
	}
	return out
}
func (b *Bridge) dispatch(q Envelope) (any, error) {
	bad := func(s string) (any, error) { return nil, errors.New(s) }
	input := q.Input
	if len(input) == 0 {
		return bad("input required (null for no payload)")
	}
	empty := func() error {
		if !bytes.Equal(input, []byte("null")) {
			return errors.New("unexpected operation payload")
		}
		return nil
	}
	var command Command
	var programCommand ProgramCommand
	var budget struct {
		Steps int
		DT    float64
	}
	switch q.Operation {
	case "core.apply":
		if err := decode(input, &command); err != nil {
			return nil, err
		}
		if command.Config.Seed > maxJSInteger {
			return bad("seed exceeds exact JSON integer range")
		}
	case "program.apply":
		if err := decode(input, &programCommand); err != nil {
			return nil, err
		}
	case "core.advance", "program.advance":
		if err := decode(input, &budget); err != nil {
			return nil, err
		}
		if budget.Steps < 1 || budget.Steps > MaxBridgeSteps || !finite(budget.DT) || budget.DT <= 0 || budget.DT > 1 {
			return bad("Worker budget requires 1–256 steps and finite dt in (0,1]")
		}
	}
	if b.analysisRunning && (q.Operation == "core.apply" || q.Operation == "core.advance" || q.Operation == "program.new" || q.Operation == "program.apply" || q.Operation == "program.advance" || q.Operation == "analysis.new") {
		return bad("analysis owns active job; cancel first")
	}
	switch q.Operation {
	case "session.new":
		var c Config
		if err := decode(input, &c); err != nil {
			return nil, err
		}
		if c.Seed > maxJSInteger {
			return bad("seed exceeds exact JSON integer range")
		}
		s, err := NewSession(c)
		if err != nil {
			return nil, err
		}
		b.session = s
		return struct {
			Core                              Snapshot
			Operations                        []Operation
			MaxEnvelopeBytes, MaxAdvanceSteps int
		}{s.Snapshot(), Operations(), MaxEnvelopeBytes, MaxBridgeSteps}, nil
	case "core.snapshot":
		if err := empty(); err != nil {
			return nil, err
		}
		return b.session.Snapshot(), nil
	case "core.apply":
		return b.session.Apply(command)
	case "core.advance":
		return b.session.Advance(Budget{Steps: budget.Steps, DT: budget.DT})
	case "core.page":
		var p struct{ Offset int }
		if err := decode(input, &p); err != nil {
			return nil, err
		}
		return b.session.Page(PageQuery{Revision: b.session.Snapshot().Revision, Offset: p.Offset})
	case "program.new":
		if b.programmer != nil && b.programmer.Snapshot().OwnsDrive {
			return bad("programming owns active job; cancel first")
		}
		var c ProgramConfig
		if err := decode(input, &c); err != nil {
			return nil, err
		}
		p, err := NewProgrammer(b.session, c)
		if err != nil {
			return nil, err
		}
		b.programmer = p
		return p.Snapshot(), nil
	case "program.apply", "program.advance", "program.snapshot":
		if b.programmer == nil {
			return bad("program.new required")
		}
		if q.Operation == "program.apply" {
			return b.programmer.Apply(programCommand)
		}
		if q.Operation == "program.advance" {
			return b.programmer.Advance(Budget{Steps: budget.Steps, DT: budget.DT})
		}
		if err := empty(); err != nil {
			return nil, err
		}
		return b.programmer.Snapshot(), nil
	case "analysis.new":
		if b.session.Snapshot().State == "running" || (b.programmer != nil && b.programmer.Snapshot().OwnsDrive) {
			return bad("pause/cancel current computation before analysis")
		}
		var p struct {
			Kind       AnalysisKind
			FORC       FORCParameters
			References []ReferenceInput
		}
		if err := decode(input, &p); err != nil {
			return nil, err
		}
		j, err := NewAnalysis(AnalysisRequest{Input: b.session.Snapshot(), Kind: p.Kind, FORC: p.FORC, References: p.References})
		if err != nil {
			return nil, err
		}
		b.analysis = j
		b.analysisRunning = true
		return j.progress(), nil
	case "analysis.advance":
		var p struct{ Steps int }
		if err := decode(input, &p); err != nil {
			return nil, err
		}
		if p.Steps != 1 {
			return bad("Worker analysis budget requires one work unit")
		}
		if b.analysis == nil {
			return bad("analysis.new required")
		}
		v, err := b.analysis.Advance(AnalysisBudget{Steps: 1})
		b.analysisRunning = v.State == "running"
		return v, err
	case "analysis.page":
		if b.analysis == nil {
			return bad("analysis.new required")
		}
		var p struct {
			Table  AnalysisTable
			Offset int
		}
		if err := decode(input, &p); err != nil {
			return nil, err
		}
		return b.analysis.Page(AnalysisPageQuery{Revision: b.analysis.progress().Revision, Table: p.Table, Offset: p.Offset})
	case "analysis.result":
		if err := empty(); err != nil {
			return nil, err
		}
		if b.analysis == nil {
			return bad("analysis.new required")
		}
		return b.analysis.Result()
	case "job.cancel":
		if err := empty(); err != nil {
			return nil, err
		}
		if b.analysisRunning {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			v, _ := b.analysis.Advance(AnalysisBudget{Steps: 1, Context: ctx})
			b.analysisRunning = false
			return v, nil
		}
		if b.programmer != nil && b.programmer.Snapshot().OwnsDrive {
			return b.programmer.Apply(ProgramCommand{Kind: CancelProgram})
		}
		return b.session.Apply(Command{Kind: Cancel})
	}
	return bad("unsupported operation")
}
