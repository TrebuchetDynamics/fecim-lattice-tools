package visualization

import (
	"io"
	"os"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
)

func TestModule2VisualizationE2EWideTerminalPanels(t *testing.T) {
	arr, err := crossbar.NewArray(&crossbar.Config{Rows: 5, Cols: 6, NoiseLevel: 0, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray error = %v", err)
	}
	defer arr.Destroy()
	if err := arr.ProgramWeightMatrix([][]float64{
		{0, 0.2, 0.4, 0.6, 0.8, 1},
		{1, 0.8, 0.6, 0.4, 0.2, 0},
		{0.1, 0.3, 0.5, 0.7, 0.9, 0.2},
		{0.9, 0.7, 0.5, 0.3, 0.1, 0.4},
		{0.25, 0.5, 0.75, 1, 0.5, 0.25},
	}); err != nil {
		t.Fatalf("ProgramWeightMatrix error = %v", err)
	}
	input := []float64{0.1, 0.3, 0.5, 0.7, 0.9, 1}
	ideal, err := arr.MVM(input)
	if err != nil {
		t.Fatalf("MVM error = %v", err)
	}
	actual, ir, err := arr.MVMWithIRDrop(input, crossbar.DefaultWireParams())
	if err != nil {
		t.Fatalf("MVMWithIRDrop error = %v", err)
	}
	sneak := arr.AnalyzeSneakPaths(2, 3)
	trace := arr.GenerateMVMSneakTrace(input, crossbar.DefaultMVMOptions(), 2)

	for _, useColor := range []bool{false, true} {
		t.Run(map[bool]string{false: "plain", true: "color"}[useColor], func(t *testing.T) {
			vis := NewTerminalVisualizer(arr, useColor)
			out := captureVisualizationE2E(t, func() {
				vis.ShowCrossbarState()
				vis.ShowMVMOperation(input, ideal)
				vis.ShowIRDropAnalysis(ir)
				vis.ShowSneakPathAnalysis(sneak, 2, 3)
				vis.ShowMVMSneakTrace(trace)
				vis.ShowMVMWithNonidealities(input, ideal, actual, ir)
				vis.ShowNeuralNetworkInference(2, makeDigitSevenE2E(), [][]float64{{0.1, 0.2, 0.7}, {0.05, 0.9, 0.05}}, 1, 0.9)
			})
			for _, marker := range []string{"Crossbar Array State", "Matrix-Vector Multiplication", "IR Drop Analysis", "Sneak Path Analysis", "MVM Sneak Path Current Trace", "MVM with Non-Idealities", "Neural Network Inference", "Predicted digit: 1"} {
				if !strings.Contains(out, marker) {
					t.Fatalf("terminal output missing %q\n%s", marker, out)
				}
			}
			if useColor && !strings.Contains(out, "\033[") {
				t.Fatalf("color output missing ANSI escapes\n%s", out)
			}
			if !useColor && strings.Contains(out, "\033[") {
				t.Fatalf("plain output unexpectedly contains ANSI escapes\n%s", out)
			}
		})
	}
}

func TestModule2VisualizationE2ELargeArrayAndNilTraceBoundaries(t *testing.T) {
	arr, err := crossbar.NewArray(&crossbar.Config{Rows: 18, Cols: 35, NoiseLevel: 0, ADCBits: 8, DACBits: 8})
	if err != nil {
		t.Fatalf("NewArray error = %v", err)
	}
	defer arr.Destroy()
	for r := 0; r < arr.Rows(); r++ {
		for c := 0; c < arr.Cols(); c++ {
			if err := arr.ProgramWeight(r, c, float64((r+c)%30)/29); err != nil {
				t.Fatalf("ProgramWeight(%d,%d): %v", r, c, err)
			}
		}
	}
	vis := NewTerminalVisualizer(arr, false)
	out := captureVisualizationE2E(t, func() {
		vis.ShowCrossbarState()
		vis.ShowMVMSneakTrace(nil)
	})
	for _, marker := range []string{"... (more rows)", "...|", "No trace data available"} {
		if !strings.Contains(out, marker) {
			t.Fatalf("large/boundary output missing %q\n%s", marker, out)
		}
	}
}

func captureVisualizationE2E(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = original
	data, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatalf("ReadAll captured stdout: %v", err)
	}
	return string(data)
}

func makeDigitSevenE2E() []float64 {
	img := make([]float64, 784)
	for col := 8; col < 22; col++ {
		img[5*28+col] = 1
	}
	for i := 0; i < 20; i++ {
		row := 6 + i
		col := 20 - i/2
		if row < 28 && col >= 0 {
			img[row*28+col] = 0.8
		}
	}
	return img
}
