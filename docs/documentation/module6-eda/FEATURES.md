# Module 6: EDA - Features

## What This Module Does

- Crossbar compiler and export tools

## Primary Components

- `module6-eda/pkg/compiler/compiler.go`
- `module6-eda/pkg/compiler/types.go`
- `module6-eda/pkg/export/spice.go`

## Key Workflows

- Set compile config -> run compile -> export artifacts.

## Extension Points

- Compiler with configurable crossbar sizes.
- Export to CSV, JSON, SPICE netlist formats.
- GUI tabs for compiler, export, and layout view.

## Known Limitations

- No timing closure or PDK-aware layout rules.
- Exports are illustrative, not fab-ready by default.

