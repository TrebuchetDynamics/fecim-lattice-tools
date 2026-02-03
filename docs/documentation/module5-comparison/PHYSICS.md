# Module 5: Comparison - Physics

## Prerequisites

- Basic performance metrics
- Power and energy concepts
- Log-scale charts and ratios

## Core Model

- Each architecture is described by compute, memory, and energy parameters.
- Workloads estimate latency, throughput, and energy.
- Comparisons focus on relative differences, not absolute claims.

## Key Equations (Simplified)

```
Energy = Power * Time
Throughput = Ops / Time
Efficiency = Ops / Energy
```

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| P | Power | Watts |
| t | Time | seconds |
| E | Energy | Joules |
| T | Throughput | ops/s |

## Assumptions And Limits

- Modeled numbers depend on configuration assumptions.
- Benchmarks are representative subsets.
- Comparisons should be interpreted with the honesty audit.

## Where It Lives In Code

- `module5-comparison/pkg/comparison/architecture.go`
- `module5-comparison/pkg/comparison/render.go`
- `module5-comparison/pkg/gui/widgets.go`

## Sources

- `docs/comparison/HONESTY_AUDIT.md`
- `docs/development/scriptReference.md#demo-5-comparison-module5-comparison`
