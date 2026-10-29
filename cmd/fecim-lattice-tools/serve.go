package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"fecim-lattice-tools/internal/browserui"
	"fecim-lattice-tools/workbench/hysteresis"
	localserver "fecim-lattice-tools/workbench/server"
)

// App cancellation precedes HTTP drain: open SSE and root-owning jobs must be
// stopped before Shutdown waits for their handlers to return.
func shutdownBrowserHTTP(ctx context.Context, app interface{ Close() }, server *http.Server) error {
	app.Close()
	return server.Shutdown(ctx)
}

// Validate immutable assets before claiming even a loopback socket. New repeats
// validation before root admission after the listener supplies its exact origin.
func listenBrowser(c localserver.Config, port int) (net.Listener, error) {
	if err := localserver.ValidateAssets(c); err != nil {
		return nil, err
	}
	return net.Listen("tcp4", "127.0.0.1:"+strconv.Itoa(port))
}

type serveRoots []string

func (r *serveRoots) String() string         { return strings.Join(*r, ", ") }
func (r *serveRoots) Set(value string) error { *r = append(*r, value); return nil }
func runServeSubcommand(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var roots serveRoots
	fs.Var(&roots, "root", "DIRECTORY granted for rooted research projects and explicit local file operations (repeatable; required)")
	port := fs.Int("port", 0, "Loopback port (0 selects an available port)")
	fs.Usage = func() {
		fmt.Fprintln(stdout, "Usage: fecim-lattice-tools serve --root DIRECTORY [--port PORT]")
		fmt.Fprintln(stdout, "Additive React browser UI; default launch remains Fyne. Type renew then Enter in this terminal for a fresh 60-second launch link.")
		fs.SetOutput(stdout)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if len(roots) == 0 || fs.NArg() != 0 {
		return fmt.Errorf("serve requires --root DIRECTORY and no positional arguments")
	}
	if *port < 0 || *port > 65535 {
		return fmt.Errorf("port must be between 0 and 65535")
	}
	assets, err := browserui.Files()
	if err != nil {
		return err
	}
	listener, err := listenBrowser(localserver.Config{Assets: assets, UI: browserui.UIID, Identity: hysteresis.BridgeIdentity()}, *port)
	if err != nil {
		return err
	}
	defer listener.Close()
	origin := "http://" + listener.Addr().String()
	app, err := localserver.New(localserver.Config{Origin: origin, Assets: assets, UI: browserui.UIID, Identity: hysteresis.BridgeIdentity(), Roots: roots})
	if err != nil {
		return err
	}
	defer app.Close()
	// Never log request URLs, bodies, cookies or secrets. Only these terminal
	// launch links intentionally contain the short-lived fragment bootstrap.
	fmt.Fprintln(stdout, "Local browser simulation with rooted research projects. Supplied browser snapshots remain unverified; stored-run verification is an explicit research operation.")
	fmt.Fprintln(stdout, app.BootstrapURL())
	fmt.Fprintln(stdout, "Type renew then Enter for a new link. Ctrl-C stops the server and revokes the session.")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		scan := bufio.NewScanner(os.Stdin)
		for scan.Scan() {
			if strings.TrimSpace(scan.Text()) == "renew" {
				fmt.Fprintln(stdout, app.BootstrapURL())
			}
		}
	}()
	httpServer := &http.Server{Handler: app, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10}
	done := make(chan error, 1)
	go func() { done <- httpServer.Serve(listener) }()
	select {
	case err = <-done:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return shutdownBrowserHTTP(shutdown, app, httpServer)
	}
}
