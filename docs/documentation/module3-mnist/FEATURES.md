# Module 3: MNIST - Features

## What This Module Does

- Neural inference: full-precision vs CIM

## Primary Components

- `module3-mnist/pkg/core/network.go`
- `module3-mnist/pkg/core/quantize.go`
- `module3-mnist/pkg/gui/dualmode.go`

## Key Workflows

- Load weights -> draw digit -> compare FP vs CIM prediction.
- Adjust ADC/DAC bits -> observe accuracy shift.

## Extension Points

- Dual-mode inference (FP vs CIM) with side-by-side metrics.
- Interactive drawing canvas for digit input.
- Quantization stats and activation visualizations.

## Known Limitations

- Weights are pre-trained and not re-trained inside the GUI.
- Model size is fixed for the demo workflow.

