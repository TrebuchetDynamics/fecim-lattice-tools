# Module 3: MNIST

Hardware-aware MNIST inference and training for FeCIM crossbar arrays. Compares floating-point baselines against quantized CIM inference with configurable noise, conductance levels, and non-idealities.

## Overview

Module 3 bridges neural network workloads with the ferroelectric crossbar simulator. It loads MNIST digit images, runs inference through multi-layer networks mapped onto crossbar arrays, and quantifies accuracy degradation from quantization, device noise, and analog non-idealities. Supports both full-precision (software) and CIM (hardware-aware) inference modes side-by-side.

## Local web candidate

Build the browser-enabled executable from the repository root. See the
[web build guide](../web/README.md) for the pinned Go, Node and npm versions.

```bash
make build-browser
./fecim-lattice-tools serve --root "$PWD/module3-mnist/data"
```

Open the terminal launch link, then choose **MNIST Pipeline** (`/modules/mnist`).
The default desktop launch remains legacy Fyne. This candidate is not a qualified
release. Real Orca + Firefox/Linux acceptance is **NOT RUN**.

1. Select a granted root and load an exact model path, such as `pretrained_weights.json`.
   Select an explicit QAT file to change models. Missing or invalid files do not select a substitute.
2. Apply uniform PTQ, per-layer PTQ or the model's declared quantization levels.
   PTQ requantizes existing weights. It does not train a new model.
   Ideal, Hardware and Noisy presets set the simulated peripherals and noise.
3. Draw with a pointer, set pixels with the keyboard, or enter 784 JSON pixels.
   **Use drawing** applies the unchanged native normalization once in Go.
   **Use unlabelled input** bypasses it. Dataset samples also bypass drawing normalization.
4. Run paired inference. Read FP/CIM predictions, all ten probabilities and logits,
   hidden activations, quantization statistics and paged weight values.
   Unlabelled inference has no accuracy value.
5. Load ordered IDX image/label files, select a sample or an explicit contiguous
   subset, then evaluate or sweep levels, ADC bits or noise. No samples are silently dropped.
   Read actual sample/correct counts and both confusion matrices. Select a sweep
   case to inspect its confusion counts.
6. Export the complete Go result JSON. It includes exact model/corpus SHA-256
   fingerprints, input fingerprints where applicable, source, preprocessing,
   subset, seed, requested quantization and resolved Go network configuration.
   File digests identify the bytes
   selected by the user, not their rights or independent scientific validity.

The [native port](../workbench/modules/mnist/) uses authenticated same-origin
module sessions and opaque rooted resource readers. No constructor opens cwd/home
files. Model loading is transactional. Each job builds an isolated CPU network.
Changing inputs marks retained results stale. Cancelling or a numerical failure
keeps the previous complete result. A cancelled job reports partial counts, not
completed accuracy. Navigation is not cancellation. If transport is lost, use
**Reconnect authoritative state** or **Cancel MNIST job**. Do not resubmit an
unknown job as if it had failed. In local mode, the tab stores only an opaque
module-session identifier in `sessionStorage`. Reload recovers the authenticated
server session, active job and retained results. Model, input and result bytes
are not stored there. A visible warning explains when browser storage is unavailable.

Hosted MNIST supplies help and escaped, bounded, read-only result text inspection.
Imported results remain **unverified**. Hosted mode has no inference, model/data
upload, server probing, scientific persistence or result export. Ordinary static
asset requests still occur.

### Initial local admission policy

These limits are policy bounds, not measured memory or performance guarantees.
The standalone training CLI retains its existing capabilities.

| Resource | Local web limit |
|---|---|
| Model | 32 MiB, 784→H→10, H=1..1024, at most 2,000,000 stored weight/bias scalars |
| Calibration | Optional exact 10×784 direct weights; unavailable when absent, never random fallback |
| Values | Finite matrices/scales/offsets; non-ragged shapes and matching optional biases |
| Input | Exactly 784 finite pixels in [0,1] |
| Quantization/peripherals | 2..`shared/neural.MaxDemoLevels` levels (currently 31), noise 0..0.20, ADC/DAC 3..16 bits |
| Corpus | At most 10000 labelled 28×28 images; combined files at most 32 MiB, decoded IDX at most 16 MiB |
| Work | Estimated allocation at most 256 MiB/job; `2 * images * cases * (784*H + 10*H)` at most 1e11 MACs |
| Sweep | At most 64 cases and an explicitly selected subset of at most 500 images |
| Lifetime | 15-minute job deadline; cancellation checked at each image and case |

All seven committed H=128 model files and the full committed 10000-image test
corpus are reachable in the native tests. Synthetic structural fixtures test
bounds only. They are not accuracy evidence. There is no fixed 80% baseline,
synthetic fallback accuracy or required accuracy threshold.

### Training workflow classification

New web training controls are excluded. Existing `cmd/train-network`,
`cmd/train-single-layer`, `cmd/train-ptq` and `cmd/mnist --train` remain intact.
The mounted dual-mode legacy GUI provides inference, weights and sweeps. Its
“Expert Mode → Train Weights” message is an existing stale instruction, not an
implemented mounted training control or a promised web workflow. The older GUI's
training-package dependency also supplies network inference; it does not establish
a mounted training UI. This task does not retire Fyne or change launch defaults.

## Package Structure

### `pkg/core/` — Compatibility Engine

