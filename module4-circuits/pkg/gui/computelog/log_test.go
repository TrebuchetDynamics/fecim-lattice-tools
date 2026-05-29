//go:build legacy_fyne

package computelog

import "testing"

func TestLogControls_ClearEnableAndEntriesCopy(t *testing.T) {
	log := New()
	log.Clear()
	log.Enable(true)
	if !log.Enabled() {
		t.Fatal("expected compute log to be enabled")
	}

	log.Append(Entry{ArraySize: "2x2"}, 100)
	entries := log.Entries()
	if len(entries) != 1 || entries[0].ArraySize != "2x2" {
		t.Fatalf("unexpected entries: %#v", entries)
	}
	entries[0].ArraySize = "mutated"
	if got := log.Entries()[0].ArraySize; got != "2x2" {
		t.Fatalf("entries should be copied, got %q", got)
	}

	log.Clear()
	if got := log.Entries(); len(got) != 0 {
		t.Fatalf("expected clear to remove entries, got %#v", got)
	}
}

func TestLogAppend_AppliesLimit(t *testing.T) {
	log := New()
	for i := 0; i < 3; i++ {
		log.Append(Entry{QuantLevels: i}, 2)
	}
	entries := log.Entries()
	if len(entries) != 2 || entries[0].QuantLevels != 1 || entries[1].QuantLevels != 2 {
		t.Fatalf("unexpected limited entries: %#v", entries)
	}
}
