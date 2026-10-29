package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"fecim-lattice-tools/workbench/hysteresis"
	localserver "fecim-lattice-tools/workbench/server"
)

func serveTestAddress(offset int) string {
	base, err := strconv.Atoi(os.Getenv("AUTOWORK_PORT_BASE"))
	if err != nil {
		base = 20200
	}
	return "127.0.0.1:" + strconv.Itoa(base+offset)
}

func TestBrowserShutdownCancelsBeforeHTTPDrain(t *testing.T) {
	cancelled := make(chan struct{})
	entered := make(chan struct{})
	listener, err := net.Listen("tcp4", serveTestAddress(94))
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-cancelled; w.WriteHeader(200) })}
	go server.Serve(listener)
	defer server.Close()
	responseDone := make(chan struct{})
	go func() {
		defer close(responseDone)
		response, err := http.Get("http://" + listener.Addr().String())
		if err == nil {
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := shutdownBrowserHTTP(ctx, closeFunction(func() { close(cancelled) }), server); err != nil {
		t.Fatal(err)
	}
	select {
	case <-responseDone:
	case <-time.After(time.Second):
		t.Fatal("active request did not drain")
	}
}

type closeFunction func()

func (f closeFunction) Close() { f() }

func TestServeCommand(t *testing.T) {
	t.Run("stale API assets rejected before listener", func(t *testing.T) {
		occupied, err := net.Listen("tcp4", serveTestAddress(95))
		if err != nil {
			t.Fatal(err)
		}
		defer occupied.Close()
		port, _ := strconv.Atoi(strings.Split(occupied.Addr().String(), ":")[1])
		identity := hysteresis.BridgeIdentity()
		identity.Build = "test-qualified"
		assets := fstest.MapFS{"index.html": {Data: []byte("index")}, "engine/hysteresis.wasm": {Data: []byte("wasm")}, "engine/wasm_exec.js": {Data: []byte("runtime")}, "engine/hysteresis-worker.mjs": {Data: []byte("worker")}}
		workerHashes := map[string]string{}
		for name, file := range assets {
			if strings.HasPrefix(name, "engine/") {
				workerHashes[strings.TrimPrefix(name, "engine/")] = fmt.Sprintf("%x", sha256.Sum256(file.Data))
			}
		}
		const staleAPI = "0000000000000000000000000000000000000000000000000000000000000000"
		engine, _ := json.Marshal(map[string]any{"UI": "test-ui", "API": staleAPI, "identity": identity, "assets": workerHashes})
		assets["engine/manifest.json"] = &fstest.MapFile{Data: engine}
		hashes := map[string]string{}
		for name, file := range assets {
			hashes[name] = fmt.Sprintf("%x", sha256.Sum256(file.Data))
		}
		manifest, _ := json.Marshal(localserver.Manifest{Profile: "local", UI: "test-ui", API: staleAPI, Identity: identity, Assets: hashes})
		assets["local-manifest.json"] = &fstest.MapFile{Data: manifest}
		listener, err := listenBrowser(localserver.Config{Assets: assets, UI: "test-ui", Identity: identity}, port)
		if listener != nil {
			listener.Close()
		}
		if err == nil || !strings.Contains(err.Error(), "browser-assets-unavailable") {
			t.Fatalf("listener attempted before stale API rejection: %v", err)
		}
	})

	for _, tc := range []struct {
		args []string
		want string
		ok   bool
	}{
		{[]string{"serve", "--help"}, "rooted research projects", true},
		{[]string{"serve"}, "--root", false},
		{[]string{"serve", "--root", t.TempDir()}, "browser-assets-unavailable", false},
		{[]string{"serve", "--root", t.TempDir(), "--port", "-1"}, "port", false},
	} {
		var out, errs bytes.Buffer
		err := dispatchSubcommandWithWriters(tc.args, &out, &errs)
		if (err == nil) != tc.ok {
			t.Fatalf("%v: %v", tc.args, err)
		}
		text := out.String() + errs.String()
		if err != nil {
			text += err.Error()
		}
		if !strings.Contains(text, tc.want) {
			t.Fatalf("%v: %q lacks %q", tc.args, text, tc.want)
		}
	}
	var out bytes.Buffer
	printRootUsage(&out)
	if !strings.Contains(out.String(), "serve") {
		t.Fatal("serve absent from help")
	}
}
