# Module 6: EDA - Physics

## Prerequisites

- Matrix dimensions
- Basic mapping constraints

## Core Model

- Mapping assigns each weight to a specific crossbar coordinate.
- Compile stats summarize utilization and fragmentation.

## Key Equations (Simplified)

```
Utilization = assigned_cells / total_cells
```

## Parameters and Units

| Symbol | Meaning | Units |
|---|---|---|
| Rows | Crossbar rows | count |
| Cols | Crossbar cols | count |

## Assumptions and Limits

- No placement-and-route; mapping is logical not physical.

## Where It Lives in Code

- `module6-eda/pkg/compiler/compiler.go`
- `module6-eda/pkg/compiler/types.go`
- `module6-eda/pkg/export/spice.go`

## Sources

- `docs/eda/README.md`
- `docs/eda/references/cli-reference.md`

