# Module3 MNIST GPU Acceleration Plan

## Overview

GPU-accelerate the MNIST neural network inference in module3-mnist using the shared Vulkan compute infrastructure (`shared/compute/`).

## Current State Analysis

### Compute Profile (Per Inference)

| Operation | Count | % Time | GPU Benefit |
|-----------|-------|--------|-------------|
| MVM Layer 1 | 100,352 MACs | ~70% | **HIGH** |
| MVM Layer 2 | 1,280 MACs | ~15% | **HIGH** |
| ReLU | 128 ops | ~5% | LOW |
| Softmax | 10 exp + norm | ~5% | NONE |
| Quantization | 138 ops | ~3% | LOW |
| Noise | 138 ops | ~2% | LOW |

**Total**: 101,632 MACs per inference (two-layer: 784→128→10)

### Existing GPU Infrastructure

| Component | Location | Status |
|-----------|----------|--------|
| VulkanContext | shared/compute/context.go | ✅ Working |
| ComputePipeline | shared/compute/compute_pipeline.go | ✅ Working |
| GPUBuffer | shared/compute/buffer.go | ✅ Working |
| MVM Shader | shared/compute/shaders/mvm.comp.spv | ✅ Compiled |
| Activation Shader | shared/compute/shaders/activation.comp.spv | ✅ Compiled |
| Crossbar GPU | module2-crossbar/pkg/crossbar/gpu_mvm.go | ✅ Working (bugs fixed) |

### Current MNIST Code (CPU-only)

```go
// module3-mnist/pkg/core/network_inference.go:190-212
func (n *DualModeNetwork) forwardFP(input []float64) []float64 {
    // Layer 1: Pure CPU MVM
    for i := 0; i < len(weights); i++ {
        sum := bias[i]
        for j := 0; j < len(input); j++ {
            sum += weights[i][j] * input[j]  // O(rows × cols)
        }
        output[i] = sum
    }
    // ...ReLU, Layer 2, Softmax...
}
```

## Design Decisions

### 1. Integration Strategy

**Decision: Leverage crossbar GPU infrastructure directly.**

Rather than duplicating GPU code, module3-mnist will:
1. Create crossbar arrays for each layer
2. Use `crossbar.Array.MVM()` which auto-dispatches to GPU
3. Benefit from crossbar non-ideality modeling (IR drop, noise, sneak paths)

**Rationale:**
- Crossbar already has working GPU path
- Unified physics model (FeCIM non-idealities)
- Single maintenance point for GPU code

### 2. Batch Processing

**Decision: Add batch inference API for evaluation workloads.**

Single-inference GUI will remain responsive (CPU path acceptable).
Test set evaluation (10,000 samples) will use batch GPU dispatch.

```go
// New API
func (n *DualModeNetwork) InferBatch(images [][]float64) []InferenceResult
```

### 3. Activation Fusion

**Decision: Fuse ReLU with MVM output processing.**

The activation shader already exists. We can optionally fuse it with MVM output to reduce memory bandwidth.

## Proposed Architecture

### GPU-Accelerated Network Struct

```go
// module3-mnist/pkg/core/gpu_network.go
type GPUNetwork struct {
    layer1Array *crossbar.Array  // 784×128 crossbar
    layer2Array *crossbar.Array  // 128×10 crossbar
    useGPU      bool
    batchSize   int
}

// NewGPUNetwork creates a GPU-accelerated MNIST network
func NewGPUNetwork(config NetworkConfig) (*GPUNetwork, error) {
    // Create crossbar arrays with GPU enabled
    layer1Config := crossbar.Config{
        Rows:   784,
        Cols:   128,
        UseGPU: config.UseGPU,
    }
    layer1, err := crossbar.NewArray(layer1Config)
    if err != nil {
        return nil, fmt.Errorf("failed to create layer 1: %w", err)
    }

    layer2Config := crossbar.Config{
        Rows:   128,
        Cols:   10,
        UseGPU: config.UseGPU,
    }
    layer2, err := crossbar.NewArray(layer2Config)
    if err != nil {
        layer1.Destroy()
        return nil, fmt.Errorf("failed to create layer 2: %w", err)
    }

    return &GPUNetwork{
        layer1Array: layer1,
        layer2Array: layer2,
        useGPU:      config.UseGPU,
        batchSize:   config.BatchSize,
    }, nil
}
```

