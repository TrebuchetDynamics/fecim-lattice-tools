# Module 1: Hysteresis - Physics

## Prerequisites

- Basic E-field concepts
- Units and plotting
- Simple integrals

## Core Model

- Polarization responds with history dependence, creating a loop in P-E space.
- Preisach model approximates the loop as a weighted sum of bistable elements.

## Key Equations (Simplified)

```
P(E) = integral integral mu(alpha,beta) * gamma_{alpha,beta}(E) d alpha d beta
Ec = coercive field; Pr = remanent polarization
```

## Parameters and Units

| Symbol | Meaning | Units |
|---|---|---|
| E | Electric field | V/m (or MV/cm) |
| P | Polarization | C/m^2 (or uC/cm^2) |
| Ec | Coercive field | V/m |
| Pr | Remanent polarization | C/m^2 |

## Assumptions and Limits

- Quasi-static loops (no high-frequency switching dynamics).
- Single-material parameterization (no gradients or grain effects).

## Where It Lives in Code

- `module1-hysteresis/pkg/ferroelectric/preisach.go`
- `module1-hysteresis/pkg/ferroelectric/material.go`
- `module1-hysteresis/pkg/gui/embedded.go`

## Sources

- `docs/video-transcripts/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md`
- `docs/development/scriptReference.md#quick-function-lookups`

