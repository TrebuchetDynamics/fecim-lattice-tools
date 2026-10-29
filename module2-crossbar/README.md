# Module 2: Crossbar Array Simulation

## Overview

Module 2 simulates ferroelectric crossbar arrays for Compute-in-Memory (CIM) operations. It models analog matrix-vector multiplication (MVM) and includes non-idealities such as IR drop, sneak paths, drift, temperature scaling, and variation.

**Simulation defaults:**
- Discrete conductance levels are **configurable** with a default of 30.
- Architectures supported: passive (0T1R) and transistor-isolated (1T1R).

All values in this module are **model parameters** for education and exploration, not device measurements.

---

## Key Features

- **MVM simulation** with DAC/ADC quantization
- **Differential arrays** for signed weights
- **Non-idealities**: IR drop, sneak paths, drift, variation, half-select disturb
- **Temperature scaling** for model parameters
- **Visualization**: heatmaps, parametric sweeps, and per-tab analysis
- **Weight I/O**: JSON/binary serialization for reuse
- **Optional training/inference helpers** for MNIST-style demos

---

## Quick Start

### Build and Run

```bash
# Build the main unified application
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools

# Or use the launch script
./launch.sh
```

### Browser crossbar candidate

Build either profile with the pinned tools listed in [web/README.md](../web/README.md).
Open `/modules/crossbar`. The existing CLI and Fyne launch defaults are unchanged.

- Hosted: Go CPU computation in a dedicated Worker. Maximum array: **32×32**.
- Local: authenticated same-origin Go sessions from `fecim-lattice-tools serve --root DIRECTORY`. Maximum array: **128×128**.
- Full sneak analysis is capped at **32×32** in both profiles. Disable sneak for larger local arrays.
- No GPU/WebGPU, hosted scientific uploads, persistence or automatic downloads are used.

Apply configuration explicitly. Select a zero-based row/column, then program a finite normalized cell state in `[0,1]`.
Programming uses 30 levels by default. Resistive profiles accept 2–256 programming levels.
FeCAP uses its native continuous capacitance state, not the resistive programming grid.
Inputs must contain exactly one finite voltage in `[0,1]` per matrix column.
Invalid settings, ragged matrices and invalid vectors do not replace the previous configuration or result.

| Profile | Production Go path | Output meaning |
|---|---|---|
| `basic` | `Array.MVM` | DAC/ADC-quantized normalized result, divided by input columns. Not amperes. Ideal reference uses a separate noiseless array. |
| `enhanced` | `MVMWithNonIdealities` | Normalized ideal/actual vectors with selected IR, sneak, seeded device variation and wire-temperature effects. Not amperes. |
| `parasitic` | `ParasiticSolver.SolveMVMContext` | Native **column** current sums in normalized units. Device voltages/currents, iterations, convergence and voltage residual are retained. No ADC/DAC, temperature or device-noise correction is applied to this solver. |
| `fecap` | `MVMCharge` | Raw charge in coulombs. The adapter explicitly transposes into the core's row-drive orientation, so web inputs still correspond to matrix columns. Ideal charge comes from a separate noiseless core array. No DC IR/sneak model is claimed. |

The tables expose programmed state, physical conductance in siemens or capacitance in farads, input/output vectors, IR maps, sneak traces and solver cells.
Use **Analysis mode** and **Previous/Next cells** to inspect complete data.
Numerical RMSE is not dataset classification accuracy. Latency, throughput and hardware/GPU efficiency claims are not reproduced.
Changed inputs label retained results **STALE**. A failed or cancelled job keeps earlier results.
Solver nonconvergence retains its diagnostics and reports failure, not success.

The stack view preserves the native illustrative layer-index gradient/noise workflow with a deterministic seed.
Choose 1–512 layers and select any layer through the table controls.
Only the selected layer is generated. Layers are not electrically coupled or measured-device data.
Matrix palettes change presentation only. Tables remain the primary cell-selection and reading surface.

**Cancel MVM** requests Go cancellation. Wait for Go's terminal state.
A lost native response leaves execution unknown and retains scientific ownership.
Use **Reconnect crossbar state** to recover authoritative state.
**Discard and restart crossbar** closes/drains the native session or terminates its Worker before releasing ownership.

**Files:** explicitly import a crossbar JSON document into memory, or load/save it through an opened local project.
The format contains `config`, a `rows × cols` `matrix`, and a `cols`-element `input`.
Imports are unverified and limited to 1 MiB. Local reads admit bounded regular files only.
Local writes use an exclusive temporary file and rooted rename.
Hosted Load/Save controls give local-run instructions and make no localhost request.

Automated checks cover real native/Node/Firefox scientific-byte parity, caps, cancellation and both browser transports.
These are candidate checks, not release or accessibility qualification.
Human Orca + Firefox/Linux acceptance remains **NOT RUN**.

### Command-Line Tools

