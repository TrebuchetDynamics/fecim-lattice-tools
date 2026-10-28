### Build Instructions

**Prerequisites:**
- Go 1.26.10+ (patched minimum in [`go.mod`](go.mod); https://go.dev/dl/)
- Git
- Node.js/npm (only for the Astro web landing page)

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

### Web landing page (Astro)
The `web/` folder is a static Astro landing page for the project website. It no longer uses the retired Go WASM/Fyne browser demo.

Build + run locally:
```bash
cd web
npm install
npm run dev
# Then open the local Astro URL printed by the dev server.
```

Production build:
```bash
cd web
npm run build
```

### Minimum Hardware
- 4 GB RAM (8 GB recommended for large arrays)
- 200 MB disk space
