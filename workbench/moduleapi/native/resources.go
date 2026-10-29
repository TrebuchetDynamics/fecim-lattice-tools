// Package native defines optional native capabilities. Worker/browser DTOs
// carry only moduleapi.ResourceReference; they cannot acquire filesystem power.
package native

import (
	"context"
	"fecim-lattice-tools/workbench/moduleapi"
	"io"
)

type ResourceInfo struct {
	ID, Kind string
	Size     int64
}

// Resolver opens a previously bound opaque reference, never a path. A caller
// must close every reader before its Port.Close returns. Bytes are unverified
// input: scientific adapters remain responsible for format/hash validation.
type Resolver interface {
	Open(context.Context, moduleapi.ResourceReference) (io.ReadCloser, ResourceInfo, error)
}
type key struct{}

func WithResources(ctx context.Context, resolver Resolver) context.Context {
	return context.WithValue(ctx, key{}, resolver)
}
func Resources(ctx context.Context) (Resolver, bool) {
	r, ok := ctx.Value(key{}).(Resolver)
	return r, ok
}
