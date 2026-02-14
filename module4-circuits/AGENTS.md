<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module4-circuits

## Purpose

Peripheral circuits simulation module for ferroelectric compute-in-memory systems. Implements DAC (Digital-to-Analog Converter), ADC (Analog-to-Digital Converter), and TIA (Trans-Impedance Amplifier) models with non-ideality simulation (INL/DNL, noise, bandwidth limitations). Includes array simulation with ISPP control, half-select modeling, and GPU-accelerated peripheral computation. Features unified device state management and comprehensive circuit visualization GUI.

## Key Files

| File | Purpose |
|------|---------|
| `pkg/gui/device_state.go` | Unified device state (68KB, core state machine) |
| `pkg/gui/tab_unified.go` | Main simulation tab with voltage and waveform controls |
| `pkg/gui/tab_unified_voltage.go` | Voltage mode selection and DAC/ADC configuration |
| `pkg/gui/app.go` | Main Fyne app with 4-tab interface (Unified, Reference, Comparison, Analysis) |
| `pkg/gpuperiph/gpu_peripherals.go` | GPU-accelerated DAC, ADC, TIA via Vulkan compute |
| `pkg/arraysim/types.go` | Crossbar array simulation types and interfaces |

## Subdirectories

| Directory | Purpose | Key Files |
|-----------|---------|-----------|
| `cmd/circuits/` | Standalone CLI entry point | `main.go` |
| `pkg/gui/` | Fyne GUI components and unified simulation | `app.go`, `device_state.go` (69KB), `tab_*.go` (6 tabs) |
| `pkg/gui/unified/` | Reusable unified simulation components | `renderer.go`, `state.go` |
| `pkg/arraysim/` | Array simulation with ISPP and write verify | `types.go`, `tier_a.go`, `tier_b.go`, `array_ispp.go` (40+ files) |
| `pkg/gpuperiph/` | GPU-accelerated peripheral circuits | `gpu_peripherals.go`, `gpu_peripherals_test.go` |
| `shaders/` | Vulkan compute shaders for peripherals | `dac.comp`, `adc.comp`, `tia.comp` |

## For AI Agents

### Working In This Directory

- **Operation modes**: READ (0-1V), WRITE (Vc to 1.3*Vc), COMPUTE (all rows active, input vector 0-1V).
- **Word line selection**: Single (one row), All (MVM), or Custom pattern.
- **DAC modes**: Manual, ReadPreset, WritePreset, InputVector, Random.
- **ISPP engine selection**: Level-based state machine (simple) or Landau-Khalatnikov solver (physics-accurate).
- **Device state**: Unified `DeviceState` struct manages all circuit parameters, crossbar config, voltage ranges.
- **Voltage ranges**: Derived from material physics (Ec, thickness) via physics.yaml calibration.

### Testing Requirements

```bash
go test ./module4-circuits/...
go test ./module4-circuits/pkg/gui/...
go test ./module4-circuits/pkg/arraysim/...
go test ./module4-circuits/pkg/gpuperiph/...
```

Key test files (90+ tests):
- `pkg/gui/device_state_test.go` - Device state transitions and validation
- `pkg/gui/physics_test.go` - Physics calculations and voltage ranges
- `pkg/gui/tab_unified_extended_test.go` - Unified tab operations
- `pkg/arraysim/tier_a_test.go` - Basic array simulation
- `pkg/arraysim/tier_b_test.go` - Advanced features (ISPP, write verify)
- `pkg/arraysim/kirchhoff_test.go` - Kirchhoff law solver for sneak paths
- `pkg/gui/module1_module4_integration_test.go` - Module1 (ISPP) ↔ Module4 integration
- `pkg/gpuperiph/gpu_peripherals_test.go` - GPU peripheral simulation

Pre-existing failures (do NOT fix, not in scope):
- `pkg/gui/TestUnifiedTabISPPEngine` - GUI tests don't build on clean main
- `pkg/gui/TestUnifiedActionButtons` - GUI tests pre-existing failure

### Common Patterns

