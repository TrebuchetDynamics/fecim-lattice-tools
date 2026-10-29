# Module 4: Peripheral Circuits


## Overview

Module 4 provides **simulation models** of peripheral circuits essential for ferroelectric CIM systems. Each cell in a crossbar array requires voltage control (via DAC) for writes and current sensing (via TIA/ADC) for reads. The values below reflect **default parameters** from `shared/peripherals/` and are not device measurements.

- **DAC**: Maps digital codes (0-31) to programmable voltages (-1.5V to +1.5V); demo uses 30 of 32 levels
- **TIA**: Converts crossbar column currents to measurable voltages (10 kOhm gain default)
- **ADC**: Quantizes sensed voltages back to digital codes (6-bit, 64 codes in the native and browser device defaults)
- **Charge Pump**: Boosts 1V CMOS supply to +/-1.5V write voltages (Dickson topology model)

The visualization supports interactive simulation with a Mode-First UX that automatically configures word lines and DAC ranges based on operation (READ, WRITE, or COMPUTE).

## Browser circuit workflow

Open `/modules/circuits` in the [shared React application](../web/README.md).
The screen uses a persistent Go session, not the legacy synthetic viewmodel matrix.
The default array is 8×8 with the FeCIM educational preset, 5-bit DAC and 6-bit ADC.
No hardware, accuracy, GPU or accessibility qualification is implied.

1. Select the material, node, architecture, coupling and sense settings. Apply the configuration explicitly.
2. Select a cell with the row and column fields. Set its target level and activate **WRITE selected cell**.
3. Inspect the actual pulse/verify trace. A pulse-limit result is not verified convergence.
4. Activate **READ selected cell**, then **COMPUTE vector**. Both operations use the retained programmed array.
5. Inspect the complete cell, vector and sense tables. Independent paging keeps retained results accessible after resizing.
6. Use Undo for the last device change, or explicitly reset or randomize the array. Changing material or dimensions resets levels.

Go supplies canonical values, display text, speech text, evidence, control choices and validation.
The plot is supplementary. It has no unique actions or per-point keyboard stops.
Timing playback uses native WRITE phase reference durations. It does not advance physical simulation time.
Logs retain the latest 64 actions. Stale results identify their original material, coupling, architecture and selected cell.

### Controls and scientific limits

- Native size choices are **1, 2, 4, 8, 16, 32, 64 and 128**, independently for rows and columns.
- Hosted admission is **at most 8×8** for all three solvers. Larger hosted ideal profiles are not enabled.
- Local Ideal and sparse TierB admit the native size choices. Dense TierA remains capped at **8×8** in this port.
  TierA uses the native dense reference solver. A 128×128 dense network would require impractical memory and cubic work.
- Nodes are **130, 65, 28 and 14 nm**. DAC resolution is **4–8 bits**, and ADC resolution is **5–8 bits**.
- TierB has a **1–4000** iteration cap. Circuit level ISPP has a **1–40** pulse cap. Jobs have a two-minute timeout.
- Vector codes are signed integers in **[-255, 255]**. The vector must match the column count exactly.
  Optional manual DAC voltages override vector codes for COMPUTE. Go rejects nonfinite values and voltages outside the material read range.
- Native sense presets, TIA gain and ADC references are explicit controls. Supply voltage is a peripheral reference setting, not a new array equation.
- 0T1R keeps every word line enabled. WRITE affects the whole selected column, not an isolated cell.
  The selected ISPP cell uses its coupled voltage. Neighbor updates preserve the native GUI assumption of applied column-drive voltage.
- 1T1R selects a row for READ/WRITE. COMPUTE accepts an explicit row mask.
  2T1R additionally accepts exact rectangular read and write connection masks.
- Cell currents retain their signs. Coupled row sensing retains the native magnitude polarity; ideal row currents remain signed.
  ADC codes are not programmed material levels.
- SNR reuses the native input-referred TIA, shot-noise and ADC-quantization composition.
  Column tables expose signed current sums and spatial voltage/coercive ratios, not endurance predictions.
  Nonfinite diagnostics have an explicit unavailable value, never a fabricated finite SNR.

The shared numerical path is [`arraysim/device.go`](pkg/arraysim/device.go).
Native GUI wrappers and the web port call these same helpers.
Circuit level ISPP uses the native deterministic LK pulse update and circuit calculator.
It is distinct from Hysteresis waveform programming. The separate animated native LK-controller workflow is not exposed as a browser engine.
Portable exponential evaluation reuses the existing Go Preisach/NLS helper to prevent platform-specific one-ULP drift.
No browser-side electronics model or result rounding is used.

