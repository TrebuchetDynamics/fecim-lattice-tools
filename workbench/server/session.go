package server

import (
	"context"

	"fecim-lattice-tools/workbench/moduleapi/httpapi"
)

type requestSessionKey struct{}

// Caller holds s.mu. Binding to the captured cookie also rejects old requests
// after terminal bootstrap renewal replaces a session without cancelling jobs.
func (s *Server) currentRequestSession(ctx context.Context) bool {
	value, _ := ctx.Value(requestSessionKey{}).(string)
	return !s.closed && !s.draining && ctx.Err() == nil && equal(s.session, value)
}

// Acquire only after decoding, never while reading a potentially stalled body.
// Logout first cancels/revokes under s.mu, then exclusively drains these leases.
// Readers run concurrently; this is not the global metadata lock or the per-
// project mutation gate. SSE and the scientific runner do not hold this lease.
func (s *Server) beginSession(ctx context.Context) (func(), error) {
	s.operations.RLock()
	s.mu.Lock()
	valid := s.currentRequestSession(ctx)
	s.mu.Unlock()
	if !valid {
		s.operations.RUnlock()
		return nil, httpapi.Fail(401, "Session ended or replaced")
	}
	return s.operations.RUnlock, nil
}