### GPU Forward Pass

```go
// Forward performs GPU-accelerated inference
func (n *GPUNetwork) Forward(input []float64) ([]float64, error) {
    // Layer 1: 784 → 128
    hidden, err := n.layer1Array.MVM(input)
    if err != nil {
        return nil, fmt.Errorf("layer 1 failed: %w", err)
    }

    // ReLU activation (CPU for now, could be GPU)
    for i := range hidden {
        if hidden[i] < 0 {
            hidden[i] = 0
        }
    }

    // Layer 2: 128 → 10
    output, err := n.layer2Array.MVM(hidden)
    if err != nil {
        return nil, fmt.Errorf("layer 2 failed: %w", err)
    }

    // Softmax (CPU - only 10 elements)
    return softmax(output), nil
}
```

### Batch Inference (Evaluation)

```go
// batch_inference.go - New file for batch GPU processing

// InferBatch performs batched GPU inference for evaluation
func (n *GPUNetwork) InferBatch(images [][]float64) ([]int, error) {
    predictions := make([]int, len(images))

    // Process in batches
    batchSize := n.batchSize
    if batchSize <= 0 {
        batchSize = 64 // default batch size
    }

    for i := 0; i < len(images); i += batchSize {
        end := i + batchSize
        if end > len(images) {
            end = len(images)
        }
        batch := images[i:end]

        // Process batch (GPU-accelerated)
        for j, img := range batch {
            output, err := n.Forward(img)
            if err != nil {
                return nil, err
            }
            predictions[i+j] = argmax(output)
        }
    }

    return predictions, nil
}
```

## Implementation Tasks

### Phase 1: Core Integration

#### Task 1.1: Add GPU Config to NetworkConfig

**File**: `module3-mnist/pkg/core/network.go`

Add UseGPU field to existing NetworkConfig:
```go
type NetworkConfig struct {
    // ... existing fields ...
    UseGPU    bool // Enable GPU acceleration
    BatchSize int  // Batch size for evaluation (default 64)
}
```

#### Task 1.2: Create GPUNetwork Wrapper

**File**: `module3-mnist/pkg/core/gpu_network.go` (NEW)

Create GPU-accelerated network using crossbar arrays.

#### Task 1.3: Update DualModeNetwork

**File**: `module3-mnist/pkg/core/network.go`

Add GPU path option:
```go
type DualModeNetwork struct {
    // ... existing fields ...
    gpuNetwork *GPUNetwork  // Lazy-initialized
    useGPU     bool
}

func (n *DualModeNetwork) initGPU() {
    if n.useGPU && n.gpuNetwork == nil {
        n.gpuNetwork, _ = NewGPUNetwork(n.config)
    }
}
```

### Phase 2: Batch Evaluation

#### Task 2.1: Add Batch Inference API

**File**: `module3-mnist/pkg/core/batch_inference.go` (NEW)

Implement `InferBatch()` for test set evaluation.

#### Task 2.2: Update Evaluation Loop

**File**: `module3-mnist/pkg/training/network.go`

Modify `Evaluate()` to use batch inference:
```go
func (n *TrainingNetwork) Evaluate(images, labels [][]float64) float64 {
    if n.network.useGPU {
        predictions, _ := n.network.InferBatch(images)
        // Count correct...
    } else {
        // Current serial evaluation
    }
}
```

### Phase 3: GUI Integration

#### Task 3.1: Add GPU Toggle to GUI

**File**: `module3-mnist/pkg/gui/mnist_tab.go`

Add checkbox to enable/disable GPU acceleration.

#### Task 3.2: Display GPU Status

Show whether GPU is being used and estimated speedup.