### Deployment, recovery and exports

Hosted computation stays in a dedicated CPU Worker. Science is ephemeral.
There is no scientific upload, persistence, automatic download, file export or localhost discovery.
Cancel requests are nonterminal until Go confirms that the job stopped.
Cancelled jobs discard the private transaction and retain the previous committed array and result.
A lost response is an unknown outcome. Reconnect authoritative state before starting competing science.
Explicit session discard loses ephemeral state.

The local profile uses the authenticated same-origin Go service.
Start `fecim-lattice-tools serve --root DIRECTORY` with the source-built local assets.
Create or open a project beneath that granted root. In **Files and capabilities**, open the export project and supply a relative JSON path.
The explicit export contains configuration, persistent device state, retained results and operation logs.
A static local preview grants no filesystem capability. GPU execution and instrument access are unavailable in this port.

### Focused browser and numerical checks

```bash
go test ./workbench/modules/circuits ./shared/viewmodel/circuits ./module4-circuits/pkg/arraysim -count=1
env CGO_ENABLED=1 xvfb-run -a go test ./module4-circuits/pkg/gui -run 'TestCircuitsWebNativeDeviceParity|TestHeadlessPhysicsParity_GUIVsHeadless_ReadComputeWriteStep_MaterialAware' -count=1
env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -run 'TestModuleProductionParityCircuits|TestCircuitsHostedDependencyBoundary' -count=1 -v
cd web
npm test
npm run typecheck
npm run build:hosted
npm run build:local
npm run test:e2e -- test/e2e/circuits.spec.ts
FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/circuits.spec.ts
```

`TestCircuitsWebWorkflow` exercises the real port. `TestCircuitsWebNativeDeviceParity` compares actual GUI pulse, neighbor-state, READ and COMPUTE results.
The production parity gate compares original successful scientific bytes across native Go, Node and Firefox Workers.
Its receipts measure each admitted solver profile, resource observations and payload size.
Initial test gates are 15 seconds per command scenario, 128 MiB observed Go heap, 128 MiB Node WASM memory,
and combined Worker/runtime/module payloads of 6 MiB raw, 1.5 MiB Brotli and 2 MiB gzip.
These are private-host admission gates, not real-time guarantees or peak process-RSS measurements.
The original example-results byte gate remains separate.
Keyboard and automated accessibility checks do not qualify real Orca + Firefox/Linux use.
That manual acceptance remains **NOT RUN** for both profiles.

## Key Features

### ADC: 6-Bit Successive Approximation Register (Default Model)

- Resolution: 6 bits (64 codes, distinct from the 30-level FeCIM material preset)
- Input range: 0V to 1.0V (safe sensing below coercive voltage)
- Quantization: **round-to-nearest code (ties half-up)** after clamping to [VrefLow, VrefHigh]
  - Ideal code: `code = round((V - VrefLow)/(VrefHigh - VrefLow) * (2^bits - 1))`
  - LSB: `(VrefHigh - VrefLow)/(2^bits - 1)`
- INL: 0.5 LSB (Integral Nonlinearity)
- DNL: 0.25 LSB (Differential Nonlinearity)
- Conversion time: 50 ns (SAR architecture)
- ENOB: computed by `ADC.ENOB()` from resolution and INL/DNL
- Energy: ~30 fJ/conversion for the 6-bit SAR default (model estimate)

**Modeling**: ADC includes quantization errors via INL/DNL and optional SAR noise modeling.

### DAC: 5-Bit Voltage Generation (Default Model)

- Resolution: 5 bits (32 levels, demo uses 30)
- Output range: -1.5V to +1.5V (write voltage window)
- Settling time: 10 ns
- INL: 0.5 LSB
- DNL: 0.25 LSB
- Energy: ~15 fJ/conversion (model estimate)
- Supports both positive and negative write pulses

**Modeling**: Maps discrete levels to the configured write window.

### TIA: 10 kOhm Transimpedance Amplifier (Default Model)

- Gain: 10 kΩ (10 mV/µA)
- Bandwidth: 100 MHz
- Input-referred noise: 1 pA/sqrt(Hz)
- Output offset: 5 mV
- Maximum input current: 100 µA
- Maximum output voltage: 1.0V (matches ADC input range)
- Dynamic range: computed from defaults via `TIA.DynamicRange()`

