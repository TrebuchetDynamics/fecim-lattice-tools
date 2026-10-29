//go:build orca_fixture

package server

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"fecim-lattice-tools/internal/browserui"
	"fecim-lattice-tools/workbench/experiment"
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/report"
	"fecim-lattice-tools/workbench/study"
)

var orcaScenarios = []string{"success", "design-failure", "cancel-resume", "completion-wins", "systemic-failure", "report-failure"}

type orcaExpectedRecord struct {
	Index    int     `json:"index"`
	Energy   float64 `json:"energy_pj"`
	Latency  float64 `json:"latency_ns"`
	Area     float64 `json:"area_um2"`
	Feasible bool    `json:"feasible"`
	Pareto   bool    `json:"pareto"`
}
type orcaExpectedScenario struct {
	State     string         `json:"state"`
	Completed int            `json:"completed"`
	Failed    int            `json:"failed"`
	Cancelled int            `json:"cancelled"`
	Results   string         `json:"results"`
	Counts    map[string]int `json:"counts"`
}
type orcaExpectations struct {
	Revision  string                          `json:"revision"`
	Records   []orcaExpectedRecord            `json:"records"`
	Scenarios map[string]orcaExpectedScenario `json:"scenarios"`
	RunIDs    map[string][]string             `json:"run_ids"`
}

// Cancellation deliberately does not unblock this gate. Acknowledgement can be
// observed while the real worker still owns its root. Release permits settlement.
type orcaGate struct {
	mu       sync.Mutex
	index    int
	waiting  bool
	advance  chan struct{}
	released chan struct{}
	entered  chan int
	once     sync.Once
}

func newOrcaGate() *orcaGate {
	return &orcaGate{advance: make(chan struct{}, 1), released: make(chan struct{}), entered: make(chan int, 8)}
}
func (g *orcaGate) release() { g.once.Do(func() { close(g.released) }) }
func (g *orcaGate) evaluate(ctx context.Context, scenario string, d project.Design, seed int64) (experiment.Result, error) {
	index := 0
	if d.Device.ConductanceLevels == 30 {
		index += 4
	}
	if d.Array.Rows == 64 {
		index += 2
	}
	if d.Circuit.ADCBits == 6 {
		index++
	}
	if seed != int64(17+index) {
		return experiment.Result{}, errors.New("fixture input/seed mismatch")
	}
	g.mu.Lock()
	g.index = index
	g.waiting = true
	g.mu.Unlock()
	select {
	case g.entered <- index:
	default:
	}
	select {
	case <-g.advance:
	case <-g.released:
	}
	g.mu.Lock()
	g.waiting = false
	g.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return experiment.Result{}, err
	}
	if scenario == "systemic-failure" && index == 3 {
		return experiment.Result{}, errors.New("SYNTHETIC systemic evaluator failure")
	}
	if scenario == "design-failure" && index == 7 {
		return experiment.Result{Status: experiment.StatusFailed, Failure: &experiment.Failure{Kind: "synthetic-domain", Message: "SYNTHETIC intentional design-point failure"}}, nil
	}
	// Independently defined arithmetic, not read from the expected-value manifest.
	area := float64((index + 1) * 100)
	if index >= 6 {
		area = float64(index * 1000)
	}
	metrics := []experiment.Metric{}
	for i, name := range []string{"energy_pj", "latency_ns", "area_um2"} {
		metrics = append(metrics, experiment.Metric{Name: name, Value: []float64{float64(index + 1), float64(index + 10), area}[i], Unit: []string{"pJ", "ns", "um2"}[i], Model: "SYNTHETIC ORCA FIXTURE v1", Assumptions: []string{"Invented acceptance values; not physics or measurements"}, Evidence: experiment.EvidenceDefault})
	}
	return experiment.Result{Status: experiment.StatusSuccess, Metrics: metrics}, nil
}

type orcaFixture struct {
	t                    *testing.T
	base, origin, assets string
	generation           int
	mu                   sync.RWMutex
	s                    *Server
	old                  *Server
	gates                map[string]*orcaGate
	http                 *http.Server
	listener             net.Listener
	serveDone            chan error
	expected             orcaExpectations
	expectedBytes        []byte
	manifest             Manifest
	closed               bool
	projectNames         map[string]string
	binaryHash           string
	originalHashes       map[string]map[string]string
}

