package recentfiles

import (
	"testing"
)

func TestListProjectsAndPresets(t *testing.T) {
	m := NewManager(nil)
	m.AddProject("/proj1.proj", "hysteresis")
	m.AddProject("/proj2.proj", "crossbar")
	m.AddPreset("/preset1.json", "circuits")

	projs := m.ListProjects()
	if len(projs) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projs))
	}

	presets := m.ListPresets()
	if len(presets) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(presets))
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Manager uses fyne.Preferences for persistence.
	// For testing, we use in-memory (nil prefs).
	m := NewManager(nil)
	m.AddConfig("/config.json", "test")
	m.AddExport("/export.csv", "test")

	// Save to in-memory (no-op with nil prefs)
	m.Save()

	if m.Count(FileTypeAny) != 2 {
		t.Fatalf("expected 2 files, got %d", m.Count(FileTypeAny))
	}
}

func TestFormatDurationHelper(t *testing.T) {
	tests := []struct {
		n        int
		unit     string
		expected string
	}{
		{1, "second", "1 second ago"},
		{10, "second", "10 seconds ago"}, // function adds 's' when n > 1
		{1, "minute", "1 minute ago"},
		{5, "minute", "5 minutes ago"}, // function adds 's' when n > 1
	}
	for _, tc := range tests {
		got := formatDuration(tc.n, tc.unit)
		if got != tc.expected {
			t.Errorf("formatDuration(%d, %q)=%q, want %q", tc.n, tc.unit, got, tc.expected)
		}
	}
}

func TestIntToStrSmall(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{5, "5"},
		{12, "12"},
		{99, "99"},
	}
	for _, tc := range tests {
		got := intToStrSmall(tc.input)
		if got != tc.expected {
			t.Errorf("intToStrSmall(%d)=%q, want %q", tc.input, got, tc.expected)
		}
	}
}
