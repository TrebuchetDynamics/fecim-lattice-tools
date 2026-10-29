package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/fecim"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/study"
)

func TestWebResearchWorkflow(t *testing.T) {
	s := testServer(t)
	cookie := login(t, s)
	roots := request(s, "GET", "/api/v1/roots", "", cookie, nil)
	var grants struct {
		Roots []struct {
			ID string `json:"id"`
		} `json:"roots"`
	}
	if err := json.Unmarshal(roots.Body.Bytes(), &grants); err != nil || len(grants.Roots) != 1 {
		t.Fatalf("roots: %s", roots.Body)
	}
	body, _ := json.Marshal(map[string]string{"root": grants.Roots[0].ID, "name": "eight-point"})
	created := request(s, "POST", "/api/v1/projects", string(body), cookie, nil)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body)
	}
	var p struct {
		ID       string `json:"id"`
		Revision string `json:"revision"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &p); err != nil || p.ID == "" || p.Revision == "" {
		t.Fatalf("project: %s", created.Body)
	}
	base := "/api/v1/projects/" + p.ID
	get := request(s, "GET", base+"/configuration", "", cookie, nil)
	if get.Code != 200 || get.Header().Get("ETag") != p.Revision {
		t.Fatalf("configuration: %d %s", get.Code, get.Body)
	}
	patch := `{"document":"project.yaml","changes":[{"path":"/name","value":"Edited study"}]}`
	if w := request(s, "PUT", base+"/configuration", patch, cookie, nil); w.Code != 428 {
		t.Fatalf("missing If-Match: %d %s", w.Code, w.Body)
	}
	saved := request(s, "PUT", base+"/configuration", patch, cookie, map[string]string{"If-Match": p.Revision})
	if saved.Code != 200 || saved.Header().Get("ETag") == p.Revision {
		t.Fatalf("save: %d %s", saved.Code, saved.Body)
	}
	stale := request(s, "PUT", base+"/configuration", patch, cookie, map[string]string{"If-Match": p.Revision})
	if stale.Code != 412 || stale.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("stale: %d %s", stale.Code, stale.Body)
	}
	numericConflict := request(s, "PUT", base+"/configuration", `{"document":"design.yaml","changes":[{"path":"/array/read_voltage_v","value":"0.3"}]}`, cookie, map[string]string{"If-Match": p.Revision})
	var conflict struct {
		Fields []struct {
			Unit               string                     `json:"unit"`
			OnDiskPresentation struct{ Speakable string } `json:"on_disk_presentation"`
		} `json:"fields"`
	}
	json.Unmarshal(numericConflict.Body.Bytes(), &conflict)
	if numericConflict.Code != 412 || len(conflict.Fields) != 1 || conflict.Fields[0].Unit != "V" || !strings.Contains(conflict.Fields[0].OnDiskPresentation.Speakable, "volts") {
		t.Fatalf("Go conflict units/speech: %d %s", numericConflict.Code, numericConflict.Body)
	}
	invalid := request(s, "PUT", base+"/configuration", `{"document":"design.yaml","changes":[{"path":"/array/read_voltage_v","value":"not-a-voltage"}]}`, cookie, map[string]string{"If-Match": saved.Header().Get("ETag")})
	var invalidProblem struct {
		Issues []struct{ Document, Path, Label, Reason string } `json:"issues"`
	}
	json.Unmarshal(invalid.Body.Bytes(), &invalidProblem)
	if invalid.Code != 400 || len(invalidProblem.Issues) != 1 || invalidProblem.Issues[0].Document != "design.yaml" || invalidProblem.Issues[0].Path != "/array/read_voltage_v" || invalidProblem.Issues[0].Label != "Read voltage" || invalidProblem.Issues[0].Reason == "" || strings.Contains(invalid.Body.String(), "not-a-voltage") {
		t.Fatalf("safe field validation issue missing: %d %s", invalid.Code, invalid.Body)
	}
	validated := request(s, "POST", base+"/validate", `{}`, cookie, nil)
	var validation struct {
		Points int `json:"points"`
	}
	json.Unmarshal(validated.Body.Bytes(), &validation)
	if validated.Code != 200 || validation.Points != 8 {
		t.Fatalf("validation: %d %s", validated.Code, validated.Body)
	}
	// Opening the same directory returns the same exclusion/handle, not a second editor.
	opened := request(s, "POST", "/api/v1/projects/open", `{"root":"`+grants.Roots[0].ID+`","path":"eight-point"}`, cookie, nil)
	var duplicate struct {
		ID string `json:"id"`
	}
	json.Unmarshal(opened.Body.Bytes(), &duplicate)
	if opened.Code != 200 || duplicate.ID != p.ID {
		t.Fatalf("duplicate open: %d %s", opened.Code, opened.Body)
	}
	wait := func(id string) experiment.ProgressSnapshot {
		t.Helper()
		deadline := time.Now().Add(20 * time.Second)
		for time.Now().Before(deadline) {
			w := request(s, "GET", "/api/v1/experiments/"+id, "", cookie, nil)
			var job struct {
				State    string                      `json:"state"`
				Progress experiment.ProgressSnapshot `json:"progress"`
			}
			json.Unmarshal(w.Body.Bytes(), &job)
			if w.Code != 200 {
				t.Fatalf("job: %d %s", w.Code, w.Body)
			}
			if job.State == "succeeded" {
				return job.Progress
			}
			if job.State == "failed" || job.State == "cancelled" {
				t.Fatalf("job failed: %s", w.Body)
			}
			time.Sleep(5 * time.Millisecond)
		}
		t.Fatal("job timeout")
		return experiment.ProgressSnapshot{}
	}
	start := func(key string) string {
		t.Helper()
		w := request(s, "POST", base+"/experiments", `{"workers":2}`, cookie, map[string]string{"Idempotency-Key": key, "If-Match": saved.Header().Get("ETag")})
		var job struct {
			ID string `json:"id"`
		}
		json.Unmarshal(w.Body.Bytes(), &job)
		if w.Code != 202 || job.ID == "" {
			t.Fatalf("start: %d %s", w.Code, w.Body)
		}
		return job.ID
	}
	if w := request(s, "POST", base+"/experiments", `{"workers":2}`, cookie, map[string]string{"Idempotency-Key": "missing-revision"}); w.Code != 428 {
		t.Fatalf("start missing revision: %d", w.Code)
	}
	if w := request(s, "POST", base+"/experiments", `{"workers":2}`, cookie, map[string]string{"Idempotency-Key": "stale-revision", "If-Match": p.Revision}); w.Code != 412 {
		t.Fatalf("start stale revision: %d", w.Code)
	}
	// Hold the actual rooted runner after acceptance, then change the files as
	// an external editor would. Accepted run identities/analysis must remain bound
	// to the validated snapshot, not a second asynchronous filesystem load.
	normalRun := s.runBundle
	entered, resume := make(chan struct{}), make(chan struct{})
	s.runBundle = func(ctx context.Context, root, citations *os.Root, bundle project.Bundle, opts study.RunOptions) (study.RunOutcome, error) {
		close(entered)
		<-resume
		return normalRun(ctx, root, citations, bundle, opts)
	}
	acceptedRoot := s.projects[p.ID].root
	acceptedBundle, _, err := study.ValidateProjectRoot(acceptedRoot, nil)
	if err != nil {
		t.Fatal(err)
	}
	originalSweep, err := acceptedRoot.ReadFile("sweep.yaml")
	if err != nil {
		t.Fatal(err)
	}
	first := start("first")
	<-entered
	if err := acceptedRoot.WriteFile("sweep.yaml", []byte(strings.Replace(string(originalSweep), "seed: 17", "seed: 999", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	close(resume)
	if again := start("first"); again != first {
		t.Fatal("idempotency did not return original execution")
	}
	if progress := wait(first); progress.Total != 8 || progress.Completed != 8 || progress.Failed != 0 {
		t.Fatalf("terminal counts: %+v", progress)
	}
	s.workers.Wait()
	s.runBundle = normalRun
	terminalReply := request(s, "GET", "/api/v1/experiments/"+first, "", cookie, nil)
	var firstJob moduleapi.Job
	json.Unmarshal(terminalReply.Body.Bytes(), &firstJob)
	acceptedRecords, err := experiment.LoadRunsRoot(acceptedRoot)
	if err != nil {
		t.Fatal(err)
	}
	expectedPoints, err := experiment.Expand(acceptedBundle)
	if err != nil {
		t.Fatal(err)
	}
	expectedIDs := map[string]bool{}
	for _, point := range expectedPoints {
		id, err := experiment.ID(point, fecim.EvaluatorVersion, acceptedBundle.Inputs)
		if err != nil {
			t.Fatal(err)
		}
		expectedIDs[id] = true
	}
	for _, record := range acceptedRecords {
		if !expectedIDs[record.Manifest.RunID] {
			t.Fatalf("unexpected accepted run ID: %s", record.Manifest.RunID)
		}
	}
	// Verify exact scientific identities via a read-only analysis of accepted data.
	expectedAnalysis := experiment.Analyze(acceptedBundle, acceptedRecords)
	expectedPage, err := study.QueryResults(expectedAnalysis, study.ResultsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if firstJob.InputRevision != saved.Header().Get("ETag") || !firstJob.AnalysisAvailable || firstJob.ResultsRevision != expectedPage.Revision || !reflect.DeepEqual(firstJob.Counts, expectedAnalysis.Counts) || !strings.Contains(firstJob.Message, "8 of 8") {
		t.Fatalf("authoritative terminal snapshot: %+v", firstJob)
	}
	// A later configuration/analysis read cannot rewrite the retained job counts.
	originalProject, err := acceptedRoot.ReadFile("project.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := acceptedRoot.WriteFile("project.yaml", []byte(strings.Replace(string(originalProject), "value: 5000", "value: -1", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	request(s, "GET", base+"/analysis", "", cookie, nil)
	var retained moduleapi.Job
	json.Unmarshal(request(s, "GET", "/api/v1/experiments/"+first, "", cookie, nil).Body.Bytes(), &retained)
	if !reflect.DeepEqual(firstJob, retained) {
		t.Fatal("later configuration changed terminal job analysis")
	}
	if err := acceptedRoot.WriteFile("sweep.yaml", originalSweep, 0600); err != nil {
		t.Fatal(err)
	}
	if err := acceptedRoot.WriteFile("project.yaml", originalProject, 0600); err != nil {
		t.Fatal(err)
	}
	if progress := wait(start("reuse")); progress.Completed != 8 || progress.Reused != 8 {
		t.Fatalf("reuse counts: %+v", progress)
	}
	// Reconnect returns the current counts, without replay promises or paths/job
	// errors. Terminal lifecycle remains independently readable via GET state.
	for i := 0; i < 2; i++ {
		r := httptest.NewRequest("GET", testOrigin+"/api/v1/experiments/"+first+"/events", nil)
		r.AddCookie(cookie)
		w := &deadlineRecorder{httptest.NewRecorder()}
		s.ServeHTTP(w, r)
		if w.Code != 200 || w.Header().Get("Content-Type") != "text/event-stream" || !strings.Contains(w.Body.String(), `"completed":8`) || strings.Contains(w.Body.String(), first) || strings.Contains(w.Body.String(), "message") {
			t.Fatalf("reconnect counts: %d %s", w.Code, w.Body)
		}
	}
	// GET must not call report generation, restore missing artifacts, or touch metadata.
	root := s.roots[grants.Roots[0].ID]
	reportPath := filepath.Join(root.Name(), "eight-point", "reports", "results.json")
	if err := os.Remove(reportPath); err != nil {
		t.Fatal(err)
	}
	before := fileSnapshot(t, filepath.Join(root.Name(), "eight-point"))
	runs := request(s, "GET", base+"/runs?page_size=2&sort=energy_pj", "", cookie, nil)
	var page study.ResultsPage
	json.Unmarshal(runs.Body.Bytes(), &page)
	if runs.Code != 200 || page.Total != 8 || len(page.Rows) != 2 || page.Rows[0].RunID == page.Rows[1].RunID {
		t.Fatalf("runs: %d %s", runs.Code, runs.Body)
	}
	for _, endpoint := range []string{base, base + "/runs/" + page.Rows[0].RunID, base + "/analysis?compare=" + page.Rows[0].RunID + "," + page.Rows[1].RunID, base + "/experiments?state=active"} {
		if w := request(s, "GET", endpoint, "", cookie, nil); w.Code != 200 {
			t.Fatalf("read %s: %d %s", endpoint, w.Code, w.Body)
		}
	}
	if after := fileSnapshot(t, filepath.Join(root.Name(), "eight-point")); !reflect.DeepEqual(before, after) {
		t.Fatal("GET changed project files")
	}
	for _, endpoint := range []string{"verify", "reports"} {
		if w := request(s, "POST", base+"/"+endpoint, `{}`, cookie, nil); w.Code != 200 {
			t.Fatalf("%s: %d %s", endpoint, w.Code, w.Body)
		}
	}
	if w := request(s, "GET", base+"/reports/results.json", "", cookie, nil); w.Code != 200 {
		t.Fatalf("report download: %d %s", w.Code, w.Body)
	}
	if w := request(s, "POST", base+"/exports", `{"run_id":"`+page.Rows[0].RunID+`"}`, cookie, nil); w.Code != 200 {
		t.Fatalf("export: %d %s", w.Code, w.Body)
	}
	if w := request(s, "GET", base+"/exports/"+page.Rows[0].RunID+"/manifest.json", "", cookie, nil); w.Code != 200 {
		t.Fatalf("export download: %d %s", w.Code, w.Body)
	}
	// Export downloads must verify the enumerated bundle, not just a static filename.
	if err := root.WriteFile(filepath.Join("eight-point", "exports", page.Rows[0].RunID, "design.spice"), []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if w := request(s, "GET", base+"/exports/"+page.Rows[0].RunID+"/design.spice", "", cookie, nil); w.Code != 409 {
		t.Fatalf("tampered download: %d %s", w.Code, w.Body)
	}
	// Cancellation acknowledgement is not terminal cancellation. A fast finished
	// study may succeed; otherwise cancellation must settle with exact counts.
	saved = request(s, "PUT", base+"/configuration", `{"document":"sweep.yaml","changes":[{"path":"/seed","value":"18"}]}`, cookie, map[string]string{"If-Match": saved.Header().Get("ETag")})
	if saved.Code != 200 {
		t.Fatalf("seed patch: %s", saved.Body)
	}
	cancelledID := start("cancel")
	if w := request(s, "POST", "/api/v1/experiments/"+cancelledID+"/cancel", `{}`, cookie, nil); w.Code != 200 {
		t.Fatalf("cancel ack: %d %s", w.Code, w.Body)
	}
	deadline := time.Now().Add(20 * time.Second)
	for {
		w := request(s, "GET", "/api/v1/experiments/"+cancelledID, "", cookie, nil)
		var job moduleapi.Job
		json.Unmarshal(w.Body.Bytes(), &job)
		if !activeState(job.State) {
			if job.State != "cancelled" && job.State != "succeeded" {
				t.Fatalf("cancel terminal: %s", w.Body)
			}
			if job.Progress.Completed+job.Progress.Cancelled != 8 {
				t.Fatalf("cancel settled counts: %+v", job.Progress)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("cancellation did not settle")
		}
		time.Sleep(time.Millisecond)
	}
	s.workers.Wait()
	// Fault the orchestration boundary after real rooted records are produced.
	// Cancellation plus a systemic/integrity failure must not advertise a plain
	// cancellation or usable analysis. Late pure cancellation may be complete.
	for _, mode := range []string{"late cancellation", "cancelled integrity"} {
		s.runBundle = func(ctx context.Context, r, c *os.Root, b project.Bundle, opts study.RunOptions) (study.RunOutcome, error) {
			out, err := normalRun(ctx, r, c, b, opts)
			if err != nil {
				return out, err
			}
			if mode == "cancelled integrity" {
				out.AnalysisAvailable = false
				out.Analysis = experiment.Analysis{}
				out.ResultsState = study.ResultsPartial
				return out, errors.Join(context.Canceled, experiment.ErrUnverifiableRun)
			}
			return out, context.Canceled
		}
		id := start(strings.ReplaceAll(mode, " ", "-"))
		deadline := time.Now().Add(10 * time.Second)
		var terminal moduleapi.Job
		for {
			json.Unmarshal(request(s, "GET", "/api/v1/experiments/"+id, "", cookie, nil).Body.Bytes(), &terminal)
			if !activeState(terminal.State) {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("terminal fault timeout")
			}
			time.Sleep(time.Millisecond)
		}
		s.workers.Wait()
		if mode == "late cancellation" {
			if terminal.State != "cancelled" || terminal.ResultsState != "complete" || terminal.Progress.Cancelled != 0 || !terminal.AnalysisAvailable || terminal.ResultsRevision == "" {
				t.Fatalf("late cancellation: %+v", terminal)
			}
		} else {
			if terminal.State != "failed" || terminal.AnalysisAvailable || terminal.Counts != nil || terminal.ResultsRevision != "" {
				t.Fatalf("joined cancellation/integrity: %+v", terminal)
			}
		}
	}
	s.runBundle = normalRun
	if w := request(s, "POST", base+"/close", `{}`, cookie, nil); w.Code != 200 {
		t.Fatalf("close: %d %s", w.Code, w.Body)
	}
	if w := request(s, "GET", base, "", cookie, nil); w.Code != 404 {
		t.Fatalf("closed handle: %d", w.Code)
	}
}

func fileSnapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	if err := filepath.WalkDir(dir, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() {
			info, e := entry.Info()
			if e != nil {
				return e
			}
			data, e := os.ReadFile(name)
			if e != nil {
				return e
			}
			out[name] = info.ModTime().String() + string(data)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return out
}

// Recorder with net/http-equivalent deadline support for finite SSE responses.
type deadlineRecorder struct{ *httptest.ResponseRecorder }

func (w *deadlineRecorder) SetWriteDeadline(time.Time) error { return nil }
