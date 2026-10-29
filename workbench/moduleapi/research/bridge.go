// Package research is the bounded, memory-only production research Worker port.
package research

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/study"
)

//go:embed example/*.yaml
var examples embed.FS

func ExampleDocuments() map[string][]byte {
	docs := map[string][]byte{}
	for _, name := range []string{"project.yaml", "design.yaml", "sweep.yaml"} {
		docs[name], _ = examples.ReadFile("example/" + name)
	}
	return docs
}
func Identity() hysteresis.Identity {
	id := hysteresis.BridgeIdentity()
	id.Protocol = moduleapi.SchemaVersion
	id.Model = "research-memory/1"
	return id
}

type Bridge struct {
	mu                 sync.Mutex
	revision, sequence uint64
	session            string
	bundle             project.Bundle
	configuration      *moduleapi.ResearchConfiguration
	points             int
	outcome            study.RunOutcome
	job                moduleapi.Job
	cancel             context.CancelFunc
	done               chan struct{}
}

func New() *Bridge { return &Bridge{job: moduleapi.Job{State: "idle", ResultsState: "unavailable"}} }
func (b *Bridge) Close() {
	b.mu.Lock()
	if b.cancel != nil {
		b.cancel()
	}
	done := b.done
	b.mu.Unlock()
	if done != nil {
		<-done
	}
}
func (b *Bridge) Dispatch(data []byte) []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	var req moduleapi.ResearchRequest
	reply := moduleapi.ResearchReply{Identity: Identity(), Revision: b.revision}
	finish := func(err error) []byte {
		if err != nil {
			reply.Error = err.Error()
		}
		out, _ := json.Marshal(reply)
		return out
	}
	if len(data) > 12<<20 {
		return finish(errors.New("request exceeds memory envelope limit"))
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&req); err != nil {
		return finish(errors.New("invalid research request"))
	}
	if d.Decode(new(any)) != io.EOF {
		return finish(errors.New("invalid trailing request"))
	}
	reply.Session = req.Session
	reply.Request = req.Request
	reply.Operation = req.Operation
	if req.Identity != Identity() || req.Session == "" || b.session != "" && b.session != req.Session || req.Request <= b.sequence || req.Revision != b.revision {
		return finish(errors.New("identity, session, sequence or revision mismatch"))
	}
	b.session = req.Session
	b.sequence = req.Request
	active := b.job.State == "running" || b.job.State == "cancelling"
	switch req.Operation {
	case "example", "import":
		if active {
			return finish(errors.New("one active scientific computation allowed"))
		}
		docs := req.Input.Documents
		if req.Operation == "example" {
			docs = ExampleDocuments()
		}
		bundle, err := project.LoadBytesLimited(docs, 256, 32)
		if err != nil {
			return finish(err)
		}
		snapshot, err := project.ReadConfigurationBytes(docs)
		if err != nil {
			return finish(err)
		}
		fields, err := snapshot.Fields()
		if err != nil {
			return finish(err)
		}
		b.bundle = bundle
		b.configuration = &moduleapi.ResearchConfiguration{ID: "memory", Name: bundle.Project.Name, Revision: snapshot.Revision(), Evidence: "Simulation only; imported bytes are unverified"}
		for _, field := range fields {
			b.configuration.Fields = append(b.configuration.Fields, moduleapi.FieldFromConfiguration(field, moduleapi.ConfigurationUnits(field, bundle)))
		}
		points, err := experiment.Expand(bundle)
		if err != nil {
			return finish(err)
		}
		b.points = len(points)
		b.outcome = study.RunOutcome{}
		b.job = moduleapi.Job{State: "idle", ResultsState: "unavailable"}
	case "validate":
		if b.configuration == nil {
			return finish(errors.New("open a bounded study first"))
		}
	case "start":
		if active {
			return finish(errors.New("one active scientific computation allowed"))
		}
		if b.configuration == nil {
			return finish(errors.New("open a bounded study first"))
		}
		ctx, cancel := context.WithCancel(context.Background())
		b.cancel = cancel
		b.done = make(chan struct{})
		b.job = moduleapi.Job{ID: fmt.Sprintf("memory-%d", b.revision+1), InputRevision: b.configuration.Revision, State: "running", ResultsState: "unavailable", Progress: moduleapi.Progress{Total: b.points}, Message: "Experiment started"}
		bundle := b.bundle
		revision := req.Input.RepositoryRevision
		done := b.done
		go b.run(ctx, bundle, revision, done)
	case "cancel":
		if active {
			b.cancel()
			b.job.State = "cancelling"
			b.job.CancellationRequested = true
			b.job.Message = "Cancellation requested; waiting for Go to settle"
		}
	case "state", "query", "compare", "inspect":
	default:
		return finish(errors.New("unavailable research operation"))
	}
	result := moduleapi.ResearchResult{Configuration: b.configuration, Job: b.job, Points: b.points}
	if b.outcome.AnalysisAvailable {
		page, err := study.QueryResults(b.outcome.Analysis, req.Input.Query)
		if err != nil {
			return finish(err)
		}
		result.Page = &page
		plot, err := study.PlotResults(b.outcome.Analysis, req.Input.Query)
		if err != nil {
			return finish(err)
		}
		result.Plot = &plot
		result.Counts = b.outcome.Analysis.Counts
		if req.Operation == "compare" {
			comparison, err := study.CompareResults(b.outcome.Analysis, req.Input.Compare)
			if err != nil {
				return finish(err)
			}
			result.Comparison = &comparison
		}
		if req.Operation == "inspect" {
			found := false
			for _, run := range b.outcome.Summary.Runs {
				if run.Manifest.RunID == req.Input.RunID {
					data, _ := json.MarshalIndent(run, "", "  ")
					result.RunJSON = string(data)
					found = true
					break
				}
			}
			if !found {
				return finish(errors.New("run is unavailable"))
			}
		}
		result.ResultsJSON = string(b.outcome.ResultsJSON)
	}
	b.revision++
	reply.Revision = b.revision
	reply.Committed = true
	reply.Result = result
	return finish(nil)
}
func (b *Bridge) run(ctx context.Context, bundle project.Bundle, revision string, done chan struct{}) {
	defer close(done)
	progress := make(chan experiment.ProgressSnapshot, 1)
	// Yield is inside Go's production point loop, before each Evaluate. Timers
	// return control to the JS host so real cancel/state messages can be serviced.
	yield := func(ctx context.Context) error {
		b.mu.Lock()
		select {
		case p := <-progress:
			b.job.Progress = moduleapi.Progress{Total: p.Total, Completed: p.Completed, Reused: p.Reused, Failed: p.Failed, Cancelled: p.Cancelled}
		default:
		}
		b.mu.Unlock()
		timer := time.NewTimer(time.Millisecond)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}
	outcome, err := study.RunMemory(ctx, bundle, study.RunOptions{Workers: 1, RepositoryRevision: revision, Progress: progress, PointYield: yield})
	b.mu.Lock()
	defer b.mu.Unlock()
	b.outcome = outcome
	b.job = moduleapi.CompleteJob(b.job, outcome, err)
	// Keep the existing production Worker terminal token; both transports share
	// authoritative counts/revisions/completeness rather than JS reconstruction.
	if b.job.State == "succeeded" {
		b.job.State = "finished"
	}
}
