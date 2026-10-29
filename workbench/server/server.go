// Package server serves the additive, loopback-only browser profile. Rooted
// research operations never treat supplied browser snapshots as verified studies.
package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi"
	"fecim-lattice-tools/workbench/moduleapi/httpapi"
	"fecim-lattice-tools/workbench/project"
	"fecim-lattice-tools/workbench/study"
)

const MaxBodyBytes = 8 << 20

// Manifest is generated after the local Vite build and binds all served bytes.
type Manifest struct {
	Profile  string              `json:"profile"`
	UI       string              `json:"UI"`
	API      string              `json:"API"`
	Identity hysteresis.Identity `json:"identity"`
	Assets   map[string]string   `json:"assets"`
}
type Config struct {
	Origin   string
	Assets   fs.FS
	UI       string
	Identity hysteresis.Identity
	Roots    []string
}
type Server struct {
	mu                               sync.Mutex
	operations                       sync.RWMutex // decoded session operations, never bodies or SSE
	origin, host                     string
	assets                           fs.FS
	manifest                         Manifest
	roots                            map[string]*os.Root
	rootNames                        map[string]string
	cookieName, session, secret      string
	expires                          time.Time
	now                              func() time.Time
	closed                           bool
	admission                        chan struct{}
	streams                          chan struct{}
	requests, workers, streamWorkers sync.WaitGroup
	closeDone                        chan struct{}
	epoch                            context.Context
	cancelEpoch                      context.CancelFunc
	draining                         bool
	api                              *httpapi.API
	projects                         map[string]*openedProject
	jobs                             map[string]*execution
	jobOrder                         []string
	expiredJobs                      []string
	modules                          map[string]*moduleSession
	moduleDescriptors                []moduleapi.ModuleDescriptor
	resources                        map[string]*grantedResource
	runBundle                        func(context.Context, *os.Root, *os.Root, project.Bundle, study.RunOptions) (study.RunOutcome, error)
}

func randomSecret() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
func equal(a, b string) bool { return a != "" && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }
func unavailable() error {
	return errors.New("browser-assets-unavailable: build and embed the matching local browser profile (make build-browser)")
}
func New(c Config) (*Server, error) {
	u, err := url.Parse(c.Origin)
	if err != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.Port() == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return nil, errors.New("canonical loopback origin required")
	}
	if _, _, err = net.SplitHostPort(u.Host); err != nil {
		return nil, errors.New("canonical loopback authority required")
	}
	m, err := validateAssets(c)
	if err != nil {
		return nil, err
	}
	s := &Server{origin: c.Origin, host: u.Host, assets: c.Assets, manifest: m, roots: map[string]*os.Root{}, rootNames: map[string]string{}, cookieName: "fecim_" + randomSecret()[:16], now: time.Now, admission: make(chan struct{}, 8), streams: make(chan struct{}, 4), closeDone: make(chan struct{}), projects: map[string]*openedProject{}, jobs: map[string]*execution{}, modules: map[string]*moduleSession{}, resources: map[string]*grantedResource{}}
	canonicalRoots := []string{}
	for _, name := range c.Roots {
		canonical, err := filepath.EvalSymlinks(name)
		if err == nil {
			canonical, err = filepath.Abs(canonical)
		}
		if err != nil {
			s.Close()
			return nil, errors.New("cannot open granted root")
		}
		for _, prior := range canonicalRoots {
			if canonical == prior || strings.HasPrefix(canonical, strings.TrimSuffix(prior, string(os.PathSeparator))+string(os.PathSeparator)) || strings.HasPrefix(prior, strings.TrimSuffix(canonical, string(os.PathSeparator))+string(os.PathSeparator)) {
				s.Close()
				return nil, errors.New("overlapping granted roots are not supported")
			}
		}
		canonicalRoots = append(canonicalRoots, canonical)
		root, err := os.OpenRoot(name)
		if err != nil {
			s.Close()
			return nil, errors.New("cannot open granted root")
		}
		id := randomSecret()
		s.roots[id] = root
		s.rootNames[id] = path.Base(name)
	}
	s.runBundle = study.RunBundleRoot
	s.epoch, s.cancelEpoch = context.WithCancel(context.Background())
	s.initModules()
	s.api = httpapi.New(s)
	return s, nil
}

