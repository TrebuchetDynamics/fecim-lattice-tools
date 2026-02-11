package crossbarcmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunGUIHelpTextReflectsImplementedEntryPoints(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	err = RunGUI([]string{"-help"})

	_ = w.Close()
	os.Stdout = oldStdout
	if err != nil {
		t.Fatalf("RunGUI(-help): %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read help output: %v", err)
	}
	out := buf.String()

	mustContain := []string{
		"fecim-lattice-tools crossbar [options]",
		"fecim-lattice-tools crossbar gui [options]",
		"fecim-lattice-tools crossbar inference [options]",
		"Implemented GUI capabilities:",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Fatalf("help text missing %q\noutput:\n%s", s, out)
		}
	}
}
