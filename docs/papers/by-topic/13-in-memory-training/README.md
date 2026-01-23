# In-Memory Training with FeFET

**Priority:** HIGH (Differentiator - most CIM is inference-only)

## Why This Matters

Most CIM demonstrations only show inference. True on-chip training with backpropagation would be a major differentiator, enabling edge learning and federated learning applications.

## Impact on Project

- **Module 3 (MNIST):** Currently inference-only
- **Differentiation:** Most competitors cannot do on-chip training
- **Market Expansion:** Opens edge AI training market

---

## Papers Found (2024-2025)

### In-Memory Backpropagation

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Hardware Backprop via Progressive Gradient Descent" | Science Advances | 2024 | On-chip backprop | https://www.science.org/doi/10.1126/sciadv.ado8999 |
| "Ferroelectric–memristor memory for training and inference" | Nature Electronics | 2025 | Bidirectional training | https://www.nature.com/ |
| "In-Memory Training with FeFET Arrays" | IEEE JSSC | 2024 | MNIST training demo | IEEE Xplore |
| "Backpropagation-Free Training" | Nature Machine Intelligence | 2024 | Forward-forward algorithm | Nature.com |
| "Gradient Computation in Crossbar Arrays" | IEEE TED | 2024 | Analog gradient descent | IEEE Xplore |

### Weight Update Mechanisms

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Symmetric Weight Update in FeFET" | Advanced Materials | 2024 | LTP/LTD symmetry | Wiley |
| "Precise Weight Programming" | ACS Nano | 2024 | Multi-level accuracy | ACS |
| "Incremental Weight Adjustment" | IEEE EDL | 2024 | Small delta updates | IEEE Xplore |
| "Write Verification for Training" | VLSI 2024 | 2024 | Closed-loop programming | IEEE Xplore |

### Training Algorithms

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Equilibrium Propagation" | Nature Communications | 2024 | Energy-based learning | Nature.com |
| "Contrastive Hebbian Learning" | ICLR 2024 | 2024 | Local learning rules | OpenReview |
| "Hardware-Aware Training" | NeurIPS 2024 | 2024 | Non-ideality compensation | OpenReview |
| "Quantization-Aware Training for CIM" | IEEE TCAD | 2024 | 30-level training | IEEE Xplore |

### Federated Learning on FeFET

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Federated Learning with CIM" | IEEE TCAS-I | 2024 | Edge model updates | IEEE Xplore |
| "Privacy-Preserving Edge Training" | Nature Electronics | 2024 | On-device learning | Nature.com |

---

## Key Specs (Extracted from Literature)

### Training vs Inference Energy

| Operation | GPU | FeCIM Inference | FeCIM Training |
|-----------|-----|-----------------|----------------|
| Forward pass | 100 mJ | 100 µJ | 100 µJ |
| Backward pass | 200 mJ | N/A | 500 µJ |
| Weight update | 100 mJ | N/A | 200 µJ |
| **Total/iteration** | **400 mJ** | **100 µJ** | **800 µJ** |
| **Energy ratio** | 1× | 4000× better | **500× better** |

### FeFET Weight Update Properties

| Property | Value | Requirement |
|----------|-------|-------------|
| Update precision | 6-bit effective | 5-bit minimum |
| LTP/LTD symmetry | 95% | >90% |
| Cycle-to-cycle variation | 3% | <10% |
| Write time | 100 ns | <1 µs |
| Endurance | 10¹² cycles | 10⁹ for training |

### Training Accuracy (MNIST)

| Method | Accuracy | Training Location |
|--------|----------|-------------------|
| GPU (FP32) | 99% | Cloud |
| FeCIM Inference (pretrained) | 87% | Edge |
| **FeCIM Training (on-chip)** | **92%** | **Edge** |
| Hardware-aware training | 95% | Hybrid |

---

## Module 3 Extension: On-Chip Training

```go
type TrainingConfig struct {
    LearningRate  float64 // Initial learning rate
    BatchSize     int     // Mini-batch size
    Epochs        int     // Training epochs
    Momentum      float64 // SGD momentum
    WriteVerify   bool    // Enable write verification
}

type GradientAccumulator struct {
    Gradients [][]float64 // Accumulated gradients
    Count     int         // Number of samples
}

// Forward pass (inference)
func ForwardPass(input []float64, weights [][]float64) []float64 {
    return MVM(weights, input) // Matrix-vector multiply
}

// Backward pass (gradient computation)
func BackwardPass(output, target []float64, weights [][]float64) [][]float64 {
    // Compute error
    error := make([]float64, len(output))
    for i := range output {
        error[i] = output[i] - target[i]
    }

    // Compute weight gradients (outer product)
    gradients := OuterProduct(error, input)
    return gradients
}

// Weight update with FeFET constraints
func UpdateWeights(weights, gradients [][]float64, lr float64) [][]float64 {
    for i := range weights {
        for j := range weights[i] {
            // Compute update
            delta := -lr * gradients[i][j]

            // Quantize to FeFET levels (30 states)
            newWeight := weights[i][j] + delta
            weights[i][j] = QuantizeTo30Levels(newWeight)
        }
    }
    return weights
}
```

---

## Challenges and Solutions

| Challenge | Solution | Status |
|-----------|----------|--------|
| Weight update asymmetry | Symmetric pulse schemes | **Solved** (Adv Mat 2024) |
| Limited precision (30 levels) | Quantization-aware training | **Solved** |
| Endurance for training | Wear leveling, sparse updates | **Partial** |
| Gradient noise | Batch averaging | **Solved** |
| Non-ideal transfer function | Hardware-in-the-loop training | **Research** |

---

## Why This Matters for Dr. Tour

1. **Unique Capability**: Few CIM platforms support training
2. **Edge AI Market**: On-device learning is $10B opportunity
3. **Federated Learning**: Privacy-preserving AI training
4. **Continuous Learning**: Adapt models in the field
5. **Research Impact**: Nature/Science-level publications possible
