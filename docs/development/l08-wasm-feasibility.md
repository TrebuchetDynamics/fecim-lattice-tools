# L08 — WASM feasibility spike (Fyne)

Date: 2026-02-14

## Goal
Assess whether `fecim-lattice-tools` can be compiled to WebAssembly (GOOS=js GOARCH=wasm) for a browser demo build, using Fyne’s experimental web driver.

## What I tried

### 1) Build the main desktop entrypoint

Command:

```bash
cd <local-path>
GOOS=js GOARCH=wasm go build ./cmd/fecim-lattice-tools
```

Result: **FAIL**.

Key errors (verbatim excerpt):

- Vulkan binding fails under `js/wasm`:

```
# github.com/vulkan-go/vulkan
.../errors.go:8:19: undefined: Result
.../vk_null64.go:11:16: undefined: Semaphore
... (many more undefined Vulkan types)
```

- Bubbletea TUI fails under `js/wasm`:

```
# github.com/charmbracelet/bubbletea
.../tea.go:287:8: p.listenForResize undefined
.../tea.go:497:13: undefined: openInputTTY
.../tty.go:19:2: undefined: suspendProcess
```

Full build log captured at:
- `/tmp/l08_wasm_build_cmd_fecim_lattice_tools.out`

### Dependency provenance

To confirm why these deps are pulled in:

```bash
go mod why github.com/vulkan-go/vulkan
# github.com/vulkan-go/vulkan
fecim-lattice-tools/module1-hysteresis/pkg/render
github.com/vulkan-go/vulkan

go mod why github.com/charmbracelet/bubbletea
# github.com/charmbracelet/bubbletea
fecim-lattice-tools/module1-hysteresis/pkg/tui
github.com/charmbracelet/bubbletea
```

So the wasm failure is **not** currently Fyne itself; it’s that the **desktop app build graph includes packages that are inherently non-wasm**.

## Blockers (categorized)

### B1 — Non-wasm graphics backend (Vulkan)
- **Root:** `module1-hysteresis/pkg/render` imports `github.com/vulkan-go/vulkan`.
- **Why it breaks:** Vulkan bindings rely on platform APIs not provided in the JS/WASM environment.
- **Fix class:** build-tag exclusion or renderer abstraction.

### B2 — Non-wasm terminal IO (bubbletea)
- **Root:** `module1-hysteresis/pkg/tui` imports `github.com/charmbracelet/bubbletea`.
- **Why it breaks:** terminal suspend/TTY resize/input code paths don’t exist in browser.
- **Fix class:** build-tag exclusion for wasm builds.

### B3 — Monolithic entrypoint imports everything
- `./cmd/fecim-lattice-tools` appears to pull in more than the minimum needed to run a web UI.
- Even if Fyne web driver works, we need a **web-specific entrypoint** that avoids non-wasm packages.

## Remediation plan (proposed)

### Step 1 — Make wasm build graph minimal
Create a new entrypoint:
- `cmd/fecim-web/main.go`

It should:
- build a minimal Fyne app
- expose **only** safe GUI modules (hysteresis GUI, crossbar GUI, MNIST GUI, circuits GUI, EDA GUI, docs GUI)
- explicitly **not** import any Vulkan renderer or terminal UI packages.

### Step 2 — Add build tags to exclude non-wasm packages
For example:
- add `//go:build !js` to files in:
  - `module1-hysteresis/pkg/render` (Vulkan renderer)
  - `module1-hysteresis/pkg/tui` (bubbletea)

Optionally add stub files:
- `*_js.go` that provide minimal no-op APIs so imports compile if any remain.

### Step 3 — Produce `web/` artifacts
Once `cmd/fecim-web` builds:

```bash
GOOS=js GOARCH=wasm go build -o web/fecim.wasm ./cmd/fecim-web
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/
# write web/index.html that loads fecim.wasm
python3 -m http.server --directory web 8080
```

### Step 4 — Define reduced web scope (first milestone)
To de-risk:
- Milestone web demo = docs viewer + one physics module (hysteresis) + one system module (MNIST).
- Then add crossbar/circuits/EDA iteratively.

## Current status
- WASM build feasibility: **blocked** (B1, B2, B3) for `cmd/fecim-lattice-tools`.
- No test regressions introduced by this spike.

## Next action
Implement `cmd/fecim-web` + build-tag exclusions for Vulkan/TUI, then reattempt wasm build.
