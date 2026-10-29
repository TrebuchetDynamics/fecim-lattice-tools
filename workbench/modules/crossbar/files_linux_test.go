//go:build linux

package crossbar

import (
	"context"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/native"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

type projectLease struct{ root *os.Root }

func (p projectLease) WithProject(ctx context.Context, _ moduleapi.ResourceReference, _ bool, f func(*os.Root) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return f(p.root)
}

func TestCrossbarLoadRejectsNonregularFilesWithoutBlocking(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	pipe := filepath.Join(dir, "input.json")
	if err = syscall.Mkfifo(pipe, 0600); err != nil {
		t.Fatal(err)
	}
	p := New(128)
	defer p.Close()
	ctx := native.WithProjects(context.Background(), projectLease{root})
	done := make(chan error, 1)
	go func() {
		_, err := p.Apply(ctx, moduleapi.Command{Action: "load", Input: map[string]string{"path": "input.json"}, Resources: []moduleapi.ResourceReference{{Kind: "project", ID: "opened"}}})
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("FIFO accepted as JSON input")
		}
	case <-time.After(time.Second):
		// Unblock the intentionally broken reader before failing; no leaked callback.
		f, e := os.OpenFile(pipe, os.O_RDWR|syscall.O_NONBLOCK, 0)
		if e != nil {
			t.Fatal(e)
		}
		f.Write([]byte("{}"))
		f.Close()
		<-done
		t.Fatal("load blocked opening a nonregular file; cancellation/Close cannot drain")
	}
}
