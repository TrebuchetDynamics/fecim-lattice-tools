# Module 6: EDA - Features

## What This Module Does

- Compiles networks into crossbar mappings.
- Exports mappings to CSV, JSON, and SPICE formats.
- Provides a GUI for compiler configuration and visualization.

## Primary Components

- `module6-eda/pkg/compiler/compiler.go`
- `module6-eda/pkg/export/csv.go`
- `module6-eda/pkg/export/json.go`
- `module6-eda/pkg/export/spice.go`

## Key Workflows

- Configure compiler settings and run compile.
- Export mapping to CSV/JSON for analysis.
- Generate SPICE netlists for downstream tools.

## Extension Points

- Add new export formats or mapping constraints.
- Improve tiling or placement strategies.
- Integrate with external layout workflows.

## Known Limitations

- No physical placement or routing optimization.
- SPICE export is structural, not calibrated to a PDK.
- Assumes array sizes and levels are already defined.
