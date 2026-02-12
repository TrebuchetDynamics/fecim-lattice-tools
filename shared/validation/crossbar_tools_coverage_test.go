package validation

import (
	"testing"
)

func TestToolStatus_String(t *testing.T) {
	cases := []struct {
		s    ToolStatus
		want string
	}{
		{StatusUnknown, "Unknown"},
		{StatusInstalled, "Installed"},
		{StatusNotInstalled, "Not Installed"},
		{StatusError, "Error"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("ToolStatus(%d).String() = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestToolStatus_Symbol(t *testing.T) {
	cases := []struct {
		s    ToolStatus
		want string
	}{
		{StatusInstalled, "✓"},
		{StatusNotInstalled, "✗"},
		{StatusError, "⚠"},
		{StatusUnknown, "○"},
	}
	for _, c := range cases {
		if got := c.s.Symbol(); got != c.want {
			t.Errorf("ToolStatus(%d).Symbol() = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestCheckAllTools_ReturnsEntries(t *testing.T) {
	tools := CheckAllTools()
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}
	names := map[string]bool{}
	for _, ti := range tools {
		names[ti.Name] = true
		if ti.Description == "" {
			t.Errorf("tool %s has empty description", ti.Name)
		}
		if ti.InstallCmd == "" {
			t.Errorf("tool %s has empty install command", ti.Name)
		}
	}
	if !names["CrossSim"] || !names["BadCrossbar"] {
		t.Error("expected CrossSim and BadCrossbar in tool list")
	}
}

func TestGetProjectRoot_Succeeds(t *testing.T) {
	root, err := GetProjectRoot()
	if err != nil {
		t.Fatalf("GetProjectRoot: %v", err)
	}
	if root == "" {
		t.Fatal("empty project root")
	}
}

func TestGetLocalClonePaths_Coverage(t *testing.T) {
	cs, bc, err := GetLocalClonePaths()
	if err != nil {
		t.Fatalf("GetLocalClonePaths: %v", err)
	}
	if cs == "" || bc == "" {
		t.Error("expected non-empty clone paths")
	}
}

func TestHasLocalClone_FalseForTmpDir(t *testing.T) {
	if HasLocalClone(t.TempDir()) {
		t.Error("expected false for empty temp dir")
	}
}

func TestValidateAllTools_ReturnsResults(t *testing.T) {
	results := ValidateAllTools()
	if len(results) != 2 {
		t.Errorf("expected 2 validation results, got %d", len(results))
	}
	for _, r := range results {
		if r.Tool == "" {
			t.Error("empty tool name in result")
		}
	}
}

func TestInstallToolsIfNeeded_DoesNotPanic(t *testing.T) {
	results := InstallToolsIfNeeded()
	if len(results) != 2 {
		t.Errorf("expected 2 install results, got %d", len(results))
	}
}