`shim.go` aliases the engine in [`shared/neural`](../shared/neural/).
That package owns paired inference, quantization, noise/peripheral simulation,
metrics and sweeps. `model_bytes.go` admits local web models before constructing a
network. `preprocess.go` holds the unchanged drawing normalization shared by both UIs.

### `pkg/mnist/` — Dataset

- **loader.go** — Existing native/CLI MNIST IDX file parser
- **local_bytes.go** — Bounded local-web reader decoder with paired-label validation and file identity

### `pkg/training/` — Training Utilities

- **network.go** — Backpropagation training loop
- **single_layer.go** — Single-layer perceptron trainer
- **trainer_foundation.go** — Foundation trainer with learning rate scheduling
- **seed.go** — Reproducible random seeding

### `pkg/gui/` — Legacy Fyne Compatibility

- **embedded.go** — Embeddable app for unified launcher
- **dualmode.go** — Dual-mode (FP vs CIM) comparison view
- **dualmode_controls.go** — Parameter controls for dual-mode
- **dualmode_inference.go** — Inference execution and display
- **dualmode_weights.go** — Weight visualization (FP vs quantized)
- **dualmode_demo.go** — Demo/walkthrough mode
- **canvas.go** — Digit drawing canvas for live inference
- **energy_widget.go** — Energy breakdown visualization
- **quantization_widget.go** — Quantization sweep controls
- **metrics.go** — Accuracy/loss display
- **activations.go** — Layer activation visualization
- **comparison_card.go** — Architecture comparison cards
- **weight_comparison.go** — Weight distribution comparison
- **network_controller.go** — GUI↔network bridge
- **preprocess.go** — Image preprocessing (center, normalize)
- **export.go** — Result export utilities

### `cmd/` — Entry Points

- **mnist/main.go** — CLI inference runner
- **../../cmd/fecim-lattice-tools** — Canonical Fyne launcher (`mnist` subcommand or `--module mnist`)
- **train-network/main.go** — Full network training
- **train-single-layer/main.go** — Single-layer training
- **train-ptq/main.go** — Post-training quantization workflow

## Key Types and Functions

| Type / Function | Package | Description |
|---|---|---|
| `DualModeNetwork`, `InferenceResult` | `shared/neural` | Paired FP/CIM inference |
| `NetworkConfig` | `shared/neural` | Resolved quantization, noise and peripheral configuration |
| `MNISTModel`, `DecodeMNISTModel` | `shared/neural` | Immutable admitted model bytes and isolated network construction |
| `EvaluateDualModeDataset` | `shared/neural` | Native confusion/metric oracle |
| `QuantizeWeights` | `shared/neural` | Weight quantization to N discrete levels |
| `DecodeLocalDataset` | `pkg/mnist` | Bounded image/label readers and exact corpus identity |
| `Port` | `workbench/modules/mnist` | Native sessions, jobs, inspection and evidence |

## Testing

```bash
# Run all module 3 tests
go test ./module3-mnist/...

# With race detector
go test -race ./module3-mnist/...

# Specific packages
go test -v ./module3-mnist/pkg/core/...
go test -v ./module3-mnist/pkg/gui/...
```

Focused web checks, from the repository root:

```bash
go test ./workbench/modules/mnist ./shared/viewmodel/mnist ./shared/neural ./module3-mnist/pkg/mnist -count=1
env CGO_ENABLED=1 go test -race ./workbench/modules/mnist ./shared/neural -count=1
env CGO_ENABLED=1 xvfb-run -a go test ./module3-mnist/pkg/gui -run 'Test.*Preprocess|TestAccuracySweepRequiresLabelledDataset' -count=1
go test ./shared/viewmodel -run TestAllModulesE2E -count=1
cd web
npm test && npm run typecheck && npm run build:hosted && npm run build:local
npm run test:e2e -- test/e2e/mnist.spec.ts
FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/mnist.spec.ts
```

`TestMNISTWebWorkflow` compares paired results against direct shared/neural
execution. `TestMNISTWebEvidence` checks actual IDX sample accounting, native
sweep equivalence, rooted rejection and cancellation. `TestMNISTWebAdmission`,
`TestMNISTWebWorkAdmission` and `TestLocalDatasetAdmission` check model/data/work
bounds. The preprocessing characterization preserves the native pixel convention.
The Playwright case uses an actual authenticated Go server and committed model/IDX
fixtures locally. Its direct Go oracle compares all inference values and both
confusion matrices. Hosted checks assert the absence of API requests and inference.
Automated keyboard/browser checks do not qualify human screen-reader speech.

## Physics Context

**Quantization:** Weights are mapped from floating-point to N discrete conductance levels (default 30), mimicking the finite states of a ferroelectric memory cell. Quantization error grows as levels decrease.

**CIM Inference:** Matrix-vector multiplication is performed in the analog domain: input voltages × stored conductances = output currents. Non-idealities (noise, IR drop, sneak paths, drift) degrade the analog computation.

**Energy Model:** Energy estimates include DAC conversions, crossbar read currents,
TIA amplification and ADC quantization. They are simulation outputs, not measured
hardware energy or evidence of an energy advantage over digital MACs.

**Key metrics:** Top-1 accuracy (%), accuracy delta (FP − CIM), energy per inference (pJ), throughput (inferences/s).

## Related Documentation

- [MNIST learning material](../docs/modules/mnist/) — Concepts and model context
- [API reference](../docs/internals/api-reference.md) — Development notes
- [CLI reference](../docs/guides/cli-reference.md) — Command-line use
- [Web build and transport guide](../web/README.md) — Both profiles and qualification limits
