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

type allModulesEvidenceManifestE2E struct {
	Schema  string                        `json:"schema"`
	Modules []allModulesEvidenceRecordE2E `json:"modules"`
}

type allModulesEvidenceRecordE2E struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	BoundaryNotice string   `json:"boundary_notice"`
	MetricCount    int      `json:"metric_count"`
	SectionCount   int      `json:"section_count"`
	ActionCount    int      `json:"action_count"`
	PlotCount      int      `json:"plot_count"`
	SectionIDs     []string `json:"section_ids"`
	ActionIDs      []string `json:"action_ids"`
	PlotIDs        []string `json:"plot_ids"`
}

func TestAllModulesE2EEvidenceManifestArtifactCoversTrustAndSurfaces(t *testing.T) {
	modules := []viewmodel.ModulePort{
		hystvm.New(),
		crossbarvm.New(10, 6),
		mnistvm.New(),
		circuitsvm.New(),
		comparisonvm.New(),
		edavm.New(),
		docsvm.New(),
	}
	preps := map[viewmodel.ModuleID][]viewmodel.Action{
		viewmodel.ModuleHysteresis: {{ID: hystvm.EventRunPUND}, {ID: hystvm.EventRunFORC}, {ID: hystvm.EventRunLevelCalibration}},
		viewmodel.ModuleCrossbar:   {{ID: "toggle_ir"}, {ID: "run_mvm"}},
		viewmodel.ModuleMNIST:      {{ID: "sweep_levels", Payload: map[string]string{"levels": "30"}}, {ID: "run_inference"}},
		viewmodel.ModuleCircuits:   {{ID: circuitsvm.ActionRunRead}, {ID: circuitsvm.ActionRunWrite}, {ID: circuitsvm.ActionRunCompute}},
		viewmodel.ModuleEDA:        {{ID: "generate_all"}},
		viewmodel.ModuleDocs:       {{ID: "search", Payload: map[string]string{"query": "trust evidence manifest"}}, {ID: "start_curriculum"}},
	}

	manifest := allModulesEvidenceManifestE2E{Schema: "fecim.all_modules.e2e.evidence_manifest.v1"}
	for _, module := range modules {
		module.Start()
		defer module.Stop()
		id := module.Descriptor().ID
		for _, action := range preps[id] {
			if err := module.ApplyAction(action); err != nil {
				t.Fatalf("%s prep %s: %v", id, action.ID, err)
			}
		}
		snap := module.Snapshot()
		assertManifestSnapshotTrustE2E(t, snap)
		manifest.Modules = append(manifest.Modules, allModulesEvidenceRecordE2E{
			ID:             string(snap.Descriptor.ID),
			Title:          snap.Descriptor.Title,
			BoundaryNotice: snap.Descriptor.BoundaryNotice,
			MetricCount:    len(snap.Metrics),
			SectionCount:   len(snap.Sections),
			ActionCount:    len(snap.Actions),
			PlotCount:      len(snap.Plots),
			SectionIDs:     sectionIDsManifestE2E(snap),
			ActionIDs:      actionIDsManifestE2E(snap),
			PlotIDs:        plotIDsManifestE2E(snap),
		})
	}
	if len(manifest.Modules) != len(viewmodel.KnownDescriptors()) {
		t.Fatalf("manifest modules=%d", len(manifest.Modules))
	}

	path := filepath.Join(t.TempDir(), "all-modules-evidence-manifest.json")
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	decodedRaw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var decoded allModulesEvidenceManifestE2E
	if err := json.Unmarshal(decodedRaw, &decoded); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	assertManifestRoundTripE2E(t, decoded)
}

func assertManifestSnapshotTrustE2E(t *testing.T, snap viewmodel.ModuleSnapshot) {
	t.Helper()
	if snap.Descriptor.ID == "" || snap.Descriptor.Title == "" || snap.Descriptor.Status != viewmodel.StatusFunctional {
		t.Fatalf("bad descriptor: %+v", snap.Descriptor)
	}
	lowerNotice := strings.ToLower(snap.Descriptor.BoundaryNotice)
	if lowerNotice == "" {
		t.Fatalf("%s missing boundary notice", snap.Descriptor.ID)
	}
	if !strings.Contains(lowerNotice, "simulation") && !strings.Contains(lowerNotice, "educational") && !strings.Contains(lowerNotice, "estimated") && !strings.Contains(lowerNotice, "documentation") {
		t.Fatalf("%s boundary notice lacks evidence qualifier: %q", snap.Descriptor.ID, snap.Descriptor.BoundaryNotice)
	}
	if strings.Contains(lowerNotice, "production silicon") && !strings.Contains(lowerNotice, "not") {
		t.Fatalf("%s unsafe production silicon notice: %q", snap.Descriptor.ID, snap.Descriptor.BoundaryNotice)
	}
	if len(snap.Metrics) == 0 || len(snap.Sections) == 0 {
		t.Fatalf("%s insufficient observable surface", snap.Descriptor.ID)
	}
}

func assertManifestRoundTripE2E(t *testing.T, manifest allModulesEvidenceManifestE2E) {
	t.Helper()
	if manifest.Schema != "fecim.all_modules.e2e.evidence_manifest.v1" {
		t.Fatalf("schema=%q", manifest.Schema)
	}
	seen := map[string]bool{}
	for _, rec := range manifest.Modules {
		if rec.ID == "" || rec.Title == "" || rec.BoundaryNotice == "" || rec.MetricCount <= 0 || rec.SectionCount <= 0 {
			t.Fatalf("invalid manifest record: %+v", rec)
		}
		if seen[rec.ID] {
			t.Fatalf("duplicate manifest module %s", rec.ID)
		}
		seen[rec.ID] = true
		if len(rec.SectionIDs) != rec.SectionCount || len(rec.ActionIDs) != rec.ActionCount || len(rec.PlotIDs) != rec.PlotCount {
			t.Fatalf("manifest count mismatch for %s: %+v", rec.ID, rec)
		}
	}
	for _, desc := range viewmodel.KnownDescriptors() {
		if !seen[string(desc.ID)] {
			t.Fatalf("manifest missing module %s", desc.ID)
		}
	}
}

func sectionIDsManifestE2E(s viewmodel.ModuleSnapshot) []string {
	ids := make([]string, 0, len(s.Sections))
	for _, sec := range s.Sections {
		ids = append(ids, sec.ID)
	}
	return ids
}

func actionIDsManifestE2E(s viewmodel.ModuleSnapshot) []string {
	ids := make([]string, 0, len(s.Actions))
	for _, action := range s.Actions {
		ids = append(ids, action.ID)
	}
	return ids
}

func plotIDsManifestE2E(s viewmodel.ModuleSnapshot) []string {
	ids := make([]string, 0, len(s.Plots))
	for _, plot := range s.Plots {
		ids = append(ids, plot.ID)
	}
	return ids
}
