package logging

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHysteresisLogRetentionPrunesExpiredAndOversizedFiles(t *testing.T) {
	dir := t.TempDir()
	if err := ensureHysteresisRetentionMarker(dir); err != nil {
		t.Fatal(err)
	}
	baseline := time.Now().Add(-300 * 24 * time.Hour)
	marker := filepath.Join(dir, hysteresisRetentionMarker)
	if err := os.Chtimes(marker, baseline, baseline); err != nil {
		t.Fatal(err)
	}
	expired := filepath.Join(dir, "hysteresis-expired.csv")
	if err := os.WriteFile(expired, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-200 * 24 * time.Hour)
	if err := os.Chtimes(expired, old, old); err != nil {
		t.Fatal(err)
	}
	oversized := filepath.Join(dir, "hysteresis-oversized.csv")
	if err := os.WriteFile(oversized, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Truncate(oversized, maxHysteresisLogBytes+1); err != nil {
		t.Fatal(err)
	}
	recent := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(oversized, recent, recent); err != nil {
		t.Fatal(err)
	}
	if err := RegisterHysteresisLog(filepath.Join(dir, "hysteresis-active.csv")); err != nil {
		t.Fatal(err)
	}
	UnregisterHysteresisLog(filepath.Join(dir, "hysteresis-active.csv"))
	for _, path := range []string{expired, oversized} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("managed log %s was not pruned: %v", filepath.Base(path), err)
		}
	}
}

func TestHysteresisLogRetentionPreservesPreexistingAndActiveFiles(t *testing.T) {
	dir := t.TempDir()
	preexisting := filepath.Join(dir, "hysteresis-old.csv")
	if err := os.WriteFile(preexisting, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(preexisting, time.Now().Add(-365*24*time.Hour), time.Now().Add(-365*24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	marker := filepath.Join(dir, hysteresisRetentionMarker)
	if err := ensureHysteresisRetentionMarker(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal(err)
	}
	baseline := time.Now().Add(-72 * time.Hour)
	if err := os.Chtimes(marker, baseline, baseline); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < maxHysteresisLogFiles+1; i++ {
		path := filepath.Join(dir, "hysteresis-managed-"+time.Unix(int64(i), 0).Format("150405")+".csv")
		if err := os.WriteFile(path, []byte("managed"), 0o644); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(-48 * time.Hour)
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	active := filepath.Join(dir, "hysteresis-active.csv")
	if err := RegisterHysteresisLog(active); err != nil {
		t.Fatal(err)
	}
	defer UnregisterHysteresisLog(active)
	if err := os.WriteFile(active, []byte("active"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(preexisting); err != nil {
		t.Fatalf("pre-existing log was pruned: %v", err)
	}
	if _, err := os.Stat(active); err != nil {
		t.Fatalf("active log was pruned: %v", err)
	}
	entries, err := filepath.Glob(filepath.Join(dir, "hysteresis-*.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) > maxHysteresisLogFiles+2 {
		t.Fatalf("managed logs remain unbounded: %d files", len(entries))
	}
}
