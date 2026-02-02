# Module 5: Comparison - Physics

## Prerequisites

- Units (J, W, ops/s)
- Basic scaling intuition

## Core Model

- Energy per inference = energy per op * ops per inference.
- System scaling uses linear or simple aggregate models.

## Key Equations (Simplified)

```
E_infer = E_op * Ops_infer
Efficiency = Ops/s / W
```

## Parameters and Units

| Symbol | Meaning | Units |
|---|---|---|
| E_op | Energy per op | J |
| Ops/s | Throughput | operations/second |

## Assumptions and Limits

- High-level models; does not replace detailed power analysis.
- Comparisons depend on workload choice.

## Where It Lives in Code

- `module5-comparison/pkg/comparison/architecture.go`
- `module5-comparison/pkg/comparison/render.go`

## Sources

- `docs/development/scriptReference.md#demo-5-comparison-module5-comparison`
- `docs/comparison/HONESTY_AUDIT.md`

