# Example 02: MNIST First Layer

Compile the first layer of a trained MNIST classifier to FeCIM crossbar.

## Overview

This example demonstrates compiling a realistic neural network weight matrix - the first fully connected layer of an MNIST digit classifier (784→128).

**Note:** For practical demonstration, we use a 32x32 subset representing the first 32 weights of 32 output neurons.

## Files

| File | Description |
|------|-------------|
| `weights.json` | 32x32 subset of MNIST layer 1 weights |
| `run.sh` | Compilation and simulation script |
| `testbench.sp` | ngspice testbench for analog validation |

## Running the Example

```bash
cd module6-eda

# Compile weights
./examples/02-mnist-layer/run.sh

# Or manually:
go run ./cmd/eda-cli \
  -input examples/02-mnist-layer/weights.json \
  -output examples/02-mnist-layer/output \
  -rows 32 -cols 32 -levels 30
```

## Simulation with ngspice

After compilation, validate the crossbar behavior:

```bash
cd examples/02-mnist-layer/output

# Run simulation
ngspice -b ../testbench.sp -o sim_results.log

# View results
cat sim_results.log | grep "ibl"
```

### Expected Simulation Output

The testbench applies input voltages to wordlines and measures bitline currents:

```
Bitline currents for input vector [1, 0, 1, 0, ...]:
  BL[0] = 12.3 μA (sum of row 0,2 conductances × VDD)
  BL[1] = 8.7 μA
  ...
```

This demonstrates the **matrix-vector multiply (MVM)** operation:
```
I_out = G × V_in
```

## Weight Matrix Origin

The weights come from a simple MNIST classifier:

```python
# Training code (PyTorch)
class MNISTClassifier(nn.Module):
    def __init__(self):
        super().__init__()
        self.fc1 = nn.Linear(784, 128)  # 784×128 weights
        self.fc2 = nn.Linear(128, 10)

model = MNISTClassifier()
# ... train on MNIST ...
weights = model.fc1.weight.detach().numpy()  # Shape: (128, 784)

# Extract 32x32 subset for this example
subset = weights[:32, :32]
```

## Interpreting Results

### Level Distribution

For trained neural network weights, expect:
- Most weights near zero → levels 14-15 (middle range)
- Some positive outliers → levels 20-29
- Some negative outliers → levels 0-9

### PSNR

Quantization error is measured as PSNR:
- **> 40 dB:** Excellent - negligible accuracy loss
- **30-40 dB:** Good - minor accuracy degradation
- **< 30 dB:** Poor - may need more levels or retraining

### Conductance Mapping

```
Weight: -0.9 → Level: 0  → G: 1.0 μS   (high resistance)
Weight:  0.0 → Level: 15 → G: 50.5 μS  (mid resistance)
Weight: +0.9 → Level: 29 → G: 100.0 μS (low resistance)
```

## Extending to Full Network

For a complete MNIST classifier on FeCIM:

1. **Layer 1 (784×128):** 6 crossbars of 128×128 + partial
2. **Layer 2 (128×10):** 1 crossbar of 128×16 (padded)

Total: ~101,000 FeCIM cells

## Next Steps

1. Measure inference accuracy with quantized weights
2. Compare with floating-point baseline
3. Integrate with Module 3 (MNIST visualization) for end-to-end demo
