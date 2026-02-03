# Module 6: EDA - Open-Source Tools

## When To Use External Tools

- Running full synthesis, placement, and routing flows.
- Verifying SPICE netlists at device or circuit level.
- Visualizing layout and parasitics.

## Recommended Tools (With Rationale)

- OpenROAD or OpenLane for open-source RTL to GDS flows.
- KLayout for layout inspection and editing.
- Yosys for synthesis and netlist generation.
- ngspice or Xyce for circuit simulation.

## Integration Notes

- EDA docs live in `docs/eda/README.md` and `docs/eda/guides/`.
- Export formats are implemented in `module6-eda/pkg/export/`.
- Use `docs/eda/references/cli-reference.md` for CLI usage.