**Modeling**: Current-to-voltage conversion with noise and output clamping.

### Charge Pump: 1V to +/-1.5V Boost (Default Model)

- Topology: 2-stage Dickson pump
- Input: 1V (standard CMOS supply)
- Ideal output: 3.0V (1V × 3 stages)
- Actual output: clamped to target voltage after modeled drops
- Clock frequency: 50 MHz
- Efficiency: 70%
- Output ripple: computed by `ChargePump.OutputRipple()` (default output cap = 10x flying cap)

**Modeling**: Dickson pump charge redistribution enables write voltage generation without off-chip supplies.

## Quick Start

### Build from Source

```bash
cd <repo-root>
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools
```

### Run CLI Demo

```bash
# Show all peripheral details
go run ./cmd/fecim-lattice-tools circuits cli -all

# Show all details with file logging
go run ./cmd/fecim-lattice-tools circuits cli -all -logger -verbosity 2

# Show specific circuit
go run ./cmd/fecim-lattice-tools circuits cli -dac
go run ./cmd/fecim-lattice-tools circuits cli -adc
go run ./cmd/fecim-lattice-tools circuits cli -tia
go run ./cmd/fecim-lattice-tools circuits cli -pump

# Show detailed analysis
go run ./cmd/fecim-lattice-tools circuits cli -linearity
go run ./cmd/fecim-lattice-tools circuits cli -timing
go run ./cmd/fecim-lattice-tools circuits cli -power

# Run ISPP write/verify demo (shared hysteresis physics)
go run ./cmd/fecim-lattice-tools circuits cli -ispp
```

### Run GUI Visualization

```bash
# Start unified application with Module 4 tab
go run ./cmd/fecim-lattice-tools/main.go

# Or run Module 4 GUI via subcommand
go run ./cmd/fecim-lattice-tools circuits
```

## Package Structure

### `shared/peripherals/` — Physics Models (Shared)

Core peripheral circuit implementations with realistic behavior:

- **adc.go** (123 lines)
  - `ADC` struct: 5-bit SAR configuration
  - `Convert()`: Ideal quantization
  - `ConvertWithNonlinearity()`: Realistic INL/DNL errors
  - `ENOB()`: Effective number of bits
  - `EffectiveSNR()`: Signal-to-noise accounting for errors
  - Methods: `Levels()`, `Resolution()`, `EnergyPerConversion()`, `TheoreticalSNR()`

- **dac.go** (90 lines)
  - `DAC` struct: 5-bit converter
  - `Convert()`: Level to voltage mapping
  - `ConvertWithNonlinearity()`: With INL/DNL errors
  - `VoltageRange()`: Min/max output
  - Energy and resolution calculations

- **tia.go** (101 lines)
  - `TIA` struct: Transimpedance amplifier
  - `Convert()`: Current to voltage with offset/clamping
  - `ConvertWithNoise()`: Thermal noise injection
  - `SNR()`: Signal-to-noise ratio for input current
  - Performance: `MinDetectableCurrent()`, `DynamicRange()`, `SettlingTime()`, `PowerConsumption()`

- **chargepump.go** (127 lines)
  - `ChargePump` struct: Dickson topology
  - `IdealOutputVoltage()`: Theoretical (N+1)×Vin
  - `ActualOutputVoltage()`: With threshold drops and IR losses
  - Efficiency analysis: `BoostFactor()`, `PowerInput()`, `PowerOutput()`, `PowerLoss()`
  - Dynamic behavior: `OutputRipple()`, `RiseTime()`, `MaxCurrentCapability()`

- **analysis.go** — Advanced Analysis
  - `AnalyzeINLDNL()`: Linearity metrics for DAC/ADC
  - `AnalyzeTiming()`: Critical path analysis
  - `AnalyzePower()`: Energy breakdown across components



- **embedded.go** (99 lines)
  - `EmbeddedCircuitsApp`: Pluggable app for unified visualizer
  - `BuildContent()`: Create tab content
  - `Start()`: Resume simulation
  - `Stop()`: Pause simulation

- **device_state.go** (616 lines)
  - `DeviceState`: Unified simulation state
  - `OpMode`: READ/WRITE/COMPUTE operation modes
  - `VoltageRange`: Material-derived safe voltage regions
  - `CalibrationParams`: Loaded from physics.yaml
  - **Passive Mode (0T1R)**: Forces all WLs on; ignores manual WL selection
  - **1T1R Mode**: Individual row control via word lines
  - `Compute()`: MVM computation with TIA/ADC signal chain

