# Module 1: Hysteresis - Features

## What This Module Does

- Ferroelectric memory cell physics (P-E curves, Preisach model)

## Primary Components

- `module1-hysteresis/pkg/ferroelectric/preisach.go`
- `module1-hysteresis/pkg/ferroelectric/material.go`
- `module1-hysteresis/pkg/gui/embedded.go`

## Key Workflows

- Adjust material parameters -> regenerate loop -> compare to reference curves.
- Switch between basic and advanced Preisach models.

## Extension Points

- Interactive P-E loop visualization with adjustable parameters.
- Preisach model variants (basic and advanced).
- Embedded app interface for the unified GUI.

## Known Limitations

- No spatial domain wall modeling.
- No direct fit to device-specific datasets in this module.