// ValidateAssets checks all immutable profile/API/core identities and asset
// bytes without opening roots or a listener. Launchers call it before Listen;
// New rechecks the same policy before grant admission. Origin/Roots are ignored.
func ValidateAssets(c Config) error { _, err := validateAssets(c); return err }
func validateAssets(c Config) (Manifest, error) {
	if c.Assets == nil {
		return Manifest{}, unavailable()
	}
	data, err := fs.ReadFile(c.Assets, "local-manifest.json")
	if err != nil || len(data) > MaxBodyBytes {
		return Manifest{}, unavailable()
	}
	var m Manifest
	if json.Unmarshal(data, &m) != nil || m.Profile != "local" || m.UI == "" || m.UI != c.UI || m.Identity != c.Identity || m.Identity.Build == "unqualified" || m.Identity.Build == "" || m.Identity.Go != hysteresis.BridgeIdentity().Go || len(m.Assets) == 0 {
		return Manifest{}, unavailable()
	}
	schema, err := httpapi.New(nil).JSON()
	if err != nil {
		return Manifest{}, unavailable()
	}
	apiHash := sha256.Sum256(append(schema, '\n'))
	if m.API != hex.EncodeToString(apiHash[:]) {
		return Manifest{}, unavailable()
	}
	if _, ok := m.Assets["index.html"]; !ok {
		return Manifest{}, unavailable()
	}
	for name, want := range m.Assets {
		if !fs.ValidPath(name) || name == "local-manifest.json" {
			return Manifest{}, unavailable()
		}
		b, err := fs.ReadFile(c.Assets, name)
		if err != nil {
			return Manifest{}, unavailable()
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			return Manifest{}, unavailable()
		}
	}
	var engine struct {
		UI       string
		API      string
		Identity hysteresis.Identity
		Assets   map[string]string
	}
	engineData, err := fs.ReadFile(c.Assets, "engine/manifest.json")
	if err != nil || json.Unmarshal(engineData, &engine) != nil || engine.UI != m.UI || engine.API != m.API || engine.Identity != m.Identity {
		return Manifest{}, unavailable()
	}
	if _, ok := m.Assets["engine/manifest.json"]; !ok {
		return Manifest{}, unavailable()
	}
	for _, name := range []string{"hysteresis.wasm", "wasm_exec.js", "hysteresis-worker.mjs"} {
		if engine.Assets[name] == "" || engine.Assets[name] != m.Assets["engine/"+name] {
			return Manifest{}, unavailable()
		}
	}
	return m, nil
}

// BootstrapURL rotates a single-use 60-second launch secret. Only an explicit
// terminal action may call this; there is no unauthenticated renewal endpoint.
func (s *Server) BootstrapURL() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secret = randomSecret()
	s.expires = s.now().Add(time.Minute)
	return s.origin + "/#bootstrap=" + s.secret
}

