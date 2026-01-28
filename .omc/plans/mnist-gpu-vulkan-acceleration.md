# Plan: Module 3 MNIST GPU Vulkan Acceleration

## Overview

### Goals
Enable GPU-accelerated inference for Module 3 MNIST neural network, providing 5-10x speedup for batch inference while maintaining identical numerical results to CPU.

### Approach Summary
Create a new `shared/gpu/` package with clean neural network layer abstractions. This avoids inheriting Module 2's crossbar-specific physics (drift, IR drop, sneak paths) which are overkill for FP path inference. The new package provides:
1. Simple dense layer GPU operations (MVM + bias + activation)
2. Reusable by future neural network modules
3. Clean separation from crossbar physics simulation

### Key Decisions
- **Option B: Dedicated `shared/gpu/` package** (not reusing Module 2's mvm.comp)
- Rationale: Module 2's shader models crossbar physics; MNIST FP path needs pure matrix math
- Module 2's known MVM/VMM semantic issues (see `.omc/plans/module2-gpu-flawless.md`) avoided
- CIM path can optionally use Module 2's crossbar GPU once fixed

---

## Architecture

### New Package Structure

```
shared/
  gpu/                              # NEW: Neural network GPU acceleration
    gpu.go                          # GPUNetwork - high-level API
    dense_layer.go                  # DenseLayerGPU - single layer operations
    batch_inference.go              # Batch inference orchestration
    shaders/
      dense_mvm.comp                # Simple W*x + b matrix-vector multiply
      dense_mvm.comp.spv            # Compiled SPIR-V
      relu.comp                     # ReLU activation (in-place)
      relu.comp.spv                 # Compiled SPIR-V
      softmax.comp                  # Softmax activation (requires reduction)
      softmax.comp.spv              # Compiled SPIR-V
```

### Integration Points

```
module3-mnist/pkg/core/
  network_inference.go              # MODIFIED: Add GPU inference path
  network.go                        # MODIFIED: Add GPU accelerator field
  network_gpu.go                    # NEW: GPU-specific inference methods
```

### Interface Design

```go
// shared/gpu/gpu.go
type GPUNetwork interface {
    // Single-sample inference
    Forward(input []float32, weights []LayerWeights) ([]float32, error)

    // Batch inference (main speedup path)
    ForwardBatch(inputs [][]float32, weights []LayerWeights) ([][]float32, error)

    // Resource management
    IsAvailable() bool
    Destroy()
}

type LayerWeights struct {
    Weights []float32  // Row-major [OutFeatures x InFeatures]
    Bias    []float32  // [OutFeatures]
    Rows    int        // OutFeatures
    Cols    int        // InFeatures
    Activation ActivationType // RELU, SOFTMAX, NONE
}
```

---

## New Files

### Shaders

| File | Lines | Purpose |
|------|-------|---------|
| `shared/gpu/shaders/dense_mvm.comp` | ~60 | Simple MVM: out[i] = sum(W[i,j]*x[j]) + b[i] |
| `shared/gpu/shaders/relu.comp` | ~25 | In-place ReLU: x[i] = max(0, x[i]) |
| `shared/gpu/shaders/softmax.comp` | ~80 | Numerically stable softmax with reduction |

### Go Files

| File | Lines | Purpose |
|------|-------|---------|
| `shared/gpu/gpu.go` | ~150 | GPUNetwork struct, high-level API, context management |
| `shared/gpu/dense_layer.go` | ~200 | DenseLayerGPU: pipeline creation, buffer management |
| `shared/gpu/batch_inference.go` | ~120 | BatchInference: orchestrates multi-layer forward pass |
| `shared/gpu/params.go` | ~50 | Shader parameter structs with std140 alignment |
| `module3-mnist/pkg/core/network_gpu.go` | ~180 | GPU integration for DualModeNetwork |

### Test Files

| File | Lines | Purpose |
|------|-------|---------|
| `shared/gpu/gpu_test.go` | ~250 | GPU vs CPU parity tests |
| `module3-mnist/pkg/core/network_gpu_test.go` | ~200 | MNIST inference parity tests |

---

## Code Changes

### 1. module3-mnist/pkg/core/network.go

**Location:** Lines 40-78 (DualModeNetwork struct)

**Change:** Add GPU accelerator field

```go
// Before:
type DualModeNetwork struct {
    // ... existing fields ...
    rngMu sync.Mutex
}

// After:
type DualModeNetwork struct {
    // ... existing fields ...
    rngMu sync.Mutex

    // GPU acceleration
    gpuAccelerator *gpu.GPUNetwork  // nil if unavailable
    useGPU         bool             // Enable/disable GPU path
}
```

### 2. module3-mnist/pkg/core/network.go

**Location:** Lines 104-146 (NewDualModeNetwork function)

**Change:** Initialize GPU accelerator

```go
// After net := &DualModeNetwork{...} initialization:

// Try to initialize GPU acceleration
net.gpuAccelerator = gpu.NewGPUNetwork()
if net.gpuAccelerator != nil && net.gpuAccelerator.IsAvailable() {
    net.useGPU = true
}
```

### 3. module3-mnist/pkg/core/network_inference.go

**Location:** Lines 191-203 (forwardFP function)

**Change:** Add GPU dispatch option

```go
// Before:
func (net *DualModeNetwork) forwardFP(input []float64, weights [][]float64, bias []float64) []float64 {
    output := make([]float64, len(bias))
    // CPU nested loops...
}

// After:
func (net *DualModeNetwork) forwardFP(input []float64, weights [][]float64, bias []float64) []float64 {
    // Try GPU path if available and input is large enough
    if net.useGPU && len(input) >= 128 {
        result, err := net.forwardFPGPU(input, weights, bias)
        if err == nil {
            return result
        }
        // Fall back to CPU on GPU error
    }

    // CPU path (existing code)
    output := make([]float64, len(bias))
    for i := 0; i < len(weights); i++ {
        sum := bias[i]
        for j := 0; j < len(input); j++ {
            sum += weights[i][j] * input[j]
        }
        output[i] = sum
    }
    return output
}
```

---

## Shader Design

### dense_mvm.comp

```glsl
#version 450

layout(local_size_x = 256) in;

// Layer parameters (std140 layout)
layout(std140, binding = 0) uniform LayerParams {
    int rows;        // Output features
    int cols;        // Input features
    int activation;  // 0=none, 1=relu
    int padding;     // std140 alignment
};

// Weight matrix W[rows][cols] in row-major order
layout(std430, binding = 1) buffer WeightBuffer {
    float W[];
};

// Bias vector b[rows]
layout(std430, binding = 2) buffer BiasBuffer {
    float b[];
};

// Input vector x[cols]
layout(std430, binding = 3) buffer InputBuffer {
    float x[];
};

// Output vector y[rows]
layout(std430, binding = 4) buffer OutputBuffer {
    float y[];
};

void main() {
    uint rowIdx = gl_GlobalInvocationID.x;
    if (rowIdx >= rows) return;

    // Compute y[i] = sum(W[i,j] * x[j]) + b[i]
    float sum = b[rowIdx];
    for (int j = 0; j < cols; j++) {
        sum += W[rowIdx * cols + j] * x[j];
    }

    // Optional inline ReLU
    if (activation == 1) {
        sum = max(0.0, sum);
    }

    y[rowIdx] = sum;
}
```

### softmax.comp

```glsl
#version 450

layout(local_size_x = 256) in;

layout(std140, binding = 0) uniform SoftmaxParams {
    int size;        // Vector length
    int padding1;
    int padding2;
    int padding3;
};

layout(std430, binding = 1) buffer DataBuffer {
    float data[];
};

// Shared memory for reduction
shared float sharedMax;
shared float sharedSum;

void main() {
    uint idx = gl_GlobalInvocationID.x;

    // Pass 1: Find max for numerical stability
    if (idx == 0) {
        float maxVal = data[0];
        for (int i = 1; i < size; i++) {
            maxVal = max(maxVal, data[i]);
        }
        sharedMax = maxVal;
    }
    barrier();

    // Pass 2: Compute exp(x - max) and sum
    if (idx < size) {
        data[idx] = exp(data[idx] - sharedMax);
    }
    barrier();

    if (idx == 0) {
        float sum = 0.0;
        for (int i = 0; i < size; i++) {
            sum += data[i];
        }
        sharedSum = sum;
    }
    barrier();

    // Pass 3: Normalize
    if (idx < size) {
        data[idx] /= sharedSum;
    }
}
```

---

## Implementation Tasks

### Task 1: Create shared/gpu package structure
**Files:** `shared/gpu/*.go`, `shared/gpu/shaders/*.comp`

1.1. Create `shared/gpu/params.go` with LayerParams struct (std140 aligned)
1.2. Create `shared/gpu/shaders/dense_mvm.comp` shader
1.3. Create `shared/gpu/shaders/relu.comp` shader
1.4. Create `shared/gpu/shaders/softmax.comp` shader
1.5. Compile shaders: `glslangValidator -V dense_mvm.comp -o dense_mvm.comp.spv`

**Acceptance Criteria:**
- [ ] All .comp files compile to .spv without errors
- [ ] params.go struct sizes verified with compile-time assertions

### Task 2: Implement DenseLayerGPU
**File:** `shared/gpu/dense_layer.go`

2.1. Create DenseLayerGPU struct with pipeline and buffer management
2.2. Implement `NewDenseLayerGPU(ctx, maxRows, maxCols)` constructor
2.3. Implement `Forward(weights, bias, input []float32) ([]float32, error)`
2.4. Implement `Destroy()` for resource cleanup

**Acceptance Criteria:**
- [ ] DenseLayerGPU compiles without errors
- [ ] Buffers pre-allocated for max dimensions
- [ ] Proper error handling for GPU failures

### Task 3: Implement GPUNetwork
**File:** `shared/gpu/gpu.go`

3.1. Create GPUNetwork struct wrapping VulkanContext and DenseLayerGPU
3.2. Implement `NewGPUNetwork()` with graceful fallback
3.3. Implement `Forward(input, layers)` for single inference
3.4. Implement `ForwardBatch(inputs, layers)` for batch inference
3.5. Implement `IsAvailable()` and `Destroy()`

**Acceptance Criteria:**
- [ ] GPUNetwork initializes without error on systems with/without GPU
- [ ] IsAvailable() correctly reports GPU status
- [ ] Destroy() releases all resources

### Task 4: Add GPU parity tests
**File:** `shared/gpu/gpu_test.go`

4.1. TestDenseLayerParity: GPU vs CPU dense layer comparison
4.2. TestSoftmaxParity: GPU vs CPU softmax comparison
4.3. TestTwoLayerNetworkParity: Full 784->128->10 network test
4.4. TestGracefulFallback: Verify behavior when GPU unavailable

**Acceptance Criteria:**
- [ ] All tests pass on GPU systems
- [ ] All tests skip gracefully on non-GPU systems
- [ ] GPU/CPU parity within 1e-5 tolerance (float32 precision)

### Task 5: Integrate GPU into DualModeNetwork
**File:** `module3-mnist/pkg/core/network_gpu.go` (new file)

5.1. Add `InitGPU()` method to initialize GPU accelerator
5.2. Add `forwardFPGPU()` method for GPU forward pass
5.3. Add `SetUseGPU(bool)` to enable/disable GPU path
5.4. Add `DestroyGPU()` for cleanup

**Acceptance Criteria:**
- [ ] network_gpu.go compiles without errors
- [ ] GPU path produces results matching CPU path

### Task 6: Modify network.go and network_inference.go
**Files:**
- `module3-mnist/pkg/core/network.go` (lines 40-78, 104-146)
- `module3-mnist/pkg/core/network_inference.go` (lines 191-203)

6.1. Add gpuAccelerator and useGPU fields to DualModeNetwork struct
6.2. Initialize GPU in NewDualModeNetwork constructor
6.3. Modify forwardFP to dispatch to GPU when available
6.4. Add cleanup in any Destroy() method

**Acceptance Criteria:**
- [ ] All existing tests pass
- [ ] GPU path activates when available
- [ ] CPU fallback works when GPU unavailable

### Task 7: Add MNIST GPU integration tests
**File:** `module3-mnist/pkg/core/network_gpu_test.go` (new file)

7.1. TestMNISTInferenceParity: Compare GPU vs CPU inference results
7.2. TestBatchInferenceSpeedup: Benchmark batch processing
7.3. TestSingleLayerModeParity: Verify Tour mode works with GPU

**Acceptance Criteria:**
- [ ] GPU produces same predictions as CPU
- [ ] Probability distributions match within 1e-4
- [ ] Batch inference shows measurable speedup

### Task 8: Run full test suite
**Command:** `go test ./...`

8.1. Run all module3-mnist tests
8.2. Run all shared/gpu tests
8.3. Run all shared/compute tests
8.4. Verify no regressions

**Acceptance Criteria:**
- [ ] All 117+ existing tests pass
- [ ] New GPU tests pass (or skip gracefully)
- [ ] No build errors

---

## Testing Strategy

### Unit Tests

| Test | Description | Tolerance |
|------|-------------|-----------|
| DenseLayerParity | Single layer GPU vs CPU | 1e-5 |
| SoftmaxParity | Softmax GPU vs CPU | 1e-5 |
| ReluParity | ReLU GPU vs CPU | 0 (exact) |

### Integration Tests

| Test | Description | Tolerance |
|------|-------------|-----------|
| TwoLayerNetwork | Full 784->128->10 pass | 1e-4 |
| MNISTInference | Real MNIST sample inference | Same prediction |
| BatchInference | 100-sample batch | Same predictions |

### Benchmark Tests

| Benchmark | Expected Speedup |
|-----------|------------------|
| Single inference (small batch) | 1-2x (GPU overhead may dominate) |
| Batch inference (100 samples) | 5-10x |
| Batch inference (1000 samples) | 10-20x |

### GPU Availability Handling

```go
func TestWithGPU(t *testing.T) {
    gpu := NewGPUNetwork()
    if !gpu.IsAvailable() {
        t.Skip("GPU not available, skipping GPU test")
    }
    defer gpu.Destroy()
    // ... test code ...
}
```

---

## Acceptance Criteria

### Functional Requirements

| Criterion | Measurement |
|-----------|-------------|
| GPU/CPU parity | Predictions match on 1000 MNIST samples |
| Numerical precision | Probabilities match within 1e-4 |
| Graceful fallback | Works on CPU-only systems |
| No regressions | All 117+ existing tests pass |

### Performance Requirements

| Criterion | Target |
|-----------|--------|
| Batch inference (100 samples) | 5x speedup over CPU |
| Single inference overhead | <2x CPU time (acceptable for small batches) |
| Memory overhead | <100MB GPU memory |

### Code Quality

| Criterion | Requirement |
|-----------|-------------|
| Test coverage | >80% for new code |
| Error handling | All GPU errors caught with fallback |
| Resource cleanup | No GPU memory leaks |
| Documentation | All public APIs documented |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| GPU not available in CI | High | Low | Tests skip gracefully |
| Float32 precision loss | Medium | Low | 1e-4 tolerance; GPU uses float32 |
| Softmax reduction correctness | Medium | Medium | Extensive testing; use known-good CPU reference |
| Shader compilation issues | Low | High | Pre-compile shaders; fallback to CPU |
| Buffer size mismatches | Medium | Medium | Runtime dimension validation |
| Performance regression on small batches | High | Low | Only use GPU for batches >= threshold |

---

## Commit Strategy

### Commit 1: Add shared/gpu package with shaders
```
feat(gpu): add shared/gpu package for neural network GPU acceleration

- Add dense_mvm.comp shader for matrix-vector multiply
- Add relu.comp shader for ReLU activation
- Add softmax.comp shader with reduction
- Add params.go with std140-aligned structs
- Pre-compile shaders to SPIR-V
```

### Commit 2: Implement DenseLayerGPU and GPUNetwork
```
feat(gpu): implement DenseLayerGPU and GPUNetwork

- DenseLayerGPU handles single-layer GPU operations
- GPUNetwork provides high-level inference API
- Graceful fallback when Vulkan unavailable
- Add GPU parity unit tests
```

### Commit 3: Integrate GPU into Module 3 MNIST
```
feat(mnist): integrate GPU acceleration into DualModeNetwork

- Add network_gpu.go with GPU-specific inference methods
- Modify forwardFP to dispatch to GPU when available
- Add GPU accelerator initialization and cleanup
- CPU fallback on GPU errors
```

### Commit 4: Add comprehensive tests and benchmarks
```
test(mnist): add GPU inference tests and benchmarks

- TestMNISTInferenceParity: GPU vs CPU prediction parity
- TestBatchInferenceSpeedup: Performance benchmarks
- All tests skip gracefully when GPU unavailable
```

---

## Dependencies

### Existing Infrastructure (Reused)
- `shared/compute/context.go` - VulkanContext
- `shared/compute/compute_pipeline.go` - ComputePipeline
- `shared/compute/buffer.go` - GPUBuffer
- `shared/compute/shader_loader.go` - LoadSPIRV, CreateShaderModule
- `shared/compute/dispatcher.go` - CalculateDispatchSize

### External Dependencies
- `github.com/vulkan-go/vulkan` - Already in go.mod
- `glslangValidator` - For shader compilation (build-time only)

---

## Notes

1. **CIM Path GPU**: The CIM path could optionally use Module 2's crossbar GPU once the MVM/VMM issues are fixed. For now, CIM stays CPU-only since quantization effects dominate accuracy.

2. **Batch Size Threshold**: GPU overhead dominates for small batches. Only use GPU when batch size >= 8 or input dimension >= 128.

3. **Softmax Special Case**: Softmax requires reduction across all elements. The shader uses a simple sequential reduction since output dimension is small (10 classes).

4. **Memory Pre-allocation**: Pre-allocate GPU buffers for maximum expected dimensions (784 input, 128 hidden, 10 output) to avoid per-inference allocation.

5. **Error Recovery**: All GPU errors should fall back to CPU silently. Never fail inference due to GPU issues.
