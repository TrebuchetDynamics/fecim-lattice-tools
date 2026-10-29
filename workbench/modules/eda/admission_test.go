package eda

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/workbench/edaexport"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/native"
)

type weightGrant struct {
	size int64
	data string
	root *os.Root
}

func (g weightGrant) Open(context.Context, moduleapi.ResourceReference) (io.ReadCloser, native.ResourceInfo, error) {
	return io.NopCloser(strings.NewReader(g.data)), native.ResourceInfo{Size: g.size}, nil
}
func (g weightGrant) WithProject(ctx context.Context, ref moduleapi.ResourceReference, _ bool, fn func(*os.Root) error) error {
	if ref.ID != "opened" || ref.Kind != "project" {
		return os.ErrPermission
	}
	return fn(g.root)
}

// Root-selected file byte bounds are distinct from JSON-body and freeform cell limits.
func TestEDAWebAdmission(t *testing.T) {
	for _, tc := range []struct {
		name  string
		size  int64
		data  string
		state string
	}{
		{"at root weight limit", edaexport.MaxWeightsBytes, "[[0,1]]", "finished"},
		{"over root weight limit", edaexport.MaxWeightsBytes + 1, "[[0,1]]", "failed"},
		{"shape mismatch", 5, "[[0]]", "failed"},
		{"unverified path as project", 5, "[[0]]", "rejected"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, err := os.OpenRoot(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			g := weightGrant{tc.size, tc.data, root}
			ctx := native.WithProjects(native.WithResources(context.Background(), g), g)
			p := New()
			defer p.Close()
			cfg := edaexport.Defaults(compiler.ModeCompute, 1, 2, compiler.ArchPassive)
			data, _ := json.Marshal(cfg)
			ref := moduleapi.ResourceReference{ID: "opened", Kind: "project"}
			if tc.state == "rejected" {
				ref.ID = "/etc"
			}
			job, err := p.Start(ctx, moduleapi.Command{Action: "generate_all", Payload: data, Resources: []moduleapi.ResourceReference{ref, {ID: "bound", Kind: "input"}}})
			if err != nil {
				if tc.state == "rejected" {
					return
				}
				t.Fatal(err)
			}
			deadline := time.Now().Add(3 * time.Second)
			for job.State == "running" || job.State == "queued" {
				if time.Now().After(deadline) {
					t.Fatal("not bounded")
				}
				time.Sleep(time.Millisecond)
				job, err = p.Job(ctx, job.ID)
				if err != nil {
					t.Fatal(err)
				}
			}
			want := tc.state
			if want == "rejected" {
				want = "failed"
			}
			if job.State != want {
				t.Fatalf("%+v", job)
			}
		})
	}
}
