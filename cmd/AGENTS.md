<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# Command Entry Points

## Purpose

The `cmd/` directory contains all executable entry points for FeCIM Lattice Tools:

1. **fecim-lattice-tools** - Main Fyne application (all 7 modules)
3. **fecim-screenshotter** - Fyne desktop Fyne screenshot capture tool for documentation and testing
5. **latex-svg** - LaTeX equation to SVG converter utility

Each is a standalone executable that can be built and run independently.

## Directory Structure

```
cmd/
├── fecim-lattice-tools/         # Main Fyne GUI app (7 modules + orchestration)
│   ├── main.go
│   ├── mode_*.go                # CLI mode handlers (eda, preisach, etc)
│   ├── *_test.go                # Integration and headless tests
│   └── testdata/                # Test fixtures and golden outputs
├── fecim-screenshotter/         # Fyne screenshot automation tool
│   └── main.go
├── latex-svg/                   # LaTeX to SVG converter
│   ├── main.go
│   └── main_test.go
└── AGENTS.md                   # This file
```

## For AI Agents

### fecim-lattice-tools

**Purpose:** Main application entry point. Provides the Fyne GUI with navigation across 7 simulation and design modules.

**Key Files:**
- `cmd/fecim-lattice-tools/main.go` - App initialization, window setup, module loading
- `cmd/fecim-lattice-tools/mode_*.go` - CLI mode handlers (mode_engine_matrix_test.go, mode_preisach_target_progression_test.go, etc)
- Test files: Integration tests for headless operation, physics validation, golden data regression

**Usage:**
```bash
CGO_ENABLED=0 go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools                           # Launch GUI with all modules

# CLI modes (for headless testing/validation)
./fecim-lattice-tools eda gui                   # EDA GUI route (Module 6)
./fecim-lattice-tools eda cli                   # EDA CLI with sample design

# Flags
./fecim-lattice-tools -h                        # Show help
./fecim-lattice-tools --module hysteresis       # Start on a specific module
./fecim-lattice-tools --logger debug            # Debug logging
```

**Architecture:**
- Shell: `internal/Fyneapp`
- View-model boundary: `shared/viewmodel`

**State Management:**
- Shell navigation state in `internal/Fyneapp`
- Module state exposed through `shared/viewmodel` snapshots/actions
- Recording state: FFmpeg subprocess with pipe-based frame writing
- Recent files tracking for quick access

**Testing:**
```bash
go test ./cmd/fecim-lattice-tools/...           # All integration tests
go test -race ./cmd/fecim-lattice-tools/...     # Race condition detection
```

Key test patterns:
- Headless mode tests (no GUI, physics-only validation)
- Golden data regression (physics outputs match known-good results)
- ISPP convergence tests across multiple materials
- UI state transitions and event handling

**Common Tasks:**
- Add new module: add a UI-neutral viewmodel, register its `ModulePort`, and add a renderer in `internal/Fyneapp`
- Add CLI mode: Create new mode_*.go file with command handling
- Fix physics bug: Update ISPP or material model, regenerate golden data (FECIM_UPDATE_PHYSICS_GOLDEN=1)
- Improve UI: update the `Fyne` renderer and dispatch typed actions through `shared/viewmodel`

---

### fecim-screenshotter


**Key Files:**
- `cmd/fecim-screenshotter-fyne/main.go` - Thin wrapper for the Fyne screenshot generator
- `internal/Fynescreenshot/` - Shared screenshot rendering, option parsing, and file output

**Usage:**
```bash
go run ./cmd/fecim-screenshotter-fyne \
  -out docs/screenshots \
  -w 1200 -h 800 \
  -tag initial

# Capture single module
go run ./cmd/fecim-screenshotter-fyne \
  -only hysteresis \
  -tag after_fix
```

**Flags:**
- `-out DIR` - Output directory (default: screenshots)
- `-w WIDTH` - Output image width (default: 1400)
- `-h HEIGHT` - Output image height (default: 900)
- `-only MODULE` - Capture single module (hysteresis|crossbar|mnist|circuits|comparison|eda|docs)
- `-tag SUFFIX` - Filename suffix

**Output:**
Generates PNG files named for each rendered view, with optional `_{tag}` suffix.

Examples:
```
hysteresis-p-e-loop_initial.png
crossbar-heatmap-8x8_initial.png
mnist-accuracy-sweep_initial.png
circuits-ispp-convergence_initial.png
comparison-architecture-bars_initial.png
eda-design-overview_initial.png
docs-overview_initial.png
```

**Implementation:**
- Renders real `Fyne` app frames through `internal/Fyneapp`
- Uses the requested `-w` and `-h` as the headless app frame size
- Saves deterministic PNG files for docs and visual review

