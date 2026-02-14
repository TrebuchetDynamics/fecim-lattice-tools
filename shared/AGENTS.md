<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# shared/ — Cross-Module Utilities and Physics

**Purpose:** Core physics engines, peripheral circuits, UI components, theming, presets, and infrastructure shared across all modules. This package prevents duplication by centralizing code used by 2+ modules.

**Status:** Production
**Stability:** High (mature APIs)
**Test Coverage:** 85%+ (see `go test ./shared/...`)

## Key Files

| File | Purpose | Key Type/Function |
|------|---------|-------------------|
| `physics/preisach.go` | Preisach hysteron model | `PreisachModel` |
| `physics/landau.go` | Landau-Khalatnikov solver | `LKSolver` |
| `physics/ispp_write.go` | Physics-accurate ISPP WriteController | `WriteController` |
| `physics/ispp.go` | Legacy waveform-based ISPP | Deprecated; use `ispp_write.go` |
| `physics/material.go` | Material parameter database | `HZOMaterial`, `AlScNMaterial` |
| `physics/calibration.go` | Fitting routines for P-E loops | `CalibrationPoint` |
| `physics/cell_geometry.go` | E_c and thickness → voltage mapping | `CellGeometry` |
| `physics/conductance.go` | Polarization ↔ conductance mapping | `ConductanceModel` |
| `physics/quantization.go` | Discrete level quantization (30-level default) | `QuantizeTo30Levels()` |
| `physics/transfer.go` | Transfer curve modeling | `TransferCurve` |
| `physics/units.go` | Unit conversion helpers (V/nm, µC/cm², etc.) | Conversion functions |
| `physics/write_verify_stats.go` | Write-verify convergence tracking | `WriteVerifyStats` |
| `peripherals/adc.go` | 5-bit SAR ADC with INL/DNL | `ADC` |
| `peripherals/dac.go` | 5-bit DAC with nonlinearity | `DAC` |
| `peripherals/tia.go` | Transimpedance amplifier | `TIA` |
| `peripherals/chargepump.go` | Dickson charge pump | `ChargePump` |
| `presets/manager.go` | Preset CRUD and persistence | `PresetManager` |
| `presets/builtin.go` | Built-in presets (HZO, AlScN, cryogenic) | Registration helpers |
| `widgets/embedded_base.go` | Base class for embeddable module apps | `EmbeddedAppBase` |
| `widgets/material_picker.go` | Material selection UI | `MaterialPicker` |
| `widgets/educational_panel.go` | Learning-mode UI components | `EducationalPanel` |
| `theme/theme.go` | Theme manager (light, dark, high-contrast) | `ThemeManager` |
| `logging/logging.go` | Structured logging | `Logger` |
| `export/export.go` | Multi-format export (PNG, CSV, JSON) | `Exporter` |

## Subdirectories

| Directory | Purpose | Key Exports |
|-----------|---------|-------------|
| `physics/` | Core physics engines: Preisach, Landau-K, ISPP, material params | 80+ Go files; regression test golden data |
| `peripherals/` | Peripheral circuit models: ADC, DAC, TIA, charge pump | Circuit behavioral simulators + SPICE export |
| `widgets/` | Reusable Fyne GUI components: material picker, tutorials, accessibility | 60+ UI widget files |
| `presets/` | Preset system: CRUD, builtin definitions, per-module providers | Manager + registry |
| `themes/` | Theme definitions: light, dark, high-contrast | `ThemeManager` |
| `theme/` | Legacy single-file theme (deprecated) | — |
| `logging/` | Structured logging with ring-buffer capture | `Logger` interface |
| `export/` | Export pipeline: PNG, CSV, JSON, progress tracking | `Exporter` interface |
| `cli/` | Shared CLI flag parsing and output formatting | Flag helpers |
| `io/` | I/O utilities: JSON read/write helpers | JSON marshal/unmarshal |
| `compute/` | GPU compute pipeline abstraction | `ComputePipeline` |
| `gpu/` | GPU neural network kernels | GPU dense layers |
| `recording/` | Screen/audio recording with FFmpeg | `RecordingManager` |
| `keyboard/` | Global keyboard shortcuts and help dialog | Shortcut registration |
| `progress/` | CLI and GUI progress bars | Progress reporters |
| `recentfiles/` | Recent file tracking and menu | Menu builder |
| `accessibility/` | Accessibility preferences (font size, high-contrast) | `AccessibilityPrefs` |
| `errors/` | Error types and panic recovery | Error helpers |
| `undo/` | Undo/redo command framework | `UndoManager` |
| `utils/` | Utilities: drawing, fonts, path discovery, PNG metadata | Helper functions |
| `assets/` | Embedded static assets: equation images, etc. | Asset loaders |
| `validation/` | Crossbar validation tools (shared with `validation/` module) | Validators |

## For AI Agents

### Working in This Directory

**Physics layers:** When implementing physics features:
1. Always use `physics/ispp_write.go` (WriteController) for ISPP—it uses the L-K solver and is physics-accurate
2. Never modify material parameters directly; use `material.go` and test with `physics/material_test.go`
3. Quantization is centralized in `physics/quantization.go`—all conductances must pass through `QuantizeTo30Levels()`

