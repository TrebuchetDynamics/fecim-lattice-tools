Role

- You are an expert software engineer and ferroelectrics scientist.
- Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
- If an ambiguity remains, choose the most reasonable default and proceed; document the choice.
- Keep scope tight: only change files required to satisfy the objectives.
- Default to headless-only work unless a GUI change is required for correctness.

Objective

- Maintain high-quality, readable **equation widgets** for Module 1: **Frankestein (L-K)** and **Preisach**.
- Equations must render as **LaTeX-derived SVGs** (single source of truth in `.tex` files).
- Frankestein equation uses **interactive hotspots**; Preisach is **static display** (no hotspots).
- Editing either equation should be as simple as updating a `.tex` file and regenerating the SVG.

Primary Focus (ranked)

1) Equation rendering quality (Frankestein + Preisach)
- SVG renders crisply in Fyne (no pixelated raster output).
- LaTeX source is the single source of truth for both equations.

2) Hotspot correctness (Frankestein only)
- Hotspot regions align with visible terms.
- Tap/click selects a term and shows details in the panel below the equation (no hover tooltips).
- Debug overlay can be enabled to tune hotspot positions.

3) Preisach display quality
- Preisach SVG renders in the info panel and stays in sync with `preisach.tex`.
- No hotspots; purely visual reference.

4) Performance hygiene
- Avoid per-frame SVG parsing/reloading; prefer cached image objects.
- Keep SVGs lean (no raster fallbacks). Debug overlays must be opt-in.

5) Safe fallback
- If SVG is missing, the widget should gracefully fall back to the text layout.

Scope / Files of Interest

- Widget: `module1-hysteresis/pkg/gui/widgets/frankestein_equation.go`
- Term info panel: `module1-hysteresis/pkg/gui/widgets/frankestein_equation_info.go`
- LaTeX source: `shared/assets/equations/frankestein.tex`
- LaTeX source: `shared/assets/equations/preisach.tex`
- Hotspots: `shared/assets/equations/frankestein.hotspots.json`
- SVG output: `shared/assets/equations/frankestein.svg`
- SVG output: `shared/assets/equations/preisach.svg`
- CLI generator: `cmd/latex-svg`

Tasks

1) LaTeX → SVG pipeline (Frankestein + Preisach)
- Use `cmd/latex-svg` to generate SVG from `shared/assets/equations/frankestein.tex`.
- Use `cmd/latex-svg` to generate SVG from `shared/assets/equations/preisach.tex`.
- Ensure SVGs write to their respective targets in `shared/assets/equations/`.

2) Hotspot alignment (Frankestein only)
- Enable `FECIM_EQUATION_DEBUG=1` to visualize hotspot boxes.
- Adjust `shared/assets/equations/frankestein.hotspots.json` (x/y/w/h normalized to SVG bounds).
- Validate that each selection label matches the correct term and the LK nonlinearity row is grouped.

3) Widget behavior
- Verify Frankestein SVG renders in the equation dialog.
- Verify Preisach SVG renders in the info panel (Preisach tab).
- Confirm tap/click updates the selection detail panel below the equation.
- Confirm fallback to text layout when SVG is absent.

4) Performance sanity (no code unless necessary)
- Ensure SVG loading is not repeated per-frame (no hot loop reloads).
- Prefer caching and one-time parse if touching widget code.

Validation

- Run (regenerate SVGs):
  - `go run ./cmd/latex-svg -in shared/assets/equations/frankestein.tex -out shared/assets/equations/frankestein.svg`
  - `go run ./cmd/latex-svg -in shared/assets/equations/preisach.tex -out shared/assets/equations/preisach.svg`
- Visual check (debug overlay):
  - `FECIM_EQUATION_DEBUG=1 ./launch.sh`

Deliverable

- Concise report:
  - SVG generation status (commands + success).
  - Hotspot alignment changes (file + summary).
  - Widget verification (Frankestein SVG, Preisach SVG, tap/click, fallback).
  - Performance note (any SVG load/caching concerns).
  - Any blockers.

Baseline (update each run)

- SVG generation attempt: 2026-02-03 via `go run ./cmd/latex-svg -in shared/assets/equations/frankestein.tex -out shared/assets/equations/frankestein.svg` and `go run ./cmd/latex-svg -in shared/assets/equations/preisach.tex -out shared/assets/equations/preisach.svg` (failed: `latex` binary not found; no system package install permissions).
- Hotspots alignment: 2026-02-03 not re-verified (headless-only; no GUI validation run).
- Widget status: GUI validation still pending (needs `FECIM_EQUATION_DEBUG=1 ./launch.sh` for selection and fallback confirmation). Preisach SVG in info panel not re-verified in GUI.
- SVG generated (Frankestein): 2026-02-02 via `go run ./cmd/latex-svg -in shared/assets/equations/frankestein.tex -out shared/assets/equations/frankestein.svg` (success; viewBox normalized to 0,0 and `<use>` glyphs inlined for Fyne SVG rendering).
- SVG generated (Preisach): 2026-02-02 via `go run ./cmd/latex-svg -in shared/assets/equations/preisach.tex -out shared/assets/equations/preisach.svg`.
- Hotspots aligned: 2026-02-02 updated `shared/assets/equations/frankestein.hotspots.json` using font-based SVG bounds; includes LK row and alpha definition hotspot.
- Materials/References: 2026-02-02 expanded Materials tab fields (Pr, Ps, Ec, Vc, C, switching energy, NLS params, etc.) and references list in `module1-hysteresis/pkg/gui/widgets/frankestein_equation_info.go`.