```bash
# Crossbar GUI (standard mode)
go run ./cmd/fecim-lattice-tools crossbar

# Crossbar GUI (enhanced mode)
go run ./cmd/fecim-lattice-tools crossbar -enhanced

# Neural network inference (terminal)
go run ./cmd/fecim-lattice-tools crossbar inference -size=64 -layers=3
```

### Simple Example

```go
package main

import (
    "fmt"
    "math/rand"
    "fecim-lattice-tools/shared/crossbar"
)

func main() {
    cfg := &crossbar.Config{
        Rows:       64,
        Cols:       784,
        NoiseLevel: 0.02,
        ADCBits:    8,
        DACBits:    8,
    }
    array, _ := crossbar.NewArray(cfg)

    // Program random weights (quantized to the configured discrete levels)
    weights := make([][]float64, 64)
    for i := range weights {
        weights[i] = make([]float64, 784)
        for j := range weights[i] {
            weights[i][j] = rand.Float64() // [0, 1]
        }
    }
    array.ProgramWeightMatrix(weights)

    input := make([]float64, 784)
    for i := range input {
        input[i] = rand.Float64()
    }
    output, _ := array.MVM(input)
    fmt.Printf("Output shape: [%d]\n", len(output))

    opts := crossbar.DefaultMVMOptions()
    opts.Architecture = "0T1R"
    result, _ := array.MVMWithNonIdealities(input, opts)

    fmt.Printf("Ideal output:  %v\n", result.IdealOutput[:4])
    fmt.Printf("Actual output: %v\n", result.ActualOutput[:4])
    fmt.Printf("RMSE:          %.4f\n", result.RMSE)
}
```

---

## Core Concepts

### Discrete Conductance Levels

The simulator stores weights as discrete conductance levels. The default is 30 levels, but this is configurable.

```go
// Quantize any value to the nearest level
quantized := crossbar.QuantizeToLevels(0.542)  // Returns 16/29 on the default grid

// Get discrete level index (0-29 for default 30 levels)
level := crossbar.GetLevel(0.5)
```

### Matrix-Vector Multiplication (MVM)

```
I[i] = Sum(G[i][j] * V[j])
```

- Inputs are voltages
- Stored weights are conductances
- Physical Ohm/Kirchhoff currents and normalized API outputs are different quantities.
- Basic/enhanced MVM APIs return normalized, ADC-quantized outputs. Do not label them as amperes.
- Physical per-cell conductance is available through `GetPhysicalConductanceForCell`.

---

## API Quick Reference

### Creating Arrays

```go
cfg := &crossbar.Config{
    Rows:       128,
    Cols:       784,
    NoiseLevel: 0.02,
    ADCBits:    8,
    DACBits:    8,
}

cfg.ConductanceModel = crossbar.ConductanceExponential
cfg.Endurance = crossbar.DefaultEnduranceConfig()
cfg.ProcessVariation = crossbar.DefaultProcessVariationConfig()
cfg.HalfSelect = crossbar.DefaultHalfSelectConfig()

array, err := crossbar.NewArray(cfg)
if err != nil {
    log.Fatal(err)
}
```

### Programming Weights

```go
err := array.ProgramWeight(row, col, 0.75)  // Quantized to configured levels

weights := make([][]float64, 128)
for i := range weights {
    weights[i] = make([]float64, 784)
}
err := array.ProgramWeightMatrix(weights)
```

### Ideal MVM

```go
output, err := array.MVM(input)
```

### MVM with Non-Idealities

```go
opts := &crossbar.MVMOptions{
    EnableIRDrop:     true,
    EnableSneakPaths: true,
    EnableVariation:  true,
    EnableDrift:      false,
    Temperature:      300.0,
    Architecture:     "0T1R", // "0T1R" or "1T1R"
}

result, err := array.MVMWithNonIdealities(input, opts)
```

### Differential Arrays (Signed Weights)

```go
pos, _ := crossbar.NewArray(cfg)
neg, _ := crossbar.NewArray(cfg)

diffArr := crossbar.NewDifferentialArray(pos, neg)
diffArr.ProgramWeights(signedWeights)

output, _ := diffArr.MVM(input)
```

---

## Testing

```bash
cd <repo-root>
go test ./module2-crossbar/...
go test ./workbench/modules/crossbar ./shared/viewmodel/crossbar ./shared/crossbar -count=1
env CGO_ENABLED=1 go test -race ./workbench/modules/crossbar ./shared/viewmodel/crossbar -count=1
env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -run TestModuleProductionParityCrossbar -count=1 -v
# From web/, after both production builds:
# npm run test:e2e -- test/e2e/crossbar.spec.ts
# FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/crossbar.spec.ts
```

---

## Related Documentation

- [Crossbar learning notes](../docs/modules/crossbar/README.md)
- [Test conventions](../docs/internals/testing/TESTING.md)
- `CLAUDE.md` for project-wide guidance

Note: Some research notes under `docs/` include external claims and literature summaries. Treat them as references only; verified claims are tracked in [the honesty audit](../docs/research/honesty-audit.md).
