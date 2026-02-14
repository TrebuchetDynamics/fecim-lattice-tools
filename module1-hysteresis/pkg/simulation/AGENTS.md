<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module1-hysteresis/pkg/simulation

## Purpose

Provides time-stepping simulation engine for hysteresis module. Integrates Preisach model over user-defined waveforms (sine, triangle, sawtooth), manages simulation lifecycle (play/pause/reset), and maintains state history for plotting. Runs in goroutine and communicates with GUI via channels or Fyne data binding.

## Key Files

| File | Description |
|------|-------------|
| `engine.go` | Simulation engine (7.6KB). Manages waveform generation, time-stepping, Preisach integration, state history. |
| `multicell.go` | Multi-cell simulation wrapper (3.5KB). Applies waveforms to multiple cells with device variation. |
| `engine_test.go` | Unit tests for time-stepping, waveform generation, state management. |

## For AI Agents

### Working In This Directory

**Simulation Loop Structure:**

- Engine runs in goroutine from GUI's `Start()` method
- Each tick: (1) Generate voltage from waveform, (2) Convert to field, (3) Integrate Preisach model, (4) Store history
- Loop respects pause flag and can be stopped via context or internal state
- Time step (`dt`) is user-configurable; typical range 1e-4 to 1e-2 seconds

**Waveform Types:**

- `Sine`: Simple sinusoidal voltage
- `Triangle`: Linear ramps up/down
- `Sawtooth`: Asymmetric ramps (charge-discharge)
- Generate via `voltage(t)` function based on `waveform` enum and frequency

**State History:**

- Maintains circular history of voltage and polarization points
- Max history size: configurable (typically 1000-5000 points)
- Older points discarded when buffer full
- Accessed by GUI for plot rendering

**Thread Safety:**

- State protected by `sync.RWMutex`
- `running` and `paused` flags must be atomic or protected
- Preisach model is stateful; must not be accessed concurrently

### Testing Requirements

```bash
# Run all simulation tests
go test ./module1-hysteresis/pkg/simulation -v

# Run engine tests (time-stepping, waveform)
go test ./module1-hysteresis/pkg/simulation -run TestEngine -v

# Run multicell tests
go test ./module1-hysteresis/pkg/simulation -run TestMulticell -v
```

### Common Patterns

- **Waveform generation**: `voltage = amplitude * waveformFunc(time, frequency)`
- **Field conversion**: `field = voltage / thickness` (or thickness from material)
- **Preisach integration**: Each tick calls `model.Calculate(field)` to get P
- **History management**: Ring buffer with max capacity; discard old points when full
- **Pause/resume**: Flag-based pause; resume continues from same waveform phase

## Dependencies

### Internal

- `module1-hysteresis/pkg/ferroelectric` - Preisach model, material presets
- `shared/logging` - Engine logging
- `shared/physics` - Material models

### External

- `math` (Go stdlib) - Waveform generation (sin, etc.)
- `sync` (Go stdlib) - RWMutex for thread-safe state
- `time` (Go stdlib) - Simulation timing

<!-- MANUAL: Last edited 2026-02-13. Time-stepping logic stable; waveforms are standard. -->