func orcaFixtureAssets(t *testing.T) string {
	t.Helper()
	path := os.Getenv("FECIM_ORCA_ASSETS")
	if path == "" {
		path = "../../internal/browserui/assets"
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
func orcaFixtureOrigin(t *testing.T) string {
	t.Helper()
	port := os.Getenv("FECIM_ORCA_PORT")
	if port == "" {
		port = os.Getenv("AUTOWORK_PORT_BASE")
	}
	if port == "" {
		t.Fatal("set FECIM_ORCA_PORT to an assigned private port (see testdata/orca/README.md)")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		t.Fatal("invalid FECIM_ORCA_PORT")
	}
	if base := os.Getenv("AUTOWORK_PORT_BASE"); base != "" {
		b, _ := strconv.Atoi(base)
		if n < b || n > b+99 {
			t.Fatal("port outside assigned range")
		}
	}
	return "http://127.0.0.1:" + port
}
func newOrcaFixture(t *testing.T, base, origin, assets string) *orcaFixture {
	t.Helper()
	h := &orcaFixture{t: t, base: base, origin: origin, assets: assets, projectNames: map[string]string{}}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	h.binaryHash = orcaHash(binary)
	h.expectedBytes, err = os.ReadFile("testdata/orca/expected.json")
	if err != nil || json.Unmarshal(h.expectedBytes, &h.expected) != nil {
		t.Fatalf("fixed manifest: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(assets, "local-manifest.json"))
	if err != nil || json.Unmarshal(b, &h.manifest) != nil {
		t.Fatalf("build real local assets first (README): %v", err)
	}
	if h.manifest.Identity != hysteresis.BridgeIdentity() || h.manifest.UI != browserui.UIID {
		t.Fatal("candidate/linker identity mismatch; follow README matching -ldflags")
	}
	h.initialize()
	h.listener, err = net.Listen("tcp", strings.TrimPrefix(origin, "http://"))
	if err != nil {
		h.close()
		t.Fatal(err)
	}
	h.http = &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.mu.RLock()
		s := h.s
		h.mu.RUnlock()
		s.ServeHTTP(w, r)
	}), ReadHeaderTimeout: 5 * time.Second}
	h.serveDone = make(chan error, 1)
	go func() { h.serveDone <- h.http.Serve(h.listener) }()
	return h
}
func (h *orcaFixture) projectPath(scenario string) string {
	return filepath.Join(h.base, fmt.Sprintf("generation-%d", h.generation), "granted", scenario)
}
func (h *orcaFixture) initialize() {
	h.generation++
	granted := filepath.Dir(h.projectPath("success"))
	if err := os.MkdirAll(granted, 0700); err != nil {
		h.t.Fatal(err)
	}
	// Outside exists for negative path selection; it is never granted to the server.
	if err := os.MkdirAll(filepath.Join(h.base, "outside-ungranted"), 0700); err != nil {
		h.t.Fatal(err)
	}
	h.gates = map[string]*orcaGate{}
	h.originalHashes = map[string]map[string]string{}
	for _, name := range orcaScenarios {
		path := h.projectPath(name)
		if err := study.CreateStarterProject(path, study.DefaultStarterSweepOptions()); err != nil {
			h.t.Fatal(err)
		}
		data := fmt.Sprintf("schema_version: 1\nid: orca-fixture\nname: SYNTHETIC ORCA FIXTURE - %s\nhypothesis: Invented acceptance values, not physics or measurements.\nmodel_version: orca-fixture-v1-%s\nobjectives:\n  - metric: energy_pj\n    direction: minimize\n  - metric: latency_ns\n    direction: minimize\nconstraints:\n  - metric: area_um2\n    operator: '<='\n    value: 5000\n    unit: um2\ncitations: []\ninputs: []\n", name, name)
		if err := os.WriteFile(filepath.Join(path, "project.yaml"), []byte(data), 0600); err != nil {
			h.t.Fatal(err)
		}
		h.originalHashes[name] = map[string]string{}
		for _, file := range []string{"project.yaml", "design.yaml", "sweep.yaml"} {
			b, err := os.ReadFile(filepath.Join(path, file))
			if err != nil {
				h.t.Fatal(err)
			}
			h.originalHashes[name][file] = orcaHash(b)
		}
		h.gates[name] = newOrcaGate()
	}
	s, err := New(Config{Origin: h.origin, Assets: os.DirFS(h.assets), UI: browserui.UIID, Identity: hysteresis.BridgeIdentity(), Roots: []string{granted}})
	if err != nil {
		h.t.Fatal(err)
	}
	gates := h.gates
	// Assign once, before publishing the server or accepting any requests.
	s.runBundle = func(ctx context.Context, r, c *os.Root, b project.Bundle, opts study.RunOptions) (study.RunOutcome, error) {
		name := strings.TrimPrefix(b.Project.ModelVersion, "orca-fixture-v1-")
		g := gates[name]
		if g == nil {
			return study.RunOutcome{}, errors.New("fixture accepts only generated synthetic projects")
		}
		opts.Workers = 1
		writer := report.GenerateRoot
		if name == "report-failure" {
			writer = func(root *os.Root, b project.Bundle, a experiment.Analysis) error {
				// Make a genuine filesystem error inside the actual report writer: the
				// reports directory cannot be created over this regular file.
				if err := root.WriteFile("reports", []byte("SYNTHETIC writer obstruction\n"), 0600); err != nil {
					return err
				}
				return report.GenerateRoot(root, b, a)
			}
		}
		return study.RunOrcaFixtureBundleRoot(ctx, r, c, b, opts, func(ctx context.Context, d project.Design, seed int64) (experiment.Result, error) {
			return g.evaluate(ctx, name, d, seed)
		}, writer)
	}
	h.mu.Lock()
	h.s = s
	h.mu.Unlock()
	h.writeIdentity()
}
func (h *orcaFixture) settle() {
	// Revoke admission and cancel BEFORE opening the gates; no fresh jobs can race
	// the WaitGroup join. Close also drains requests/streams and releases roots.
	h.s.mu.Lock()
	for id, p := range h.s.projects {
		h.projectNames[id] = p.path
	}
	h.s.draining = true
	h.s.cancelEpoch()
	h.s.mu.Unlock()
	for _, g := range h.gates {
		g.release()
	}
	h.s.Close()
	h.old = h.s
}
func (h *orcaFixture) close() {
	if h.closed {
		return
	}
	h.closed = true
	h.settle()
	h.writeIdentity()
	if h.http != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.http.Shutdown(ctx); err != nil {
			h.t.Error(err)
			_ = h.http.Close()
		}
		if err := <-h.serveDone; !errors.Is(err, http.ErrServerClosed) {
			h.t.Errorf("fixture serve: %v", err)
		}
	}
}
func (h *orcaFixture) command(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	switch fields[0] {
	case "status":
		h.writeIdentity()
		lines := []string{"identity.json updated (no credentials)"}
		for _, name := range orcaScenarios {
			g := h.gates[name]
			g.mu.Lock()
			lines = append(lines, fmt.Sprintf("%s: point index %d, checkpoint held=%t", name, g.index, g.waiting))
			g.mu.Unlock()
		}
		h.s.mu.Lock()
		for _, j := range h.s.jobs {
			v := j.snapshot()
			lines = append(lines, v.Message+fmt.Sprintf(" State=%s completed=%d reused=%d", v.State, v.Progress.Completed, v.Progress.Reused))
		}
		h.s.mu.Unlock()
		return strings.Join(lines, "\n")
	case "renew":
		return h.s.BootstrapURL()
	case "advance", "release":
		if len(fields) != 2 || h.gates[fields[1]] == nil {
			return "usage: advance SCENARIO | release SCENARIO"
		}
		g := h.gates[fields[1]]
		if fields[0] == "release" {
			g.release()
		} else {
			select {
			case g.advance <- struct{}{}:
			default:
				return "advance already pending"
			}
		}
		return "ok"
	case "reset":
		h.settle()
		h.writeIdentity()
		h.initialize()
		return "settled; fresh generation ready; renew and sign in again"
	case "shutdown":
		h.close()
		return "settled and stopped"
	default:
		return "commands: status | advance SCENARIO | release SCENARIO | reset | renew | shutdown"
	}
}
func (h *orcaFixture) checkpoint(t *testing.T, name string) int {
	t.Helper()
	select {
	case i := <-h.gates[name].entered:
		return i
	case <-time.After(10 * time.Second):
		t.Fatal("fixture checkpoint timeout")
		return -1
	}
}
func (h *orcaFixture) oldJob(id string) moduleapi.Job { return h.old.jobs[id].snapshot() }
func (h *orcaFixture) recordPath() string             { return filepath.Join(h.base, "identity.json") }
func orcaHash(b []byte) string                        { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
func (h *orcaFixture) writeIdentity() {
	// Paths, cookies, bootstrap secrets and session tokens are intentionally absent.
	type binding struct {
		Documents map[string]string      `json:"original_document_sha256"`
		RunIDs    []string               `json:"expected_run_ids"`
		Jobs      []moduleapi.Job        `json:"jobs"`
		Runs      []experiment.RunRecord `json:"committed_runs"`
	}
	scenarios := map[string]binding{}
	for _, name := range orcaScenarios {
		b := binding{Documents: h.originalHashes[name], RunIDs: h.runIDs(h.t, name), Runs: h.records(h.t, name)}
		h.s.mu.Lock()
		for _, j := range h.s.jobs {
			v := j.snapshot()
			if p := h.s.projects[v.Project]; p != nil {
				h.projectNames[v.Project] = p.path
			}
			if h.projectNames[v.Project] == name {
				b.Jobs = append(b.Jobs, v)
			}
		}
		h.s.mu.Unlock()
		scenarios[name] = b
	}
	revision, err := exec.Command("git", "-C", "../..", "rev-parse", "HEAD").Output()
	if err != nil {
		h.t.Fatalf("candidate revision: %v", err)
	}
	// Git diffs omit untracked files; the direct hashes below bind all five
	// fixture sources regardless of their index state.
	source, err := exec.Command("git", "-C", "../..", "diff", "--binary", "HEAD", "--", "workbench/study/orca_fixture.go", "workbench/server/orca_fixture_test.go", "workbench/server/orca_fixture_contract_test.go", "workbench/server/testdata/orca/expected.json", "workbench/server/testdata/orca/README.md").Output()
	if err != nil {
		h.t.Fatalf("tracked fixture diff: %v", err)
	}
	terminal := []moduleapi.Job{}
	h.s.mu.Lock()
	for _, j := range h.s.jobs {
		terminal = append(terminal, j.snapshot())
	}
	h.s.mu.Unlock()
	sources := map[string]string{}
	for _, file := range []string{"../study/orca_fixture.go", "orca_fixture_test.go", "orca_fixture_contract_test.go", "testdata/orca/expected.json", "testdata/orca/README.md"} {
		b, e := os.ReadFile(file)
		if e != nil {
			h.t.Fatal(e)
		}
		sources[file] = orcaHash(b)
	}
	data, err := json.MarshalIndent(map[string]any{"qualification": "SYNTHETIC automated engineering fixture; human Orca NOT RUN", "fixture_revision": h.expected.Revision, "expected_sha256": orcaHash(h.expectedBytes), "candidate_revision": strings.TrimSpace(string(revision)), "tracked_fixture_diff_sha256": orcaHash(source), "candidate_binary_sha256": h.binaryHash, "fixture_source_sha256": sources, "local_manifest": h.manifest, "generation": h.generation, "scenarios": scenarios, "job_snapshots": terminal}, "", "  ")
	if err != nil {
		h.t.Fatal(err)
	}
	if err = os.WriteFile(h.recordPath(), append(data, '\n'), 0600); err != nil {
		h.t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(h.base, fmt.Sprintf("generation-%d", h.generation), "identity.json"), append(data, '\n'), 0600); err != nil {
		h.t.Fatal(err)
	}
}
func (h *orcaFixture) records(t *testing.T, name string) []experiment.RunRecord {
	t.Helper()
	r, err := os.OpenRoot(h.projectPath(name))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	records, err := experiment.LoadRunsRoot(r)
	if err != nil {
		t.Fatal(err)
	}
	return records
}
func (h *orcaFixture) runIDs(t *testing.T, name string) []string {
	t.Helper()
	return h.expected.RunIDs[name]
}

// These clients use the real listener, auth, rooted project and job routes, not
// private Job manufacturing. Controls below never become HTTP endpoints.
type orcaClient struct {
	h      *orcaFixture
	cookie *http.Cookie
	client *http.Client
	serial int
}
type orcaProject struct{ ID, Revision string }

func (h *orcaFixture) client(t *testing.T) *orcaClient {
	t.Helper()
	c := &orcaClient{h: h, client: &http.Client{Timeout: 15 * time.Second}}
	token := strings.Split(h.command("renew"), "#bootstrap=")[1]
	res := c.call(t, "POST", "/api/v1/session", map[string]string{"secret": token}, nil)
	if res.StatusCode != 200 {
		t.Fatalf("fixture login: %d", res.StatusCode)
	}
	c.cookie = res.Cookies()[0]
	res.Body.Close()
	return c
}
func (c *orcaClient) call(t *testing.T, method, path string, body any, headers map[string]string) *http.Response {
	t.Helper()
	data, _ := json.Marshal(body)
	req, err := http.NewRequest(method, c.h.origin+path, bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if method != "GET" {
		req.Header.Set("Origin", c.h.origin)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-FeCIM-Request", "1")
	}
	if c.cookie != nil {
		req.AddCookie(c.cookie)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := c.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}
func (c *orcaClient) get(t *testing.T, path string) *http.Response {
	return c.call(t, "GET", path, nil, nil)
}
func orcaDecode(t *testing.T, res *http.Response, want int, out any) {
	t.Helper()
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil || res.StatusCode != want {
		t.Fatalf("HTTP %d want %d: %s %v", res.StatusCode, want, b, err)
	}
	if out != nil {
		if err = json.Unmarshal(b, out); err != nil {
			t.Fatal(err)
		}
	}
}
func (c *orcaClient) open(t *testing.T, name string) orcaProject {
	t.Helper()
	var roots struct{ Roots []struct{ ID string } }
	orcaDecode(t, c.get(t, "/api/v1/roots"), 200, &roots)
	var p orcaProject
	orcaDecode(t, c.call(t, "POST", "/api/v1/projects/open", map[string]string{"root": roots.Roots[0].ID, "path": name}, nil), 200, &p)
	return p
}
func (c *orcaClient) start(t *testing.T, p orcaProject) string {
	t.Helper()
	c.serial++
	var j moduleapi.Job
	orcaDecode(t, c.call(t, "POST", "/api/v1/projects/"+p.ID+"/experiments", map[string]int{"workers": 1}, map[string]string{"If-Match": p.Revision, "Idempotency-Key": fmt.Sprintf("fixture-%d", c.serial)}), 202, &j)
	return j.ID
}
func (c *orcaClient) job(t *testing.T, id string) moduleapi.Job {
	t.Helper()
	var j moduleapi.Job
	orcaDecode(t, c.get(t, "/api/v1/experiments/"+id), 200, &j)
	return j
}
func (c *orcaClient) cancel(t *testing.T, id string) moduleapi.Job {
	t.Helper()
	var j moduleapi.Job
	orcaDecode(t, c.call(t, "POST", "/api/v1/experiments/"+id+"/cancel", struct{}{}, nil), 200, &j)
	return j
}
func (c *orcaClient) wait(t *testing.T, id string) moduleapi.Job {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		j := c.job(t, id)
		if !activeState(j.State) {
			return j
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("fixture settlement timeout")
	return moduleapi.Job{}
}

func TestOrcaFixtureServer(t *testing.T) {
	if os.Getenv("FECIM_ORCA_FIXTURE") != "1" {
		t.Skip("interactive fixture is opt-in; finite TestOrcaFixtureContract requires no tester")
	}
	base := os.Getenv("FECIM_ORCA_ROOT")
	if base == "" || !filepath.IsAbs(base) {
		t.Fatal("set absolute disposable FECIM_ORCA_ROOT")
	}
	entries, err := os.ReadDir(base)
	if err != nil || len(entries) != 0 {
		t.Fatal("FECIM_ORCA_ROOT must exist and be empty")
	}
	h := newOrcaFixture(t, base, orcaFixtureOrigin(t), orcaFixtureAssets(t))
	defer h.close()
	fmt.Println("SYNTHETIC only. Launch link is secret: do not record it.")
	fmt.Println(h.command("renew"))
	fmt.Println("commands: status | advance SCENARIO | release SCENARIO | reset | renew | shutdown")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(h.command(scanner.Text()))
		if h.closed {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}
