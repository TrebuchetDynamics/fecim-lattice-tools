# Module 2: Crossbar - Physics

## Prerequisites

- Ohm's law
- Matrix-vector multiplication
- Basic circuit networks

## Core Model

- A crossbar computes y = G * v, where G is conductance.
- Non-idealities such as IR drop and sneak paths perturb results.
- Quantization and noise approximate device limits.

## Key Equations (Simplified)

```
I = G * V
I_i = sum_j G_ij * V_j
V_drop ≈ I * R_wire
```

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| G | Conductance | Siemens |
| V | Input voltage | Volts |
| I | Output current | Amps |
| R_wire | Wire resistance | Ohms |

## Assumptions And Limits

- Linear conductance model for ideal MVM.
- Non-idealities are simplified and not device-calibrated.
- Noise is modeled as additive perturbations.

## Where It Lives In Code

- `module2-crossbar/pkg/crossbar/array.go`
- `module2-crossbar/pkg/crossbar/nonidealities.go`
- `module2-crossbar/pkg/crossbar/irdrop.go`
- `module2-crossbar/pkg/crossbar/sneakpath.go`

## Sources

- `docs/development/scriptReference.md#demo-2-crossbar-module2-crossbar`
- `docs/ELI5.md`
