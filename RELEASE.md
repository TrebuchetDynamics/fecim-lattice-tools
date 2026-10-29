### Build Instructions

**Prerequisites:**
- Go 1.26.10+ (patched minimum in [`go.mod`](go.mod); https://go.dev/dl/)
- Git
- Node 24.19.0 and npm 11.17.0 only when building the optional React assets

The Fyne desktop app requires CGO, a C compiler, and native graphics development libraries. See [desktop build dependencies](docs/guides/installation.md#desktop-build-dependencies). This page contains source-build instructions, not an announcement of a published release. See the [release and JOSS readiness checklist](docs/internals/release/research-software-readiness.md) for publication gates.

### Build & Run
```bash
git clone https://github.com/TrebuchetDynamics/fecim-lattice-tools.git
cd fecim-lattice-tools
CGO_ENABLED=1 go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools
```

Alternatively, `./launch.sh` builds and starts the default Fyne desktop app. It defaults unset or empty `CGO_ENABLED` to `1` and preserves explicit caller settings. Use `CGO_ENABLED=0 ./launch.sh --help` for headless use, not desktop startup. See [installation and launcher behavior](docs/guides/installation.md#build-and-run) for headless subcommands, argument forwarding, and `--clear` cleanup scope.

### Verify Installation
```bash
go test ./...
```

### Optional shared React candidate

`web/` builds the same seven module views and research routes for hosted-static and authenticated local use. Hosted provides Hysteresis, bounded studies, Crossbar, Circuits, Comparison and read-only Documentation. Go/WASM owns browser science. MNIST inference and EDA generation require the authenticated local application; their hosted views provide read-only help/imports. The local profile embeds matching assets and adds rooted projects, revision-checked edits, persistent experiments and explicit reports/exports. Real installed Icarus/Yosys qualification remains NOT RUN; fixture checks do not qualify tools or signoff artifacts.

```bash
make build-browser
mkdir -p "$HOME/fecim-projects"
./fecim-lattice-tools serve --root "$HOME/fecim-projects"
```

Open the short-lived terminal launch link. `make build-browser` installs locked frontend dependencies, regenerates contracts, builds both profiles and builds the matching embedded binary. Browser assets require the exact Go version in `go.mod` (currently 1.26.10), including its runtime shim. Users of the built binary need no Node installation. Plain Go builds retain the CLI/Fyne surfaces but `serve` refuses missing browser assets.

The current `build:native` packaging script explicitly selects CGO=1 and needs the desktop build dependencies. The Go command supports CGO=0 headless subcommand builds, but those plain builds have no matched browser packaging. Default desktop startup fails explicitly without CGO. Legacy Fyne remains the default and retains its regression gates.

This is a source-built candidate, not a published or qualified release. Real Orca + Firefox/Linux acceptance is **NOT RUN** for both profiles. No Fyne retirement, deployment, publication or default switch follows from automated checks. See [web build/check instructions](web/README.md) and the [remaining manual qualification limits](docs/internals/testing/all-modules-web-recheck.md#verdict-and-scope).

### Minimum Hardware
- 4 GB RAM (8 GB recommended for large arrays)
- 200 MB disk space