// Close first revokes admission and cancels jobs/streams, then drains every
// root-owning request before releasing roots. Safe and idempotent concurrently.
func (s *Server) Close() {
	s.mu.Lock()
	if s.closed {
		done := s.closeDone
		s.mu.Unlock()
		<-done
		return
	}
	s.closed = true
	s.secret = ""
	s.session = ""
	if s.cancelEpoch != nil {
		s.cancelEpoch()
	}
	for _, j := range s.jobs {
		j.cancel()
	}
	s.mu.Unlock()
	s.workers.Wait()
	s.streamWorkers.Wait()
	s.requests.Wait()
	s.closeProjects()
	s.closeModules()
	s.closeResources()
	s.mu.Lock()
	for _, root := range s.roots {
		_ = root.Close()
	}
	close(s.closeDone)
	s.mu.Unlock()
}
func problem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(httpapi.Fail(status, detail))
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Content-Security-Policy", "default-src 'none'; script-src 'self' blob: 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; worker-src 'self' blob:; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
	// Applies to ordinary responses too: no stalled writer can pin a grant forever.
	controller := http.NewResponseController(w)
	_ = controller.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if r.Host != s.host || (r.Header.Get("Origin") != "" && r.Header.Get("Origin") != s.origin) || (r.Header.Get("Sec-Fetch-Site") != "" && r.Header.Get("Sec-Fetch-Site") != "same-origin" && r.Header.Get("Sec-Fetch-Site") != "none") {
		problem(w, 403, "Foreign request rejected")
		return
	}
	s.mu.Lock()
	if s.closed || s.draining {
		s.mu.Unlock()
		problem(w, 503, "Session ended or draining")
		return
	}
	s.requests.Add(1)
	authenticated := false
	requestSession := ""
	if c, err := r.Cookie(s.cookieName); err == nil {
		authenticated = equal(s.session, c.Value)
		if authenticated {
			requestSession = c.Value
		}
	}
	epoch := s.epoch
	s.mu.Unlock()
	defer s.requests.Done()
	r = r.WithContext(context.WithValue(r.Context(), requestSessionKey{}, requestSession))
	if epoch != nil {
		ctx, cancel := context.WithCancel(r.Context())
		stop := context.AfterFunc(epoch, cancel)
		defer stop()
		defer cancel()
		r = r.WithContext(ctx)
	}
	if path.Clean(r.URL.Path) != r.URL.Path || r.URL.RawPath != "" {
		problem(w, 400, "Noncanonical request path")
		return
	}
	queries, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		problem(w, 400, "Invalid query")
		return
	}
	known, allowed, queryOK := false, false, len(queries) == 0
	if s.api != nil {
		known, allowed, queryOK = s.api.Policy(r.Method, r.URL.Path, queries)
	}
	legacy := r.URL.Path == "/api/v1/roots" || r.URL.Path == "/api/v1/files" || r.URL.Path == "/api/v1/session" || r.URL.Path == "/api/v1/session/logout"
	if legacy {
		known = true
		allowed = (r.URL.Path == "/api/v1/roots" && r.Method == "GET") || (r.URL.Path != "/api/v1/roots" && r.Method == "POST")
		queryOK = len(queries) == 0
	}
	if !queryOK {
		problem(w, 400, "Unsupported or malformed query parameters")
		return
	}
	if (known && !allowed) || (!known && r.Method != "GET" && r.Method != "POST") {
		problem(w, 405, "Method not allowed")
		return
	}
	if r.Method == "POST" || r.Method == "PUT" {
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if r.Header.Get("Origin") != s.origin || r.Header.Get("X-FeCIM-Request") != "1" || err != nil || media != "application/json" {
			problem(w, 403, "Origin, JSON and X-FeCIM-Request required")
			return
		}
		if r.ContentLength > MaxBodyBytes {
			problem(w, 413, "Request exceeds 8 MiB")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	}
	stream := r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/api/v1/experiments/") && strings.HasSuffix(r.URL.Path, "/events")
	if stream {
		select {
		case s.streams <- struct{}{}:
			defer func() { <-s.streams }()
		default:
			problem(w, 503, "Progress stream capacity reached")
			return
		}
	} else {
		select {
		case s.admission <- struct{}{}:
			defer func() { <-s.admission }()
		default:
			problem(w, 503, "Busy; retry")
			return
		}
	}
	if r.URL.Path == "/api/v1/session" && r.Method == "POST" {
		var input struct {
			Secret string `json:"secret"`
		}
		if !readJSON(w, r, &input) {
			return
		}
		s.mu.Lock()
		if s.closed || s.draining || !equal(s.secret, input.Secret) || !s.now().Before(s.expires) {
			s.mu.Unlock()
			problem(w, 401, "Launch link expired or used; renew in terminal")
			return
		}
		s.secret = ""
		s.session = randomSecret()
		if s.epoch == nil || s.epoch.Err() != nil {
			s.epoch, s.cancelEpoch = context.WithCancel(context.Background())
		}
		cookie := &http.Cookie{Name: s.cookieName, Value: s.session, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode}
		s.mu.Unlock()
		http.SetCookie(w, cookie)
		jsonReply(w, map[string]string{"status": "authenticated"})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/") {
		if !authenticated {
			problem(w, 401, "Authentication required; renew launch link in terminal")
			return
		}
		switch r.URL.Path {
		case "/api/v1/session/logout":
			var input struct{}
			if !readJSON(w, r, &input) {
				return
			}
			if err := s.logout(r.Context()); err != nil {
				problem(w, 401, "Session ended or replaced")
				return
			}
			http.SetCookie(w, &http.Cookie{Name: s.cookieName, Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
			jsonReply(w, map[string]string{"status": "logged out"})
		case "/api/v1/roots":
			roots := []map[string]string{}
			for id, name := range s.rootNames {
				roots = append(roots, map[string]string{"id": id, "name": name})
			}
			jsonReply(w, map[string]any{"roots": roots, "identity": s.manifest.Identity, "UI": s.manifest.UI, "kinds": fileKinds})
		case "/api/v1/files":
			s.files(w, r)
		default:
			if !known {
				problem(w, 404, "Resource not found")
				return
			}
			s.api.Mux.ServeHTTP(w, r)
		}
		return
	}
	if r.Method != "GET" {
		problem(w, 405, "Method not allowed")
		return
	}
	if r.URL.Path == "/bootstrap.js" {
		h.Set("Content-Type", "text/javascript")
		_, _ = io.WriteString(w, bootstrapJS)
		return
	}
	route := appRoute(r.URL.Path)
	if route && !authenticated {
		h.Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, bootstrapHTML)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/")
	if route {
		name = "index.html"
	}
	if _, ok := s.manifest.Assets[name]; !ok {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(s.assets, name)
	if err != nil {
		problem(w, 503, "Asset unavailable")
		return
	}
	content := mime.TypeByExtension(path.Ext(name))
	if path.Ext(name) == ".wasm" {
		content = "application/wasm"
	}
	h.Set("Content-Type", content)
	if name == "index.html" {
		data = bytes.Replace(data, []byte("<head>"), []byte(`<head><script src="/bootstrap.js"></script>`), 1)
	}
	_, _ = w.Write(data)
}
func appRoute(p string) bool {
	switch p {
	case "/", "/hysteresis", "/projects", "/configuration", "/experiment", "/results", "/docs":
		return true
	}
	for _, f := range nativeFactories {
		if p == "/modules/"+f.id {
			return true
		}
	}
	return false
}
func (s *Server) logout(ctx context.Context) error {
	s.mu.Lock()
	if !s.currentRequestSession(ctx) {
		s.mu.Unlock()
		return httpapi.Fail(401, "Session ended or replaced")
	}
	s.session = ""
	s.secret = ""
	s.draining = true
	if s.cancelEpoch != nil {
		s.cancelEpoch()
	}
	for _, j := range s.jobs {
		j.cancel()
	}
	s.mu.Unlock()
	// Bodies may still be stalled. Only already decoded, authorized operations
	// must drain before registry cleanup; late old bodies fail beginSession.
	s.operations.Lock()
	defer s.operations.Unlock()
	s.workers.Wait()
	s.streamWorkers.Wait()
	s.closeProjects()
	s.mu.Lock()
	modules := s.modules
	s.modules = map[string]*moduleSession{}
	s.jobs = map[string]*execution{}
	s.jobOrder = nil
	s.expiredJobs = nil
	s.mu.Unlock()
	for _, m := range modules {
		if m.cancel != nil {
			m.cancel()
		}
		m.mu.Lock()
		if m.port != nil {
			m.port.Close()
			m.port = nil
		}
		m.mu.Unlock()
	}
	s.closeResources()
	s.mu.Lock()
	s.draining = false
	s.mu.Unlock()
	return nil
}
func readJSON(w http.ResponseWriter, r *http.Request, into any) bool {
	data, err := io.ReadAll(r.Body)
	if err == nil {
		d := json.NewDecoder(bytes.NewReader(data))
		d.DisallowUnknownFields()
		err = d.Decode(into)
		if err == nil && d.Decode(new(any)) != io.EOF {
			err = errors.New("expected one JSON value")
		}
	}
	if err != nil {
		var large *http.MaxBytesError
		if errors.As(err, &large) {
			problem(w, 413, "Request exceeds 8 MiB")
		} else {
			problem(w, 400, "Invalid typed JSON request")
		}
		return false
	}
	return true
}
func jsonReply(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}

const bootstrapHTML = `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>FeCIM local session</title><script src="/bootstrap.js"></script></head><body><main><h1>FeCIM local session</h1><p id="session-status" role="status">Authenticating local launch link…</p><p>Use the launch link printed by fecim-lattice-tools serve. If expired or already used, type renew then Enter in its terminal. No file is written by signing in.</p></main></body></html>`
const bootstrapJS = `let fragment = location.hash; history.replaceState(null, "", location.pathname); let secret = new URLSearchParams(fragment.slice(1)).get("bootstrap"); fragment = ""; document.addEventListener("DOMContentLoaded", async () => { let status = document.getElementById("session-status"); if (!secret) { if(status) status.textContent = "Authentication required. Renew the launch link in the terminal."; return; } try { const body = JSON.stringify({secret}); secret = null; const response = await fetch("/api/v1/session", {method:"POST", credentials:"same-origin", redirect:"error", headers:{"Content-Type":"application/json","X-FeCIM-Request":"1"},body}); if (!response.ok) throw new Error("Launch link expired, used or rejected. Renew in the terminal."); location.replace("/hysteresis"); } catch (error) { if(!status) {status=document.createElement("p");status.setAttribute("role","alert");document.body.prepend(status);} status.textContent = String(error);} });`
