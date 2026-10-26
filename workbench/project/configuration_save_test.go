package project

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// These cases catch rewriting no-ops, trusting stale bundles, replacing before
// persistence succeeds, and treating cancellation after the commit as rollback.
func TestSaveConfiguration(t *testing.T) {
	fixture := func(t *testing.T, project string) (*os.Root, *ConfigurationEditor, ConfigurationSnapshot) {
		t.Helper()
		dir := t.TempDir()
		writeBundle(t, dir, project, validDesignYAML, validSweepYAML)
		root, err := os.OpenRoot(dir)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { root.Close() })
		editor, err := NewConfigurationEditor(root, nil)
		if err != nil {
			t.Fatal(err)
		}
		base, err := ReadConfiguration(root)
		if err != nil {
			t.Fatal(err)
		}
		return root, editor, base
	}
	put := func(t *testing.T, root *os.Root, name, body string) {
		t.Helper()
		if err := root.WriteFile(name, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	unchanged := func(t *testing.T, root *os.Root, base ConfigurationSnapshot) {
		t.Helper()
		got, err := ReadConfiguration(root)
		if err != nil || got != base {
			t.Fatalf("disk changed: revision=%s err=%v", got.Revision(), err)
		}
		file, err := root.Open(".")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		names, err := file.Readdirnames(-1)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			if strings.HasPrefix(name, ".fecim-config-") {
				t.Fatalf("leaked temporary file: %s", name)
			}
		}
	}
	conflict := func(t *testing.T, err error, revision string, want []ConfigurationFieldConflict) {
		t.Helper()
		var got *ConfigurationConflict
		if !errors.As(err, &got) || got.CurrentRevision != revision || !reflect.DeepEqual(got.Fields, want) {
			t.Fatalf("conflict=%+v err=%v, want revision=%s fields=%+v", got, err, revision, want)
		}
	}

	for _, newline := range []string{"LF", "CRLF", "no final newline"} {
		for _, changes := range [][]FieldChange{nil, {{Path: "/name", Value: "HZO study"}}} {
			t.Run("no-op "+newline+" "+string(rune('0'+len(changes))), func(t *testing.T) {
				body := strings.Replace(validProjectYAML, "name: HZO study", "name: 'HZO study'   # retained", 1)
				if newline == "CRLF" {
					body = strings.ReplaceAll(body, "\n", "\r\n")
				}
				if newline == "no final newline" {
					body = strings.TrimSuffix(body, "\n")
				}
				root, editor, base := fixture(t, body)
				stamp := time.Unix(123456789, 0)
				for _, name := range configurationDocuments {
					if err := root.Chtimes(name, stamp, stamp); err != nil {
						t.Fatal(err)
					}
				}
				// No-op must not even create a temporary file.
				editor.fs.open = func(string, int, os.FileMode) (*os.File, error) {
					t.Error("no-op created temp")
					return nil, errors.New("unexpected create")
				}
				got, err := editor.Save(context.Background(), base.Revision(), "project.yaml", changes)
				if err != nil || got.Changed || got.Revision != base.Revision() {
					t.Fatalf("result=%+v err=%v", got, err)
				}
				unchanged(t, root, base)
				for _, name := range configurationDocuments {
					info, err := root.Stat(name)
					if err != nil || !info.ModTime().Equal(stamp) {
						t.Fatalf("mtime changed: %s %v", name, err)
					}
				}
			})
		}
	}

	for _, document := range configurationDocuments {
		t.Run("replace only "+document, func(t *testing.T) {
			root, editor, base := fixture(t, "# header\r\n"+strings.ReplaceAll(strings.TrimSuffix(validProjectYAML, "\n"), "\n", "\r\n"))
			if err := root.Chmod(document, 0o640); err != nil {
				t.Fatal(err)
			}
			patch := map[string][]FieldChange{
				"project.yaml": {{"/name", "Updated"}},
				"design.yaml":  {{"/array/rows", "64"}},
				"sweep.yaml":   {{"/seed", "18"}},
			}[document]
			got, err := editor.Save(context.Background(), base.Revision(), document, patch)
			if err != nil || !got.Changed || got.Revision == base.Revision() {
				t.Fatalf("result=%+v err=%v", got, err)
			}
			after, err := ReadConfiguration(root)
			if err != nil || got.Revision != after.Revision() {
				t.Fatalf("revision=%s err=%v", after.Revision(), err)
			}
			for i, name := range configurationDocuments {
				if name != document && after.documents[i] != base.documents[i] {
					t.Fatalf("rewrote %s", name)
				}
			}
			if !strings.Contains(after.documents[0], "# header") || !strings.Contains(after.documents[0], "model_version: fecim-system-v1") {
				t.Fatal("lost comment or unexposed value")
			}
			info, err := root.Stat(document)
			if err != nil || info.Mode().Perm() != 0o640 {
				t.Fatalf("permissions: %v %v", info, err)
			}
			bundle, err := LoadRoot(root, nil)
			if err != nil {
				t.Fatal(err)
			}
			if document == "project.yaml" && bundle.Project.Name != "Updated" || document == "design.yaml" && bundle.Design.Array.Rows != 64 || document == "sweep.yaml" && bundle.Sweep.Seed != 18 {
				t.Fatalf("patch not saved: %+v", bundle)
			}
			unchanged(t, root, after)
		})
	}

	t.Run("required and exact strong revision", func(t *testing.T) {
		root, editor, base := fixture(t, validProjectYAML)
		for _, revision := range []string{"", "*", "W/" + base.Revision(), strings.Trim(base.Revision(), `"`), base.Revision() + ", " + base.Revision(), `"cfg-v1-abcd"`, `"cfg-v1-` + strings.Repeat("G", 64) + `"`} {
			_, err := editor.Save(context.Background(), revision, "project.yaml", nil)
			want := ErrInvalidRevision
			if revision == "" {
				want = ErrRevisionRequired
			}
			if !errors.Is(err, want) {
				t.Fatalf("revision=%s err=%v want=%v", revision, err, want)
			}
		}
		unchanged(t, root, base)
	})

	for _, name := range configurationDocuments {
		t.Run("stale bundle "+name, func(t *testing.T) {
			root, editor, base := fixture(t, validProjectYAML)
			i := configurationDocumentIndex(name)
			put(t, root, name, base.documents[i]+"# external edit\n")
			current, _ := ReadConfiguration(root)
			patch := []FieldChange{{"/hypothesis", "Yours"}, {"/name", "Your name"}}
			_, err := editor.Save(context.Background(), base.Revision(), "project.yaml", patch)
			conflict(t, err, current.Revision(), []ConfigurationFieldConflict{
				{Path: "/hypothesis", Yours: "Yours", OnDisk: "Increasing ADC resolution trades area for energy fidelity.", OnDiskAvailable: true},
				{Path: "/name", Yours: "Your name", OnDisk: "HZO study", OnDiskAvailable: true},
			})
			_, err = editor.Save(context.Background(), base.Revision(), "project.yaml", nil)
			conflict(t, err, current.Revision(), []ConfigurationFieldConflict{})
			unchanged(t, root, current)
		})
	}

	for _, body := range []string{strings.Replace(validProjectYAML, "name: HZO study\n", "", 1), "not: [valid", validProjectYAML + "name: duplicate\n", strings.Replace(validProjectYAML, "name: HZO study", "name: [not, scalar]", 1)} {
		t.Run("unavailable conflict "+body[:8], func(t *testing.T) {
			root, editor, base := fixture(t, validProjectYAML)
			put(t, root, "project.yaml", body)
			current, _ := ReadConfiguration(root)
			_, err := editor.Save(context.Background(), base.Revision(), "project.yaml", []FieldChange{{"/name", "Yours"}})
			var got *ConfigurationConflict
			if !errors.As(err, &got) || got.CurrentRevision != current.Revision() || len(got.Fields) != 1 || got.Fields[0].Path != "/name" || got.Fields[0].Yours != "Yours" || got.Fields[0].OnDiskAvailable || got.Fields[0].Reason == "" {
				t.Fatalf("conflict=%+v err=%v", got, err)
			}
			unchanged(t, root, current)
		})
	}

	t.Run("participating saves serialize", func(t *testing.T) {
		root, editor, base := fixture(t, validProjectYAML)
		entered, release, secondStarted := make(chan struct{}), make(chan struct{}), make(chan struct{})
		write := editor.fs.write
		editor.fs.write = func(f *os.File, data []byte) (int, error) { close(entered); <-release; return write(f, data) }
		first, second := make(chan error, 1), make(chan error, 1)
		go func() {
			_, err := editor.Save(context.Background(), base.Revision(), "project.yaml", []FieldChange{{"/name", "First"}})
			first <- err
		}()
		<-entered
		go func() {
			close(secondStarted)
			_, err := editor.Save(context.Background(), base.Revision(), "project.yaml", []FieldChange{{"/name", "Second"}})
			second <- err
		}()
		<-secondStarted
		close(release)
		if err := <-first; err != nil {
			t.Fatal(err)
		}
		current, _ := ReadConfiguration(root)
		conflict(t, <-second, current.Revision(), []ConfigurationFieldConflict{{Path: "/name", Yours: "Second", OnDisk: "First", OnDiskAvailable: true}})
		unchanged(t, root, current)
	})

	for _, name := range configurationDocuments {
		t.Run("external edit before replacement "+name, func(t *testing.T) {
			root, editor, base := fixture(t, validProjectYAML)
			syncFile := editor.fs.sync
			editor.fs.sync = func(f *os.File) error {
				put(t, root, name, base.documents[configurationDocumentIndex(name)]+"# observed external edit\n")
				return syncFile(f)
			}
			_, err := editor.Save(context.Background(), base.Revision(), "project.yaml", []FieldChange{{"/name", "Yours"}})
			current, _ := ReadConfiguration(root)
			conflict(t, err, current.Revision(), []ConfigurationFieldConflict{{Path: "/name", Yours: "Yours", OnDisk: "HZO study", OnDiskAvailable: true}})
			unchanged(t, root, current)
		})
	}

	for _, stage := range []string{"create", "stat", "write", "short write", "sync", "close", "rename", "cancel before", "cancel precommit", "cancel after rename"} {
		t.Run(stage, func(t *testing.T) {
			root, editor, base := fixture(t, validProjectYAML)
			put(t, root, ".unrelated-temp", "keep")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			fault := errors.New("injected persistence error")
			switch stage {
			case "create":
				editor.fs.open = func(string, int, os.FileMode) (*os.File, error) { return nil, fault }
			case "stat":
				editor.fs.stat = func(*os.File) (os.FileInfo, error) { return nil, fault }
			case "write":
				editor.fs.write = func(f *os.File, data []byte) (int, error) { n, _ := f.Write(data[:8]); return n, fault }
			case "short write":
				editor.fs.write = func(f *os.File, data []byte) (int, error) { return f.Write(data[:8]) }
			case "sync":
				editor.fs.sync = func(*os.File) error { return fault }
			case "close":
				editor.fs.close = func(f *os.File) error { f.Close(); return fault }
			case "rename":
				editor.fs.rename = func(string, string) error { return fault }
			case "cancel before":
				cancel()
			case "cancel precommit":
				syncFile := editor.fs.sync
				editor.fs.sync = func(f *os.File) error { cancel(); return syncFile(f) }
			case "cancel after rename":
				rename := editor.fs.rename
				editor.fs.rename = func(from, to string) error { err := rename(from, to); cancel(); return err }
			}
			got, err := editor.Save(ctx, base.Revision(), "project.yaml", []FieldChange{{"/name", "Saved"}})
			if stage == "cancel after rename" {
				current, readErr := ReadConfiguration(root)
				if err != nil || !got.Changed || got.Revision != current.Revision() || readErr != nil || !strings.Contains(current.documents[0], "name: Saved") {
					t.Fatalf("committed save=%+v err=%v read=%v", got, err, readErr)
				}
			} else {
				want := fault
				if strings.HasPrefix(stage, "cancel") {
					want = context.Canceled
				}
				if stage == "short write" {
					want = io.ErrShortWrite
				}
				if !errors.Is(err, want) || got.Changed {
					t.Fatalf("result=%+v err=%v want=%v", got, err, want)
				}
				unchanged(t, root, base)
			}
			if body, err := root.ReadFile(".unrelated-temp"); err != nil || string(body) != "keep" {
				t.Fatalf("removed unrelated temp: %q %v", body, err)
			}
		})
	}

	t.Run("temporary creation and ownership", func(t *testing.T) {
		root, editor, base := fixture(t, validProjectYAML)
		var temp string
		open := editor.fs.open
		editor.fs.open = func(name string, flags int, mode os.FileMode) (*os.File, error) {
			temp = name
			if flags&os.O_EXCL == 0 || flags&os.O_CREATE == 0 || mode&0o077 != 0 {
				t.Errorf("temporary not exclusive/private: flags=%d mode=%v", flags, mode)
			}
			return open(name, flags, mode)
		}
		fault := errors.New("temporary substituted by external writer")
		editor.fs.sync = func(f *os.File) error {
			if err := root.Remove(temp); err != nil {
				t.Fatal(err)
			}
			put(t, root, temp, "foreign temporary")
			return fault
		}
		_, err := editor.Save(context.Background(), base.Revision(), "project.yaml", []FieldChange{{"/name", "Yours"}})
		if !errors.Is(err, fault) {
			t.Fatalf("err=%v", err)
		}
		if body, err := root.ReadFile(temp); err != nil || string(body) != "foreign temporary" {
			t.Fatalf("removed another writer's temporary: %q %v", body, err)
		}
		if err := root.Remove(temp); err != nil {
			t.Fatal(err)
		}
		unchanged(t, root, base)
	})

	for _, tc := range []struct {
		document string
		patch    []FieldChange
	}{
		{"../outside", nil}, {"/project.yaml", nil}, {"design.yaml", []FieldChange{{"/device/g_min_s", "1"}}}, {"project.yaml", []FieldChange{{"/id", "new-id"}}},
	} {
		t.Run("validation "+tc.document, func(t *testing.T) {
			root, editor, base := fixture(t, validProjectYAML)
			editor.fs.open = func(string, int, os.FileMode) (*os.File, error) {
				t.Error("invalid patch created temp")
				return nil, errors.New("unexpected create")
			}
			if _, err := editor.Save(context.Background(), base.Revision(), tc.document, tc.patch); err == nil {
				t.Fatal("invalid patch accepted")
			}
			unchanged(t, root, base)
		})
	}

	for _, kind := range []string{"escape", "directory", "closed root"} {
		t.Run(kind, func(t *testing.T) {
			root, editor, base := fixture(t, validProjectYAML)
			outside := filepath.Join(t.TempDir(), "outside")
			if err := os.WriteFile(outside, []byte("sentinel"), 0o600); err != nil {
				t.Fatal(err)
			}
			if kind == "closed root" {
				root.Close()
			} else {
				if err := root.Remove("project.yaml"); err != nil {
					t.Fatal(err)
				}
				if kind == "escape" {
					if err := root.Symlink(outside, "project.yaml"); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := root.Mkdir("project.yaml", 0o700); err != nil {
						t.Fatal(err)
					}
				}
			}
			if _, err := editor.Save(context.Background(), base.Revision(), "project.yaml", []FieldChange{{"/name", "Yours"}}); err == nil {
				t.Fatal("unsafe root accepted")
			}
			if data, err := os.ReadFile(outside); err != nil || string(data) != "sentinel" {
				t.Fatalf("outside changed: %q %v", data, err)
			}
		})
	}

	t.Run("caller handles and renamed roots", func(t *testing.T) {
		root, _, base := fixture(t, validProjectYAML)
		citationsDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(citationsDir, "park2015_advmat_hzo.md"), []byte("# Park"), 0o600); err != nil {
			t.Fatal(err)
		}
		citations, err := os.OpenRoot(citationsDir)
		if err != nil {
			t.Fatal(err)
		}
		defer citations.Close()
		editor, err := NewConfigurationEditor(root, citations)
		if err != nil {
			t.Fatal(err)
		}
		for _, dir := range []string{root.Name(), citationsDir} {
			if err := os.Rename(dir, dir+"-moved"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { os.RemoveAll(dir + "-moved") })
			if err := os.Mkdir(dir, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := editor.Save(context.Background(), base.Revision(), "project.yaml", []FieldChange{{"/name", "Saved"}}); err != nil {
			t.Fatal(err)
		}
		for _, handle := range []*os.Root{root, citations} {
			if _, err := handle.Stat("."); err != nil {
				t.Fatalf("caller handle closed: %v", err)
			}
		}
		citations.Close()
		current, _ := ReadConfiguration(root)
		if _, err := editor.Save(context.Background(), current.Revision(), "project.yaml", []FieldChange{{"/name", "Not saved"}}); err == nil {
			t.Fatal("closed citation root accepted")
		}
		unchanged(t, root, current)
	})

	t.Run("constructor requires live root", func(t *testing.T) {
		if _, err := NewConfigurationEditor(nil, nil); err == nil {
			t.Fatal("nil root accepted")
		}
		root, _, _ := fixture(t, validProjectYAML)
		root.Close()
		if _, err := NewConfigurationEditor(root, nil); err == nil {
			t.Fatal("closed root accepted")
		}
	})
}
