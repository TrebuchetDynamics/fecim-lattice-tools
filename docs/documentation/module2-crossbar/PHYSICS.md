# Module 2: Crossbar - Physics

## Prerequisites

- Ohm's law
- Matrix-vector multiplication
- Basic circuit networks

## Core Model

- Crossbar performs y = G * v where G is conductance matrix.
- Non-idealities (IR drop, sneak paths, drift) perturb ideal results.

## Key Equations (Simplified)

```
I = G * V
y_i = sum_j G_ij * v_j
```

## Parameters and Units

| Symbol | Meaning | Units |
|---|---|---|
| G | Conductance | Siemens |
| V | Input voltage | Volts |
| I | Output current | Amps |

## Assumptions and Limits

- Linear conductance model for ideal MVM.
- Noise injected as simple additive perturbations.

## Where It Lives in Code

- `module2-crossbar/pkg/crossbar/array.go`
- `module2-crossbar/pkg/crossbar/nonidealities.go`
- `module2-crossbar/pkg/gui/tabs/irdrop_tab.go`

## Sources

- `docs/development/scriptReference.md#demo-2-crossbar-module2-crossbar`
- `docs/ELI5.md`

