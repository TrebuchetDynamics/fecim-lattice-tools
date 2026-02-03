# Module 1: Hysteresis - Features

## What This Module Does

- Simulates ferroelectric hysteresis loops with a Preisach-based model.
- Exposes discrete polarization levels for memory state modeling.
- Provides GUI visualization and lab-bench controls.

## Primary Components

- `module1-hysteresis/pkg/ferroelectric/preisach.go`
- `module1-hysteresis/pkg/ferroelectric/material.go`
- `module1-hysteresis/pkg/gui/gui.go`
- `module1-hysteresis/pkg/render/plot.go`

## Key Workflows

- Sweep electric field to generate a P-E loop.
- Sample discrete states for quantized memory levels.
- Compare baseline and advanced Preisach behavior.

## Extension Points

- Add new material parameter sets in `material.go`.
- Extend hysteron distributions in `preisach_advanced.go`.
- Add alternative renderers or plotting styles.

## Known Limitations

- No full domain-level switching dynamics or fatigue modeling.
- Parameters are not device-calibrated by default.
- GPU renderer is specialized and may not cover all platforms.