**UI patterns:** When adding widgets:
1. All widgets go in `widgets/`; follow the `EmbeddedAppBase` interface for module integration
2. Use `fyne.Do(func() {...})` for all UI updates from goroutines (thread-safety is critical)
3. Tooltips and glossary terms use `widgets/tooltips.go` and `widgets/glossary.go`

**Presets:** When adding new preset types:
1. Register in `presets/builtin.go`
2. Define provider in `presets/providers.go`
3. Add tests in `presets/presets_test.go`

**Exports:** When adding export formats:
1. Use `export/export.go` pipeline (PNG, CSV, JSON supported)
2. Check `export/export_progress.go` for batch export progress tracking

**Peripherals:** When modeling circuits:
1. Implement both behavioral and SPICE modes (see `peripherals/adc.go`)
2. Add noise models (INL/DNL for ADC/DAC)
3. Test with `peripherals/*_test.go`

### Testing Requirements

**Physics tests are critical:**
- All physics changes must pass `physics/landau_equation_test.go`, `physics/preisach_test.go`, `physics/ispp_writecontroller_test.go`
- Regression golden data is in `validation/testdata/physics_regression/` and regenerated via `FECIM_UPDATE_PHYSICS_GOLDEN=1`
- Use `TestISPPConverges_LandauK_Ensemble_Superlattice` for ISPP validation across 9 materials

**Widget tests:**
- All widget changes must pass their corresponding `_test.go` file
- GUI stress tests in `widgets/gui_stress_test.go` detect frame drops and hangs
- Use `go test -race ./shared/...` to catch concurrency bugs

**Peripheral tests:**
- ADC/DAC INL/DNL must stay within spec (see `peripherals/adc_test.go`)
- SPICE export must be syntactically valid (check with `ngspice -b`)

### Common Patterns

**ISPP convergence loop** (in `physics/ispp_write.go`):
```
WriteController.WriteLevel(targetLevel) → binary search over voltage
→ verify actual level via LK solver → apply pulses → repeat until converged or max iterations
```

**Material definition** (see `physics/material.go`):
```
HZOMaterial struct holds Landau coefficients (alpha, beta, gamma), Preisach params, thickness, etc.
Always test with calibration data from `data/calibrations/`
```

**Widget embedding** (see `widgets/embedded_base.go`):
```
Module apps inherit EmbeddedAppBase
Implement BuildContent(fyneApp, window) → fyne.CanvasObject
Implement Start() and Stop() for goroutine lifecycle
```

**Preset loading** (see `presets/manager.go`):
```
PresetManager.Load("preset-name") → returns Preset struct
Presets are YAML/JSON in `config/` and cached in memory
```

## Dependencies

### Internal
- `validation/` — Config validation rules
- `module1-hysteresis/` — Uses physics engines
- `module2-crossbar/` — Uses peripherals and physics
- `module3-mnist/` — Uses presets and export
- `module4-circuits/` — Uses peripherals and ISPP
- `module5-comparison/` — Uses all packages
- `module6-eda/` — Uses validation and export

### External
- `fyne.io/fyne/v2` — GUI framework (all widgets)
- `golang.org/x/image` — Image processing
- Standard library: `math`, `math/cmplx`, `sync`, `encoding/json`

## MANUAL

**Regression Testing:**
To regenerate physics golden data (after validated physics changes):
```bash
FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./validation/...
```
This updates `validation/testdata/physics_regression/preisach_loop_default_hzo.json` and similar files.

**Physics Debugging:**
- Use `physics/research_trace.go` to log intermediate solver steps
- ISPP overshoots are tracked in `physics/write_verify_stats.go`
- L-K solver convergence is logged to `isppLog` (set via `logging.NewLogger("ispp")`)

**Widget Thread Safety:**
- NEVER call `widget.Refresh()` directly from goroutines
- ALWAYS wrap UI updates in `fyne.Do(func() { widget.Refresh() })`
- Use `widgets/ui_lock.go` for critical sections

**Material Calibration:**
- Calibration files in `data/calibrations/` are JSON
- Schema: version, material_name, num_levels, calibrations[temp][fields]
- Load via `physics/calibration.go` → `LoadCalibration()` function

**Preset Persistence:**
- Presets are saved to `~/.fecim/presets/` or per-project directory
- Use `PresetManager.Save()` to persist custom presets
- Builtin presets are read-only (embedded in binary via `presets/builtin.go`)

**Export Progress:**
- Batch exports show progress via `export/export_progress.go`
- PNG export uses multi-threaded rendering
- CSV/JSON exports are streamed (low memory footprint)

**Logging:**
- Create loggers via `logging.NewLogger("component-name")`
- Log levels: Debug, Info, Warn, Error
- Ring buffer available for capturing last 1000 logs (see `logging/buffer.go`)

**Keyboard Shortcuts:**
- Register global shortcuts in `keyboard/keyboard.go`
- Test with `keyboard/keyboard_test.go`
- Shortcuts persist in `~/.fecim/shortcuts.json`

**Accessibility:**
- Check `accessibility/preferences.go` for user's font size, contrast preference
- All widgets must respect `AccessibilityPrefs`
- High-contrast theme in `themes/accessibility_theme.go`

