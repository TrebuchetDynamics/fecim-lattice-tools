// Package eda adapts real native EDA services to the shared authenticated module transport.
package eda

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/workbench/edaexport"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/native"
)

func Descriptor() moduleapi.ModuleDescriptor {
	return moduleapi.ModuleDescriptor{ID: "eda", Title: "EDA", Available: true, SchemaVersion: moduleapi.SchemaVersion, ModelVersion: "eda-native/1", Features: []string{"builder", "selected-run", "artifacts", "structural-validation", "installed-tools", "learn"}, Capabilities: []moduleapi.Capability{{Profile: "local", Available: true}, {Profile: "hosted", Available: true, Reason: "Learn and unverified read-only import only; generation and tools require local mode"}}}
}

type Port struct {
	mu                   sync.Mutex
	wg                   sync.WaitGroup
	closed               bool
	revision             uint64
	config               edaexport.Builder
	jobs                 map[string]moduleapi.Job
	cancel               context.CancelFunc
	active               string
	last                 edaexport.Result
	project              moduleapi.ResourceReference
	preview, previewName string
	tool                 ToolResult
}

func New() *Port {
	return &Port{config: edaexport.Defaults(compiler.ModeCompute, 8, 8, compiler.ArchPassive), jobs: map[string]moduleapi.Job{}}
}
func (p *Port) Descriptor() moduleapi.ModuleDescriptor { return Descriptor() }
func (p *Port) Snapshot(ctx context.Context) (moduleapi.Snapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return moduleapi.Snapshot{}, errors.New("EDA session closed")
	}
	return p.snapshot(), ctx.Err()
}
func jsonText(v any) string { b, _ := json.MarshalIndent(v, "", "  "); return string(b) }
func (p *Port) snapshot() moduleapi.Snapshot {
	status := "not generated"
	if p.last.ID != "" {
		status = "generated; tools not run"
	}
	if p.active != "" {
		status = p.jobs[p.active].State
	}
	if p.tool.State != "" {
		status = p.tool.State + ": " + p.tool.Summary
	}
	return moduleapi.Snapshot{Revision: p.revision, Descriptor: Descriptor(), Status: status, Evidence: []string{edaexport.Warning}, Fields: []moduleapi.Field{{ID: "config", Label: "Resolved builder config (conductance in µS)", Input: jsonText(p.config)}, {ID: "result", Label: "Enumerated artifacts and provenance", Input: jsonText(p.last)}, {ID: "project", Label: "Opened project reference", Input: p.project.ID}, {ID: "preview", Label: p.previewName, Input: p.preview}, {ID: "tool", Label: "Tool evidence", Input: jsonText(p.tool)}}, Actions: []moduleapi.Action{{ID: "generate_all", Label: "Generate all formats", Available: true}, {ID: "selected_run", Label: "Export verified selected run", Available: true}, {ID: "validate", Label: "Validate structure", Available: p.last.ID != "", Reason: "Requires an integrity-verified generated bundle"}, {ID: "tool", Label: "Run installed structural tool", Available: p.last.ID != "", Reason: "Only Icarus and Yosys are supported by this bounded profile"}}, Tables: []moduleapi.TablePage{}, Plots: []moduleapi.Plot{}}
}
func (p *Port) Apply(ctx context.Context, c moduleapi.Command) (moduleapi.Snapshot, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || c.Revision != p.revision {
		return moduleapi.Snapshot{}, errors.New("closed or stale EDA revision")
	}
	if p.active != "" {
		return moduleapi.Snapshot{}, errors.New("EDA job is active")
	}
	switch c.Action {
	case "defaults":
		rows, e := strconv.Atoi(c.Input["rows"])
		if e != nil {
			return moduleapi.Snapshot{}, e
		}
		cols, e := strconv.Atoi(c.Input["cols"])
		if e != nil {
			return moduleapi.Snapshot{}, e
		}
		mode, e := strconv.Atoi(c.Input["mode"])
		if e != nil {
			return moduleapi.Snapshot{}, e
		}
		b := edaexport.Defaults(compiler.OperationMode(mode), rows, cols, c.Input["architecture"])
		if e = b.Validate(); e != nil {
			return moduleapi.Snapshot{}, e
		}
		p.config = b
	case "preview", "validate":
		if p.last.ID == "" {
			return moduleapi.Snapshot{}, errors.New("no completed export")
		}
		projects, ok := native.Projects(ctx)
		if !ok {
			return moduleapi.Snapshot{}, errors.New("local project capability required")
		}
		err := projects.WithProject(ctx, p.project, false, func(root *os.Root) error {
			if c.Action == "validate" {
				return edaexport.ValidateRoot(root, p.last.ID)
			}
			if c.Input["artifact"] == "package.zip" {
				return errors.New("binary package: use enumerated download")
			}
			data, err := edaexport.ReadArtifactRoot(root, p.last.ID, c.Input["artifact"])
			if err != nil {
				return err
			}
			offset, e := strconv.Atoi(c.Input["offset"])
			if c.Input["offset"] == "" {
				offset = 0
				e = nil
			}
			if e != nil || offset < 0 || offset > len(data) {
				return errors.New("invalid preview offset")
			}
			end := min(offset+64*1024, len(data))
			p.preview = string(data[offset:end])
			p.previewName = fmt.Sprintf("%s bytes %d..%d of %d; escaped text", c.Input["artifact"], offset, end, len(data))
			return nil
		})
		if err != nil {
			p.tool = ToolResult{State: "failed", Summary: "Integrity/structural review failed; no validation pass"}
			return moduleapi.Snapshot{}, err
		}
		if c.Action == "validate" {
			p.tool = ToolResult{State: "structural-only", Summary: "DEF counts checked; freeform cell LEF/Liberty/Verilog names and ports checked. Selected-run bundles lack cell.v: cross-format check not run. Not signoff or device validation."}
		}
	case "tools":
		p.tool = DiscoverTools()
	default:
		return moduleapi.Snapshot{}, errors.New("unsupported EDA action")
	}
	p.revision++
	return p.snapshot(), nil
}
func (p *Port) Start(ctx context.Context, c moduleapi.Command) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed || c.Revision != p.revision {
		return moduleapi.Job{}, errors.New("closed or stale EDA revision")
	}
	if p.active != "" || len(p.jobs) >= 16 {
		return moduleapi.Job{}, errors.New("EDA job capacity reached; close the session after jobs drain")
	}
	if len(c.Payload) > edaexport.MaxConfigBytes {
		return moduleapi.Job{}, errors.New("EDA payload exceeds 8 MiB")
	}
	if c.Action != "generate_all" && c.Action != "selected_run" && c.Action != "tool" {
		return moduleapi.Job{}, errors.New("unsupported EDA job")
	}
	projects, ok := native.Projects(ctx)
	if !ok {
		return moduleapi.Job{}, errors.New("local-only generation requires project capability")
	}
	ref := p.project
	var weights *moduleapi.ResourceReference
	for _, r := range c.Resources {
		switch r.Kind {
		case "project":
			ref = r
		case "input":
			rr := r
			weights = &rr
		default:
			return moduleapi.Job{}, errors.New("unsupported EDA resource")
		}
	}
	if ref.ID == "" || ref.Kind != "project" {
		return moduleapi.Job{}, errors.New("opened project reference required")
	}
	var b edaexport.Builder
	var err error
	if c.Action == "generate_all" {
		b, err = edaexport.DecodeBuilder(c.Payload)
		if err != nil {
			return moduleapi.Job{}, err
		}
	}
	if c.Action == "selected_run" && !validRun(c.Input["run_id"]) {
		return moduleapi.Job{}, errors.New("verified stored run ID required")
	}
	if c.Action == "tool" && (p.last.ID == "" || ref != p.project || (c.Input["tool"] != "iverilog" && c.Input["tool"] != "yosys")) {
		return moduleapi.Job{}, errors.New("choose a completed bundle and allowlisted installed tool")
	}
	if weights != nil && c.Action != "generate_all" {
		return moduleapi.Job{}, errors.New("weights apply only to explicit freeform compute")
	}
	id := randomID()
	jobCtx, cancel := context.WithTimeout(ctx, edaexport.JobTimeout)
	p.cancel = cancel
	p.active = id
	p.revision++
	job := moduleapi.Job{ID: id, Project: ref.ID, State: "queued", ResultsState: "unavailable", Counts: map[string]int{}}
	p.jobs[id] = job
	source := p.last.ID
	p.tool = ToolResult{}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer cancel()
		p.setState(id, "running")
		var result edaexport.Result
		var tool ToolResult
		if weights != nil {
			b, err = loadWeights(jobCtx, b, *weights)
		}
		if err == nil {
			err = projects.WithProject(jobCtx, ref, true, func(root *os.Root) error {
				switch c.Action {
				case "generate_all":
					var e error
					result, e = edaexport.GenerateBuilderRoot(jobCtx, root, id, b)
					return e
				case "selected_run":
					run := c.Input["run_id"]
					if _, e := edaexport.ExportSelectedRunRoot(root, run); e != nil {
						return e
					}
					var e error
					result, e = edaexport.InspectRoot(root, run)
					return e
				case "tool":
					var e error
					tool, e = RunTool(jobCtx, root, source, c.Input["tool"])
					return e
				}
				return errors.New("unsupported job")
			})
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		j := p.jobs[id]
		j.State = "finished"
		j.ResultsState = "available"
		j.Message = "Generated educational artifacts; external tools not run"
		if c.Action == "tool" {
			j.Message = tool.Summary
			p.tool = tool
		}
		if err != nil {
			j.State = "failed"
			j.ResultsState = "unavailable"
			j.Message = "EDA job failed: " + err.Error()
			if jobCtx.Err() != nil {
				j.State = "cancelled"
			}
			p.tool = tool
			p.tool.State, p.tool.Summary = j.State, j.Message
		}
		if err == nil && jobCtx.Err() != nil {
			j.State = "cancelled"
			j.ResultsState = "unavailable"
			j.Message = "Cancellation settled; inspect disk before retry; completion was not claimed"
		}
		if j.State == "finished" && c.Action != "tool" {
			p.last = result
			p.preview, p.previewName = "", ""
			p.project = ref
			if c.Action == "generate_all" {
				p.config = b
			}
		}
		p.jobs[id] = j
		p.active = ""
		p.revision++
	}()
	return job, nil
}
func loadWeights(ctx context.Context, b edaexport.Builder, ref moduleapi.ResourceReference) (edaexport.Builder, error) {
	r, ok := native.Resources(ctx)
	if !ok {
		return b, errors.New("root weights capability unavailable")
	}
	if b.Config.Mode != compiler.ModeCompute {
		return b, errors.New("weights require compute mode")
	}
	reader, info, err := r.Open(ctx, ref)
	if err != nil {
		return b, err
	}
	defer reader.Close()
	if info.Size > edaexport.MaxWeightsBytes {
		return b, errors.New("root weights exceed 32 MiB")
	}
	data, err := io.ReadAll(io.LimitReader(reader, edaexport.MaxWeightsBytes+1))
	if err != nil {
		return b, err
	}
	if len(data) > edaexport.MaxWeightsBytes {
		return b, errors.New("root weights exceed 32 MiB")
	}
	var weights [][]float64
	d := json.NewDecoder(strings.NewReader(string(data)))
	if err = d.Decode(&weights); err != nil {
		return b, err
	}
	if d.Decode(new(any)) != io.EOF {
		return b, errors.New("one weights matrix required")
	}
	b.Config.ComputeConfig.InitialWeights = weights
	return b, b.Validate()
}
func validRun(id string) bool {
	if len(id) != 64 {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'f' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
func randomID() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
func (p *Port) setState(id, state string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	j := p.jobs[id]
	j.State = state
	p.jobs[id] = j
}
func (p *Port) Job(ctx context.Context, id string) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	j, ok := p.jobs[id]
	if !ok {
		return j, errors.New("unknown EDA job")
	}
	return j, ctx.Err()
}
func (p *Port) Cancel(ctx context.Context, id string) (moduleapi.Job, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	j, ok := p.jobs[id]
	if !ok {
		return j, errors.New("unknown EDA job")
	}
	if p.active == id {
		p.cancel()
		j.CancellationRequested = true
		j.State = "cancelling"
		p.jobs[id] = j
	}
	return j, ctx.Err()
}
func (p *Port) Close() {
	p.mu.Lock()
	p.closed = true
	if p.cancel != nil {
		p.cancel()
	}
	p.mu.Unlock()
	p.wg.Wait()
}
