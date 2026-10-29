//go:build !js

package comparison

import (
	"bytes"
	"context"
	"encoding/json"
	"fecim-lattice-tools/workbench/moduleapi"
	nativeapi "fecim-lattice-tools/workbench/moduleapi/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type projectLease struct{ root *os.Root }

func (l projectLease) WithProject(ctx context.Context, ref moduleapi.ResourceReference, mutation bool, fn func(*os.Root) error) error {
	if ref.ID != "opened" || ref.Kind != "project" || !mutation {
		return os.ErrPermission
	}
	return fn(l.root)
}
func TestComparisonLocalExportIsRootedAndCanonical(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	ctx := nativeapi.WithProjects(context.Background(), projectLease{root})
	p := New()
	defer p.Close()
	snapshot, err := p.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range snapshot.Actions {
		if action.ID == "export" && !action.Available {
			t.Error("granted local context advertises export as unavailable")
		}
	}
	for _, format := range []string{"json", "csv"} {
		before, _ := p.Canonical(format)
		s, err := p.Apply(ctx, moduleapi.Command{Action: "export", Revision: p.revision, Input: map[string]string{"format": format}, Resources: []moduleapi.ResourceReference{{ID: "opened", Kind: "project"}}})
		if err != nil {
			t.Fatal(err)
		}
		name := strings.Fields(strings.TrimPrefix(s.Status, "Exported "))[0]
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || !bytes.Equal(data, before) {
			t.Fatalf("not original canonical %s: %v", format, err)
		}
		if format == "json" {
			var artifact map[string]any
			if json.Unmarshal(data, &artifact) != nil || artifact["schema_version"] != "fecim-comparison/2" {
				t.Fatal("missing artifact version")
			}
		}
		// Even an in-root symlink with identical bytes is not an immutable artifact.
		if err := os.Rename(filepath.Join(dir, name), filepath.Join(dir, "alias-source")); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("../../alias-source", filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
		if _, err := p.Apply(ctx, moduleapi.Command{Action: "export", Revision: p.revision, Input: map[string]string{"format": format}, Resources: []moduleapi.ResourceReference{{ID: "opened", Kind: "project"}}}); err == nil {
			t.Fatal("in-root artifact symlink admitted")
		}
		os.Remove(filepath.Join(dir, name))
		os.Remove(filepath.Join(dir, "alias-source"))
	}
	// An escaping directory link is refused by the root through the entire write.
	outside := t.TempDir()
	if err := os.RemoveAll(filepath.Join(dir, "exports")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "exports")); err != nil {
		t.Fatal(err)
	}
	_, err = p.Apply(ctx, moduleapi.Command{Action: "export", Revision: p.revision, Input: map[string]string{"format": "json"}, Resources: []moduleapi.ResourceReference{{ID: "opened", Kind: "project"}}})
	if err == nil {
		t.Fatal("escaping artifact directory admitted")
	}
	entries, _ := os.ReadDir(outside)
	if len(entries) != 0 {
		t.Fatal("export wrote outside grant")
	}
}
