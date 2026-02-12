# Module 5: Comparison

Architecture comparison and benchmarking module for FeCIM versus baseline computing paradigms. Visualizes trade-offs across energy, performance, area, and market positioning.

## Overview

Module 5 provides structured comparisons between FeCIM (Ferroelectric Compute-in-Memory) and competing architectures: SRAM-CIM, ReRAM/memristor, Flash-CIM, and conventional digital (GPU/TPU). It generates summary tables, radar charts, and market-positioning slides for presentations and educational use.

## Package Structure

### `pkg/comparison/` — Comparison Logic

- **architecture.go** — Architecture definitions: FeCIM, SRAM-CIM, ReRAM, Flash, Digital. Stores specs (energy/op, density, endurance, retention, speed) and comparison scoring
- **render.go** — Text/table rendering of comparison results for CLI output

### `pkg/gui/` — Fyne GUI

- **embedded.go** — Embeddable app for unified launcher
- **app.go** — Main comparison GUI layout
- **hero.go** — Hero/summary cards with key differentiators
- **market.go** — Market positioning and competitive landscape widgets
- **fabrication_reality.go** — Fabrication readiness and maturity assessment
- **widgets.go** — Shared comparison widgets (radar, bar charts, stat cards)
- **liveslide.go** — Live slide presentation mode
- **export.go** — Export comparison results
- **keyboard.go** — Keyboard shortcuts

### `cmd/` — Entry Points

- **comparison/main.go** — CLI comparison runner
- **comparison-gui/main.go** — Standalone GUI launcher

## Key Types and Functions

| Type / Function | Package | Description |
|---|---|---|
| `Architecture` | `pkg/comparison` | Architecture spec (energy, density, endurance, speed) |
| `ComparisonResult` | `pkg/comparison` | Scored comparison between two architectures |
| `Render` | `pkg/comparison` | Text/table output formatting |
| `HeroCard` | `pkg/gui` | Summary visualization widget |
| `MarketView` | `pkg/gui` | Competitive landscape display |

## Testing

```bash
# Run all module 5 tests
go test ./module5-comparison/...

# With race detector
go test -race ./module5-comparison/...

# Verbose
go test -v ./module5-comparison/pkg/comparison/...
go test -v ./module5-comparison/pkg/gui/...
```

Key test suites:
- `pkg/comparison/` — Architecture scoring, comparison logic, additional edge cases
- `pkg/gui/` — Hero card rendering, market widget calculations

## Physics Context

**Energy efficiency:** FeCIM achieves ~1-10 fJ/MAC in-memory, compared to ~100 fJ/MAC for SRAM-CIM and ~1 pJ/MAC for digital accelerators. The ferroelectric polarization switching is inherently low-energy.

**Endurance:** FeCIM (HZO) targets 10^10–10^12 cycles, exceeding Flash (~10^5) but below SRAM (unlimited). ReRAM sits at ~10^6–10^9.

**Density:** Passive (0T1R) FeCIM achieves 4F² cell size — the theoretical minimum — versus 120–150F² for SRAM-based CIM.

**Retention:** Ferroelectric remnant polarization provides non-volatile retention (>10 years at 85°C for mature HZO processes).

## Related Documentation

- `docs/documentation/module5-comparison/` — ELI5, features, physics, open-source tools
- `docs/comparison/` — Detailed physics/math comparisons, honesty audit
