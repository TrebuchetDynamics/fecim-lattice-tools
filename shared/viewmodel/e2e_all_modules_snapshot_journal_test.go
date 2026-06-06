package viewmodel_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
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

type snapshotJournalRecordE2E struct {
	Module     string            `json:"module"`
	Phase      string            `json:"phase"`
	Title      string            `json:"title"`
	Metrics    map[string]string `json:"metrics"`
	Sections   []string          `json:"sections"`
	Actions    []string          `json:"actions"`
	Plots      []string          `json:"plots"`
	TrustWords []string          `json:"trust_words"`
}

func TestAllModulesE2ESnapshotJournalRoundTripAndDeltas(t *testing.T) {
	cases := []struct {
		name         string
		module       viewmodel.ModulePort
		actions      []viewmodel.Action
		wantMetric   map[string]string
		wantSection  string
		wantChanged  string
		wantTrustAny []string
	}{
		{name: "hysteresis", module: hystvm.New(), actions: []viewmodel.Action{{ID: hystvm.EventSetWaveform, Payload: map[string]string{"waveform": hystvm.WaveformSquare}}, {ID: hystvm.EventRunPUND}}, wantMetric: map[string]string{"waveform": "square"}, wantSection: "diagnostic_pund", wantChanged: "waveform", wantTrustAny: []string{"simulation", "educational"}},
		{name: "crossbar", module: crossbarvm.New(6, 6), actions: []viewmodel.Action{{ID: "resize", Payload: map[string]string{"rows": "10", "cols": "6"}}, {ID: "run_mvm"}}, wantMetric: map[string]string{"rows": "10", "cols": "6"}, wantSection: "mvm", wantChanged: "rows", wantTrustAny: []string{"simulation", "educational"}},
		{name: "mnist", module: mnistvm.New(), actions: []viewmodel.Action{{ID: "sweep_levels", Payload: map[string]string{"levels": "64"}}, {ID: "run_inference"}}, wantMetric: map[string]string{"levels": "64 levels"}, wantSection: "research_benchmark", wantChanged: "levels", wantTrustAny: []string{"educational", "simulation"}},
		{name: "circuits", module: circuitsvm.New(), actions: []viewmodel.Action{{ID: circuitsvm.ActionSetOperationMode, Payload: map[string]string{"mode": circuitsvm.OperationCompute}}, {ID: circuitsvm.ActionSetADCBits, Payload: map[string]string{"bits": "8"}}, {ID: circuitsvm.ActionRunCompute}}, wantMetric: map[string]string{"mode": "COMPUTE", "adc": "8-bit SAR"}, wantSection: "compute_run_log", wantChanged: "mode", wantTrustAny: []string{"simulation", "educational"}},
		{name: "comparison", module: comparisonvm.New(), wantMetric: map[string]string{"count": "3"}, wantSection: "fecim-cim", wantChanged: "", wantTrustAny: []string{"estimated", "educational"}},
		{name: "eda", module: edavm.New(), actions: []viewmodel.Action{{ID: "generate_all"}}, wantMetric: map[string]string{"spice": "ready", "verilog": "ready"}, wantSection: "verilog_content", wantChanged: "", wantTrustAny: []string{"educational", "simulation"}},
		{name: "docs", module: docsvm.New(), actions: []viewmodel.Action{{ID: "search", Payload: map[string]string{"query": "snapshot journal trust"}}, {ID: "start_curriculum"}}, wantMetric: map[string]string{"active_page": "curriculum", "search_query": "snapshot journal trust"}, wantSection: "search_results", wantChanged: "active_page", wantTrustAny: []string{"documentation", "educational"}},
	}

	journalPath := filepath.Join(t.TempDir(), "all-modules-snapshot-journal.jsonl")
	journal, err := os.Create(journalPath)
	if err != nil {
		t.Fatalf("create journal: %v", err)
	}
	enc := json.NewEncoder(journal)

	for _, tc := range cases {
		tc.module.Start()
		baseline := tc.module.Snapshot()
		if err := enc.Encode(snapshotJournalRecordFromE2E("baseline", baseline)); err != nil {
			t.Fatalf("encode baseline %s: %v", tc.name, err)
		}
		for _, action := range tc.actions {
			if err := tc.module.ApplyAction(action); err != nil {
				t.Fatalf("%s action %s: %v", tc.name, action.ID, err)
			}
		}
		final := tc.module.Snapshot()
		if err := enc.Encode(snapshotJournalRecordFromE2E("final", final)); err != nil {
			t.Fatalf("encode final %s: %v", tc.name, err)
		}
		tc.module.Stop()

		for id, want := range tc.wantMetric {
			if got := metricValueAllE2E(final, id); !strings.Contains(got, want) {
				t.Fatalf("%s final metric %s=%q want contains %q", tc.name, id, got, want)
			}
		}
		if sectionBodyAllE2E(final, tc.wantSection) == "" {
			t.Fatalf("%s final missing section %s", tc.name, tc.wantSection)
		}
		if tc.wantChanged != "" && metricValueAllE2E(baseline, tc.wantChanged) == metricValueAllE2E(final, tc.wantChanged) {
			t.Fatalf("%s metric %s did not change: %q", tc.name, tc.wantChanged, metricValueAllE2E(final, tc.wantChanged))
		}
		if !noticeHasAnyJournalTrustWordE2E(final.Descriptor.BoundaryNotice, tc.wantTrustAny) {
			t.Fatalf("%s boundary notice %q lacks any of %v", tc.name, final.Descriptor.BoundaryNotice, tc.wantTrustAny)
		}
	}
	if err := journal.Close(); err != nil {
		t.Fatalf("close journal: %v", err)
	}

	records := readSnapshotJournalE2E(t, journalPath)
	if len(records) != len(cases)*2 {
		t.Fatalf("journal records=%d want %d", len(records), len(cases)*2)
	}
	seenPhase := map[string]map[string]bool{}
	for _, rec := range records {
		if rec.Module == "" || rec.Phase == "" || rec.Title == "" || len(rec.Metrics) == 0 || len(rec.Sections) == 0 || len(rec.TrustWords) == 0 {
			t.Fatalf("invalid journal record: %+v", rec)
		}
		if seenPhase[rec.Module] == nil {
			seenPhase[rec.Module] = map[string]bool{}
		}
		if seenPhase[rec.Module][rec.Phase] {
			t.Fatalf("duplicate journal phase for %s/%s", rec.Module, rec.Phase)
		}
		seenPhase[rec.Module][rec.Phase] = true
	}
	for _, desc := range viewmodel.KnownDescriptors() {
		phases := seenPhase[string(desc.ID)]
		if !phases["baseline"] || !phases["final"] {
			t.Fatalf("journal missing phases for %s: %+v", desc.ID, phases)
		}
	}
}

