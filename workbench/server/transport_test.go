package server

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	nativeapi "fecim-lattice-tools/workbench/moduleapi/native"
)

func TestModuleTransportContract(t *testing.T) {
	s := testServer(t)
	cookie := login(t, s)
	t.Run("API manifest binding before grants", func(t *testing.T) {
		withoutNewline, err := s.api.JSON()
		if err != nil {
			t.Fatal(err)
		}
		for _, api := range []string{"", strings.Repeat("0", 64), fmt.Sprintf("%x", sha256.Sum256(withoutNewline))} {
			assets := testAssets(t).(fstest.MapFS)
			var manifest map[string]any
			json.Unmarshal(assets["local-manifest.json"].Data, &manifest)
			manifest["API"] = api
			assets["local-manifest.json"].Data, _ = json.Marshal(manifest)
			candidate, err := New(Config{Origin: testOrigin, Assets: assets, UI: "ui-test", Identity: testIdentity(), Roots: []string{filepath.Join(t.TempDir(), "does-not-exist")}})
			if err == nil {
				candidate.Close()
				t.Fatal("missing/drifting API digest admitted")
			}
			if !strings.Contains(err.Error(), "browser-assets-unavailable") {
				t.Fatalf("API drift reached grant admission: %v", err)
			}
		}
	})
	for _, tc := range []struct {
		method, path, body string
		status             int
		headers            map[string]string
	}{
		{"GET", "/api/v1/modules", "", 200, nil},
		{"GET", "/api/v1/modules?secret=not-a-query-token", "", 400, nil},
		{"GET", "/api/v1/roots?page=1", "", 400, nil},
		{"GET", "/api/v1/projects?page=1&page_size=2", "", 200, nil},
		{"GET", "/api/v1/projects?page=-1", "", 400, nil},
		{"GET", "/api/v1/projects?page=9223372036854775807&page_size=100", "", 200, nil},
		{"POST", "/api/v1/health", `{}`, 405, nil},
		{"PUT", "/api/v1/roots", `{}`, 405, nil},
		{"POST", "/api/v1/files", `{}`, 409, nil},
		{"PUT", "/api/v1/projects/unknown/configuration", `{}`, 403, map[string]string{"Origin": ""}},
		{"PUT", "/api/v1/projects/unknown/configuration", strings.Repeat(" ", MaxBodyBytes+1), 413, nil},
		{"GET", "/api/v1/experiments/missing/events?token=secret", "", 400, nil},
		{"GET", "/api/v1/experiments/missing/events", "", 404, nil},
		{"GET", "/api/v1/openapi.json", "", 404, nil},
		{"GET", "/api/v1/docs", "", 404, nil},
		{"GET", "/api/missing", "", 404, nil},
		{"GET", "/missing-asset.js", "", 404, nil},
		{"GET", "/projects", "", 200, nil},
		{"GET", "/configuration", "", 200, nil},
		{"GET", "/experiment", "", 200, nil},
		{"GET", "/results", "", 200, nil},
		{"GET", "/modules/crossbar", "", 200, nil},
	} {
		w := request(s, tc.method, tc.path, tc.body, cookie, tc.headers)
		if w.Code != tc.status {
			t.Errorf("%s %s: %d want %d: %s", tc.method, tc.path, w.Code, tc.status, w.Body)
		}
		if strings.HasPrefix(tc.path, "/api/") && tc.status >= 400 && w.Header().Get("Content-Type") != "application/problem+json" {
			t.Errorf("problem content type: %s", w.Header().Get("Content-Type"))
		}
	}
	if w := request(s, "GET", "/api/v1/modules", "", nil, nil); w.Code != 401 {
		t.Fatal("unauthenticated module access")
	}
	w := request(s, "GET", "/api/v1/modules", "", cookie, nil)
	var modules struct {
		Modules []struct {
			ID        string `json:"id"`
			Available bool   `json:"available"`
			Reason    string `json:"reason"`
		} `json:"modules"`
	}
	if json.Unmarshal(w.Body.Bytes(), &modules) != nil || len(modules.Modules) != 7 {
		t.Fatalf("module registry: %s", w.Body)
	}
	// Feature-owned Providers now include Documentation. Availability must be
	// backed by an actual port, never by a legacy functional status badge.
	available := map[string]bool{}
	for _, factory := range nativeFactories {
		port := nativePort(factory)
		if port != nil {
			available[factory.id] = port.Descriptor().Available
			port.Close()
		}
	}
	for _, m := range modules.Modules {
		switch m.ID {
		case "hysteresis", "crossbar", "eda", "mnist", "comparison":
			if !m.Available {
				t.Fatalf("implemented port unavailable: %+v", m)
			}
		}
		if m.Available != available[m.ID] || m.Available && m.Reason != "" || !m.Available && m.Reason == "" {
			t.Fatalf("fabricated capability: %+v", m)
		}
	}
	if w := request(s, "POST", "/api/v1/modules/crossbar/sessions", `{}`, cookie, nil); w.Code != 201 {
		t.Fatalf("real crossbar module session: %d %s", w.Code, w.Body)
	} else {
		var snapshot moduleapi.Snapshot
		if err := json.Unmarshal(w.Body.Bytes(), &snapshot); err != nil || snapshot.Descriptor.ID != "crossbar" || len(snapshot.Tables) == 0 {
			t.Fatalf("missing typed crossbar state: %s", w.Body)
		}
	}
	t.Run("generated safe problem schema", func(t *testing.T) {
		schema, err := s.api.JSON()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(schema), `"FieldConflict"`) || !strings.Contains(string(schema), `#/components/schemas/Problem`) {
			t.Fatal("OpenAPI does not describe actual conflict/problem responses")
		}
	})
	t.Run("safe errors and chunked PUT cap", func(t *testing.T) {
		var grant string
		for id := range s.roots {
			grant = id
		}
		w := request(s, "POST", "/api/v1/projects", `{"root":"`+grant+`","name":"safe-errors","secret":"/private/never-reflect"}`, cookie, nil)
		if w.Code < 400 || strings.Contains(w.Body.String(), "never-reflect") {
			t.Fatalf("unsafe decoder problem: %d %s", w.Code, w.Body)
		}
		r := httptest.NewRequest("PUT", testOrigin+"/api/v1/projects/missing/configuration", strings.NewReader(`{}`+strings.Repeat(" ", MaxBodyBytes)))
		r.ContentLength = -1
		r.Header.Set("Origin", testOrigin)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-FeCIM-Request", "1")
		r.AddCookie(cookie)
		recorder := httptest.NewRecorder()
		s.ServeHTTP(recorder, r)
		if recorder.Code != 413 {
			t.Fatalf("chunked PUT cap: %d %s", recorder.Code, recorder.Body)
		}
	})
	t.Run("functional HYS adapter", func(t *testing.T) {
		old := hysteresis.BuildID
		hysteresis.BuildID = "transport-test"
		defer func() { hysteresis.BuildID = old }()
		w := request(s, "POST", "/api/v1/modules/hysteresis/sessions", `{}`, cookie, nil)
		var snapshot moduleapi.Snapshot
		json.Unmarshal(w.Body.Bytes(), &snapshot)
		if w.Code != 201 || snapshot.Bridge == nil || snapshot.Bridge.Error != "" {
			t.Fatalf("HYS adapter: %d %s", w.Code, w.Body)
		}
		if w := request(s, "GET", "/api/v1/module-sessions/"+snapshot.Session, "", cookie, nil); w.Code != 200 {
			t.Fatal(w.Code)
		}
		send := func(action string, input any, want int) {
			t.Helper()
			payload, _ := json.Marshal(input)
			body, _ := json.Marshal(moduleapi.Command{Action: action, Revision: snapshot.Revision, Payload: payload})
			w := request(s, "POST", "/api/v1/module-sessions/"+snapshot.Session+"/actions", string(body), cookie, nil)
			if w.Code != want {
				t.Fatalf("%s: %d %s", action, w.Code, w.Body)
			}
			if want == 200 {
				if err := json.Unmarshal(w.Body.Bytes(), &snapshot); err != nil {
					t.Fatal(err)
				}
			}
		}
		config := hysteresis.DefaultConfig()
		config.Mode = hysteresis.Manual
		send("core.apply", hysteresis.Command{Kind: hysteresis.Configure, Config: config}, 200)
		program := hysteresis.DefaultProgramConfig()
		program.WorkLimit = 1
		send("program.new", program, 200)
		send("program.apply", hysteresis.ProgramCommand{Kind: hysteresis.WriteTarget, Target: 22}, 200)
		send("program.advance", map[string]any{"Steps": 1, "DT": 0.01}, 200)
		before := snapshot.Revision
		send("program.advance", map[string]any{"Steps": 1, "DT": 0.01}, 200)
		var terminal hysteresis.ProgrammingSnapshot
		data, _ := json.Marshal(snapshot.Bridge.Result)
		json.Unmarshal(data, &terminal)
		if snapshot.Revision != before+1 || snapshot.Status != "failed" || snapshot.Bridge.Error == "" || !snapshot.Bridge.Committed || terminal.Status != hysteresis.ProgramWorkLimit || terminal.OwnsDrive {
			t.Fatalf("committed terminal error lost: %+v / %+v", snapshot, terminal)
		}
		send("program.apply", hysteresis.ProgramCommand{Kind: hysteresis.ResetProgram}, 200)
		send("core.apply", hysteresis.Command{Kind: hysteresis.Run}, 200)
	})
	t.Run("functional comparison adapter", func(t *testing.T) {
		w := request(s, "POST", "/api/v1/modules/comparison/sessions", `{}`, cookie, nil)
		var snapshot moduleapi.Snapshot
		if err := json.Unmarshal(w.Body.Bytes(), &snapshot); err != nil || w.Code != 201 || snapshot.Session == "" || !snapshot.Descriptor.Available {
			t.Fatalf("comparison session: %d %s", w.Code, w.Body)
		}
		path := "/api/v1/module-sessions/" + snapshot.Session
		body, _ := json.Marshal(moduleapi.Command{Action: "calculate", Revision: snapshot.Revision, Input: map[string]string{"workload": "MNIST", "rate": "2", "profile": "conservative"}})
		w = request(s, "POST", path+"/actions", string(body), cookie, nil)
		if err := json.Unmarshal(w.Body.Bytes(), &snapshot); err != nil || w.Code != 200 {
			t.Fatalf("comparison calculate: %d %s", w.Code, w.Body)
		}
		var daily, unavailable bool
		for _, table := range snapshot.Tables {
			if table.ID == "energy" && len(table.Rows) == 3 {
				cell := table.Rows[0][2]
				daily = cell.Value != nil && *cell.Value == 172800 && cell.Unit == "inf/day"
			}
			if table.ID == "summary" {
				for _, row := range table.Rows {
					if row[0].Presentation.Display == "CO2 reduction" {
						unavailable = row[1].Value == nil && row[1].Presentation.Reason != ""
					}
				}
			}
		}
		if !daily || !unavailable {
			t.Fatal("comparison transport lost canonical dimensions or unavailable evidence")
		}
		before := request(s, "GET", path, "", cookie, nil).Body.String()
		if w := request(s, "POST", path+"/actions", string(body), cookie, nil); w.Code != 409 {
			t.Fatalf("stale comparison revision: %d %s", w.Code, w.Body)
		}
		if after := request(s, "GET", path, "", cookie, nil).Body.String(); after != before {
			t.Fatal("rejected comparison command changed the committed snapshot")
		}
		if w := request(s, "POST", path+"/close", `{}`, cookie, nil); w.Code != 200 {
			t.Fatalf("comparison close: %d %s", w.Code, w.Body)
		}
		if w := request(s, "GET", path, "", cookie, nil); w.Code != 404 {
			t.Fatal("closed comparison session remains accessible")
		}
	})
	t.Run("opaque bounded granted resources", func(t *testing.T) {
		var grant string
		for id := range s.roots {
			grant = id
		}
		if err := s.roots[grant].WriteFile("large-model.bin", []byte("bounded model bytes"), 0600); err != nil {
			t.Fatal(err)
		}
		w := request(s, "POST", "/api/v1/resources", `{"root":"`+grant+`","path":"large-model.bin","kind":"model"}`, cookie, nil)
		var ref moduleapi.ResourceReference
		json.Unmarshal(w.Body.Bytes(), &ref)
		if w.Code != 201 || ref.ID == "" || ref.Kind != "model" || strings.Contains(w.Body.String(), "large-model.bin") {
			t.Fatalf("opaque resource: %d %s", w.Code, w.Body)
		}
		resolver, ok := nativeapi.Resources(s.nativeContext(context.Background()))
		if !ok {
			t.Fatal("missing native context resolver")
		}
		reader, info, err := resolver.Open(context.Background(), ref)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil || string(data) != "bounded model bytes" || info.Size != int64(len(data)) {
			t.Fatalf("rooted resource read: %s %+v %v", data, info, err)
		}
		held, _, err := resolver.Open(context.Background(), ref)
		if err != nil {
			t.Fatal(err)
		}
		drained := make(chan struct{})
		go func() { s.closeResources(); close(drained) }()
		select {
		case <-drained:
		case <-time.After(time.Second):
			held.Close()
			<-drained
			t.Fatal("an idle native reader blocked resource shutdown")
		}
		if _, err := held.Read(make([]byte, 1)); err == nil {
			t.Fatal("resource reader survived revocation")
		}
		held.Close()
		if _, _, err := resolver.Open(context.Background(), moduleapi.ResourceReference{ID: "/etc/passwd", Kind: "model"}); err == nil {
			t.Fatal("nonopaque resource accepted")
		}
		if err := s.roots[grant].Symlink("large-model.bin", "alias.bin"); err != nil {
			t.Fatal(err)
		}
		if w := request(s, "POST", "/api/v1/resources", `{"root":"`+grant+`","path":"alias.bin","kind":"model"}`, cookie, nil); w.Code != 400 {
			t.Fatalf("resource symlink: %d", w.Code)
		}
	})
	t.Run("native project leases share mutation and shutdown ownership", func(t *testing.T) {
		local := testServer(t)
		auth := login(t, local)
		var grant string
		for id := range local.roots {
			grant = id
		}
		w := request(local, "POST", "/api/v1/projects", `{"root":"`+grant+`","name":"native-project"}`, auth, nil)
		var p struct {
			ID string `json:"id"`
		}
		if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
			t.Fatalf("create: %d %s", w.Code, w.Body)
		}
		projects, ok := nativeapi.Projects(local.nativeContext(context.Background()))
		if !ok {
			t.Fatal("missing native project lease capability")
		}
		ref := moduleapi.ResourceReference{ID: p.ID, Kind: "project"}
		for _, invalid := range []moduleapi.ResourceReference{{ID: p.ID}, {ID: "/etc", Kind: "project"}} {
			if err := projects.WithProject(context.Background(), invalid, false, func(*os.Root) error { t.Error("invalid lease admitted"); return nil }); err == nil {
				t.Fatal("invalid project reference accepted")
			}
		}
		entered, release := make(chan struct{}), make(chan struct{})
		leaseDone := make(chan error, 1)
		go func() {
			leaseDone <- projects.WithProject(context.Background(), ref, true, func(root *os.Root) error {
				if err := root.WriteFile("native.txt", []byte("confined"), 0600); err != nil {
					return err
				}
				if err := root.WriteFile("../escape.txt", nil, 0600); err == nil {
					return fmt.Errorf("project lease escaped its root")
				}
				close(entered)
				<-release
				_, err := root.Stat("native.txt")
				return err
			})
		}()
		select {
		case <-entered:
		case err := <-leaseDone:
			t.Fatalf("lease failed: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("lease not admitted")
		}
		if w := request(local, "POST", "/api/v1/projects/"+p.ID+"/validate", `{}`, auth, nil); w.Code != 409 {
			t.Fatalf("native mutation did not exclude study operation: %d", w.Code)
		}
		closed := make(chan struct{})
		go func() { local.Close(); close(closed) }()
		deadline := time.After(5 * time.Second)
		for {
			local.mu.Lock()
			ending := local.closed
			local.mu.Unlock()
			if ending {
				break
			}
			select {
			case <-deadline:
				t.Fatal("shutdown did not revoke admission")
			case <-time.After(time.Millisecond):
			}
		}
		if err := projects.WithProject(context.Background(), ref, false, func(*os.Root) error { return nil }); err == nil {
			t.Error("lease admitted after shutdown")
		}
		select {
		case <-closed:
			t.Error("shutdown closed a leased root")
		default:
		}
		close(release)
		if err := <-leaseDone; err != nil {
			t.Fatalf("root closed before lease returned: %v", err)
		}
		select {
		case <-closed:
		case <-time.After(5 * time.Second):
			t.Fatal("shutdown did not drain lease")
		}
	})
	t.Run("retired session bodies cannot mutate or revoke a replacement session", func(t *testing.T) {
		for _, endpoint := range []string{"/api/v1/projects", "/api/v1/session/logout"} {
			local := testServer(t)
			auth := login(t, local)
			var grant string
			for id := range local.roots {
				grant = id
			}
			reader, writer := io.Pipe()
			defer reader.Close()
			defer writer.Close()
			entered := make(chan struct{})
			r := httptest.NewRequest("POST", testOrigin+endpoint, &observedReader{Reader: reader, entered: entered})
			r.AddCookie(auth)
			r.Header.Set("Origin", testOrigin)
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("X-FeCIM-Request", "1")
			w := httptest.NewRecorder()
			done := make(chan struct{})
			go func() { local.ServeHTTP(w, r); close(done) }()
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("request did not begin decoding")
			}
			if response := request(local, "POST", "/api/v1/session/logout", `{}`, auth, nil); response.Code != 200 {
				t.Fatal(response.Code)
			}
			replacement := login(t, local)
			body := `{}`
			if endpoint == "/api/v1/projects" {
				body = `{"root":"` + grant + `","name":"after-logout"}`
			}
			if _, err := io.WriteString(writer, body); err != nil {
				t.Fatal(err)
			}
			writer.Close()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("retired request did not finish")
			}
			if w.Code != 401 {
				t.Errorf("%s retired request: %d %s", endpoint, w.Code, w.Body)
			}
			if _, err := local.roots[grant].Stat("after-logout"); !os.IsNotExist(err) {
				t.Errorf("retired request wrote project: %v", err)
			}
			if response := request(local, "GET", "/api/v1/health", "", replacement, nil); response.Code != 200 {
				t.Errorf("replacement session revoked: %d", response.Code)
			}
		}
	})
	t.Run("native job dispatch retains granted capabilities", func(t *testing.T) {
		local := testServer(t)
		auth := login(t, local)
		var grant string
		for id := range local.roots {
			grant = id
		}
		w := request(local, "POST", "/api/v1/projects", `{"root":"`+grant+`","name":"job-project"}`, auth, nil)
		var p struct {
			ID string `json:"id"`
		}
		if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &p) != nil {
			t.Fatal(w.Body)
		}
		ctx, cancel := context.WithCancel(local.epoch)
		defer cancel()
		local.modules["capability-test"] = &moduleSession{id: "capability-test", port: &capabilityJobPort{project: p.ID}, ctx: local.nativeContext(ctx), cancel: cancel}
		for _, method := range []string{"GET", "POST"} {
			url, body := "/api/v1/module-sessions/capability-test/jobs/job", ""
			if method == "POST" {
				url += "/cancel"
				body = `{}`
			}
			if response := request(local, method, url, body, auth, nil); response.Code != 200 {
				t.Errorf("%s native capabilities missing: %d %s", method, response.Code, response.Body)
			}
		}
	})
	t.Run("bounded streams independent admission logout and close", func(t *testing.T) {
		var grant string
		for id := range s.roots {
			grant = id
		}
		created := request(s, "POST", "/api/v1/projects", `{"root":"`+grant+`","name":"stream-study"}`, cookie, nil)
		var p struct {
			ID       string `json:"id"`
			Revision string `json:"revision"`
		}
		json.Unmarshal(created.Body.Bytes(), &p)
		if created.Code != 201 {
			t.Fatalf("create: %s", created.Body)
		}
		start := request(s, "POST", "/api/v1/projects/"+p.ID+"/experiments", `{"workers":1}`, cookie, map[string]string{"If-Match": p.Revision, "Idempotency-Key": "stream"})
		var job moduleapi.Job
		json.Unmarshal(start.Body.Bytes(), &job)
		if start.Code != 202 {
			t.Fatalf("start: %s", start.Body)
		}
		if w := request(s, "POST", "/api/v1/projects/"+p.ID+"/experiments", `{"workers":2}`, cookie, map[string]string{"If-Match": p.Revision, "Idempotency-Key": "stream"}); w.Code != 409 {
			t.Fatalf("idempotency changed options: %d", w.Code)
		}
		var wg sync.WaitGroup
		for i := 0; i < cap(s.streams); i++ {
			writer := &stalledWriter{header: make(http.Header), entered: make(chan struct{}, 1)}
			r := httptest.NewRequest("GET", testOrigin+"/api/v1/experiments/"+job.ID+"/events", nil)
			r.AddCookie(cookie)
			wg.Add(1)
			go func() { defer wg.Done(); s.ServeHTTP(writer, r) }()
			select {
			case <-writer.entered:
			case <-time.After(5 * time.Second):
				t.Fatal("stream did not write")
			}
		}
		if w := request(s, "GET", "/api/v1/experiments/"+job.ID+"/events", "", cookie, nil); w.Code != 503 {
			t.Fatalf("stream cap: %d", w.Code)
		}
		if w := request(s, "GET", "/api/v1/roots", "", cookie, nil); w.Code != 200 {
			t.Fatalf("stalled stream blocked request: %d", w.Code)
		}
		if w := request(s, "POST", "/api/v1/experiments/"+job.ID+"/cancel", `{}`, cookie, nil); w.Code != 200 {
			t.Fatalf("stalled stream blocked cancel: %d", w.Code)
		}
		startTime := time.Now()
		if w := request(s, "POST", "/api/v1/session/logout", `{}`, cookie, nil); w.Code != 200 {
			t.Fatalf("logout: %d", w.Code)
		}
		if time.Since(startTime) > 4*time.Second {
			t.Fatal("logout not bounded")
		}
		wg.Wait()
		done := make(chan struct{})
		go func() { s.Close(); close(done) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Close not bounded")
		}
		if _, err := s.roots[grant].Stat("."); err == nil {
			t.Fatal("grant stayed open after shutdown")
		}
	})
	t.Run("root aliases rejected", func(t *testing.T) {
		dir := t.TempDir()
		child := filepath.Join(dir, "child")
		if err := os.Mkdir(child, 0700); err != nil {
			t.Fatal(err)
		}
		duplicate, err := New(Config{Origin: testOrigin, Assets: testAssets(t), UI: "ui-test", Identity: testIdentity(), Roots: []string{dir, child}})
		if err == nil {
			duplicate.Close()
			t.Fatal("overlapping grants admitted")
		}
	})
}

