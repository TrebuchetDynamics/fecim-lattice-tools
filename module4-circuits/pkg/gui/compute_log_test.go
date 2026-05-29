//go:build legacy_fyne

package gui

import "testing"

func TestComputeLogPublicControls_ClearEnableAndEntriesCopy(t *testing.T) {
	oldEnabled := ComputeLogEnabled()
	defer EnableComputeLog(oldEnabled)
	defer ClearComputeLog()

	ClearComputeLog()
	EnableComputeLog(true)
	if !ComputeLogEnabled() {
		t.Fatal("expected compute log to be enabled")
	}

	globalComputeLog.Append(ComputeLogEntry{ArraySize: "2x2"}, 100)

	entries := GetComputeLogEntries()
	if len(entries) != 1 || entries[0].ArraySize != "2x2" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
	entries[0].ArraySize = "mutated"
	if got := GetComputeLogEntries()[0].ArraySize; got != "2x2" {
		t.Fatalf("entries should be copied, got %q", got)
	}

	ClearComputeLog()
	if got := GetComputeLogEntries(); len(got) != 0 {
		t.Fatalf("expected clear to remove entries, got %#v", got)
	}
}