- **tab_unified.go** (1000+ lines)
  - `createUnifiedView()`: Mode-First UX layout
  - `createModeBar()`: READ/WRITE/COMPUTE buttons
  - `createWriteModePanel()`: Target level selection
  - `createComputeModePanel()`: Input vector controls
  - `createDACInputSection()`: Voltage presets with material calibration
  - `createMainSimSection()`: Array visualization
  - Dynamic voltage ranges from physics.yaml

- **helpers.go**, **drawing.go**, **font.go** — UI Utilities
  - Drawing primitives for arrays, waveforms
  - Color scheme management
  - Layout helpers

## Mode-First UX

The GUI presents three operation modes, each with automatically configured word lines and DAC ranges:

### READ Mode

- **Word Lines**: Single row active (user selectable)
- **DAC Voltage**: Safe read range (0 to ~0.5V, derived from material)
- **Purpose**: Non-destructive sense of cell state
- **Signal Chain**: DAC → Array → TIA → ADC
- **Read voltage** must stay below coercive voltage Vc to avoid polarization switching

### WRITE Mode

- **Word Lines**: Single row active (user selectable)
- **DAC Voltage**: Write range (Vc to ~2.5V, derived from material)
- **Purpose**: Program cell to target state via ferroelectric polarization
- **Signal Chain**: DAC → ChargePump → Array
- **Write voltage** must exceed Vc to switch polarization
- **Target Level** slider maps 0-29 to write voltage range

### COMPUTE Mode

- **Word Lines**: All rows active
- **DAC Voltage**: Safe compute range (0 to ~0.5V)
- **Purpose**: Matrix-vector multiplication across all rows
- **Signal Chain**: DAC → Array → TIA → ADC (per row)
- Each column gets independent input voltage from input vector
- Produces one output per active row

## Architecture Support

Module 4 supports multiple crossbar array architectures:

### Passive (0T1R)

- **Topology**: Direct FeFET gate-to-bit-line, no access transistor
- **Density**: 4F² (minimum, highest density)
- **Word Lines**: Always driven (no control)
- **Sneak Paths**: 5-20% error; mitigated via design
- **Best For**: Edge inference, mobile compute, ultra-high density
- **Mode Behavior**: All WLs always on; cannot change manually

### 1T1R

- **Topology**: 1 Transistor + 1 FeFET per cell
- **Density**: 8-12F² (lower than 0T1R)
- **Word Lines**: Gated by pass transistor; precise control
- **Sneak Paths**: ~0% (transistor isolates)
- **Best For**: Training, precise programming, neural networks
- **Mode Behavior**: WL selection controls active rows

### 2T1R

The browser port exposes separate read and write connection masks over the existing arraysim selector model.
This is an educational abstraction, not a transistor-level implementation or device validation.

### Array coupling and fidelity tiers

Ideal uses the material-aware independent-cell path.
TierA uses the dense DC nodal reference network. TierB uses the sparse iterative WL/BL network solver.
The browser port reports solver errors instead of silently falling back to Ideal.
Existing native CLI and legacy GUI contracts remain intact.
See [physics validation](../docs/research/physics-validation.md) for the validation boundary.

## Signal Chain

Complete signal path from digital command to digital readout:

```
WRITE OPERATION:
  Digital Level (0-29)
       ↓
    DAC (5-bit)
       ↓
  Charge Pump (1V → ±1.5V)
       ↓
  Ferroelectric Cell
       ↓
  Polarization State Change

READ OPERATION:
  Digital Command (row/col select)
       ↓
    DAC (sense voltage ~0.1-0.5V)
       ↓
  Ferroelectric Cell
       ↓
  Column Current (picoamps to nanoamps)
       ↓
    TIA (10 kΩ gain)
       ↓
  Voltage (0-1V range)
       ↓
    ADC (5-bit quantization)
       ↓
  Digital Level (0-29)

COMPUTE OPERATION (MVM):
  Input Vector [0-255 × N cols]
       ↓
    DAC × N (maps to read range)
       ↓
  Crossbar Array [M rows × N cols]
       ↓
  Row Currents = Sum(DAC[c] × G[r][c])
       ↓
    TIA × M (current-to-voltage)
       ↓
  Row Voltages
       ↓
    ADC × M (quantize)
       ↓
  Output Vector [0-29 × M rows]
```

## Material Calibration

Voltage ranges are automatically derived from material physics via `physics.yaml`:

### Configuration (physics.yaml)

