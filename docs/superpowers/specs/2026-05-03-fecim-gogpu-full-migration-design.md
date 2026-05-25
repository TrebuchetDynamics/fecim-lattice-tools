# FeCIM Full gogpu/ui Migration & UX Enhancement — Design Spec

**Date:** 2026-05-03
**Status:** Draft (awaiting user review)
**Priority:** Design-first migration order, education + research + design layers

## Scope Summary

Port all 6 remaining placeholder modules from Fyne to the `gogpu/ui` zero-CGO shell, add a comprehensive design system with custom drawing primitives, and layer education/research/design features across every module. The end goal is a design-focused engineering workspace for FeCIM accelerator design, backed by education and research infrastructure.

## Migration Order (Design-First)

| Order | Module | Rationale |
|-------|--------|-----------|
| 1 | Module 1 Hysteresis | Physics foundation — material presets feed all other modules |
| 2 | Module 2 Crossbar | Array design core — feeds EDA and MNIST |
| 3 | Module 6 EDA | Design export — the end goal for engineering workflow |
| 4 | Module 4 Circuits | Peripheral design — ADC/DAC/TIA/ISPP |
| 5 | Module 3 MNIST | Inference pipeline — crossbar demo and accuracy analysis |
| 6 | Module 7 Docs | Education hub — tutorials, citation browser, curriculum |

## Architecture Pattern (per module)

Each module follows the established Module 5 pilot pattern:

```
shared/viewmodel/{module}/     ← UI-neutral viewmodel (no Fyne, no gogpu/ui)
  ├── viewmodel.go             ← Module struct implementing ModulePort
  ├── snapshot.go              ← Pure buildSnapshot function
  ├── state.go                 ← Typed module state
  ├── events.go                ← Events the UI sends back
  └── *_test.go                ← TDD tests (RED-first)

cmd/fecim-lattice-tools-next/  ← gogpu/ui shell
  ├── {module}_view.go         ← Adapter: ModuleSnapshot → widget tree
  └── main.go                  ← Add case to buildRoot switch
```

## Shell Navigation Redesign

Replace flat card list with sidebar + content layout:

- **Left sidebar:** Module list with icons, status badges, quick-jump
- **Right content area:** Active module rendered full-height
- **Top bar:** App title, theme toggle, export button, citation status

## Design System

### Material 3 Tokens

```
Primary:     #2F5D50 (deep green)
Surface:     #F4F5F3 (off-white)
On-surface:  #1A1C1A (near-black)
Secondary:   #58685E (muted green)
Error:       #BA1A1A (warnings)
```

### Custom gg.Context Drawing Primitives

- `PlotWidget` — 2D line/scatter plots with axes, grid, labels
- `HeatmapWidget` — 2D color grid with colorbar
- `MeterWidget` — bar/arc gauges for metrics
- `TimelineWidget` — horizontal event timeline
- `StepperWidget` — wizard-style step navigation

## Module Feature Specifications

### Module 1 — Hysteresis (Material Physics)

**Port from Fyne:** P-E loop rendering, Preisach engine, LK solver, 4 material presets, ISPP write controller, conductance quantization display.

**Education:** Annotated P-E loops with Ec/Pr/saturation labels, interactive LK parameter sliders (α/β/γ → loop deformation), Landau free energy equation overlay with live values, 2-material comparison overlay.

**Research:** Citation tags on presets (Materlik 2015, Park 2015, Alessandri 2018, Guo 2018), experimental data overlay toggle, golden data regression check in-view, CSV/JSON export with provenance.

**Design:** Parameter sweep (thickness, Ec, Pr ranges → loop families), coercive field sensitivity analysis, ISPP endurance estimation, material recommender (given target Pr/Ec → rank presets).

### Module 2 — Crossbar Array

**Port from Fyne:** N×M crossbar, conductance heatmap, MVM operation, IR drop, sneak paths, drift, WL select logic.

**Education:** Step-by-step MVM walkthrough with highlighted active row, Kirchhoff's law overlay, "break the array" mode (toggle IR drop), animated sneak path arrows.

**Research:** Drift parameterization (time/temp/cycles), array-level yield simulation, benchmark vs. ideal matmul, configurable quantization levels (8/16/30/64/128).