func snapshotJournalRecordFromE2E(phase string, snap viewmodel.ModuleSnapshot) snapshotJournalRecordE2E {
	metrics := map[string]string{}
	for _, m := range snap.Metrics {
		metrics[m.ID] = m.Value
	}
	return snapshotJournalRecordE2E{
		Module:     string(snap.Descriptor.ID),
		Phase:      phase,
		Title:      snap.Descriptor.Title,
		Metrics:    metrics,
		Sections:   sortedSectionIDsJournalE2E(snap),
		Actions:    sortedActionIDsJournalE2E(snap),
		Plots:      sortedPlotIDsJournalE2E(snap),
		TrustWords: trustWordsJournalE2E(snap.Descriptor.BoundaryNotice),
	}
}

func readSnapshotJournalE2E(t *testing.T, path string) []snapshotJournalRecordE2E {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open journal: %v", err)
	}
	defer f.Close()
	var out []snapshotJournalRecordE2E
	s := bufio.NewScanner(f)
	for s.Scan() {
		var rec snapshotJournalRecordE2E
		if err := json.Unmarshal(s.Bytes(), &rec); err != nil {
			t.Fatalf("decode journal line %q: %v", s.Text(), err)
		}
		out = append(out, rec)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("scan journal: %v", err)
	}
	return out
}

func sortedSectionIDsJournalE2E(s viewmodel.ModuleSnapshot) []string {
	ids := make([]string, 0, len(s.Sections))
	for _, sec := range s.Sections {
		ids = append(ids, sec.ID)
	}
	sort.Strings(ids)
	return ids
}

func sortedActionIDsJournalE2E(s viewmodel.ModuleSnapshot) []string {
	ids := make([]string, 0, len(s.Actions))
	for _, action := range s.Actions {
		ids = append(ids, action.ID)
	}
	sort.Strings(ids)
	return ids
}

func sortedPlotIDsJournalE2E(s viewmodel.ModuleSnapshot) []string {
	ids := make([]string, 0, len(s.Plots))
	for _, plot := range s.Plots {
		ids = append(ids, plot.ID)
	}
	sort.Strings(ids)
	return ids
}

func trustWordsJournalE2E(notice string) []string {
	lower := strings.ToLower(notice)
	words := []string{}
	for _, word := range []string{"simulation", "educational", "estimated", "documentation", "not measured", "not validated"} {
		if strings.Contains(lower, word) {
			words = append(words, word)
		}
	}
	return words
}

func noticeHasAnyJournalTrustWordE2E(notice string, words []string) bool {
	lower := strings.ToLower(notice)
	for _, word := range words {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}
