<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module1-hysteresis/pkg/gui

## Purpose

Provides Fyne-based GUI for hysteresis module. Visualizes P-E curves, manages simulation controls (play/pause/reset), material selection, waveform generation (sine, triangle, sawtooth), calibration dialog, data logging, and literature comparison overlays. Handles UI updates from goroutines using `fyne.Do()` and embedding the app as a tabbed module within the main application.

## Key Files

| File | Description |
|------|-------------|
| `gui.go` | Main GUI window (31KB). Canvas for P-E plot, control panels, material picker, waveform selector. Manages simulation lifecycle. |
| `simulation.go` | Time-stepping simulation engine (102KB). Integrates Preisach model, generates waveforms, updates plot data. |
| `controls.go` | Control panel widgets: frequency slider, amplitude input, preset selector, export button (30KB). |
| `embedded.go` | Implements embedded app interface: `BuildContent()`, `Start()`, `Stop()`. |
| `forc_panel.go` | FORC (First-Order Reversal Curve) panel for advanced analysis. |
| `literature_overlay.go` | Overlay comparison with literature curves. |
| `data_logger.go` | CSV export of simulation data. |

## For AI Agents

### Working In This Directory

**GUI Update Protocol (CRITICAL):**

- **NEVER** update UI directly from goroutine
- **ALWAYS** use `fyne.Do(func() { ... })` to marshal updates to main thread
- All simulation updates must go through `fyne.Do()` or Fyne's data binding
- See `embedded.go` for `Start()` and `Stop()` lifecycle hooks

**Embedded App Interface:**

Module1 must implement:

```go
type EmbeddedApp interface {
    BuildContent() fyne.CanvasObject  // Returns top-level widget (usually container)
    Start()                            // Called when tab activated; start simulation
    Stop()                             // Called when tab deactivated; pause simulation
}
```

**Simulation State Management:**

- `simulation.go` runs in goroutine; all P-E data updates must use `fyne.Do()`
- Plot widget listens for data changes via Fyne binding
- Material changes trigger simulation reset
- Pause state is atomic; use RWMutex for safe access

**Common UI Tasks:**

- Material picker: `preset_provider.go` loads presets
- Waveform selection: Drives sine/triangle/sawtooth time-stepping in `simulation.go`
- Export: Calls `data_logger.go` to write CSV
- Plot rendering: `render.go` in ferroelectric package handles canvas drawing

### Testing Requirements

```bash
# Run all GUI tests
go test ./module1-hysteresis/pkg/gui -v

# Run simulation tests (time-stepping validation)
go test ./module1-hysteresis/pkg/gui -run TestSimulation -v

# Run embedded app interface test
go test ./module1-hysteresis/pkg/gui -run TestEmbedded -v

# Run UI sync tests (goroutine-UI safety)
go test ./module1-hysteresis/pkg/gui -run TestUISync -v

# Run material picker tests
go test ./module1-hysteresis/pkg/gui -run TestMaterialPicker -v

# Run mode switch tests (safety of preset changes)
go test ./module1-hysteresis/pkg/gui -run TestModeSwitch -v
```

### Common Patterns

- **Goroutine-safe updates**: `fyne.Do(func() { widget.Refresh() })`
- **Material switching**: Reset simulation state, update preset provider, trigger redraw
- **Waveform control**: Change `simulation.waveform` and `frequency` atomically
- **Export data**: Collect plot history, format CSV, write to file
- **Pause/resume**: Use atomic flag for pause state; simulation goroutine respects it

## Dependencies

### Internal

- `module1-hysteresis/pkg/simulation` - Time-stepping engine
- `module1-hysteresis/pkg/ferroelectric` - Preisach model, material presets
- `module1-hysteresis/pkg/controller` - ISPP WriteController (optional, for calibration mode)
- `shared/physics` - Material models, quantization
- `shared/widgets` - Custom Fyne widgets (tooltips, material picker, etc.)
- `shared/theme` - Application theme
- `shared/logging` - GUI logging

### External

- `fyne.io/fyne/v2` - Cross-platform GUI framework
- `image/color` (Go stdlib) - Color definitions
- Standard library: `sync`, `time`, `fmt`, `os`, `path/filepath`

<!-- MANUAL: Last edited 2026-02-13. Fyne-based GUI stable; use fyne.Do() for goroutine safety. -->
