package crossbar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"fecim-lattice-tools/workbench/moduleapi"
)

type Document struct {
	Config Config      `json:"config"`
	Matrix [][]float64 `json:"matrix"`
	Input  []float64   `json:"input"`
}
type Port struct {
	mu                      sync.Mutex
	limit                   int
	revision                uint64
	cfg                     Config
	matrix                  [][]float64
	input                   []float64
	result                  *Result
	resultRevision          uint64
	inputEpoch, resultEpoch uint64
	row, col, page          int
	layers, layer           int
	mode                    string
	job                     moduleapi.Job
	cancel                  context.CancelFunc
	done                    chan struct{}
	closed                  bool
}

func New(limit int) *Port {
	c := DefaultConfig()
	return &Port{limit: limit, cfg: c, matrix: SeedMatrix(c), input: SeedInput(c), page: 1, mode: "matrix", layers: 4, job: moduleapi.Job{State: "idle", ResultsState: "unavailable"}}
}
func (p *Port) Descriptor() moduleapi.ModuleDescriptor {
	return moduleapi.ModuleDescriptor{ID: "crossbar", Title: "Crossbar", Available: true, SchemaVersion: moduleapi.SchemaVersion, ModelVersion: ModelVersion, Features: []string{"configure", "program", "select", "view", "reset", "run", "cancel", "stack", "import", "load", "save"}, Capabilities: []moduleapi.Capability{{Profile: "hosted", Available: true, Reason: "CPU only; 32 by 32; ephemeral science, no file writes"}, {Profile: "local", Available: true, Reason: "CPU only; 128 by 128; rooted file actions require explicit project capability"}}}
}
func (p *Port) active() bool { return p.job.State == "running" || p.job.State == "cancelling" }
func (p *Port) Snapshot(ctx context.Context) (moduleapi.Snapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return moduleapi.Snapshot{}, err
	}
	if p.closed {
		return moduleapi.Snapshot{}, errors.New("crossbar session closed")
	}
	return p.snapshot(), nil
}
func (p *Port) Apply(ctx context.Context, c moduleapi.Command) (moduleapi.Snapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fail := func(err error) (moduleapi.Snapshot, error) { return moduleapi.Snapshot{}, err }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if p.closed || c.Revision != p.revision {
		return fail(errors.New("closed session or stale crossbar revision"))
	}
	if p.active() {
		return fail(errors.New("cancel or wait for current computation"))
	}
	switch c.Action {
	case "configure":
		cfg := p.cfg
		for key, value := range c.Input {
			var err error
			switch key {
			case "rows":
				cfg.Rows, err = strconv.Atoi(value)
			case "cols":
				cfg.Cols, err = strconv.Atoi(value)
			case "seed":
				cfg.Seed, err = strconv.ParseInt(value, 10, 64)
			case "levels":
				cfg.Levels, err = strconv.Atoi(value)
			case "adc_bits":
				cfg.ADCBits, err = strconv.Atoi(value)
			case "dac_bits":
				cfg.DACBits, err = strconv.Atoi(value)
			case "noise":
				cfg.Noise, err = strconv.ParseFloat(value, 64)
			case "temperature":
				cfg.Temperature, err = strconv.ParseFloat(value, 64)
			case "rp_row":
				cfg.RpRow, err = strconv.ParseFloat(value, 64)
			case "rp_col":
				cfg.RpCol, err = strconv.ParseFloat(value, 64)
			case "ir":
				cfg.IR, err = strconv.ParseBool(value)
			case "sneak":
				cfg.Sneak, err = strconv.ParseBool(value)
			case "profile":
				cfg.Profile = value
			case "architecture":
				cfg.Architecture = value
			case "input", "matrix":
			default:
				return fail(fmt.Errorf("unknown setting %s", key))
			}
			if err != nil {
				return fail(fmt.Errorf("invalid %s", key))
			}
		}
		if err := cfg.validate(p.limit); err != nil {
			return fail(err)
		}
		matrix, input := clone(p.matrix), append([]float64(nil), p.input...)
		if cfg.Rows != p.cfg.Rows || cfg.Cols != p.cfg.Cols || cfg.Seed != p.cfg.Seed {
			matrix, input = SeedMatrix(cfg), SeedInput(cfg)
		}
		if raw, ok := c.Input["matrix"]; ok {
			if err := json.Unmarshal([]byte(raw), &matrix); err != nil {
				return fail(err)
			}
		}
		if raw, ok := c.Input["input"]; ok {
			if err := json.Unmarshal([]byte(raw), &input); err != nil {
				return fail(err)
			}
		}
		if err := validateData(cfg, matrix, input); err != nil {
			return fail(err)
		}
		p.cfg, p.matrix, p.input = cfg, matrix, input
		p.inputEpoch++
		p.row, p.col, p.page = 0, 0, 1
	case "program", "select":
		row, e1 := strconv.Atoi(c.Input["row"])
		col, e2 := strconv.Atoi(c.Input["col"])
		if e1 != nil || e2 != nil || row < 0 || col < 0 || row >= p.cfg.Rows || col >= p.cfg.Cols {
			return fail(errors.New("cell outside matrix"))
		}
		if c.Action == "program" {
			v, err := strconv.ParseFloat(c.Input["weight"], 64)
			if err != nil || !finite(v) || v < 0 || v > 1 {
				return fail(errors.New("cell state must be finite in [0,1]"))
			}
			p.matrix = clone(p.matrix)
			p.matrix[row][col] = v
			p.inputEpoch++
		}
		p.row, p.col = row, col
	case "view":
		mode := c.Input["mode"]
		if mode != "matrix" && mode != "vectors" && mode != "ir" && mode != "sneak" && mode != "accuracy" && mode != "fecap" && mode != "stack" && mode != "literature" && mode != "solver" {
			return fail(errors.New("unknown analysis view"))
		}
		page := 1
		pageCells := p.cfg.Rows * p.cfg.Cols
		if (mode == "ir" || mode == "solver") && p.result != nil {
			pageCells = p.result.Config.Rows * p.result.Config.Cols
		}
		if v, ok := c.Input["page"]; ok {
			var err error
			page, err = strconv.Atoi(v)
			if err != nil || page < 1 || page > (pageCells+31)/32 {
				return fail(errors.New("page outside matrix"))
			}
		}
		p.mode, p.page = mode, page
	case "stack":
		layers, e1 := strconv.Atoi(c.Input["layers"])
		layer, e2 := strconv.Atoi(c.Input["layer"])
		if e1 != nil || e2 != nil || layers < 1 || layers > 512 || layer < 0 || layer >= layers {
			return fail(errors.New("stack requires 1 through 512 layers and a valid zero-based layer"))
		}
		p.layers, p.layer, p.mode, p.page = layers, layer, "stack", 1
	case "reset":
		p.matrix, p.input = SeedMatrix(p.cfg), SeedInput(p.cfg)
		p.inputEpoch++
		p.result = nil
		p.resultRevision = 0
		p.row, p.col, p.page = 0, 0, 1
	case "import", "load":
		data, err := p.readInput(ctx, c)
		if err != nil {
			return fail(err)
		}
		doc, err := decodeDocument(data, p.limit)
		if err != nil {
			return fail(err)
		}
		p.cfg, p.matrix, p.input = doc.Config, doc.Matrix, doc.Input
		p.inputEpoch++
		p.row, p.col, p.page = 0, 0, 1
	case "save":
		data, err := json.Marshal(Document{p.cfg, p.matrix, p.input})
		if err != nil {
			return fail(err)
		}
		if err = p.save(ctx, c, data); err != nil {
			return fail(err)
		}
	default:
		return fail(errors.New("unsupported crossbar action"))
	}
	p.revision++
	return p.snapshot(), nil
}
func (p *Port) Start(ctx context.Context, c moduleapi.Command) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return moduleapi.Job{}, err
	}
	if p.closed || p.active() || c.Action != "run" || c.Revision != p.revision {
		return moduleapi.Job{}, errors.New("busy, closed or stale crossbar start")
	}
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.done = make(chan struct{})
	p.job = moduleapi.Job{ID: fmt.Sprintf("crossbar-%d", p.revision+1), State: "running", InputRevision: strconv.FormatUint(p.revision, 10), Progress: moduleapi.Progress{Total: 1}, ResultsState: "unavailable", Message: "Crossbar computation started"}
	cfg, matrix, input, done := p.cfg, clone(p.matrix), append([]float64(nil), p.input...), p.done
	go func() {
		defer close(done)
		// Timers return control to the JS event loop inside the real CPU algorithm.
		yield := func(ctx context.Context) error {
			timer := time.NewTimer(time.Millisecond)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		}
		result, err := Evaluate(ctx, cfg, matrix, input, p.limit, yield)
		p.mu.Lock()
		defer p.mu.Unlock()
		p.revision++
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				p.job.State = "cancelled"
				p.job.Progress.Cancelled = 1
				p.job.Message = "Crossbar cancelled; previous result retained"
			} else {
				p.job.State = "failed"
				p.job.Progress.Completed = 1
				p.job.Progress.Failed = 1
				p.job.Message = err.Error()
			}
			return
		}
		// Cancellation racing with a completed algorithm cannot publish after receipt.
		if ctx.Err() != nil {
			p.job.State = "cancelled"
			p.job.Progress.Cancelled = 1
			return
		}
		p.result = &result
		p.resultRevision = p.revision
		p.resultEpoch = p.inputEpoch
		p.job.State = "succeeded"
		p.job.Progress.Completed = 1
		p.job.ResultsState = "complete"
		p.job.ResultsRevision = strconv.FormatUint(p.revision, 10)
		p.job.AnalysisAvailable = true
		p.job.Message = "Crossbar computation finished"
		if result.Solver != nil && !result.Solver.Converged || result.IR != nil && !result.IR.Converged {
			p.job.State = "failed"
			p.job.Progress.Failed = 1
			p.job.Message = "Solver did not converge; inspect retained residuals, not a successful solve"
		}
	}()
	return p.job, nil
}
func (p *Port) Job(ctx context.Context, id string) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return moduleapi.Job{}, err
	}
	if p.closed || id != p.job.ID {
		return moduleapi.Job{}, errors.New("unknown crossbar job")
	}
	return p.job, nil
}
func (p *Port) Cancel(ctx context.Context, id string) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return moduleapi.Job{}, err
	}
	if p.closed || id != p.job.ID {
		return moduleapi.Job{}, errors.New("unknown crossbar job")
	}
	if p.active() {
		p.cancel()
		p.job.State = "cancelling"
		p.job.CancellationRequested = true
		p.job.Message = "Cancellation requested; waiting for Go"
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
	data, _ := p.result.Canonical()
	return string(data)
}
