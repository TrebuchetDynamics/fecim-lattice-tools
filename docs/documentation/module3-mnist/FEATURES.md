# Module 3: MNIST - Features

## What This Module Does

- Runs dual-path inference: full precision vs CIM.
- Visualizes activations and confidence differences.
- Supports calibration mode for single-layer hardware mapping.

## Primary Components

- `module3-mnist/pkg/core/network.go`
- `module3-mnist/pkg/core/network_inference.go`
- `module3-mnist/pkg/gui/dualmode.go`
- `module3-mnist/pkg/gui/canvas.go`

## Key Workflows

- Load pretrained weights and run inference.
- Adjust quantization, noise, ADC/DAC bits and compare outputs.
- Use drawing canvas for live digit input.

## Extension Points

- Add new weight files and quantization schemes.
- Extend visualization for more layers or metrics.
- Integrate other datasets for comparison.

## Known Limitations

- Inference only; training is offline.
- Small network chosen for interactivity, not SOTA accuracy.
- Hardware non-idealities are simplified.
