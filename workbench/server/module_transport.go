package server

import (
	"context"

	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/httpapi"
)

func (s *Server) moduleExecute(ctx context.Context, op string, input any) (httpapi.Result, error) {
	if op == "module-create" {
		q := input.(*httpapi.ModuleCreateInput)
		var factory *moduleFactory
		for i := range nativeFactories {
			if nativeFactories[i].id == q.Module {
				factory = &nativeFactories[i]
				break
			}
		}
		if factory == nil {
			return httpapi.Result{}, httpapi.Fail(404, "Module not found")
		}
		for _, d := range s.moduleDescriptors {
			if d.ID == q.Module && !d.Available {
				return httpapi.Result{}, httpapi.Fail(409, "Module web port unavailable")
			}
		}
		// Constructors are memory-only and never run while holding the session lock.
		port := nativePort(*factory)
		if port == nil {
			return httpapi.Result{}, httpapi.Fail(409, "Module web port unavailable")
		}
		snapshot, err := port.Snapshot(s.nativeContext(ctx))
		if err != nil {
			port.Close()
			return httpapi.Result{}, httpapi.Fail(409, "Module unavailable; verify matching build identity")
		}
		m := &moduleSession{id: randomSecret(), module: q.Module, port: port}
		s.mu.Lock()
		if s.closed || s.draining || len(s.modules) >= maxModuleSessions {
			s.mu.Unlock()
			port.Close()
			return httpapi.Result{}, httpapi.Fail(503, "Module session capacity reached or session ended")
		}
		m.ctx, m.cancel = context.WithCancel(s.epoch)
		m.ctx = s.nativeContext(m.ctx)
		s.modules[m.id] = m
		s.mu.Unlock()
		snapshot.Session = m.id
		return result(snapshot), nil
	}
	var id string
	switch q := input.(type) {
	case *httpapi.ModuleSessionInput:
		id = q.Session
	case *httpapi.ModuleActionInput:
		id = q.Session
	case *httpapi.ModuleCloseInput:
		id = q.Session
	case *httpapi.ModuleJobInput:
		id = q.Session
	case *httpapi.ModuleCancelInput:
		id = q.Session
	}
	s.mu.Lock()
	m := s.modules[id]
	s.mu.Unlock()
	if m == nil {
		return httpapi.Result{}, httpapi.Fail(404, "Module session not found")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.port == nil {
		return httpapi.Result{}, httpapi.Fail(404, "Module session closed")
	}
	var snapshot moduleapi.Snapshot
	var err error
	switch op {
	case "module-snapshot":
		snapshot, err = m.port.Snapshot(s.nativeContext(ctx))
	case "module-action":
		snapshot, err = m.port.Apply(s.nativeContext(ctx), input.(*httpapi.ModuleActionInput).Body)
	case "module-close":
		m.cancel()
		m.port.Close()
		m.port = nil
		s.mu.Lock()
		delete(s.modules, id)
		s.mu.Unlock()
		return result(httpapi.Status{Status: "closed"}), nil
	case "module-start", "module-job", "module-cancel":
		port, ok := m.port.(moduleapi.JobPort)
		if !ok {
			return httpapi.Result{}, httpapi.Fail(409, "This module has no asynchronous job capability")
		}
		var job moduleapi.Job
		switch q := input.(type) {
		case *httpapi.ModuleActionInput:
			job, err = port.Start(m.ctx, q.Body)
		case *httpapi.ModuleJobInput:
			job, err = port.Job(s.nativeContext(ctx), q.Job)
		case *httpapi.ModuleCancelInput:
			job, err = port.Cancel(s.nativeContext(ctx), q.Job)
		}
		if err != nil {
			return httpapi.Result{}, httpapi.Fail(409, "Module job request rejected")
		}
		return result(job), nil
	}
	if err != nil {
		return httpapi.Result{}, httpapi.Fail(409, "Module command rejected; check action, input and revision")
	}
	snapshot.Session = m.id
	return result(snapshot), nil
}

// Close outside the global metadata lock: future ports may drain jobs/readers.
func (s *Server) closeModules() {
	s.mu.Lock()
	modules := s.modules
	s.modules = map[string]*moduleSession{}
	s.mu.Unlock()
	for _, m := range modules {
		if m.cancel != nil {
			m.cancel()
		}
		m.mu.Lock()
		if m.port != nil {
			m.port.Close()
			m.port = nil
		}
		m.mu.Unlock()
	}
}
