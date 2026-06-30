package gui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/shared/logging"
)

// TestSaveDebugLog_JSONHasNoLinePrefix guards against an accidental
// json.MarshalIndent prefix regression: saveDebugLog must write JSON
// indented the same way json.MarshalIndent(v, "", "  ") would, with no
// extra per-line prefix.
func TestSaveDebugLog_JSONHasNoLinePrefix(t *testing.T) {
	if log == nil {
		log = logging.NewLogger("gui-test")
	}

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	defer func() { _ = os.Chdir(origWD) }()

	a := &App{
		wrdDebugLog: &WriteReadDebugLog{
			Timestamp: "2026-06-30T00:00:00Z",
			Material:  "fecim_hzo",
			Ec:        1e6,
			EcMVcm:    10,
			Ps:        20,
			Cycles: []WriteReadCycle{
				{CycleNum: 1, TargetLevel: 5, StartLevel: 0, ReadLevel: 5, Success: true},
			},
		},
	}

	a.saveDebugLog()

	matches, err := filepath.Glob(filepath.Join(tmp, "logs", "hysteresis-*.json"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly one debug log file, got %d", len(matches))
	}

	got, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read debug log: %v", err)
	}

	want, err := json.MarshalIndent(a.wrdDebugLog, "", "  ")
	if err != nil {
		t.Fatalf("marshal expected debug log: %v", err)
	}

	if string(got) != string(want) {
		t.Fatalf("debug log JSON has unexpected indentation\ngot:\n%s\nwant:\n%s", got, want)
	}
}
