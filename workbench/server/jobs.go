package server

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"sync"
	"time"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/httpapi"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/study"
)

const maxJobs = 64
const maxActiveJobs = 4

type execution struct {
	mu               sync.Mutex
	value            moduleapi.Job
	cancel           context.CancelFunc
	key, fingerprint string
}

func (j *execution) snapshot() moduleapi.Job {
	j.mu.Lock()
	defer j.mu.Unlock()
	v := j.value
	v.Counts = maps.Clone(v.Counts)
	return v
}
func activeState(s string) bool { return s == "queued" || s == "running" || s == "cancelling" }
func (s *Server) execution(id string, cancel bool) (httpapi.Result, error) {
	s.mu.Lock()
	j := s.jobs[id]
	expired := false
	for _, old := range s.expiredJobs {
		expired = expired || old == id
	}
	s.mu.Unlock()
	if j == nil {
		if expired {
			return httpapi.Result{}, httpapi.Fail(410, "Execution snapshot expired; scientific runs remain on disk")
		}
		return httpapi.Result{}, httpapi.Fail(404, "Execution not found; inspect stored runs after reopening the project")
	}
	if cancel {
		j.mu.Lock()
		if activeState(j.value.State) {
			j.value.CancellationRequested = true
			j.value.State = "cancelling"
			j.cancel()
		}
		j.mu.Unlock()
	}
	return result(j.snapshot()), nil
}
func validKey(k string) bool {
	if len(k) < 1 || len(k) > 128 {
		return false
	}
	for _, c := range k {
		if c < 33 || c > 126 {
			return false
		}
	}
	return true
}
func (s *Server) start(_ context.Context, q *httpapi.StartInput) (httpapi.Result, error) {
	if q.IfMatch == "" {
		return httpapi.Result{}, httpapi.Fail(428, "Exact If-Match configuration revision required")
	}
	if !validKey(q.IdempotencyKey) || q.Body.Workers < 1 || q.Body.Workers > 4 {
		return httpapi.Result{}, httpapi.Fail(400, "Require an idempotency key and one to four workers")
	}
	fingerprint := fmt.Sprintf("%s\x00%d", q.IfMatch, q.Body.Workers)
	// Retry lookup precedes busy rejection; a retried accepted command is not a
	// second mutation, even if its original execution still owns the project.
	s.mu.Lock()
	for _, j := range s.jobs {
		if j.value.Project == q.Project && j.key == q.IdempotencyKey {
			if j.fingerprint != fingerprint {
				s.mu.Unlock()
				return httpapi.Result{}, httpapi.Fail(409, "Idempotency key already bound to different inputs or options")
			}
			v := j.snapshot()
			s.mu.Unlock()
			return result(v), nil
		}
	}
	s.mu.Unlock()
	p, release, err := s.borrow(q.Project, true)
	if err != nil {
		return httpapi.Result{}, err
	}
	snapshot, err := project.ReadConfiguration(p.root)
	if err != nil {
		release()
		return httpapi.Result{}, httpapi.Fail(400, "Configuration unavailable")
	}
	if q.IfMatch != snapshot.Revision() {
		release()
		return httpapi.Result{}, httpapi.Fail(412, "Configuration changed; reload before starting")
	}
	bundle, pointCount, validationErr := study.ValidateProjectRoot(p.root, nil)
	if validationErr != nil {
		release()
		return httpapi.Result{}, httpapi.Fail(400, "Project bundle is invalid or unavailable")
	}
	current, readErr := project.ReadConfiguration(p.root)
	if readErr != nil || current.Revision() != q.IfMatch {
		release()
		return httpapi.Result{}, httpapi.Fail(412, "Configuration changed during validation; reload before starting")
	}
	s.mu.Lock()
	if s.closed || s.draining {
		s.mu.Unlock()
		release()
		return httpapi.Result{}, httpapi.Fail(503, "Session ended")
	}
	active := 0
	for _, j := range s.jobs {
		if activeState(j.snapshot().State) {
			active++
		}
	}
	if active >= maxActiveJobs {
		s.mu.Unlock()
		release()
		return httpapi.Result{}, httpapi.Fail(503, "Active execution capacity reached")
	}
	if len(s.jobs) >= maxJobs {
		evicted := false
		for i, id := range s.jobOrder {
			if !activeState(s.jobs[id].snapshot().State) {
				delete(s.jobs, id)
				s.jobOrder = append(s.jobOrder[:i], s.jobOrder[i+1:]...)
				s.expiredJobs = append(s.expiredJobs, id)
				if len(s.expiredJobs) > 128 {
					s.expiredJobs = s.expiredJobs[1:]
				}
				evicted = true
				break
			}
		}
		if !evicted {
			s.mu.Unlock()
			release()
			return httpapi.Result{}, httpapi.Fail(503, "Execution registry full")
		}
	}
	ctx, cancel := context.WithCancel(s.epoch)
	j := &execution{value: moduleapi.Job{ID: randomSecret(), Project: p.id, InputRevision: q.IfMatch, State: "queued", Progress: moduleapi.Progress{Total: pointCount}, ResultsState: string(study.ResultsUnavailable)}, cancel: cancel, key: q.IdempotencyKey, fingerprint: fingerprint}
	s.jobs[j.value.ID] = j
	s.jobOrder = append(s.jobOrder, j.value.ID)
	s.workers.Add(1)
	s.mu.Unlock()
	go s.runExecution(ctx, j, p, bundle, q.Body.Workers, release)
	return result(j.snapshot()), nil
}
func (s *Server) runExecution(ctx context.Context, j *execution, p *openedProject, bundle project.Bundle, workers int, release func()) {
	defer s.workers.Done()
	defer func() { release() }()
	defer j.cancel()
	progress := make(chan experiment.ProgressSnapshot, 16)
	type completion struct {
		out study.RunOutcome
		err error
	}
	done := make(chan completion, 1)
	j.mu.Lock()
	if !j.value.CancellationRequested {
		j.value.State = "running"
	}
	j.mu.Unlock()
	go func() {
		out, err := s.runBundle(ctx, p.root, nil, bundle, study.RunOptions{Workers: workers, Progress: progress})
		done <- completion{out, err}
	}()
	for {
		select {
		case count := <-progress:
			j.mu.Lock()
			j.value.Progress = moduleapi.Progress(count)
			j.mu.Unlock()
		case completed := <-done:
			final := moduleapi.CompleteJob(j.snapshot(), completed.out, completed.err)
			// Nothing below uses the grant. Publish terminal state only after
			// releasing mutation ownership, so a terminal observer can proceed.
			release()
			release = func() {}
			j.mu.Lock()
			final.CancellationRequested = j.value.CancellationRequested
			j.value = final
			j.mu.Unlock()
			return
		}
	}
}
func (s *Server) Events(w http.ResponseWriter, r *http.Request, id string) {
	s.mu.Lock()
	j := s.jobs[id]
	if j == nil {
		s.mu.Unlock()
		problem(w, 404, "Execution not found")
		return
	}
	if s.closed || s.draining {
		s.mu.Unlock()
		problem(w, 503, "Session ended")
		return
	}
	s.streamWorkers.Add(1)
	s.mu.Unlock()
	defer s.streamWorkers.Done()
	control := http.NewResponseController(w)
	// A writer without ResponseController deadline support is not safe for a
	// long-lived stream. Real net/http writers support it; reject unsafe wrappers.
	if err := control.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		problem(w, 503, "Progress streaming requires bounded socket writes")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("X-Accel-Buffering", "no")
	send := func(snapshot moduleapi.Job) bool {
		_ = control.SetWriteDeadline(time.Now().Add(2 * time.Second))
		data, _ := json.Marshal(snapshot.Progress)
		if _, err := fmt.Fprintf(w, "event: progress\ndata: %s\n\n", data); err != nil {
			return false
		}
		return control.Flush() == nil
	}
	last := j.snapshot()
	if !send(last) {
		return
	}
	if !activeState(last.State) {
		return
	}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	heartbeat := time.NewTicker(10 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			current := j.snapshot()
			if current.Progress != last.Progress || current.State != last.State {
				if !send(current) {
					return
				}
				last = current
			}
			if !activeState(current.State) {
				return
			}
		case <-heartbeat.C:
			_ = control.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if _, err := w.Write([]byte(": heartbeat\n\n")); err != nil {
				return
			}
			if control.Flush() != nil {
				return
			}
		}
	}
}
