# Shared Package

Cross-module utilities, physics engines, UI components, and infrastructure shared across all FeCIM modules. This is the largest support package in the repository.

## Overview

The `shared/` package prevents duplication by centralizing code used by two or more modules: physics models (Preisach, Landau, ISPP), peripheral circuit simulations (ADC, DAC, TIA, charge pump), GUI widgets, theming, accessibility, export pipelines, logging, and more.

## Package Structure

### `physics/` — Core Physics Library

- **preisach.go** — Shared Preisach model implementation
- **landau.go** — Landau-Khalatnikov solver
- **ispp.go**, **ispp_write.go**, **ispp_legacy.go** — ISPP write-verify algorithms
- **material.go** — Material parameter database
- **calibration.go** — Calibration fitting routines
- **cell_geometry.go** — Cell geometry and voltage derivation from E_c and thickness
- **conductance.go** — Conductance ↔ polarization mapping
- **quantization.go** — Discrete level quantization
- **transfer.go** — Transfer curve modeling
- **device_variation.go** — Process variation and mismatch
- **units.go** — Unit conversion helpers
- **write_verify_stats.go** — Write-verify statistics tracking

### `peripherals/` — Peripheral Circuit Models

- **adc.go** — 5-bit SAR ADC with INL/DNL
- **dac.go** — 5-bit DAC with nonlinearity
- **tia.go** — Transimpedance amplifier
- **chargepump.go** — Dickson charge pump
- **sample_hold.go** — Sample-and-hold circuit
- **voltage_regulator.go** — Voltage regulator model
- **pvt.go** — Process/Voltage/Temperature variation
- **spice.go** — SPICE netlist export
- **analysis.go** — INL/DNL, timing, power analysis
- **defaults.go** — Default configurations

### `widgets/` — Reusable Fyne Widgets

- **glossary.go** — Glossary tooltip widget
- **material_picker.go**, **material_card.go**, **material_table.go** — Material selection UI
- **architecture_selector.go**, **architecture_toggle.go** — Architecture choice widgets
- **educational_panel.go**, **educational_animations.go** — Learning-mode components
- **interactive_tutorials.go**, **tutorial_controller.go** — Step-by-step tutorials
- **preset_browser.go**, **preset_selector.go** — Preset management UI
- **notification.go** — Toast notifications
- **embedded_app.go**, **embedded_base.go** — Embedded app framework
- **demo_controller.go**, **demo_mode_selector.go** — Demo mode infrastructure
- **accessibility.go**, **accessibility_helpers.go** — A11y support
- **tooltip_helpers.go**, **tooltips.go** — Rich tooltips
- And more (layout helpers, color legends, status, undo toolbar, etc.)

### `themes/` — Theme System

- **themes.go** — Theme definitions (light, dark, high-contrast)
- **manager.go** — Theme lifecycle management
- **switcher.go** — Runtime theme switching
- **accessibility_theme.go** — High-contrast accessible theme

### `presets/` — Preset System

- **types.go** — Preset data structures
- **manager.go** — Preset CRUD and persistence
- **builtin.go** — Built-in presets (HZO, AlScN, cryogenic, etc.)
- **providers.go** — Per-module preset providers
- **global.go** — Global preset registry

### `cli/` — CLI Helpers

- **cli.go** — Shared CLI flag parsing and output formatting

### `logging/` — Logging

- **logging.go** — Structured logger
- **buffer.go** — Ring-buffer log capture

### `export/` — Export Pipeline

- **export.go** — Multi-format export (PNG, CSV, JSON)
- **export_progress.go** — Progress tracking for batch exports
- **widget.go** — Export dialog widget

### `io/` — I/O Utilities

- **json_helpers.go** — JSON read/write helpers

### `compute/` — GPU Compute

- **compute_pipeline.go**, **context.go**, **dispatcher.go** — Compute pipeline abstraction
- **shader_loader.go** — Shader loading
- **buffer.go** — GPU buffer management

### `gpu/` — GPU Neural Network

- **gpu.go** — GPU context
- **dense_layer.go** — Dense layer GPU kernel
- **params.go** — GPU parameters

### `recording/` — Screen/Audio Recording

- **manager.go** — Recording session manager
- **ffmpeg.go** — FFmpeg integration
- **audio.go**, **buffer_pool.go**, **settings.go**, **types.go**

### Other Packages

- **help/** — Help system, embedded topics, tip-of-the-day, browser launcher
- **keyboard/** — Global keyboard shortcuts and help dialog
- **progress/** — CLI and GUI progress bars
- **recentfiles/** — Recent file tracking and menu
- **accessibility/** — Accessibility preferences
- **errors/** — Error types and panic recovery
- **undo/** — Undo/redo command framework
- **utils/** — Drawing, font, path discovery, PNG metadata, recovery
- **assets/equations/** — Embedded equation images
- **validation/** — Crossbar validation tools (shared with `validation/` module)
- **theme/** — Legacy theme (single-file)

## Key Types and Functions

| Type / Function | Package | Description |
|---|---|---|
| `PreisachModel` | `physics` | Shared Preisach engine |
| `LandauSolver` | `physics` | L-K ODE solver |
| `ISPPWriter` | `physics` | Write-verify loop |
| `CellGeometry` | `physics` | Voltage from E_c × thickness |
| `ADC`, `DAC`, `TIA` | `peripherals` | Peripheral circuit models |
| `ChargePump` | `peripherals` | Dickson charge pump |
| `PresetManager` | `presets` | Preset CRUD with persistence |
| `ThemeManager` | `themes` | Runtime theme switching |
| `EmbeddedBase` | `widgets` | Base for embeddable module apps |

## Testing

```bash
# Run all shared tests
go test ./shared/...

# Key sub-packages
go test -v ./shared/physics/...
go test -v ./shared/peripherals/...
go test -v ./shared/presets/...
go test -v ./shared/widgets/...

# With race detector
go test -race ./shared/...
```

## Related Documentation

- `docs/3-develop/architecture/ARCHITECTURE.md` — Overall project architecture
- `docs/3-develop/testing/TESTING.md` — Testing conventions
- `docs/4-research/internal-analysis/circuits.CIM-fundamentals.md` — Peripheral circuit design docs
