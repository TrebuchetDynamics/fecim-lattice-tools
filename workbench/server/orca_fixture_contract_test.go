//go:build orca_fixture

package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/study"
)

// Fixed expectations are handwritten, never obtained from Analyze or the fixture evaluator.
func TestOrcaFixtureContract(t *testing.T) {
	for _, scenario := range []string{"success", "design-failure", "cancel-resume", "completion-wins", "systemic-failure", "report-failure"} {
		t.Run(scenario, func(t *testing.T) {
			h := newOrcaFixture(t, t.TempDir(), orcaFixtureOrigin(t), orcaFixtureAssets(t))
			defer h.close()
			c := h.client(t)
			p := c.open(t, scenario)
			id := c.start(t, p)
			if scenario == "cancel-resume" || scenario == "systemic-failure" {
				for range 3 {
					h.checkpoint(t, scenario)
					h.command("advance " + scenario)
				}
				if index := h.checkpoint(t, scenario); index != 3 || len(h.records(t, scenario)) != 3 {
					t.Fatalf("checkpoint is not after exactly three commits: index %d", index)
				}
			}
			if scenario == "cancel-resume" {
				ack := c.cancel(t, id)
				if ack.State != "cancelling" || !ack.CancellationRequested {
					t.Fatalf("cancel ack: %+v", ack)
				}
				if got := c.job(t, id); !activeState(got.State) {
					t.Fatalf("unreleased cancellation became terminal: %+v", got)
				}
			}
			h.command("release " + scenario)
			job := c.wait(t, id)
			want := h.expected.Scenarios[scenario]
			assertOrcaJob(t, job, want)
			records := h.records(t, scenario)
			if len(records) != want.Completed {
				t.Fatalf("committed %d, want %d", len(records), want.Completed)
			}
			var page study.ResultsPage
			orcaDecode(t, c.get(t, "/api/v1/projects/"+p.ID+"/runs"), 200, &page)
			membership := map[string]study.ResultRow{}
			for _, row := range page.Rows {
				membership[row.RunID] = row
			}
			for _, r := range records {
				e := h.expected.Records[r.Manifest.PointIndex]
				failed := scenario == "design-failure" && e.Index == 7
				row, ok := membership[r.Manifest.RunID]
				if !ok || row.Feasible != (e.Feasible && !failed) || row.Pareto != (e.Pareto && !failed) {
					t.Fatalf("fixed feasibility/Pareto membership: %+v", row)
				}
				if failed {
					if r.Result.Status != experiment.StatusFailed || r.Result.Failure == nil {
						t.Fatalf("missing committed design failure: %+v", r)
					}
				} else {
					if r.Result.Status != experiment.StatusSuccess || len(r.Result.Metrics) != 3 {
						t.Fatalf("bad record: %+v", r)
					}
					for i, v := range []float64{e.Energy, e.Latency, e.Area} {
						m := r.Result.Metrics[i]
						if m.Value != v || m.Unit != []string{"pJ", "ns", "um2"}[i] || m.Evidence != experiment.EvidenceDefault || m.Model != "SYNTHETIC ORCA FIXTURE v1" || !reflect.DeepEqual(m.Assumptions, []string{"Invented acceptance values; not physics or measurements"}) {
							t.Fatalf("synthetic metric mismatch: %+v", m)
						}
					}
				}
				if r.Manifest.RunID != h.runIDs(t, scenario)[e.Index] {
					t.Fatal("committed scientific ID drift")
				}
			}
			if scenario == "completion-wins" {
				if got := c.cancel(t, id); got.State != "succeeded" || got.CancellationRequested {
					t.Fatalf("completion lost: %+v", got)
				}
			}
			if scenario == "cancel-resume" {
				resumed := c.wait(t, c.start(t, p))
				assertOrcaJob(t, resumed, h.expected.Scenarios["success"])
				if resumed.Progress.Reused != 3 {
					t.Fatalf("resume did not reuse three genuine commits: %+v", resumed)
				}
			}
			if scenario == "report-failure" {
				obstruction, err := os.Stat(h.projectPath(scenario) + "/reports")
				if err != nil || !obstruction.Mode().IsRegular() {
					t.Fatalf("real report writer obstruction missing: %v", err)
				}
				if _, err := os.ReadFile(h.projectPath(scenario) + "/reports/results.json"); err == nil {
					t.Fatal("failed writer fabricated report")
				}
			}
			if scenario == "success" {
				var report struct {
					Runs []json.RawMessage `json:"runs"`
				}
				b, err := os.ReadFile(h.projectPath(scenario) + "/reports/results.json")
				if err != nil || json.Unmarshal(b, &report) != nil || len(report.Runs) != 8 {
					t.Fatalf("real report: %s %v", b, err)
				}
			}
			// A finite SSE reconnect uses the same authenticated server and settled counts.
			res := c.get(t, "/api/v1/experiments/"+id+"/events")
			var event strings.Builder
			scan := bufio.NewScanner(res.Body)
			for scan.Scan() {
				event.WriteString(scan.Text())
				if scan.Text() == "" {
					break
				}
			}
			res.Body.Close()
			if !strings.Contains(event.String(), fmt.Sprintf(`"completed":%d`, want.Completed)) {
				t.Fatalf("SSE counts: %s", event.String())
			}
			h.command("status")
			b, err := os.ReadFile(h.recordPath())
			if err != nil || strings.Contains(string(b), "bootstrap=") || strings.Contains(string(b), c.cookie.Value) {
				t.Fatal("missing identity record or leaked credentials")
			}
			var evidence struct {
				ExpectedHash    string            `json:"expected_sha256"`
				BinaryHash      string            `json:"candidate_binary_sha256"`
				TrackedDiffHash string            `json:"tracked_fixture_diff_sha256"`
				SourceHashes    map[string]string `json:"fixture_source_sha256"`
				Scenarios       map[string]struct {
					Documents map[string]string      `json:"original_document_sha256"`
					Runs      []experiment.RunRecord `json:"committed_runs"`
					Jobs      []moduleapi.Job        `json:"jobs"`
				} `json:"scenarios"`
			}
			if json.Unmarshal(b, &evidence) != nil || evidence.ExpectedHash != orcaHash(h.expectedBytes) || evidence.BinaryHash == "" || len(evidence.Scenarios[scenario].Runs) != len(h.records(t, scenario)) || len(evidence.Scenarios[scenario].Jobs) == 0 || len(evidence.Scenarios[scenario].Documents) != 3 {
				t.Fatal("candidate/project/manifest/results evidence binding missing")
			}
			diff, err := exec.Command("git", "-C", "../..", "diff", "--binary", "HEAD", "--", "workbench/study/orca_fixture.go", "workbench/server/orca_fixture_test.go", "workbench/server/orca_fixture_contract_test.go", "workbench/server/testdata/orca/expected.json", "workbench/server/testdata/orca/README.md").Output()
			if err != nil || evidence.TrackedDiffHash != orcaHash(diff) {
				t.Fatal("evidence does not bind the repository-relative tracked fixture diff")
			}
			for file, want := range evidence.SourceHashes {
				source, err := os.ReadFile(file)
				if err != nil || orcaHash(source) != want {
					t.Fatalf("source binding %s", file)
				}
			}
			if len(evidence.SourceHashes) != 5 {
				t.Fatal("tracked or untracked fixture sources missing from evidence")
			}
		})
	}
	t.Run("real-assets-controls", func(t *testing.T) {
		h := newOrcaFixture(t, t.TempDir(), orcaFixtureOrigin(t), orcaFixtureAssets(t))
		defer h.close()
		c := h.client(t)
		p := c.open(t, "success")
		id := c.start(t, p)
		h.checkpoint(t, "success")
		h.command("advance success")
		h.checkpoint(t, "success")
		if h.command("renew") == "" {
			t.Fatal("no renewed launch link")
		}
		if got := c.get(t, "/"); got.StatusCode != 200 {
			t.Fatalf("actual asset serving: %d", got.StatusCode)
		} else {
			got.Body.Close()
		}
		h.command("reset") // Must cancel, release and join the checkpointed worker before replacing roots.
		if h.oldJob(id).State != "cancelled" {
			t.Fatal("reset did not settle active job")
		}
		c = h.client(t)
		p = c.open(t, "success")
		id = c.start(t, p)
		h.checkpoint(t, "success")
		h.command("shutdown")
		if h.oldJob(id).State != "cancelled" {
			t.Fatal("shutdown did not settle worker")
		}
	})
}

func assertOrcaJob(t *testing.T, got moduleapi.Job, want orcaExpectedScenario) {
	t.Helper()
	if got.State != want.State || got.Progress.Total != 8 || got.Progress.Completed != want.Completed || got.Progress.Failed != want.Failed || got.Progress.Cancelled != want.Cancelled || got.ResultsState != want.Results || !got.AnalysisAvailable || got.InputRevision == "" || got.ResultsRevision == "" || !reflect.DeepEqual(got.Counts, want.Counts) {
		t.Fatalf("terminal job: %+v; want %+v", got, want)
	}
}
