package study

import (
	"os"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/workbench/project"
)

// Reopening root.Name or writing into an existing/link directory would escape
// the exclusive rooted creation contract.
func TestCreateStarterProjectRoot(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	moved := filepath.Join(t.TempDir(), "moved")
	if err := os.Rename(dir, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	opts := DefaultStarterSweepOptions()
	if err := CreateStarterProjectRoot(root, "study", opts); err != nil {
		t.Fatal(err)
	}
	child, err := root.OpenRoot("study")
	if err != nil {
		t.Fatal(err)
	}
	defer child.Close()
	bundle, err := project.LoadRoot(child, nil)
	if err != nil {
		t.Fatal(err)
	}
	points, err := project.Expand(bundle)
	if err != nil || len(points) != 8 {
		t.Fatalf("points=%d err=%v", len(points), err)
	}
	original, err := child.ReadFile("project.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := CreateStarterProjectRoot(root, "study", opts); err == nil {
		t.Fatal("existing project replaced")
	}
	got, err := child.ReadFile("project.yaml")
	if err != nil || string(got) != string(original) {
		t.Fatalf("existing changed: %v", err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(moved, "link")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"link", "../escape", "/absolute", "nested/child", ".", "", "a\\b"} {
		if err := CreateStarterProjectRoot(root, name, opts); err == nil {
			t.Fatalf("unsafe/existing %q accepted", name)
		}
	}
	opts.Rows = []int{0}
	if err := CreateStarterProjectRoot(root, "invalid", opts); err == nil {
		t.Fatal("invalid options accepted")
	}
	if _, err := root.Stat("invalid"); !os.IsNotExist(err) {
		t.Fatalf("invalid creation touched disk: %v", err)
	}
	if err := CreateStarterProjectRoot(nil, "study", DefaultStarterSweepOptions()); err == nil {
		t.Fatal("nil root accepted")
	}
	if err := root.Close(); err != nil {
		t.Fatal(err)
	}
	if err := CreateStarterProjectRoot(root, "closed", DefaultStarterSweepOptions()); err == nil {
		t.Fatal("closed root accepted")
	}
	for _, path := range []string{dir, outside} {
		entries, err := os.ReadDir(path)
		if err != nil || len(entries) != 0 {
			t.Fatalf("outside mutated: %s %v", path, err)
		}
	}
}
