package viewmodel_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"
	comparisonvm "fecim-lattice-tools/shared/viewmodel/comparison"
	crossbarvm "fecim-lattice-tools/shared/viewmodel/crossbar"
	docsvm "fecim-lattice-tools/shared/viewmodel/docs"
	edavm "fecim-lattice-tools/shared/viewmodel/eda"
	hystvm "fecim-lattice-tools/shared/viewmodel/hysteresis"
	mnistvm "fecim-lattice-tools/shared/viewmodel/mnist"
)

func TestAllModulesE2EArtifactAndSnapshotChain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping all-modules E2E artifact chain: runs physics diagnostics and exports across all modules")
	}
	tmp := t.TempDir()

	// Module 1: produce physics diagnostics and export artifacts.
	h := hystvm.New()
	for _, action := range []viewmodel.Action{
		{ID: hystvm.EventSetWaveform, Payload: map[string]string{"waveform": hystvm.WaveformSquare}},
		{ID: hystvm.EventSetFieldRange, Payload: map[string]string{"min": "-1800", "max": "1800"}},
		{ID: hystvm.EventRunPUND},
		{ID: hystvm.EventRunFORC, Payload: map[string]string{"curves": "5", "points": "9"}},
		{ID: hystvm.EventRunLevelCalibration},
	} {
		if err := h.ApplyAction(action); err != nil {
			t.Fatalf("hysteresis %s: %v", action.ID, err)
		}
	}
	hCSV := filepath.Join(tmp, "hysteresis-loop.csv")
	hLevel := filepath.Join(tmp, "hysteresis-levels.json")
	hPUND := filepath.Join(tmp, "hysteresis-pund.csv")
	hFORCMeta := filepath.Join(tmp, "hysteresis-forc-meta.json")
	for _, action := range []viewmodel.Action{
		{ID: hystvm.EventExportCSV, Payload: map[string]string{"path": hCSV}},
		{ID: hystvm.EventExportLevelCalibration, Payload: map[string]string{"path": hLevel}},
		{ID: hystvm.EventExportPUNDCSV, Payload: map[string]string{"path": hPUND}},
		{ID: hystvm.EventExportFORCMeta, Payload: map[string]string{"path": hFORCMeta}},
	} {
		if err := h.ApplyAction(action); err != nil {
			t.Fatalf("hysteresis export %s: %v", action.ID, err)
		}
	}
	assertFileHasE2E(t, hCSV, "field")
	assertFileHasE2E(t, hLevel, "level_count")
	assertFileHasE2E(t, hPUND, "pulse")
	assertJSONFileHasKeyE2E(t, hFORCMeta, "material")
	hSnap := h.Snapshot()
	if metricValueAllE2E(hSnap, "pund_switching_ratio") == "" || metricValueAllE2E(hSnap, "forc_curves") == "" || metricValueAllE2E(hSnap, "level_calibration_state") == "" {
		t.Fatalf("hysteresis diagnostics missing after chain: %+v", hSnap.Metrics)
	}

	// Module 2: run array workflow and verify matrix/result plot surfaces.
	x := crossbarvm.New(7, 5)
	for _, action := range []viewmodel.Action{
		{ID: "resize", Payload: map[string]string{"rows": "9", "cols": "6"}},
		{ID: "toggle_ir"},
		{ID: "run_mvm"},
	} {
		if err := x.ApplyAction(action); err != nil {
			t.Fatalf("crossbar %s: %v", action.ID, err)
		}
	}
	xSnap := x.Snapshot()
	if metricValueAllE2E(xSnap, "rows") != "9" || metricValueAllE2E(xSnap, "cols") != "6" || len(xSnap.Plots) < 2 {
		t.Fatalf("crossbar snapshot invalid rows=%q cols=%q plots=%+v", metricValueAllE2E(xSnap, "rows"), metricValueAllE2E(xSnap, "cols"), xSnap.Plots)
	}

	// Module 3: sweep quantization and keep benchmark/trust surface observable.
	mn := mnistvm.New()
	for _, levels := range []string{"8", "30", "64"} {
		if err := mn.ApplyAction(viewmodel.Action{ID: "sweep_levels", Payload: map[string]string{"levels": levels}}); err != nil {
			t.Fatalf("mnist sweep %s: %v", levels, err)
		}
	}
	mnSnap := mn.Snapshot()
	if metricValueAllE2E(mnSnap, "levels") != "64 levels" || !strings.Contains(sectionBodyAllE2E(mnSnap, "research_benchmark"), "Educational") || len(mnSnap.Plots) != 1 {
		t.Fatalf("mnist snapshot invalid: metrics=%+v sections=%+v plots=%+v", mnSnap.Metrics, mnSnap.Sections, mnSnap.Plots)
	}

	// Module 4: run circuits and export all reference artifacts.
	c := circuitsvm.New()
	for _, action := range []viewmodel.Action{
		{ID: circuitsvm.ActionResizeArray, Payload: map[string]string{"rows": "16", "cols": "16"}},
		{ID: circuitsvm.ActionSetArchitecture, Payload: map[string]string{"architecture": circuitsvm.Architecture2T1R}},
		{ID: circuitsvm.ActionSetCouplingTier, Payload: map[string]string{"tier": circuitsvm.CouplingTierB}},
		{ID: circuitsvm.ActionSetTimingOperation, Payload: map[string]string{"operation": "WRITE"}},
		{ID: circuitsvm.ActionRunRead},
		{ID: circuitsvm.ActionRunWrite},
		{ID: circuitsvm.ActionRunCompute},
	} {
		if err := c.ApplyAction(action); err != nil {
			t.Fatalf("circuits %s: %v", action.ID, err)
		}
	}
	cOp := filepath.Join(tmp, "circuits-operation.json")
	cSpecs := filepath.Join(tmp, "circuits-specs.json")
	cTiming := filepath.Join(tmp, "circuits-timing.json")
	cSVG := filepath.Join(tmp, "circuits-timing.svg")
	for _, action := range []viewmodel.Action{
		{ID: circuitsvm.ActionExportOperationLog, Payload: map[string]string{"path": cOp}},
		{ID: circuitsvm.ActionExportReferenceSpecs, Payload: map[string]string{"path": cSpecs}},
		{ID: circuitsvm.ActionExportReferenceTiming, Payload: map[string]string{"path": cTiming}},
		{ID: circuitsvm.ActionExportReferenceTimingSVG, Payload: map[string]string{"path": cSVG}},
	} {
		if err := c.ApplyAction(action); err != nil {
			t.Fatalf("circuits export %s: %v", action.ID, err)
		}
	}
	assertJSONFileHasKeyE2E(t, cOp, "schema")
	assertJSONFileHasKeyE2E(t, cSpecs, "schema")
	assertJSONFileHasKeyE2E(t, cTiming, "schema")
	assertFileHasE2E(t, cSVG, "<svg")

	// Module 5: comparison view remains evidence-first/read-only.
	cmp := comparisonvm.New()
	cmpSnap := cmp.Snapshot()
	if metricValueAllE2E(cmpSnap, "count") != "3" || !strings.Contains(sectionBodyAllE2E(cmpSnap, "fecim-cim"), "estimated") {
		t.Fatalf("comparison snapshot invalid: %+v", cmpSnap)
	}

	// Module 6: EDA export surfaces are present and carry generated design content.
	eda := edavm.New()
	if err := eda.ApplyAction(viewmodel.Action{ID: "generate_all"}); err != nil {
		t.Fatalf("eda generate_all: %v", err)
	}
	edaSnap := eda.Snapshot()
	for _, id := range []string{"spice_content", "verilog_content", "def_content", "lef_content", "research_validation"} {
		if sectionBodyAllE2E(edaSnap, id) == "" {
			t.Fatalf("EDA section %s missing", id)
		}
	}

	// Module 7: docs search ties generated artifacts back to trust boundaries.
	docs := docsvm.New()
	if err := docs.ApplyAction(viewmodel.Action{ID: "search", Payload: map[string]string{"query": "artifact trust export"}}); err != nil {
		t.Fatalf("docs search: %v", err)
	}
	if err := docs.ApplyAction(viewmodel.Action{ID: "start_curriculum"}); err != nil {
		t.Fatalf("docs start curriculum: %v", err)
	}
	dSnap := docs.Snapshot()
	if metricValueAllE2E(dSnap, "active_page") != "curriculum" || metricValueAllE2E(dSnap, "search_query") != "artifact trust export" || sectionBodyAllE2E(dSnap, "search_results") == "" {
		t.Fatalf("docs snapshot invalid: %+v", dSnap)
	}
}

func assertFileHasE2E(t *testing.T, path, marker string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(raw), marker) {
		t.Fatalf("%s missing %q: %s", path, marker, raw)
	}
}

func assertJSONFileHasKeyE2E(t *testing.T, path, key string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("invalid JSON %s: %v data=%s", path, err, raw)
	}
	if _, ok := payload[key]; !ok {
		t.Fatalf("JSON %s missing key %q: %+v", path, key, payload)
	}
}
