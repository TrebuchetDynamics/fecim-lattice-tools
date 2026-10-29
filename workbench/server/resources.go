package server

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"

	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/httpapi"
	nativeapi "fecim-lattice-tools/workbench/moduleapi/native"
)

const maxResources = 64
const maxResourceBytes int64 = 256 << 20

type grantedResource struct {
	mu     sync.RWMutex
	file   *os.File
	ref    moduleapi.ResourceReference
	size   int64
	closed bool
}
type resourceReader struct {
	mu       sync.Mutex
	reader   io.Reader
	resource *grantedResource
	ctx      context.Context
	closed   bool
}

func (r *resourceReader) Read(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || r.ctx.Err() != nil {
		return 0, errors.New("resource reader closed or cancelled")
	}
	r.resource.mu.RLock()
	defer r.resource.mu.RUnlock()
	if r.resource.closed {
		return 0, errors.New("resource revoked")
	}
	return r.reader.Read(data)
}
func (r *resourceReader) Close() error { r.mu.Lock(); r.closed = true; r.mu.Unlock(); return nil }

type resourceResolver struct{ s *Server }

func (s *Server) nativeContext(ctx context.Context) context.Context {
	return nativeapi.WithProjects(nativeapi.WithResources(ctx, resourceResolver{s}), resourceResolver{s})
}
func (r resourceResolver) WithProject(ctx context.Context, ref moduleapi.ResourceReference, mutation bool, fn func(*os.Root) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ref.Kind != "project" || fn == nil {
		return errors.New("opened project reference and callback required")
	}
	p, release, err := r.s.borrow(ref.ID, mutation)
	if err != nil {
		return err
	}
	defer release()
	if mutation {
		p.config.Lock()
		defer p.config.Unlock()
	} else {
		p.config.RLock()
		defer p.config.RUnlock()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(p.root)
}

func (r resourceResolver) Open(ctx context.Context, ref moduleapi.ResourceReference) (io.ReadCloser, nativeapi.ResourceInfo, error) {
	if ctx.Err() != nil {
		return nil, nativeapi.ResourceInfo{}, errors.New("resource read cancelled")
	}
	r.s.mu.Lock()
	resource := r.s.resources[ref.ID]
	ended := r.s.closed || r.s.draining
	r.s.mu.Unlock()
	if ended || resource == nil || resource.ref != ref {
		return nil, nativeapi.ResourceInfo{}, errors.New("granted resource unavailable")
	}
	resource.mu.RLock()
	if resource.closed {
		resource.mu.RUnlock()
		return nil, nativeapi.ResourceInfo{}, errors.New("granted resource closed")
	}
	// SectionReader pins both the descriptor and exact admitted size. No root.Name
	// fallback, arbitrary path, shared offset or unbounded ReadAll is exposed.
	reader := &resourceReader{reader: io.NewSectionReader(resource.file, 0, resource.size), resource: resource, ctx: ctx}
	info := nativeapi.ResourceInfo{ID: ref.ID, Kind: ref.Kind, Size: resource.size}
	resource.mu.RUnlock()
	return reader, info, nil
}
func (s *Server) bindResource(q *httpapi.BindResourceInput) (httpapi.Result, error) {
	root := s.roots[q.Body.Root]
	if root == nil || !safeRelative(q.Body.Path) {
		return httpapi.Result{}, httpapi.Fail(400, "Choose a granted regular file")
	}
	switch q.Body.Kind {
	case "model", "dataset", "reference", "input":
	default:
		return httpapi.Result{}, httpapi.Fail(400, "Unsupported resource kind")
	}
	info, err := root.Lstat(q.Body.Path)
	if err != nil || !info.Mode().IsRegular() {
		return httpapi.Result{}, httpapi.Fail(400, "Resource must be a granted regular file, not a link")
	}
	if info.Size() > maxResourceBytes {
		return httpapi.Result{}, httpapi.Fail(413, "Resource exceeds 256 MiB native read limit")
	}
	file, err := root.Open(q.Body.Path)
	if err != nil {
		return httpapi.Result{}, httpapi.Fail(400, "Resource unavailable")
	}
	after, err := file.Stat()
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(info, after) || after.Size() > maxResourceBytes {
		file.Close()
		return httpapi.Result{}, httpapi.Fail(400, "Resource changed during admission")
	}
	ref := moduleapi.ResourceReference{ID: randomSecret(), Kind: q.Body.Kind}
	s.mu.Lock()
	if s.closed || s.draining || len(s.resources) >= maxResources {
		s.mu.Unlock()
		file.Close()
		return httpapi.Result{}, httpapi.Fail(503, "Resource capacity reached or session ended")
	}
	s.resources[ref.ID] = &grantedResource{file: file, ref: ref, size: after.Size()}
	s.mu.Unlock()
	return result(ref), nil
}
func (s *Server) closeResources() {
	s.mu.Lock()
	resources := s.resources
	s.resources = map[string]*grantedResource{}
	s.mu.Unlock()
	for _, resource := range resources {
		resource.mu.Lock()
		resource.closed = true
		_ = resource.file.Close()
		resource.mu.Unlock()
	}
}
