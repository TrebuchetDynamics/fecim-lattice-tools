package main

import (
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/modules/docs"
)

func TestDocsIndex(t *testing.T) {
	if os.Getenv("FECIM_DOCS_INDEX_HELPER") == "1" {
		for i, arg := range os.Args {
			if arg == "--" {
				os.Args = append([]string{os.Args[0]}, os.Args[i+1:]...)
				break
			}
		}
		flag.CommandLine = flag.NewFlagSet("fecim-docs-index", flag.ExitOnError)
		main()
		os.Exit(0)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join("..", "..")
	fixture := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		for _, source := range docs.PublicationPolicy() {
			data, err := os.ReadFile(filepath.Join(repo, source.Path))
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(root, source.Path)
			if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(name, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
		return root
	}
	invoke := func(t *testing.T, root string, success bool, args ...string) string {
		t.Helper()
		argv := append([]string{"-test.run=^TestDocsIndex$", "--", "--root", root}, args...)
		cmd := exec.Command(exe, argv...)
		cmd.Env = append(os.Environ(), "FECIM_DOCS_INDEX_HELPER=1")
		out, err := cmd.CombinedOutput()
		if (err == nil) != success || strings.HasPrefix(string(out), "docs: ") != success {
			t.Fatalf("success=%v, err=%v, output=%s", success, err, out)
		}
		return string(out)
	}
	verify := func(t *testing.T, root string) {
		t.Helper()
		invoke(t, root, true)
		items, err := os.ReadDir(filepath.Join(root, "web/public/docs-content"))
		if err != nil || len(items) != len(docs.PublicationPolicy()) {
			t.Fatalf("document outputs: %d, err=%v", len(items), err)
		}
		for _, source := range docs.PublicationPolicy() {
			if _, err := os.Stat(filepath.Join(root, "web/public/docs-content", source.ID+".json")); err != nil {
				t.Fatal(err)
			}
		}
		for _, name := range []string{"web/public/docs-manifest.json", "workbench/modules/docs/corpus.json"} {
			if _, err := os.Stat(filepath.Join(root, name)); err != nil {
				t.Fatal(err)
			}
		}
		before := docsIndexTree(t, root)
		invoke(t, root, true, "--check")
		if after := docsIndexTree(t, root); !reflect.DeepEqual(before, after) {
			t.Fatal("--check changed the fixture tree or bytes")
		}
	}

	t.Run("escaping output parent and retry", func(t *testing.T) {
		root := fixture(t)
		external := t.TempDir()
		outside := filepath.Join(external, "outside")
		if err := os.Mkdir(outside, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(external, "sentinel"), []byte("untouched\x00\xff"), 0644); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(root, "workbench")
		if err := os.Symlink(outside, link); err != nil {
			t.Fatal(err)
		}
		before := docsIndexTree(t, external)
		out := invoke(t, root, false)
		if !strings.Contains(out, "path escapes from parent") {
			t.Fatalf("expected confinement error, got %s", out)
		}
		if after := docsIndexTree(t, external); !reflect.DeepEqual(before, after) {
			t.Fatalf("outside tree mutated: before=%v after=%v", before, after)
		}
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
		verify(t, root)
	})
	t.Run("fresh generation", func(t *testing.T) {
		verify(t, fixture(t))
	})
	t.Run("escaping source", func(t *testing.T) {
		root := fixture(t)
		name := filepath.Join(root, docs.PublicationPolicy()[0].Path)
		outside := filepath.Join(t.TempDir(), "source.md")
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(outside, data, 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(name); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, name); err != nil {
			t.Fatal(err)
		}
		before := docsIndexTree(t, root)
		if out := invoke(t, root, false); !strings.Contains(out, "path escapes from parent") {
			t.Fatalf("expected source confinement error, got %s", out)
		}
		if !reflect.DeepEqual(before, docsIndexTree(t, root)) {
			t.Fatal("rejected source caused output mutation")
		}
	})
	t.Run("unknown output", func(t *testing.T) {
		root := fixture(t)
		name := filepath.Join(root, "web/public/docs-content/unknown.json")
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte("unknown"), 0644); err != nil {
			t.Fatal(err)
		}
		before := docsIndexTree(t, root)
		if out := invoke(t, root, false); !strings.Contains(out, "unknown public output: unknown.json") {
			t.Fatalf("expected unknown-output rejection, got %s", out)
		}
		if !reflect.DeepEqual(before, docsIndexTree(t, root)) {
			t.Fatal("unknown-output rejection mutated fixture")
		}
	})
	t.Run("escaping output inventory", func(t *testing.T) {
		root := fixture(t)
		outside := t.TempDir()
		if err := os.WriteFile(filepath.Join(outside, "unknown.json"), []byte("private sentinel"), 0644); err != nil {
			t.Fatal(err)
		}
		parent := filepath.Join(root, "web/public")
		if err := os.MkdirAll(parent, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(parent, "docs-content")); err != nil {
			t.Fatal(err)
		}
		before, external := docsIndexTree(t, root), docsIndexTree(t, outside)
		if out := invoke(t, root, false); !strings.Contains(out, "path escapes from parent") {
			t.Fatalf("expected inventory confinement error, got %s", out)
		}
		if !reflect.DeepEqual(before, docsIndexTree(t, root)) || !reflect.DeepEqual(external, docsIndexTree(t, outside)) {
			t.Fatal("rejected inventory caused mutation")
		}
	})
	t.Run("missing output check", func(t *testing.T) {
		root := fixture(t)
		before := docsIndexTree(t, root)
		if out := invoke(t, root, false, "--check"); !strings.Contains(out, "generated docs drift:") {
			t.Fatalf("expected missing-output drift, got %s", out)
		}
		if !reflect.DeepEqual(before, docsIndexTree(t, root)) {
			t.Fatal("--check created output or directories")
		}
	})
}

// Record every entry without following symlinks, including directory names and file bytes.
func docsIndexTree(t *testing.T, root string) map[string]string {
	t.Helper()
	tree := map[string]string{}
	err := filepath.WalkDir(root, func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		switch {
		case d.Type()&os.ModeSymlink != 0:
			target, err := os.Readlink(name)
			if err != nil {
				return err
			}
			tree[rel] = "link:" + target
		case d.IsDir():
			tree[rel] = "dir"
		default:
			data, err := os.ReadFile(name)
			if err != nil {
				return err
			}
			tree[rel] = "file:" + string(data)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return tree
}