type observedReader struct {
	io.Reader
	once    sync.Once
	entered chan struct{}
}

func (r *observedReader) Read(data []byte) (int, error) {
	r.once.Do(func() { close(r.entered) })
	return r.Reader.Read(data)
}

// This probe tests capability propagation through real HTTP dispatch, not science.
type capabilityJobPort struct {
	moduleapi.Port
	project string
}

func (p *capabilityJobPort) Close() {}
func (p *capabilityJobPort) Start(context.Context, moduleapi.Command) (moduleapi.Job, error) {
	return moduleapi.Job{}, nil
}
func (p *capabilityJobPort) Job(ctx context.Context, id string) (moduleapi.Job, error) {
	projects, ok := nativeapi.Projects(ctx)
	if !ok {
		return moduleapi.Job{}, fmt.Errorf("missing project capability")
	}
	if _, ok := nativeapi.Resources(ctx); !ok {
		return moduleapi.Job{}, fmt.Errorf("missing resource capability")
	}
	err := projects.WithProject(ctx, moduleapi.ResourceReference{ID: p.project, Kind: "project"}, false, func(root *os.Root) error { _, err := root.Stat("project.yaml"); return err })
	return moduleapi.Job{ID: id, State: "finished"}, err
}
func (p *capabilityJobPort) Cancel(ctx context.Context, id string) (moduleapi.Job, error) {
	return p.Job(ctx, id)
}

type stalledWriter struct {
	header   http.Header
	mu       sync.Mutex
	deadline time.Time
	entered  chan struct{}
}

func (w *stalledWriter) Header() http.Header { return w.header }
func (w *stalledWriter) WriteHeader(int)     {}
func (w *stalledWriter) SetWriteDeadline(d time.Time) error {
	w.mu.Lock()
	w.deadline = d
	w.mu.Unlock()
	return nil
}
func (w *stalledWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	deadline := w.deadline
	w.mu.Unlock()
	select {
	case w.entered <- struct{}{}:
	default:
	}
	time.Sleep(time.Until(deadline))
	return 0, os.ErrDeadlineExceeded
}
func (w *stalledWriter) FlushError() error { return nil }
