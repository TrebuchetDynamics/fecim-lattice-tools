# Installation

> **Note:** This file was previously located at `docs/INSTALLATION.md`. It has moved to `docs/guides/installation.md`.

This document lists required prerequisites and optional dependencies for specific features.

## Prerequisites

- **Go 1.25+** — https://go.dev/dl/
- **Git** — For cloning the repository

The Fyne desktop app requires **CGO enabled**, a C compiler, and the platform's graphics development dependencies listed below. A non-CGO build supports headless subcommands, not desktop startup: [`fyne_nocgo.go:16–20`](../../cmd/fecim-lattice-tools/fyne_nocgo.go) returns an explicit error when a desktop workspace is requested. The desktop implementation is selected by the `cgo` build constraint in [`fyne_app.go:1–18`](../../cmd/fecim-lattice-tools/fyne_app.go).

## Optional Dependencies

- **Docker** — For Module 6 EDA tools (OpenLane/OpenROAD/KLayout)
- **Graphviz** — For Yosys circuit schematic visualization
- **LaTeX + dvisvgm** — For regenerating equation SVG assets (Frankestein equation)

## Desktop Build Dependencies

Install Go and Git first. These additional dependencies are needed for the Fyne desktop build; headless-only builds can use `CGO_ENABLED=0`.

### Linux (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y gcc libgl1-mesa-dev libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev
# Optional: run legacy GUI/layout tests on a headless server
sudo apt-get install -y xvfb
```

### Linux (Fedora/RHEL)

```bash
sudo dnf install -y gcc mesa-libGL-devel libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel
```

### macOS


```bash
xcode-select --install
```

### Windows

1. Install MSYS2 from https://www.msys2.org/ or a supported MinGW-w64 toolchain.
2. Ensure `gcc` is in your PATH.

## Build and Run

From the repository root, after installing the desktop dependencies:

```bash
CGO_ENABLED=1 go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools
```

For a headless-only installation:

```bash
CGO_ENABLED=0 go build -o fecim-lattice-tools-cli ./cmd/fecim-lattice-tools
./fecim-lattice-tools-cli --help
```

Use that binary with the [workbench guide's](research-design-workbench.md) `project validate`, `experiment run`, and `report generate` subcommands. Running it without a headless subcommand requests the desktop and fails.

**Launcher limitation:** [`launch.sh:29`](../../launch.sh) currently forces `CGO_ENABLED=0`, so it cannot launch the desktop. Use the explicit desktop build above instead. Recheck this guidance when the launcher or the CGO entry points change.

## Equation SVG (Optional, Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y texlive-latex-base texlive-latex-recommended texlive-latex-extra texlive-fonts-recommended dvisvgm ghostscript
```

Regenerate the equation SVG after edits:

```bash
go run ./cmd/latex-svg -in shared/assets/equations/frankestein.tex -out shared/assets/equations/frankestein.svg
```