## Test Cases

### Test 1: GPU vs CPU Parity

```go
func TestGPUCPUParity(t *testing.T) {
    // Same weights, same input
    cpuNet := NewDualModeNetwork(config)
    cpuNet.useGPU = false

    gpuNet := NewDualModeNetwork(config)
    gpuNet.useGPU = true

    input := randomInput(784)

    cpuOut := cpuNet.Forward(input)
    gpuOut := gpuNet.Forward(input)

    // Allow 1e-4 tolerance (float32 vs float64)
    for i := range cpuOut {
        if math.Abs(cpuOut[i]-gpuOut[i]) > 1e-4 {
            t.Errorf("Mismatch at %d: CPU=%f GPU=%f", i, cpuOut[i], gpuOut[i])
        }
    }
}
```

### Test 2: MNIST Accuracy Maintained

```go
func TestMNISTAccuracyGPU(t *testing.T) {
    net := loadPretrainedNetwork()
    net.useGPU = true

    images, labels := loadMNISTTest()
    accuracy := net.Evaluate(images, labels)

    // Should maintain >96% accuracy
    if accuracy < 0.96 {
        t.Errorf("GPU accuracy too low: %f", accuracy)
    }
}
```

### Test 3: Batch vs Single Parity

```go
func TestBatchParity(t *testing.T) {
    net := NewGPUNetwork(config)
    images := loadMNISTTest()[:100]

    // Single inference results
    single := make([]int, 100)
    for i, img := range images {
        out, _ := net.Forward(img)
        single[i] = argmax(out)
    }

    // Batch inference results
    batch, _ := net.InferBatch(images)

    for i := range single {
        if single[i] != batch[i] {
            t.Errorf("Batch mismatch at %d: single=%d batch=%d", i, single[i], batch[i])
        }
    }
}
```

## File Changes Summary

### New Files

| File | Lines (est) | Purpose |
|------|-------------|---------|
| module3-mnist/pkg/core/gpu_network.go | 150 | GPU network wrapper |
| module3-mnist/pkg/core/batch_inference.go | 80 | Batch inference API |
| module3-mnist/pkg/core/gpu_network_test.go | 120 | GPU-specific tests |

### Modified Files

| File | Changes |
|------|---------|
| module3-mnist/pkg/core/network.go | Add UseGPU config, lazy GPU init |
| module3-mnist/pkg/training/network.go | Use batch inference in Evaluate() |
| module3-mnist/pkg/gui/mnist_tab.go | Add GPU toggle |

## Performance Expectations

### Single Inference

| Mode | Latency | Notes |
|------|---------|-------|
| CPU | ~500µs | Current baseline |
| GPU | ~200µs | First inference (includes transfer) |
| GPU (warm) | ~50µs | Subsequent inferences |

### Batch Evaluation (10,000 samples)

| Mode | Time | Speedup |
|------|------|---------|
| CPU (serial) | ~5s | 1× |
| GPU (batch 64) | ~500ms | **10×** |
| GPU (batch 256) | ~200ms | **25×** |

## Dependencies

- `shared/compute` package (already implemented)
- `module2-crossbar/pkg/crossbar` (GPU MVM - bugs fixed)
- Vulkan SDK with compute support
- No new Go dependencies

## Acceptance Criteria

1. [ ] GPU path produces same results as CPU (within 1e-4 tolerance)
2. [ ] MNIST test accuracy maintained (>96%)
3. [ ] Batch inference works correctly
4. [ ] GPU fallback works when Vulkan unavailable
5. [ ] GUI toggle for GPU mode
6. [ ] 10×+ speedup for test set evaluation
7. [ ] Existing CPU tests still pass

## Limitations (Documented)

1. **Precision**: GPU uses float32, may differ from CPU float64 by up to 1e-4
2. **First inference**: ~200µs overhead for GPU initialization
3. **Small batches**: For batch <16, CPU may be faster due to transfer overhead
4. **Training**: Training remains CPU-only (SGD backward pass not GPU-accelerated)
