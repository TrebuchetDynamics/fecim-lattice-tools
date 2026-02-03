# Module 2: Crossbar - Open-Source Tools

## When To Use External Tools

- Validating array behavior with circuit-level solvers.
- Exploring layout-aware wire models.
- Scaling simulations beyond interactive GUI limits.

## Recommended Tools (With Rationale)

- ngspice or Xyce for circuit-level verification.
- KLayout for quick layout visualization and parasitic awareness.
- NumPy for large parameter sweeps and statistics.

## Integration Notes

- Crossbar parameters live in `module2-crossbar/pkg/crossbar/array.go`.
- Non-idealities are implemented in `module2-crossbar/pkg/crossbar/nonidealities.go`.
- For export flows, see `module6-eda/pkg/export/spice.go`.