```yaml
calibration:
  field_min_ratio: 0.7      # Before practical read-range clamping
  field_max_ratio: 2.5      # Write max = 2.5 × Vc (program window)
```

### Voltage Calculation

For a material with coercive voltage Vc = 2.0V:

| Parameter | Formula | Value |
|-----------|---------|-------|
| Read Min | 0 | 0V |
| Read Max | clamp(field_min_ratio × Vc, 0.1V, 1.0V) | 1.0V |
| Write Min | -min(field_max_ratio × Vc, 3.0V) | -3.0V |
| Write Max | +min(field_max_ratio × Vc, 3.0V) | +3.0V |
| DAC write limits | Default independent DAC reference window | -1.5V to +1.5V |

READ uses 0.4 × read maximum, with a 0.2V fallback when that value is below 0.1V.
COMPUTE maps input codes into the material read range.
WRITE requests use the bipolar material window, then the independent DAC limits and quantization.
A requested pulse does not imply that the target cell receives enough voltage to switch.

## Testing

All peripheral circuits include comprehensive tests.

Testing is intentionally **headless** and should remain safe to run in CI (no GUI initialization). For new work (architecture-aware coupling, fidelity tiers), prefer **table-driven** tests (`[]struct{...}` + `t.Run`) so expected behavior is explicit and reviewable.

See also: `docs/research/physics-validation.md` (validation strategy).



```bash
# Run all Module 4 tests
go test ./module4-circuits/...

# Run specific test suite
go test ./shared/peripherals -v

# Run with coverage
go test ./shared/peripherals -cover
```

### Test Coverage

- **DAC Tests**: Monotonicity, level mapping, round-trip with ADC
- **ADC Tests**: Quantization, clamping, ENOB calculation
- **TIA Tests**: Linearity, output clamping, noise behavior
- **ChargePump Tests**: Voltage boost, efficiency, power balance
- **Integration Tests**: Complete signal chain round-trips

Run `go test ./shared/peripherals/peripherals_test.go` to see all 10+ test cases.

## Related Documentation

For deeper understanding of ferroelectric physics, architecture trade-offs, and design decisions:

- **[Circuit Literature Review](../docs/research/literature-review/crossbar-circuits-literature-review-2025.md)** — Core concepts of compute-in-memory circuits
- **[Module 4 ELI5](../docs/modules/circuits/eli5.md)** — Beginner-friendly introduction to peripheral circuits
- **[Module 4 Physics](../docs/modules/circuits/physics.md)** — Physics equations and model details
- **[Physics Configuration](../config/physics/)** — Material properties and calibration parameters used for voltage range calculation

## Development Notes

### Key Conventions

- All voltages in volts (V), currents in amperes (A)
- Quantize to 30 levels: `crossbar.QuantizeTo30Levels(value)`
- Test before committing: `go test ./...`

### Converter defaults

The native and browser device paths default to a 5-bit DAC and 6-bit ADC.
The deprecated synthetic compatibility viewmodel retains its separate 5-bit ADC display default.
Its fixture tests are not material or coupling parity evidence.
All converter defaults are educational settings, not hardware-validated optima.

### Design Decisions

1. **5-bit DAC and 6-bit ADC defaults**: Keep converter codes distinct from the configurable material levels
2. **Dickson Charge Pump**: Simple, efficient topology suitable for CMOS integration
3. **10 kΩ TIA Gain**: Balances speed (100 MHz) and noise (1 pA/√Hz)
4. **Material/Geometry-Derived Ranges**: Voltage limits calculated from Ec and film thickness (via shared/physics.CellGeometry), not hardcoded
5. **Passive Mode Lock**: 0T1R architecture forces all WLs on; prevents misconfiguration

### Adding New Peripherals

To add a new circuit model (e.g., multiplexer):

1. Create `shared/peripherals/component.go` with struct and methods
2. Add tests in `shared/peripherals/component_test.go`
4. Integrate into `device_state.go` if part of signal chain
5. Add to README in this section

### Debugging Simulation

Enable detailed logging in `pkg/gui/helpers.go`:

```go
// In device_state.Compute():
if debugMode {
    log.Printf("Row %d: I=%.2e A, V=%.3f V, Level=%d", r, rowCurrent, rowVoltage, rowLevel)
}
```

## License

Part of fecim-lattice-tools. See repository LICENSE for details.

## Questions?

- See `docs/internals/api-reference.md` for function lookups
- Review `docs/research/honesty-audit.md` for physics verification
