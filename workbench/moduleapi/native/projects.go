package native

import (
	"context"
	"os"

	"fecim-lattice-tools/workbench/moduleapi"
)

// ProjectResolver leases an already opened project, never an arbitrary path.
// The callback must finish all root operations before returning and must not
// retain the root. Mutations share the research workflow's exclusion gate.
// Ports must observe cancellation and drain callbacks before Close returns.
// This native-only capability is never included in browser or Worker DTOs.
type ProjectResolver interface {
	WithProject(context.Context, moduleapi.ResourceReference, bool, func(*os.Root) error) error
}

type projectsKey struct{}

func WithProjects(ctx context.Context, resolver ProjectResolver) context.Context {
	return context.WithValue(ctx, projectsKey{}, resolver)
}

func Projects(ctx context.Context) (ProjectResolver, bool) {
	r, ok := ctx.Value(projectsKey{}).(ProjectResolver)
	return r, ok
}
