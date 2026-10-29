package circuits

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/workbench/moduleapi"
)

type Port struct {
	mu                                sync.Mutex
	limit                             int
	cfg                               Config
	state                             State
	undo                              *State
	result                            *Result
	revision, inputEpoch, resultEpoch uint64
	status                            string
	page                              int
	resultPage                        int
	logs                              []string
	job                               moduleapi.Job
	cancel                            context.CancelFunc
	done                              chan struct{}
	closed                            bool
	timing                            int
}

func New(limit int) *Port {
	c := DefaultConfig()
	return &Port{limit: limit, cfg: c, state: Initial(c, false), page: 1, resultPage: 1, status: "Ready; WRITE, READ or COMPUTE explicitly", job: moduleapi.Job{State: "idle", ResultsState: "unavailable"}}
}
func (p *Port) Descriptor() moduleapi.ModuleDescriptor {
	return moduleapi.ModuleDescriptor{ID: "circuits", Title: "Circuits", Available: true, SchemaVersion: moduleapi.SchemaVersion, ModelVersion: ModelVersion, Features: []string{"configure", "select", "write", "read", "compute", "undo", "reset", "randomize", "cancel", "timing", "view", "save"}, Capabilities: []moduleapi.Capability{{Profile: "hosted", Available: true, Reason: "Ephemeral CPU only; maximum 8 by 8; no upload, persistence, GPU or file export"}, {Profile: "local", Available: true, Reason: "Native size choices through 128; explicit rooted JSON export; GPU unavailable in this port"}}}
}
func (p *Port) active() bool { return p.job.State == "running" || p.job.State == "cancelling" }
func (p *Port) Snapshot(ctx context.Context) (moduleapi.Snapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return moduleapi.Snapshot{}, err
	}
	if p.closed {
		return moduleapi.Snapshot{}, errors.New("circuits session closed")
	}
	return p.snapshot(), nil
}
func (p *Port) record(s string) {
	p.status = s
	p.logs = append(p.logs, s)
	if len(p.logs) > 64 {
		p.logs = p.logs[len(p.logs)-64:]
	}
}
func (p *Port) checkpoint() { s := clone(p.state); p.undo = &s }
func (p *Port) execute(ctx context.Context, action string, target int) (Result, error) {
	return Execute(ctx, p.cfg, p.state, action, target, p.limit, Yield)
}
func (p *Port) commit(result Result) {
	if result.Operation == "write" {
		p.checkpoint()
		p.state = clone(result.State)
		p.inputEpoch++
	}
	p.result = &result
	p.resultPage = 1
	p.resultEpoch = p.inputEpoch
	p.revision++
	message := fmt.Sprintf("%s completed using %s %s", result.Operation, p.cfg.Material, p.cfg.Coupling)
	if result.Operation == "write" {
		if result.Verified {
			message = "WRITE verified; actual programmed state retained"
		} else {
			message = result.Reason
		}
	}
	p.record(message)
}
func (p *Port) Apply(ctx context.Context, c moduleapi.Command) (moduleapi.Snapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fail := func(err error) (moduleapi.Snapshot, error) { return moduleapi.Snapshot{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if p.closed || p.active() || c.Revision != p.revision {
		return fail(errors.New("closed, busy or stale circuits revision"))
	}
	switch c.Action {
	case "configure":
		cfg, s, err := configure(p.cfg, p.state, c.Input, p.limit)
		if err != nil {
			return fail(err)
		}
		// Configuration is a single transaction, never mutate live state on errors.
		p.checkpoint()
		p.cfg, p.state = cfg, s
		p.undo = nil
		p.inputEpoch++
		p.page = 1
	case "select":
		row, e1 := strconv.Atoi(c.Input["row"])
		col, e2 := strconv.Atoi(c.Input["col"])
		if e1 != nil || e2 != nil || row < 0 || row >= p.cfg.Rows || col < 0 || col >= p.cfg.Cols {
			return fail(errors.New("selected cell outside array"))
		}
		p.state.Row, p.state.Col = row, col
		p.inputEpoch++
	case "write", "read", "compute":
		target := 0
		var err error
		if c.Action == "write" {
			target, err = strconv.Atoi(c.Input["target"])
			if err != nil {
				return fail(errors.New("target level required"))
			}
		}
		result, err := p.execute(ctx, c.Action, target)
		if err != nil {
			return fail(err)
		}
		p.commit(result)
		return p.snapshot(), nil
	case "undo":
		if p.undo == nil {
			return fail(errors.New("nothing to undo"))
		}
		p.state = clone(*p.undo)
		p.undo = nil
		p.inputEpoch++
	case "reset", "randomize":
		p.checkpoint()
		p.state = Initial(p.cfg, c.Action == "randomize")
		p.result = nil
		p.inputEpoch++
		p.page = 1
	case "view":
		page, err := strconv.Atoi(c.Input["page"])
		count := p.cfg.Rows * p.cfg.Cols
		if c.Input["scope"] == "result" {
			if p.result == nil {
				return fail(errors.New("no retained result"))
			}
			count = p.result.Config.Rows * p.result.Config.Cols
		} else if c.Input["scope"] != "" && c.Input["scope"] != "cells" {
			return fail(errors.New("unknown page scope"))
		}
		if err != nil || page < 1 || page > (count+31)/32 {
			return fail(errors.New("page outside cells"))
		}
		if c.Input["scope"] == "result" {
			p.resultPage = page
		} else {
			p.page = page
		}
	case "timing":
		switch c.Input["intent"] {
		case "reset":
			p.timing = 0
		case "step":
			p.timing = (p.timing + 1) % 5
		default:
			return fail(errors.New("timing intent must be reset or step"))
		}
	case "save":
		if err := p.save(ctx, c); err != nil {
			return fail(err)
		}
	default:
		return fail(errors.New("unsupported circuits action"))
	}
	p.revision++
	p.record(c.Action + " applied; retained numerical results require explicit recomputation after input changes")
	return p.snapshot(), nil
}
func configure(c Config, s State, input map[string]string, limit int) (Config, State, error) {
	next := c
	state := clone(s)
	if preset, ok := input["sense_preset"]; ok && preset != "Custom" {
		found := false
		for _, v := range arraysim.NativeSensePresets() {
			if v.Name == preset {
				next.Gain, next.SenseMin, next.SenseMax = v.Rf*1e3, v.Vmin, v.Vmax
				found = true
				break
			}
		}
		if !found {
			return c, s, errors.New("unknown native sense preset")
		}
	}
	for k, v := range input {
		var err error
		switch k {
		case "rows":
			next.Rows, err = strconv.Atoi(v)
		case "cols":
			next.Cols, err = strconv.Atoi(v)
		case "node":
			next.Node, err = strconv.Atoi(v)
		case "dac_bits":
			next.DACBits, err = strconv.Atoi(v)
		case "adc_bits":
			next.ADCBits, err = strconv.Atoi(v)
		case "max_iterations":
			next.MaxIterations, err = strconv.Atoi(v)
		case "max_pulses":
			next.MaxPulses, err = strconv.Atoi(v)
		case "seed":
			next.Seed, err = strconv.ParseInt(v, 10, 64)
		case "gain":
			next.Gain, err = strconv.ParseFloat(v, 64)
		case "supply":
			next.Supply, err = strconv.ParseFloat(v, 64)
		case "sense_min":
			next.SenseMin, err = strconv.ParseFloat(v, 64)
		case "sense_max":
			next.SenseMax, err = strconv.ParseFloat(v, 64)
		case "material":
			next.Material = v
		case "architecture":
			next.Architecture = v
		case "coupling":
			next.Coupling = v
		case "input", "voltages", "sense_preset", "active_rows", "read_mask", "write_mask":
		default:
			return c, s, fmt.Errorf("unknown circuits setting %s", k)
		}
		if err != nil {
			return c, s, fmt.Errorf("invalid %s", k)
		}
	}
	if err := next.Validate(limit); err != nil {
		return c, s, err
	}
	if next.Rows != c.Rows || next.Cols != c.Cols || next.Material != c.Material {
		state = Initial(next, false)
	}
	if next.Architecture != "2T1R" {
		state.ReadMask, state.WriteMask = nil, nil
	}
	for _, binding := range []struct {
		key   string
		value any
	}{{"input", &state.Input}, {"voltages", &state.Voltages}, {"active_rows", &state.ActiveRows}, {"read_mask", &state.ReadMask}, {"write_mask", &state.WriteMask}} {
		if v, ok := input[binding.key]; ok {
			if len(v) > 1<<20 {
				return c, s, errors.New("input exceeds 1 MiB")
			}
			if err := json.Unmarshal([]byte(v), binding.value); err != nil {
				return c, s, fmt.Errorf("invalid %s JSON", binding.key)
			}
		}
	}
	if err := validateState(next, state); err != nil {
		return c, s, err
	}
	return next, state, nil
}
func (p *Port) Start(ctx context.Context, c moduleapi.Command) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return moduleapi.Job{}, err
	}
	if p.closed || p.active() || c.Revision != p.revision {
		return moduleapi.Job{}, errors.New("busy, closed or stale circuits start")
	}
	if c.Action != "write" && c.Action != "read" && c.Action != "compute" {
		return moduleapi.Job{}, errors.New("expected write, read or compute job")
	}
	target := 0
	var err error
	if c.Action == "write" {
		target, err = strconv.Atoi(c.Input["target"])
		if err != nil || target < 0 || target >= Material(p.cfg.Material).GetNumLevels() {
			return moduleapi.Job{}, errors.New("target outside material levels")
		}
	}
	cfg, state := p.cfg, clone(p.state)
	jobctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	p.cancel = cancel
	p.done = make(chan struct{})
	done := p.done
	p.revision++
	p.job = moduleapi.Job{ID: fmt.Sprintf("circuits-%d", p.revision), State: "running", ResultsState: "unavailable", InputRevision: strconv.FormatUint(p.revision, 10), Progress: moduleapi.Progress{Total: 1}, Message: "Circuit computation running; previous result retained"}
	if p.result != nil {
		p.job.ResultsState = "retained"
	}
	go func() {
		defer close(done)
		defer cancel()
		result, err := Execute(jobctx, cfg, state, c.Action, target, p.limit, Yield)
		p.mu.Lock()
		defer p.mu.Unlock()
		if err == nil {
			err = jobctx.Err()
		}
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				p.job.State = "cancelled"
				p.job.Progress.Cancelled = 1
				p.job.Message = "Cancelled; no partial device-state transaction committed"
			} else {
				p.job.State = "failed"
				p.job.Progress.Completed = 1
				p.job.Progress.Failed = 1
				p.job.Message = err.Error()
			}
			p.record(p.job.Message)
		} else {
			p.commit(result)
			p.job.State = "completed"
			p.job.Progress.Completed = 1
			p.job.ResultsState = "available"
			p.job.AnalysisAvailable = true
			p.job.ResultsRevision = strconv.FormatUint(p.revision, 10)
			p.job.Message = p.status
		}
		p.revision++
	}()
	return p.job, nil
}
func (p *Port) Job(ctx context.Context, id string) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return moduleapi.Job{}, err
	}
	if id != p.job.ID || id == "" {
		return moduleapi.Job{}, errors.New("unknown circuits job")
	}
	return p.job, nil
}
func (p *Port) Cancel(ctx context.Context, id string) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return moduleapi.Job{}, err
	}
	if id != p.job.ID || id == "" {
		return moduleapi.Job{}, errors.New("unknown circuits job")
	}
	if p.active() {
		p.job.State = "cancelling"
		p.job.CancellationRequested = true
		p.cancel()
	}
	return p.job, nil
}
func (p *Port) Close() {
	p.mu.Lock()
	p.closed = true
	if p.cancel != nil {
		p.cancel()
	}
	done := p.done
	p.mu.Unlock()
	if done != nil {
		<-done
	}
}
func (p *Port) Canonical() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.result == nil {
		return ""
	}
	return p.result.Canonical()
}
