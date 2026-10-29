package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/moduleapi/httpapi"
)

const testOrigin = "http://127.0.0.1:20200"

func testIdentity() hysteresis.Identity {
	v := hysteresis.BridgeIdentity()
	v.Build = "test-pinned"
	return v
}
func testAssets(t *testing.T) fs.FS {
	t.Helper()
	files := fstest.MapFS{"index.html": {Data: []byte("<html>local</html>")}, "engine/hysteresis.wasm": {Data: []byte("wasm")}, "engine/wasm_exec.js": {Data: []byte("runtime")}, "engine/hysteresis-worker.mjs": {Data: []byte("worker")}}
	workerHashes := map[string]string{}
	for name, file := range files {
		if strings.HasPrefix(name, "engine/") {
			sum := sha256.Sum256(file.Data)
			workerHashes[strings.TrimPrefix(name, "engine/")] = hex.EncodeToString(sum[:])
		}
	}
	schema, err := httpapi.New(nil).JSON()
	if err != nil {
		t.Fatal(err)
	}
	apiDigest := sha256.Sum256(append(schema, '\n'))
	api := hex.EncodeToString(apiDigest[:])
	workerManifest, _ := json.Marshal(map[string]any{"UI": "ui-test", "API": api, "identity": testIdentity(), "assets": workerHashes})
	files["engine/manifest.json"] = &fstest.MapFile{Data: workerManifest}
	hashes := map[string]string{}
	for name, file := range files {
		sum := sha256.Sum256(file.Data)
		hashes[name] = hex.EncodeToString(sum[:])
	}
	manifest, _ := json.Marshal(map[string]any{"profile": "local", "UI": "ui-test", "API": api, "identity": testIdentity(), "assets": hashes})
	files["local-manifest.json"] = &fstest.MapFile{Data: manifest}
	return files
}
func testServer(t *testing.T) *Server {
	t.Helper()
	s, err := New(Config{Origin: testOrigin, Assets: testAssets(t), UI: "ui-test", Identity: testIdentity(), Roots: []string{t.TempDir()}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
func request(s *Server, method, path, body string, cookie *http.Cookie, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, testOrigin+path, strings.NewReader(body))
	if method != "GET" {
		r.Header.Set("Origin", testOrigin)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-FeCIM-Request", "1")
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	for k, v := range headers {
		if k == "Host" {
			r.Host = v
		} else {
			r.Header.Set(k, v)
		}
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}
func login(t *testing.T, s *Server) *http.Cookie {
	t.Helper()
	token := strings.Split(s.BootstrapURL(), "#bootstrap=")[1]
	w := request(s, "POST", "/api/v1/session", `{"secret":"`+token+`"}`, nil, nil)
	if w.Code != 200 {
		t.Fatalf("login %d %s", w.Code, w.Body)
	}
	return w.Result().Cookies()[0]
}
func TestHysteresisLocalSession(t *testing.T) {
	// Missing/drifted assets must be rejected before any operation or root access.
	for _, assets := range []fs.FS{fstest.MapFS{}, fstest.MapFS{"local-manifest.json": {Data: []byte(`{"profile":"hosted"}`)}}} {
		if _, err := New(Config{Origin: testOrigin, Assets: assets, UI: "ui-test", Identity: testIdentity()}); err == nil || !strings.Contains(err.Error(), "browser-assets-unavailable") {
			t.Fatalf("asset gate: %v", err)
		}
	}
	good := testAssets(t).(fstest.MapFS)
	good["index.html"].Data = []byte("drifted")
	if _, err := New(Config{Origin: testOrigin, Assets: good, UI: "ui-test", Identity: testIdentity()}); err == nil {
		t.Fatal("drifted asset bytes admitted")
	}
	mismatched := testAssets(t).(fstest.MapFS)
	mismatched["engine/manifest.json"].Data = []byte(`{"UI":"other-ui"}`)
	var manifest Manifest
	json.Unmarshal(mismatched["local-manifest.json"].Data, &manifest)
	sum := sha256.Sum256(mismatched["engine/manifest.json"].Data)
	manifest.Assets["engine/manifest.json"] = hex.EncodeToString(sum[:])
	mismatched["local-manifest.json"].Data, _ = json.Marshal(manifest)
	if _, err := New(Config{Origin: testOrigin, Assets: mismatched, UI: "ui-test", Identity: testIdentity()}); err == nil {
		t.Fatal("incompatible engine manifest admitted")
	}
	s := testServer(t)
	token := strings.Split(s.BootstrapURL(), "#bootstrap=")[1]
	body := `{"secret":"` + token + `"}`
	if w := request(s, "POST", "/api/v1/session", body, nil, map[string]string{"Origin": "https://foreign.invalid"}); w.Code != 403 {
		t.Fatalf("foreign Origin: %d", w.Code)
	}
	w := request(s, "POST", "/api/v1/session", body, nil, nil)
	if w.Code != 200 {
		t.Fatalf("exchange: %d %s", w.Code, w.Body)
	}
	cookie := w.Result().Cookies()[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Domain != "" || cookie.MaxAge != 0 {
		t.Fatalf("cookie: %+v", cookie)
	}
	if bytes.Contains(w.Body.Bytes(), []byte(token)) {
		t.Fatal("secret reflected")
	}
	if request(s, "POST", "/api/v1/session", body, nil, nil).Code != 401 {
		t.Fatal("bootstrap replay")
	}
	for _, headers := range []map[string]string{{"Host": "localhost:20200"}, {"Host": "127.0.0.1:20201"}, {"Origin": "null"}, {"Origin": "https://foreign.invalid"}, {"Sec-Fetch-Site": "cross-site"}} {
		if w := request(s, "GET", "/api/v1/roots", "", cookie, headers); w.Code != 403 {
			t.Fatalf("read guard %v: %d", headers, w.Code)
		}
	}
	if request(s, "GET", "/api/v1/roots", "", nil, nil).Code != 401 {
		t.Fatal("unauthenticated data")
	}
	if request(s, "GET", "/api/v1/roots", "", cookie, nil).Code != 200 {
		t.Fatal("authenticated same-origin GET")
	}
	for _, headers := range []map[string]string{{"Origin": ""}, {"Origin": "null"}, {"X-FeCIM-Request": ""}, {"Content-Type": "text/plain"}} {
		if request(s, "POST", "/api/v1/session/logout", `{}`, cookie, headers).Code != 403 {
			t.Fatalf("mutation guard %v", headers)
		}
	}
	if request(s, "POST", "/api/v1/session/logout", strings.Repeat(" ", MaxBodyBytes+1), cookie, nil).Code != 413 {
		t.Fatal("oversize cap")
	}
	chunked := httptest.NewRequest("POST", testOrigin+"/api/v1/session/logout", strings.NewReader(`{}`+strings.Repeat(" ", MaxBodyBytes)))
	chunked.ContentLength = -1
	chunked.Header.Set("Origin", testOrigin)
	chunked.Header.Set("Content-Type", "application/json")
	chunked.Header.Set("X-FeCIM-Request", "1")
	chunked.AddCookie(cookie)
	cw := httptest.NewRecorder()
	s.ServeHTTP(cw, chunked)
	if cw.Code != 413 {
		t.Fatalf("chunked trailing byte cap: %d", cw.Code)
	}
	if request(s, "POST", "/api/v1/session/logout", `{}`, cookie, nil).Code != 200 || request(s, "GET", "/api/v1/roots", "", cookie, nil).Code != 401 {
		t.Fatal("logout revocation")
	}
	s.now = func() time.Time { return time.Now().Add(61 * time.Second) }
	expired := strings.Split(s.BootstrapURL(), "#bootstrap=")[1]
	s.now = func() time.Time { return time.Now().Add(122 * time.Second) }
	if request(s, "POST", "/api/v1/session", `{"secret":"`+expired+`"}`, nil, nil).Code != 401 {
		t.Fatal("expired link")
	}
}
