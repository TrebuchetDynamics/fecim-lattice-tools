### Build Instructions

**Prerequisites:**
- Go 1.25+ (https://go.dev/dl/)
- Git
- Node.js/npm (only for the Astro web landing page)

The default desktop app is the Fyne desktop `Fyne` shell; no C compiler or OpenGL development headers are required for the standard build.

### Build & Run
```bash
git clone https://github.com/your-org/fecim-lattice-tools.git
cd fecim-lattice-tools
CGO_ENABLED=0 go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools
```

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
