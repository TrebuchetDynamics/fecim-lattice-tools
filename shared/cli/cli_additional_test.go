package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestOutputWriter_PrintAlways(t *testing.T) {
	// PrintAlways suppresses output in JSON mode
	var buf bytes.Buffer
	flags := &CommonFlags{Quiet: true, JSON: false}
	ow := &OutputWriter{flags: flags, writer: &buf}

	ow.PrintAlways("This should always print\n")

	if !strings.Contains(buf.String(), "This should always print") {
		t.Errorf("expected PrintAlways to override quiet, got: %q", buf.String())
	}
}

func TestOutputWriter_Error(t *testing.T) {
	// Error outputs JSON when in JSON mode (writes to stderr in normal mode)
	var buf bytes.Buffer
	flags := &CommonFlags{JSON: true}
	ow := &OutputWriter{flags: flags, writer: &buf}

	ow.Error("json error")

	output := buf.String()
	if !strings.Contains(output, `"error"`) {
		t.Errorf("expected JSON error field, got: %q", output)
	}
	if !strings.Contains(output, "json error") {
		t.Errorf("expected error message in JSON, got: %q", output)
	}
}

func TestOutputWriter_IsJSON(t *testing.T) {
	ow := &OutputWriter{flags: &CommonFlags{JSON: true}}
	if !ow.IsJSON() {
		t.Error("expected IsJSON to return true")
	}

	ow2 := &OutputWriter{flags: &CommonFlags{JSON: false}}
	if ow2.IsJSON() {
		t.Error("expected IsJSON to return false")
	}
}

func TestOutputWriter_IsQuiet(t *testing.T) {
	ow := &OutputWriter{flags: &CommonFlags{Quiet: true}}
	if !ow.IsQuiet() {
		t.Error("expected IsQuiet to return true")
	}

	ow2 := &OutputWriter{flags: &CommonFlags{Quiet: false}}
	if ow2.IsQuiet() {
		t.Error("expected IsQuiet to return false")
	}
}

func TestOutputWriter_PrintMultiLine(t *testing.T) {
	var buf bytes.Buffer
	ow := &OutputWriter{flags: &CommonFlags{}, writer: &buf}

	ow.Print("%s %s %s\n", "line1", "line2", "line3")

	if !strings.Contains(buf.String(), "line1") {
		t.Error("expected line1 in output")
	}
	if !strings.Contains(buf.String(), "line2") {
		t.Error("expected line2 in output")
	}
}

func TestOutputWriter_Close(t *testing.T) {
	var buf bytes.Buffer
	ow := &OutputWriter{flags: &CommonFlags{}, writer: &buf}

	if err := ow.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
}

func TestNewOutputWriter(t *testing.T) {
	flags := &CommonFlags{JSON: true, Quiet: false}
	ow, err := NewOutputWriter(flags)
	if err != nil {
		t.Fatalf("NewOutputWriter failed: %v", err)
	}
	if ow == nil {
		t.Fatal("NewOutputWriter returned nil")
	}
	if !ow.IsJSON() {
		t.Error("expected IsJSON to reflect flags")
	}
	ow.Close()
}

func TestBatchProcessor_HasItems(t *testing.T) {
	bp := &BatchProcessor{items: []string{"a", "b"}}
	if !bp.HasItems() {
		t.Error("expected HasItems to return true")
	}

	bp2 := &BatchProcessor{items: []string{}}
	if bp2.HasItems() {
		t.Error("expected HasItems to return false")
	}
}

func TestUsageHeaderAndCommonUsage(t *testing.T) {
	header := UsageHeader("testcmd", "v1.0.0")
	if !strings.Contains(header, "testcmd") {
		t.Errorf("expected testcmd in header, got: %q", header)
	}
	if !strings.Contains(header, "v1.0.0") {
		t.Errorf("expected version in header, got: %q", header)
	}

	usage := CommonUsage()
	if !strings.Contains(usage, "--json") {
		t.Errorf("expected --json in usage, got: %q", usage)
	}
	if !strings.Contains(usage, "--quiet") {
		t.Errorf("expected --quiet in usage, got: %q", usage)
	}
}
