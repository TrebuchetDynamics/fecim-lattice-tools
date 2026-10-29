package docs

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/moduleapi"
	nativeapi "fecim-lattice-tools/workbench/moduleapi/native"
)

type grantedTestRoot struct{ root *os.Root }

func (g grantedTestRoot) WithProject(_ context.Context, ref moduleapi.ResourceReference, _ bool, f func(*os.Root) error) error {
	return f(g.root)
}

type grantedTestDocument struct{}

func (grantedTestDocument) Open(_ context.Context, ref moduleapi.ResourceReference) (io.ReadCloser, nativeapi.ResourceInfo, error) {
	text := "# Granted lesson\n\nOnly local granted text contains uniquegrantword.\n"
	return io.NopCloser(strings.NewReader(text)), nativeapi.ResourceInfo{ID: ref.ID, Kind: "reference", Size: int64(len(text))}, nil
}
func TestDocsWebGrantedHistory(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	ctx := nativeapi.WithProjects(nativeapi.WithResources(context.Background(), grantedTestDocument{}), grantedTestRoot{root})
	p := New().WebPort()
	defer p.Close()
	s, err := p.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	apply := func(action string, input map[string]string, refs ...moduleapi.ResourceReference) error {
		next, e := p.Apply(ctx, moduleapi.Command{Action: action, Input: input, Resources: refs, Revision: s.Revision})
		if e == nil {
			s = next
		}
		return e
	}
	if err = apply("history_optin", map[string]string{"enabled": "surprise"}); err == nil {
		t.Fatal("unknown history intent accepted")
	}
	if err = apply("history_save", nil); err == nil {
		t.Fatal("write without opt-in")
	}
	if err = apply("search", map[string]string{"query": "hysteresis"}); err != nil {
		t.Fatal(err)
	}
	files, _ := os.ReadDir(dir)
	if len(files) != 0 {
		t.Fatal("read wrote history")
	}
	if err = apply("history_optin", map[string]string{"enabled": "true"}, moduleapi.ResourceReference{ID: "opened-project", Kind: "project"}); err != nil {
		t.Fatal(err)
	}
	files, _ = os.ReadDir(dir)
	if len(files) != 0 {
		t.Fatal("opt-in wrote without save")
	}
	if err = apply("favorite", map[string]string{"id": "glossary"}); err != nil {
		t.Fatal(err)
	}
	if err = apply("history_save", nil); err != nil {
		t.Fatal(err)
	}
	data, err := root.ReadFile("docs-history.json")
	if err != nil || !strings.Contains(string(data), "glossary") {
		t.Fatalf("history %s %v", data, err)
	}
	other := New().WebPort()
	defer other.Close()
	secondApply := func(action string, input map[string]string, refs ...moduleapi.ResourceReference) error {
		ss, e := other.Snapshot(ctx)
		if e != nil {
			return e
		}
		_, e = other.Apply(ctx, moduleapi.Command{Action: action, Input: input, Resources: refs, Revision: ss.Revision})
		return e
	}
	if err = secondApply("history_optin", map[string]string{"enabled": "true"}, moduleapi.ResourceReference{ID: "opened-project", Kind: "project"}); err != nil {
		t.Fatal(err)
	}
	if err = secondApply("history_load", nil); err != nil {
		t.Fatal(err)
	}
	if err = apply("favorite", map[string]string{"id": "trust"}); err != nil {
		t.Fatal(err)
	}
	if err = apply("history_save", nil); err != nil {
		t.Fatal(err)
	}
	latest, _ := root.ReadFile("docs-history.json")
	if err = secondApply("history_save", nil); err == nil {
		t.Fatal("second session overwrote newer history")
	}
	afterConflict, _ := root.ReadFile("docs-history.json")
	if string(afterConflict) != string(latest) {
		t.Fatal("conflict wrote history")
	}
	if err = apply("import_granted", map[string]string{"id": "local-lesson"}, moduleapi.ResourceReference{ID: "chosen-file", Kind: "reference"}); err != nil {
		t.Fatal(err)
	}
	if err = apply("search", map[string]string{"query": "uniquegrantword"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s.Fields[0].Input, `"total":1`) {
		t.Fatalf("not real granted corpus: %s", s.Fields[0].Input)
	}
	for _, malformed := range []string{`{}`, `null`, `{"favorites":null,"recent":null}`} {
		if err = root.WriteFile("docs-history.json", []byte(malformed), 0600); err != nil {
			t.Fatal(err)
		}
		if err = apply("history_load", nil); err == nil {
			t.Fatalf("malformed history accepted: %s", malformed)
		}
		if _, err = p.Snapshot(ctx); err != nil {
			t.Fatal("malformed read broke the session", err)
		}
	}
	changed := []byte(`{"favorites":[],"recent":[]}`)
	if err = root.WriteFile("docs-history.json", changed, 0600); err != nil {
		t.Fatal(err)
	}
	if err = apply("history_save", nil); err == nil {
		t.Fatal("externally changed history overwritten")
	}
	preserved, _ := root.ReadFile("docs-history.json")
	if string(preserved) != string(changed) {
		t.Fatal("history conflict changed existing file")
	}
	p.Close()
	if err = apply("import_granted", map[string]string{"id": "after-close"}, moduleapi.ResourceReference{ID: "chosen-file", Kind: "reference"}); err == nil {
		t.Fatal("closed port accepted import")
	}
}