**Integration:**
- Used in CI/CD for regression detection
- Used in documentation building
- Used to validate visual appearance across modules
- Can detect UI bugs (layout, rendering, font issues)

**Testing:**
Visual regression testing is manual but can be automated:
```bash
# Generate baseline
go run ./cmd/fecim-screenshotter-fyne -tag baseline

# After changes, generate new screenshots
go run ./cmd/fecim-screenshotter-fyne -tag current

# Compare (diff tools, perceptual hashing, etc)
```

---

### latex-svg

**Purpose:** Utility for converting LaTeX math equations to SVG for documentation and technical graphics.

**Key Files:**
- `cmd/latex-svg/main.go` - LaTeX compilation, DVI to SVG conversion
- `cmd/latex-svg/main_test.go` - Round-trip and format validation tests

**Usage:**
```bash
# Basic conversion
go run ./cmd/latex-svg \
  -in equation.tex \
  -out equation.svg

# With custom preamble
go run ./cmd/latex-svg \
  -in equation.tex \
  -preamble my_preamble.tex \
  -out equation.svg

# Inline math mode
go run ./cmd/latex-svg \
  -in equation.tex \
  -inline

# Keep temporary files (debugging)
go run ./cmd/latex-svg \
  -in equation.tex \
  -out equation.svg \
  -keep-temp
```

**Flags:**
- `-in FILE` - Input LaTeX file (required)
- `-out FILE` - Output SVG file (default: stdout)
- `-preamble FILE` - Custom LaTeX preamble
- `-inline` - Use inline math mode ($ ... $)
- `-keep-temp` - Keep temporary files for debugging
- `-latex BINARY` - LaTeX binary path (default: latex)
- `-dvisvgm BINARY` - dvisvgm binary path (default: dvisvgm)
- `-use-fonts` - Embed fonts in SVG
- `-bbox MODE` - Bounding box mode (default: exact)

**Dependencies:**
- `latex` - LaTeX compiler
- `dvisvgm` - DVI to SVG converter

**Input Format:**
LaTeX equation file (no document structure needed):
```latex
E = mc^2
```

Or with document:
```latex
\documentclass{article}
\usepackage{amsmath}
\begin{document}
E = mc^2
\end{document}
```

**Output:**
SVG with proper scaling and clipping:
```xml
<svg xmlns="..." viewBox="0 0 100 50">
  <!-- Converted equation -->
</svg>
```

**Testing:**
```bash
go test -v ./cmd/latex-svg
```

Tests cover:
- Simple equations
- Complex math (fractions, integrals, matrices)
- Preamble handling
- SVG validity
- Round-trip (LaTeX → SVG → verify)

**Use Cases:**
- Technical documentation with math
- Educational diagrams for papers
- Visualizing physics equations
- Building equation galleries

**Integration:**
Used in:
- Docs building (convert math from sources to SVG)
- README generation (equations for GitHub)
- Paper preparation (consistent LaTeX rendering)

---

## Building All Entry Points

**From Root:**
```bash
CGO_ENABLED=0 go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
go build -o fecim-screenshotter-fyne ./cmd/fecim-screenshotter-fyne
go build -o latex-svg ./cmd/latex-svg

# Or via ./launch.sh
./launch.sh
```

**Build Tags & Flags:**
```bash
# Optimize for size
CGO_ENABLED=0 go build -ldflags="-s -w" -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Include debug symbols
CGO_ENABLED=0 go build -gcflags="all=-N -l" -o fecim-lattice-tools ./cmd/fecim-lattice-tools
```

## Testing Command Entry Points

**Unit Tests:**
```bash
go test ./cmd/...                           # All cmd tests
go test ./cmd/fecim-lattice-tools/...      # Main app tests
go test ./cmd/fecim-screenshotter-fyne      # Screenshotter tests
go test ./cmd/latex-svg/...                # LaTeX converter tests
```

**Integration Tests (headless):**
```bash
go test -v ./cmd/fecim-lattice-tools/... -run TestMode
go test -v ./cmd/fecim-lattice-tools/... -race
```

**Full Test Suite:**
```bash
go test ./...                               # All project tests
go test -race ./...                         # With race detector
```

## For New Commands

To add a new command entry point:

1. **Create Directory:** `cmd/my-tool/`
2. **Create main.go:**
   ```go
   package main

   import "flag"
   import "fmt"

   func main() {
       // Command implementation
   }
   ```
3. **Build:** `go build -o my-tool ./cmd/my-tool`
4. **Test:** Add `cmd/my-tool/*_test.go` with tests
5. **Document:** Add entry to this AGENTS.md

---

**Last Updated:** 2026-02-13
