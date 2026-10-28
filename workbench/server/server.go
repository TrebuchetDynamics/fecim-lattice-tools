// Package server serves the additive, loopback-only browser profile. It never
// computes browser science or verifies supplied snapshots as persisted studies.
package server

import (
	"bytes"
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
	"strings"
	"sync"
	"time"

	"fecim-lattice-tools/workbench/hysteresis"
)

const MaxBodyBytes = 8 << 20

// Manifest is generated after the local Vite build and binds all served bytes.
type Manifest struct {
	Profile  string              `json:"profile"`
	UI       string              `json:"UI"`
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
	mu                          sync.Mutex
	origin, host                string
	assets                      fs.FS
	manifest                    Manifest
	roots                       map[string]*os.Root
	rootNames                   map[string]string
	cookieName, session, secret string
	expires                     time.Time
	now                         func() time.Time
	closed                      bool
	admission                   chan struct{}
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
	if c.Assets == nil {
		return nil, unavailable()
	}
	data, err := fs.ReadFile(c.Assets, "local-manifest.json")
	if err != nil || len(data) > MaxBodyBytes {
		return nil, unavailable()
	}
	var m Manifest
	if json.Unmarshal(data, &m) != nil || m.Profile != "local" || m.UI == "" || m.UI != c.UI || m.Identity != c.Identity || m.Identity.Build == "unqualified" || m.Identity.Build == "" || m.Identity.Go != hysteresis.BridgeIdentity().Go || len(m.Assets) == 0 {
		return nil, unavailable()
	}
	if _, ok := m.Assets["index.html"]; !ok {
		return nil, unavailable()
	}
	for name, want := range m.Assets {
		if !fs.ValidPath(name) || name == "local-manifest.json" {
			return nil, unavailable()
		}
		b, err := fs.ReadFile(c.Assets, name)
		if err != nil {
			return nil, unavailable()
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != want {
			return nil, unavailable()
		}
	}
	var engine struct {
		UI       string
		Identity hysteresis.Identity
		Assets   map[string]string
	}
	engineData, err := fs.ReadFile(c.Assets, "engine/manifest.json")
	if err != nil || json.Unmarshal(engineData, &engine) != nil || engine.UI != m.UI || engine.Identity != m.Identity {
		return nil, unavailable()
	}
	if _, ok := m.Assets["engine/manifest.json"]; !ok {
		return nil, unavailable()
	}
	for _, name := range []string{"hysteresis.wasm", "wasm_exec.js", "hysteresis-worker.mjs"} {
		if engine.Assets[name] == "" || engine.Assets[name] != m.Assets["engine/"+name] {
			return nil, unavailable()
		}
	}
	s := &Server{origin: c.Origin, host: u.Host, assets: c.Assets, manifest: m, roots: map[string]*os.Root{}, rootNames: map[string]string{}, cookieName: "fecim_" + randomSecret()[:16], now: time.Now, admission: make(chan struct{}, 8)}
	for _, name := range c.Roots {
		root, err := os.OpenRoot(name)
		if err != nil {
			s.Close()
			return nil, errors.New("cannot open granted root")
		}
		id := randomSecret()
		s.roots[id] = root
		s.rootNames[id] = path.Base(name)
	}
	return s, nil
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
func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.secret = ""
	s.session = ""
	for _, root := range s.roots {
		_ = root.Close()
	}
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Content-Security-Policy", "default-src 'none'; script-src 'self' blob: 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; worker-src 'self' blob:; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
	if r.Host != s.host || (r.Header.Get("Origin") != "" && r.Header.Get("Origin") != s.origin) || (r.Header.Get("Sec-Fetch-Site") != "" && r.Header.Get("Sec-Fetch-Site") != "same-origin" && r.Header.Get("Sec-Fetch-Site") != "none") {
		http.Error(w, "foreign request rejected", 403)
		return
	}
	if r.URL.RawQuery != "" {
		http.Error(w, "query parameters are not supported", 400)
		return
	}
	select {
	case s.admission <- struct{}{}:
		defer func() { <-s.admission }()
	default:
		http.Error(w, "busy; retry", 503)
		return
	}
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "method not allowed", 405)
		return
	}
	if r.Method == "POST" {
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if r.Header.Get("Origin") != s.origin || r.Header.Get("X-FeCIM-Request") != "1" || err != nil || media != "application/json" {
			http.Error(w, "Origin, JSON and X-FeCIM-Request required", 403)
			return
		}
		if r.ContentLength > MaxBodyBytes {
			http.Error(w, "request exceeds 8 MiB", 413)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	}
	s.mu.Lock()
	defer s.mu.Unlock() // bounded operations; shutdown and logout wait for admitted I/O
	if s.closed {
		http.Error(w, "session ended", 503)
		return
	}
	authenticated := false
	if c, err := r.Cookie(s.cookieName); err == nil {
		authenticated = equal(s.session, c.Value)
	}
	if r.URL.Path == "/api/v1/session" && r.Method == "POST" {
		var input struct {
			Secret string `json:"secret"`
		}
		if !readJSON(w, r, &input) {
			return
		}
		if !equal(s.secret, input.Secret) || !s.now().Before(s.expires) {
			http.Error(w, "launch link expired or used; renew in terminal", 401)
			return
		}
		s.secret = ""
		s.session = randomSecret()
		http.SetCookie(w, &http.Cookie{Name: s.cookieName, Value: s.session, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
		jsonReply(w, map[string]string{"status": "authenticated"})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/") {
		if !authenticated {
			http.Error(w, "authentication required; renew launch link in terminal", 401)
			return
		}
		switch {
		case r.URL.Path == "/api/v1/session/logout" && r.Method == "POST":
			var input struct{}
			if !readJSON(w, r, &input) {
				return
			}
			s.session = ""
			http.SetCookie(w, &http.Cookie{Name: s.cookieName, Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
			jsonReply(w, map[string]string{"status": "logged out"})
		case r.URL.Path == "/api/v1/roots" && r.Method == "GET":
			roots := []map[string]string{}
			for id, name := range s.rootNames {
				roots = append(roots, map[string]string{"id": id, "name": name})
			}
			jsonReply(w, map[string]any{"roots": roots, "identity": s.manifest.Identity, "UI": s.manifest.UI, "kinds": fileKinds})
		case r.URL.Path == "/api/v1/files" && r.Method == "POST":
			s.files(w, r)
		default:
			http.NotFound(w, r)
		}
		return
	}
	if r.Method != "GET" {
		http.Error(w, "method not allowed", 405)
		return
	}
	if r.URL.Path == "/bootstrap.js" {
		h.Set("Content-Type", "text/javascript")
		io.WriteString(w, bootstrapJS)
		return
	}
	if (r.URL.Path == "/" || r.URL.Path == "/hysteresis") && !authenticated {
		h.Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, bootstrapHTML)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/")
	if name == "" || name == "hysteresis" {
		name = "index.html"
	}
	// Only the build-manifest's immutable public code is served. No filesystem
	// URL maps to a launch-granted root, nor to a client-selected file.
	if _, ok := s.manifest.Assets[name]; !ok {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(s.assets, name)
	if err != nil {
		http.Error(w, "asset unavailable", 503)
		return
	}
	content := mime.TypeByExtension(path.Ext(name))
	if path.Ext(name) == ".wasm" {
		content = "application/wasm"
	}
	h.Set("Content-Type", content)
	if name == "index.html" {
		// A renewed launch link can be opened while already authenticated.
		// Run fragment cleanup/exchange before any application module executes.
		data = bytes.Replace(data, []byte("<head>"), []byte(`<head><script src="/bootstrap.js"></script>`), 1)
	}
	w.Write(data)
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
			http.Error(w, "request exceeds 8 MiB", 413)
		} else {
			http.Error(w, "invalid typed JSON request", 400)
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
