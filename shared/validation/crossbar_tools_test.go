package validation

import (
	"testing"
)

func TestFindPython(t *testing.T) {
	python := findPython()
	if python == "" {
		t.Skip("Python 3 not available on this system")
	}
	t.Logf("Found Python at: %s", python)
}

func TestCheckPythonModule(t *testing.T) {
	python := findPython()
	if python == "" {
		t.Skip("Python 3 not available")
	}

	status, _, err := checkPythonModule("os")
	if err != nil {
		t.Errorf("Checking 'os' module failed: %v", err)
	}
	if status != StatusInstalled {
		t.Errorf("Expected 'os' module to be installed, got: %s", status)
	}

	status, _, _ = checkPythonModule("nonexistent_module_xyz_12345")
	if status != StatusNotInstalled {
		t.Errorf("Expected nonexistent module to be NotInstalled, got: %s", status)
	}
}

func TestCrossSimInfo(t *testing.T) {
	info := CrossSimInfo()
	if info.Name != "CrossSim" {
		t.Errorf("Expected name 'CrossSim', got: %s", info.Name)
	}
	if info.Status == StatusUnknown {
		t.Error("Status should not be Unknown after check")
	}
	t.Logf("CrossSim status: %s %s", info.Status.Symbol(), info.Status)
}

func TestBadCrossbarInfo(t *testing.T) {
	info := BadCrossbarInfo()
	if info.Name != "BadCrossbar" {
		t.Errorf("Expected name 'BadCrossbar', got: %s", info.Name)
	}
	if info.Status == StatusUnknown {
		t.Error("Status should not be Unknown after check")
	}
	t.Logf("BadCrossbar status: %s %s", info.Status.Symbol(), info.Status)
}

func TestToolStatusSymbol(t *testing.T) {
	tests := []struct {
		status   ToolStatus
		expected string
	}{
		{StatusInstalled, "✓"},
		{StatusNotInstalled, "✗"},
		{StatusError, "⚠"},
		{StatusUnknown, "○"},
	}

	for _, tt := range tests {
		if got := tt.status.Symbol(); got != tt.expected {
			t.Errorf("Symbol() for %s = %s, want %s", tt.status, got, tt.expected)
		}
	}
}
