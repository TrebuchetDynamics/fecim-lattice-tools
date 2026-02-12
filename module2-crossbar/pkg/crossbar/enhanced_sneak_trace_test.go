package crossbar

import (
	"strings"
	"testing"
)

func TestGenerateMVMSneakTrace(t *testing.T) {
	arr, err := NewArray(&Config{Rows: 4, Cols: 4, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}

	weights := [][]float64{
		{0.9, 0.8, 0.7, 0.6},
		{0.5, 0.4, 0.3, 0.2},
		{0.6, 0.7, 0.8, 0.9},
		{0.3, 0.5, 0.7, 0.9},
	}
	if err := arr.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("ProgramWeightMatrix failed: %v", err)
	}

	input := []float64{1.0, 0.8, 0.6, 0.4}
	opts := DefaultMVMOptions()
	opts.Architecture = "0T1R"
	trace := arr.GenerateMVMSneakTrace(input, opts, 2)
	if trace == nil {
		t.Fatal("trace is nil")
	}
	if len(trace.Rows) != 4 {
		t.Fatalf("unexpected number of trace rows: %d", len(trace.Rows))
	}
	if trace.TotalSneak <= 0 {
		t.Fatalf("expected positive sneak current, got %.8f", trace.TotalSneak)
	}
	if trace.PeakRow < 0 || trace.PeakRow >= 4 {
		t.Fatalf("invalid peak row: %d", trace.PeakRow)
	}

	text := trace.FormatText(2, 1)
	if !strings.Contains(text, "MVM Sneak Path Trace") {
		t.Fatalf("formatted text missing header: %q", text)
	}
	if !strings.Contains(text, "Architecture:") {
		t.Fatalf("formatted text missing architecture: %q", text)
	}
}

func TestMVMWithNonIdealities_PopulatesSneakTrace(t *testing.T) {
	arr, err := NewArray(&Config{Rows: 4, Cols: 4, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray failed: %v", err)
	}
	if err := arr.ProgramWeightMatrix([][]float64{{0.9, 0.9, 0.9, 0.9}, {0.8, 0.8, 0.8, 0.8}, {0.7, 0.7, 0.7, 0.7}, {0.6, 0.6, 0.6, 0.6}}); err != nil {
		t.Fatalf("ProgramWeightMatrix failed: %v", err)
	}

	opts := DefaultMVMOptions()
	opts.EnableIRDrop = false
	opts.EnableVariation = false
	opts.EnableSneakPaths = true
	opts.Architecture = "0T1R"

	res, err := arr.MVMWithNonIdealities([]float64{1, 1, 1, 1}, opts)
	if err != nil {
		t.Fatalf("MVMWithNonIdealities failed: %v", err)
	}
	if res.SneakTrace == nil {
		t.Fatal("SneakTrace should be populated when sneak paths are enabled")
	}
}
