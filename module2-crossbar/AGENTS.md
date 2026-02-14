<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module2-crossbar

## Purpose

Ferroelectric crossbar array simulation module. Implements matrix-vector multiplication (MVM), non-ideality modeling (IR drop, sneak paths, drift, half-select disturb), endurance tracking, process variation, and GPU-accelerated computation for inference workloads.

## Key Files

| File | Purpose |
|------|---------|
| `pkg/crossbar/array.go` | Main crossbar array data structure and MVM engine |
| `pkg/crossbar/enhanced.go` | Non-ideality models (IR drop, sneak paths, drift) |
| `pkg/crossbar/gpu_mvm.go` | GPU-accelerated MVM via Vulkan compute |
| `pkg/gui/tabbed_app.go` | Fyne GUI with 4-tab interface (Ideal, IR Drop, Sneak, Drift) |
| `pkg/network/network.go` | Network layer for chaining crossbars |

## Subdirectories

| Directory | Purpose | Key Files |
|-----------|---------|-----------|
| `cmd/crossbar/` | Standalone CLI entry point | `main.go` |
| `pkg/crossbar/` | Core array simulation and non-idealities | `array.go` (46 files, extensive testing) |
| `pkg/gui/` | Fyne GUI components and tabs | `tabbed_app.go`, `tabs/` subdirectory |
| `pkg/gui/tabs/` | Individual tab implementations | `ideal_tab.go`, `irdrop_tab.go`, `sneak_tab.go`, `drift_tab.go` |
| `pkg/network/` | Multi-layer network simulation | `network.go` (connects multiple arrays) |
| `pkg/training/` | Weight training utilities | `trainer.go` |
| `pkg/visualization/` | Plot rendering and analysis | `plotter.go` |
| `pkg/weights/` | Weight storage and I/O | `loader.go`, `saver.go` |
| `shaders/` | GPU compute shaders (Vulkan GLSL) | `mvm.comp`, `irdrop.comp`, `sneak.comp` |

## For AI Agents

### Working In This Directory

- **Crossbar fundamentals**: 16×16 default (configurable to 8×8, 32×32, 64×64).
- **Quantization**: Default 30 analog levels (configurable via `NumLevels`). Conference claim pending peer review.
- **Non-idealities**: IR drop, sneak paths, drift, endurance fatigue, half-select disturb, process variation.
- **GPU acceleration**: Optional Vulkan compute pipeline for MVM (enable via `UseGPU: true`).
- **Conductance models**: Linear, exponential, or lookup table based on calibration.
- **Noise injection**: Device-to-device variation (σ/μ coefficient), thermal noise, flicker noise.

### Testing Requirements

```bash
go test ./module2-crossbar/...
go test ./module2-crossbar/pkg/crossbar/...
go test ./module2-crossbar/pkg/gui/...
```

Key test files (46 tests in pkg/crossbar/):
- `array_test.go` - Basic array operations and MVM
- `boundary_test.go` - Edge cases (0V, MaxV, noise saturation)
- `coupled_nonidealities_test.go` - Interaction of IR drop + sneak + drift
- `drift_irdrop_test.go` - Drift with temperature and IR drop coupling
- `statistical_validation_test.go` - Monte Carlo validation of noise models
- `sneak_path_test.go` - Sneak path calculation and impact
- `endurance_tracking_test.go` - Fatigue modeling
- `half_select_disturb_test.go` - Half-select V/2 pulse effects
- `concurrent_stress_test.go` - Thread safety under load

### Common Patterns

- Array initialization: `crossbar.NewArray(cfg)` with config specifying rows, cols, noise level, ADC/DAC bits
- MVM operation: `array.MatrixVectorMultiply(inputVector)` returns output vector (may include non-idealities)
- Non-ideality application: Enabled via config flags in `crossbar.Config`
- GPU path: `array.MatrixVectorMultiplyGPU(inputVector)` (requires Vulkan context)
- Cell access: `array.GetConductance(row, col)` and `array.SetConductance(row, col, value)`

### Critical Bug Patterns (Do Not Re-Introduce)

1. **Sneak Path Double-Counting**: Sneak paths can interfere with IR drop calculation. Fixed: Decouple via separate models with interaction term.
2. **Half-Select Disturb Sign**: Can apply in wrong direction. Fixed: Validate direction matches V/2 mode (read vs write).
3. **Drift Bounds**: Drift can push conductance beyond [GMin, GMax]. Fixed: Clamp after drift application.
4. **Temperature Scaling**: Temperature effects on conductance and non-idealities. Fixed: Use consistent physics.yaml calibration.
5. **GPU Memory Alignment**: Vulkan std140 layout requires 4-byte alignment. Fixed: Validate struct packing in DACParams, ADCParams, TIAParams.

## Dependencies

### Internal

- `module1-hysteresis/` - Ferroelectric material models, ISPP controller
- `shared/physics/` - Conductance quantization, material presets
- `shared/widgets/` - GUI components (sliders, charts)
- `shared/theme/` - Styling
- `shared/logging/` - Structured logging
- `shared/compute/` - Vulkan context and compute pipeline support

### External

- Fyne v2 (`fyne.io/fyne/v2`) - Cross-platform native GUI
- gonum (`gonum.org/v1/gonum`) - Linear algebra (matrix operations)
- Vulkan SDK (optional, for GPU acceleration) - Compute shader compilation

## Configuration

Crossbar settings in `pkg/crossbar/Config`:
- `Rows`, `Cols` - Array dimensions
- `NoiseLevel` - Device-to-device σ/μ (typical: 0.01-0.05)
- `ADCBits`, `DACBits` - Peripheral resolution (typical: 6-8 bits)
- `UseGPU` - Enable GPU acceleration (requires Vulkan)
- `ConductanceModel` - Linear, exponential, or lookup
- `ConductanceTable` - Calibration lookup table (length 30)
- `Endurance`, `ProcessVariation`, `HalfSelect` - Sub-configs for non-idealities

## Performance Notes

- MVM operation: ~1-10ms per vector (CPU, 16×16 array)
- MVM with GPU: ~0.1-1ms per vector (Vulkan compute, 16×16 array)
- Non-idealities overhead: ~20-40% slowdown when all enabled
- Memory: ~1-5MB per array (conductance state + noise tables)
- Concurrent operations: Thread-safe via atomic operations on cell state

## GPU Acceleration

- **Vulkan compute shaders** in `shaders/` directory
- **Pipelines**: mvm.comp (MVM), irdrop.comp (IR drop), sneak.comp (sneak paths)
- **Parameters**: Passed via uniform buffers (std140 layout required)
- **Optional**: Falls back to CPU if Vulkan unavailable

<!-- MANUAL: -->
