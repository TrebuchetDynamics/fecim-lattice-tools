# Installation

> **Note:** This file was previously located at `docs/INSTALLATION.md`. It has moved to `docs/1-getting-started/installation.md`.

This document lists required prerequisites and optional dependencies for specific features.

## Prerequisites

- **Go 1.25+** — https://go.dev/dl/
- **Git** — For cloning the repository

The default gogpu/ui app requires Go and Git only; no CGO, C compiler, or OpenGL headers are required.

## Optional Dependencies

- **Docker** — For Module 6 EDA tools (OpenLane/OpenROAD/KLayout)
- **Graphviz** — For Yosys circuit schematic visualization
- **LaTeX + dvisvgm** — For regenerating equation SVG assets (Frankestein equation)
- **Legacy Fyne parity toolchain** — Only for opt-in legacy GUI checks with `-tags legacy_fyne`

### Linux (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y git
# Optional: for Module 6 Yosys schematic visualization
sudo apt-get install -y graphviz
```

### Linux (Fedora/RHEL)

```bash
sudo dnf install -y git
```

### macOS

Install Go from https://go.dev/dl/ and Git via Xcode Command Line Tools or your package manager.

### Windows

Install Go from https://go.dev/dl/ and Git for Windows.

## Legacy Fyne parity only (`-tags legacy_fyne`)

These dependencies are not required for the default gogpu/ui app. Install them only when running legacy Fyne parity commands such as `go test -tags legacy_fyne ./...` or `go run -tags legacy_fyne ./cmd/fecim-lattice-tools-fyne`.

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

Install Xcode Command Line Tools before running legacy Fyne parity builds:

```bash
xcode-select --install
```

### Windows

1. Install MSYS2 from https://www.msys2.org/ or a supported MinGW-w64 toolchain.
2. Ensure `gcc` is in your PATH.
3. Build legacy Fyne parity commands with `-tags legacy_fyne`.

## Equation SVG (Optional, Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y texlive-latex-base texlive-latex-recommended texlive-latex-extra texlive-fonts-recommended dvisvgm ghostscript
```

Regenerate the equation SVG after edits:

```bash
go run ./cmd/latex-svg -in shared/assets/equations/frankestein.tex -out shared/assets/equations/frankestein.svg
```
