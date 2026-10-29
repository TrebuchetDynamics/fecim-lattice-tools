package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/hysteresis"
)

func TestHysteresisLocalFiles(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	s, err := New(Config{Origin: testOrigin, Assets: testAssets(t), UI: "ui-test", Identity: testIdentity(), Roots: []string{root}})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cookie := login(t, s)
	var rootID string
	for id := range s.roots {
		rootID = id
	}
	snapshot := BrowserSnapshot{Config: hysteresis.DefaultConfig(), Rows: []LoopRow{{Index: 1, Time: 0.1, E: 100, P: 0.01}}}
	call := func(kind, operation, name string, snap BrowserSnapshot) int {
		t.Helper()
		body, _ := json.Marshal(FileRequest{Root: rootID, Path: name, Kind: kind, Operation: operation, UI: "ui-test", Identity: testIdentity(), Snapshot: snap})
		return request(s, "POST", "/api/v1/files", string(body), cookie, nil).Code
	}
	session, err := hysteresis.NewSession(snapshot.Config)
	if err != nil {
		t.Fatal(err)
	}
	programmer, err := hysteresis.NewProgrammer(session, hysteresis.DefaultProgramConfig())
	if err != nil {
		t.Fatal(err)
	}
	_, err = programmer.Apply(hysteresis.ProgramCommand{Kind: hysteresis.CalibrateProgram})
	if err != nil {
		t.Fatal(err)
	}
	for programmer.Snapshot().Status == hysteresis.ProgramRunning {
		if _, err = programmer.Advance(hysteresis.Budget{Steps: 256, DT: 1e-4}); err != nil {
			t.Fatal(err)
		}
	}
	cal, err := programmer.ExportCalibration()
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Calibration = &cal
	analysis, err := hysteresis.NewAnalysis(hysteresis.AnalysisRequest{Input: session.Snapshot(), Kind: hysteresis.AnalysisPUND})
	if err != nil {
		t.Fatal(err)
	}
	for {
		progress, advanceErr := analysis.Advance(hysteresis.AnalysisBudget{Steps: 256})
		if advanceErr != nil {
			t.Fatal(advanceErr)
		}
		if progress.State != "running" {
			break
		}
	}
	result, err := analysis.Result()
	if err != nil {
		t.Fatal(err)
	}
	result.ExperimentalValidation = true
	result.Evidence = "verified persisted study"
	result.References = []hysteresis.Reference{{Source: "user-reference.csv", Material: "unverified material", Evidence: "verified", ExperimentalValidation: true}}
	snapshot.Analysis = &result
	for _, kind := range []string{"preset", "calibration", "pe-json", "pe-csv", "pe-svg", "log-csv", "analysis-json", "pund-csv"} {
		name := kind + ".json"
		if code := call(kind, "save", name, snapshot); code != 200 {
			t.Fatalf("save %s: %d", kind, code)
		}
		if code := call(kind, "save", name, snapshot); code != 409 {
			t.Fatalf("overwrite %s: %d", kind, code)
		}
	}
	presetBytes, _ := os.ReadFile(filepath.Join(root, "preset.json"))
	if !strings.Contains(string(presetBytes), `"physics_engine": "preisach"`) {
		t.Fatal("legacy preset engine vocabulary not preserved")
	}
	if code := call("preset", "load", "preset.json", BrowserSnapshot{}); code != 200 {
		t.Fatalf("load preset: %d", code)
	}
	if code := call("calibration", "load", "calibration.json", BrowserSnapshot{}); code != 200 {
		t.Fatalf("load calibration: %d", code)
	}
	legacy := map[string]any{"version": 1, "material_name": session.Snapshot().Material.Parameters.Name, "num_levels": snapshot.Config.Levels, "target_range_frac": snapshot.Config.RangeFraction, "calibration_up": make([]float64, snapshot.Config.Levels), "calibration_down": make([]float64, snapshot.Config.Levels)}
	legacyBytes, _ := json.Marshal(legacy)
	if err := os.WriteFile(filepath.Join(root, "legacy.json"), legacyBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if code := call("calibration", "load", "legacy.json", BrowserSnapshot{}); code != 200 {
		t.Fatalf("legacy calibration format: %d", code)
	}
	for _, name := range []string{"calibration.json", "analysis-json.json", "pund-csv.json"} {
		b, _ := os.ReadFile(filepath.Join(root, name))
		if !strings.Contains(string(b), "browser simulation data") {
			t.Fatalf("artifact provenance missing: %s", name)
		}
		if strings.Contains(string(b), `"Evidence": "verified persisted study"`) || strings.Contains(string(b), `"ExperimentalValidation": true`) {
			t.Fatalf("client evidence promoted: %s", name)
		}
	}
	diagnosticBytes, _ := os.ReadFile(filepath.Join(root, "analysis-json.json"))
	if !strings.Contains(string(diagnosticBytes), "user-reference.csv") {
		t.Fatal("import source provenance lost")
	}
	bytes, _ := os.ReadFile(filepath.Join(root, "pe-json.json"))
	if !strings.Contains(string(bytes), "browser simulation data") || strings.Contains(string(bytes), `"experimental_validation": true`) {
		t.Fatal("unverified provenance missing")
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"../outside.json", filepath.Join(outside, "absolute.json"), "escape/write.json", "a/../write.json", "a//write.json", `C:\\write.json`, "CON", "a.", "a%2fb"} {
		if code := call("preset", "save", name, snapshot); code == 200 {
			t.Fatalf("unsafe path accepted: %s", name)
		}
	}
	if code := call("html", "save", "arbitrary.html", snapshot); code != 400 {
		t.Fatalf("kind enumeration: %d", code)
	}
	if entries, _ := os.ReadDir(outside); len(entries) != 0 {
		t.Fatal("outside write")
	}
	f, err := os.Create(filepath.Join(root, "oversize.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Truncate(MaxBodyBytes + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if code := call("preset", "load", "oversize.json", BrowserSnapshot{}); code != 413 {
		t.Fatalf("pre-read cap: %d", code)
	}
	forc, err := hysteresis.NewAnalysis(hysteresis.AnalysisRequest{Input: session.Snapshot(), Kind: hysteresis.AnalysisFORC, FORC: hysteresis.DefaultFORCParameters()})
	if err != nil {
		t.Fatal(err)
	}
	for {
		progress, advanceErr := forc.Advance(hysteresis.AnalysisBudget{Steps: 256})
		if advanceErr != nil {
			t.Fatal(advanceErr)
		}
		if progress.State != "running" {
			break
		}
	}
	forcResult, err := forc.Result()
	if err != nil {
		t.Fatal(err)
	}
	forcSnapshot := snapshot
	forcSnapshot.Analysis = &forcResult
	for _, kind := range []string{"forc-sweep-csv", "forc-density-csv"} {
		if code := call(kind, "save", kind+".csv", forcSnapshot); code != 200 {
			t.Fatalf("FORC %s: %d", kind, code)
		}
	}
	forcResult.FORC.Matrix.Cols++
	if code := call("forc-density-csv", "save", "malformed.csv", forcSnapshot); code != 400 {
		t.Fatalf("malformed diagnostic: %d", code)
	}
	nondefault := snapshot
	nondefault.Config.Model = hysteresis.Landau
	nondefault.Config.Stress = 2e9
	nondefault.Config.Seed = 17
	nondefault.Config.HistoryLimit = 1000
	if code := call("preset", "save", "roundtrip.json", nondefault); code != 200 {
		t.Fatal(code)
	}
	q, _ := json.Marshal(FileRequest{Root: rootID, Path: "roundtrip.json", Kind: "preset", Operation: "load", UI: "ui-test", Identity: testIdentity()})
	loaded := request(s, "POST", "/api/v1/files", string(q), cookie, nil)
	var roundtrip struct {
		Value struct{ Config hysteresis.Config }
	}
	if err := json.Unmarshal(loaded.Body.Bytes(), &roundtrip); err != nil || roundtrip.Value.Config != nondefault.Config {
		t.Fatalf("nondefault preset changed: %+v %v", roundtrip.Value.Config, err)
	}
	bad := snapshot
	bad.Config.Levels = 1000
	if code := call("preset", "save", "invalid.json", bad); code != 400 {
		t.Fatalf("invalid typed config: %d", code)
	}
	if _, err := os.Stat(filepath.Join(root, "invalid.json")); !os.IsNotExist(err) {
		t.Fatal("invalid payload wrote a file")
	}
}
