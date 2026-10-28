# Installation

> **Note:** This file was previously located at `docs/INSTALLATION.md`. It has moved to `docs/guides/installation.md`.

This document lists required prerequisites and optional dependencies for specific features.

## Prerequisites

- **Go 1.26.10+** (patched minimum; CI reads `go.mod`) — https://go.dev/dl/
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
sudo apt-get install -y gcc libgl1-mesa-dev libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev libwayland-dev libxkbcommon-dev
# Optional: run legacy GUI/layout tests on a headless server
sudo apt-get install -y xvfb
```

### Linux (Fedora/RHEL)

```bash
sudo dnf install -y gcc mesa-libGL-devel libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel wayland-devel libxkbcommon-devel
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

Alternatively, [`./launch.sh`](../../launch.sh) builds and starts the default Fyne desktop app. It defaults an unset or empty `CGO_ENABLED` to `1` and preserves an explicit caller value. For headless commands, use `CGO_ENABLED=0 ./launch.sh --help` or the subcommands above; there is no automatic headless fallback if the desktop build fails.

The launcher forwards arguments unchanged except for `--clear`, which deletes only `logs/` and `docs/assets/screenshots/` relative to the script before building. It replaces the root `fecim-lattice-tools` binary on each run, never launches after a failed build, and returns the application's exit status. See [`launch.sh:9–36`](../../launch.sh) and the display-free [launcher contract test](../../scripts/launch_contract_test.go). Recheck this guidance when the launcher or the CGO entry points change.

## Optional local React Hysteresis application

The additive command is **`fecim-lattice-tools serve --root DIRECTORY`**.
The design's `fecim serve` is shorthand, not a separate installed executable.
Default launch and existing CLI/Fyne entry points are unchanged.

To build the embedded browser profile, install Node 24+, npm and the exact Go
version in `go.mod`, plus the desktop build dependencies above, then run:

```bash
make build-browser
./fecim-lattice-tools serve --root ./data --port 20200
```

Only grant directories you intend to read/write. Repeat `--root` for additional
grants. Omit `--port` to select an available loopback port; Autowork checks use
only the worker's assigned port range. Open the terminal's 60-second, single-use
launch link. Type `renew` then Enter in that terminal if the link expires or was
used. Use the local file panel to log out; Ctrl-C shuts down the server and
revokes its session. Browser users need no runtime Node installation.

Plain `go build` still compiles without generated browser assets; `serve`
reports `browser-assets-unavailable` unless the matching profile is staged and
pinned. Rebuild after changing Go or frontend source. Generated assets are not
committed. See the [command and security reference](../../cmd/fecim-lattice-tools/README.md)
for explicit save/load/export support, conflict recovery and trust limits.

## Equation SVG (Optional, Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install -y texlive-latex-base texlive-latex-recommended texlive-latex-extra texlive-fonts-recommended dvisvgm ghostscript
```

Regenerate the equation SVG after edits:

```bash
go run ./cmd/latex-svg -in shared/assets/equations/frankestein.tex -out shared/assets/equations/frankestein.svg
```