**Design:** Array sizing tradeoff explorer, programming strategy comparison, topology variants (1T1R/0T1R/passive), thermal crosstalk estimation, design → EDA module bridge.

### Module 6 — EDA Design Suite

**Port from Fyne:** SPICE netlist generation, Verilog export, Liberty timing/power, DEF/LEF physical, OpenLane integration.

**Education:** "What is SPICE?" explainer, schematic ↔ netlist side-by-side, interactive cell library browser.

**Research:** Citation-tagged technology assumptions, netlist validation against golden refs, hash-verified reproducible export.

**Design (primary focus):** Multi-module design wizard (material → array → circuits → export), cross-module parameter bridge, parameter sweep netlist generation, corner analysis (TT/FF/SS/FS/SF), design report PDF with citations, DRC integration via OpenLane.

### Module 4 — Peripheral Circuits

**Port from Fyne:** 5-bit SAR ADC, 5-bit DAC, TIA, charge pump, ISPP write path, sample-hold, voltage regulator, PVT.

**UX fix:** Replace 30-40% wasted sidebar with top toolbar layout (Option B from GUI.module4.md).

**Education:** Animated ADC SAR conversion (bits flipping), ideal vs. real transfer curves, INL/DNL histograms, drag-and-drop read path builder.

**Research:** PVT corner simulation, configurable noise models, citation-linked circuit assumptions.

**Design:** ADC resolution vs. area vs. power sweep, crossbar-ADC quantization error analysis, read path latency stack-up, power budget estimator.

### Module 3 — MNIST Inference

**Port from Fyne:** Quantized weight matrix, CIM inference pipeline, 80% baseline accuracy, confusion matrix.

**Education:** Animated MVM showing weights × inputs, step-by-step pipeline visualization, CIM vs. ideal float comparison, weight heatmap per digit class.

**Research:** Accuracy vs. quantization level sweep, non-ideality impact study (IR drop + drift → accuracy), cross-validation with HZO FTJ benchmark, full inference trace export.

**Design:** Accuracy/energy Pareto frontier, array size vs. accuracy tradeoff, what-if quantization scenarios, feedback to Crossbar module spec requirements.

### Module 7 — Documentation & Education Hub

**Port from Fyne:** Markdown viewer, TOC navigation, full-text search, reference browser, equation rendering.

**Education:** Guided curriculum ("FeCIM 101" → hysteresis → crossbar → inference → design export), interactive glossary, progress tracker, "try it now" jump-to-module buttons.

**Research:** Citation browser (filter by module/paper/author), honesty audit dashboard, literature gap map, paper draft export.

**Design:** Design guide workflow, cross-module hyperlinks, documented design package export.

## Cross-Module Integration

**Design composition workspace:**
- Material (M1) → Array (M2) → Circuits (M4) → Export (M6) pipeline
- Unified "Design Snapshot" capturing state across modules
- Shared parameter propagation

**Global features:**
- Undo/redo for all parameter changes
- Screenshot capture per module state
- Dark/light theme toggle
- Export with citation frontmatter
- Status bar: current material, array size, validation status

## Screenshot Generation

- Extend `cmd/fecim-screenshotter` with gogpu/ui headless support
- Generate `screenshots/` baseline per module state
- CI integration for visual regression detection

## TDD Discipline

Per CLAUDE.md hard rule: all behavior changes start with a failing test. Each module migration follows:
1. Write viewmodel test (RED)
2. Implement viewmodel (GREEN)
3. Write adapter test (RED)
4. Implement gogpu/ui adapter (GREEN)
5. Wire into buildRoot switch
6. Full test suite verification

## UI Boundary Rule

`shared/viewmodel/` must contain zero Fyne or gogpu/ui imports. Enforced by CI grep check:
```bash
grep -r 'fyne.io\|gogpu/ui' shared/viewmodel/ && exit 1
```

## Acceptance Criteria

1. All 7 modules render functional gogpu/ui views (not placeholder cards)
2. `CGO_ENABLED=0 go build ./cmd/fecim-lattice-tools-next` succeeds
3. `make test-next-ui` passes
4. `go test ./shared/viewmodel/...` passes
5. Sidebar navigation with module switching works
6. Design system primitives (PlotWidget, HeatmapWidget) render correctly
7. Screenshotter generates gogpu/ui captures
8. Education/research/design layers present in each module
