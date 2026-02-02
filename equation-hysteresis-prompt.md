Role

- You are an expert software engineer and ferroelectrics scientist.
- Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
- If an ambiguity remains, choose the most reasonable default and proceed; document the choice.
- Keep scope tight: only change files required to satisfy the objectives.
- Default to headless-only work unless a GUI change is required for correctness.

Objective

- Maintain a high-quality, readable Frankestein equation widget for Module 1.
- The equation should render as a LaTeX-derived SVG with interactive hotspots.
- Editing the equation should be as simple as updating a `.tex` file and regenerating the SVG.

Primary Focus (ranked)

1) Equation rendering quality
- SVG renders crisply in Fyne (no pixelated raster output).
- LaTeX source is the single source of truth.

2) Hotspot correctness
- Hotspot regions align with visible terms.
- Tap/click selects a term and shows details in the panel below the equation (no hover tooltips).
- Debug overlay can be enabled to tune hotspot positions.

3) Safe fallback
- If SVG is missing, the widget should gracefully fall back to the text layout.

Scope / Files of Interest

- Widget: `module1-hysteresis/pkg/gui/widgets/frankestein_equation.go`
- Term info panel: `module1-hysteresis/pkg/gui/widgets/frankestein_equation_info.go`
- LaTeX source: `shared/assets/equations/frankestein.tex`
- Hotspots: `shared/assets/equations/frankestein.hotspots.json`
- SVG output: `shared/assets/equations/frankestein.svg`
- CLI generator: `cmd/latex-svg`

Tasks

1) LaTeX → SVG pipeline
- Use `cmd/latex-svg` to generate SVG from `shared/assets/equations/frankestein.tex`.
- Ensure the SVG writes to `shared/assets/equations/frankestein.svg`.

2) Hotspot alignment
- Enable `FECIM_EQUATION_DEBUG=1` to visualize hotspot boxes.
- Adjust `shared/assets/equations/frankestein.hotspots.json` (x/y/w/h normalized to SVG bounds).
- Validate that each selection label matches the correct term and the LK nonlinearity row is grouped.

3) Widget behavior
- Verify SVG renders in the equation dialog.
- Confirm tap/click updates the selection detail panel below the equation.
- Confirm fallback to text layout when SVG is absent.

Validation

- Run (regenerate SVG):
  - `go run ./cmd/latex-svg -in shared/assets/equations/frankestein.tex -out shared/assets/equations/frankestein.svg`
- Visual check (debug overlay):
  - `FECIM_EQUATION_DEBUG=1 ./launch.sh`

Deliverable

- Concise report:
  - SVG generation status (command + success).
  - Hotspot alignment changes (file + summary).
  - Widget verification (hover/tap/fallback).
  - Any blockers.

Baseline (update each run)

- SVG generated: 2026-02-02 via `go run ./cmd/latex-svg -in shared/assets/equations/frankestein.tex -out shared/assets/equations/frankestein.svg` (success; viewBox normalized to 0,0 and `<use>` glyphs inlined for Fyne SVG rendering).
- Hotspots aligned: 2026-02-02 updated `shared/assets/equations/frankestein.hotspots.json` using font-based SVG bounds; includes LK row and alpha definition hotspot.
- Materials/References: 2026-02-02 expanded Materials tab fields (Pr, Ps, Ec, Vc, C, switching energy, NLS params, etc.) and references list in `module1-hysteresis/pkg/gui/widgets/frankestein_equation_info.go`.
- Widget status: selection panel replaces tooltips; GUI validation pending (needs `FECIM_EQUATION_DEBUG=1 ./launch.sh` for selection and fallback confirmation).
