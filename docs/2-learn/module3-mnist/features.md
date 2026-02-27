# Module 3: MNIST - Features

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## What This Module Does

- Runs dual-path inference: full precision vs CIM.
- Visualizes activations and confidence differences in the dual-mode UI.
- Provides model-based energy visualization alongside modeled accuracy.

## Code Layout (Synced)

### Core inference path (runtime)
- `module3-mnist/pkg/core/` (dual-mode network, quantization, inference, energy model)
- `module3-mnist/pkg/gui/` (Fyne apps/widgets)
- `module3-mnist/pkg/mnist/` (MNIST loader)
- `module3-mnist/cmd/mnist-gui/main.go` (GUI entry)
- `module3-mnist/cmd/mnist/main.go` (CLI entry)

### Training/offline path
- `module3-mnist/pkg/training/` (training network utilities)
- `module3-mnist/cmd/train-network/main.go`
- `module3-mnist/cmd/train-ptq/main.go`
- `module3-mnist/cmd/train-single-layer/main.go`
- Legacy top-level scripts/utilities in `module3-mnist/*.go` are training helpers, not GUI runtime dependencies.

## Key Workflows

- Load pretrained weights and run inference.
- Adjust levels and noise from the UI (noise slider: `0.00-0.20`, clamped in core to match UI/docs).
- Use drawing canvas for live digit input.

## Accuracy Sweep Analysis (Analysis Tab)

The "Analysis" tab provides three parameter sweeps for hardware-aware accuracy degradation studies:

| Sweep | API | Description |
|-------|-----|-------------|
| Quantization levels | `SweepQuantizationLevels()` | Accuracy vs. number of conductance levels (2–30) |
| ADC bits | `SweepADCBits()` | Accuracy vs. ADC resolution (1–8 bits) |
| Noise level | `SweepNoiseLevel()` | Accuracy vs. relative noise (0–20%) |

Located in `shared/neural/accuracy_sweep.go`. GUI in `module3-mnist/pkg/gui/accuracy_sweep_panel.go`.

Results displayed as an ASCII bar chart in the GUI (no external dependencies required). Each sweep runs the inference engine with the parameter varied while holding other non-idealities at defaults.

**Interpretation notes:**
- Accuracy degradation is model-based and depends on network weights; not a silicon measurement.
- The 30-level baseline is the simulator default, not a validated hardware claim.

## Extension Points

- Add new weight files (QAT levels) and quantization schemes.
- Extend visualization for more layers or metrics.
- Add new sweep types to `accuracy_sweep.go` following the existing `SweepResult` pattern.
- Integrate other datasets for comparison.

## Known Limitations

- Training is offline; GUI focuses on inference.
- Small network is chosen for interactivity, not SOTA accuracy.
- Hardware non-idealities are simplified and modeled.
- Accuracy sweep results are for the bundled pre-trained weights and may differ with retrained networks.