- Device state init: `device_state.go:loadCalibrationParams()` reads physics.yaml
- Voltage calculation: `VoltageRange` struct with Min, Max, StepSize, NumLevels
- ISPP dispatch: `runISPPWithAnimation()` in `tab_unified_voltage.go` selects level-based or L-K engine
- Array simulation: `arraysim.NewSimulator(config)` for Tier A/B models
- GPU peripherals: `gpuperiph.NewGPUPeripherals()` creates Vulkan compute pipeline
- Operation modes: `OpMode` (Read/Write/Compute) and `WLMode` (Single/All/Custom)

### Critical Bug Patterns (Do Not Re-Introduce)

1. **OpMode / DACMode Mismatch**: OpMode controls row selection, DACMode controls column values. Can conflict. Fixed: Validate OpMode before DACMode in tab_unified.go.
2. **Voltage Range Derivation**: Must load from physics.yaml calibration (FieldMinRatio, FieldMaxRatio). Fixed: Call loadCalibrationParams() at init.
3. **Half-Select Model**: Half-select V/2 pulse effects on unselected cells. Fixed: Validate WLMode and HalfSelectConfig in arraysim.
4. **Read/Write Voltage Distinction**: Read uses safe zone (0-1V), Write uses write zone (Vc to 1.3*Vc). Fixed: OpMode gates DAC range.
5. **ISPP Engine Integration**: Level-based and L-K engines have different state representations. Fixed: Normalize output to level (0-NumLevels-1) in runISPPWithAnimation().
6. **GPU Memory Layout**: Vulkan std140 requires 4-byte alignment. Fixed: Validate DACParams, ADCParams, TIAParams struct packing.

## Dependencies

### Internal

- `module1-hysteresis/` - ISPP WriteController, Landau-Khalatnikov solver, material models
- `module2-crossbar/` - Crossbar array simulation (MVM, IR drop, sneak paths)
- `shared/physics/` - Material presets, quantization, calibration
- `shared/widgets/` - GUI components (sliders, charts, buttons)
- `shared/theme/` - Styling
- `shared/logging/` - Structured logging
- `shared/compute/` - Vulkan context and compute pipeline support
- `shared/peripherals/` - DAC, ADC, TIA model definitions

### External

- Fyne v2 (`fyne.io/fyne/v2`) - Cross-platform native GUI
- gonum (`gonum.org/v1/gonum`) - Linear algebra for array simulation
- Vulkan SDK (optional, for GPU acceleration) - Compute shader compilation and device runtime

## Configuration

Key device state config (from `device_state.go`):
- `OpMode` - Operation mode (Read, Write, Compute)
- `WLMode` - Word line selection (Single, All, Custom)
- `DACMode` - DAC voltage source (Manual, ReadPreset, WritePreset, InputVector, Random)
- `DACRangeMode` - DAC output range (Read safe zone or Write zone)
- Material selection - Via physics.yaml config file
- `VoltageRange` - Min/Max/StepSize for operation (derived from physics.yaml calibration)

Array simulation config (from `arraysim`):
- Crossbar dimensions (rows, cols)
- Quantization levels (1-30)
- Noise level (σ/μ, typical 0.01-0.05)
- Non-idealities (IR drop, sneak paths, drift, half-select disturb)
- Endurance tracking
- Process variation

Peripheral config (from `pkg/gpuperiph`):
- DAC resolution (Bits), reference voltages (VrefP, VrefN), INL/DNL
- ADC resolution, reference voltages, input-referred noise
- TIA feedback resistance (Rf), capacitance (Cf), op-amp GBW

## Performance Notes

- MVM operation: ~1-10ms per vector (CPU, 16×16 array)
- MVM with GPU: ~0.1-1ms per vector (Vulkan compute)
- ISPP write: ~50-200ms per cell (level-based or L-K engine, material-dependent)
- Peripheral simulation: ~0.1-1ms per conversion (CPU or GPU)
- GUI update rate: ~30-60 Hz (synced to Fyne render cycle)
- Memory: ~10-50MB for device state, array config, and history

## Module 4 Special Notes

- **Size**: Largest module (~90 test files, 69KB device_state.go)
- **Integration**: Bridges Module 1 (ISPP) and Module 2 (array) via unified device state
- **Circuit accuracy**: Physics-based models from literature (IEEE references in code)
- **Visualization**: 6-tab interface covering unified, reference, comparison, analysis modes
- **Flexibility**: Supports both simple level-based and complex physics-accurate ISPP

<!-- MANUAL: -->
