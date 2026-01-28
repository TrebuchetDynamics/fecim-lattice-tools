# Plan: module2-crossbar GPU Flawless Integration (REVISED)

## Context

### Original Request
Make module2-crossbar work flawlessly with Vulkan GPU acceleration.

### Previous Plan Issue
The previous plan incorrectly identified "bugs" that do not exist in the current code. The actual code at:
- Line 218: `Cols: int32(a.config.Cols)` - already correct
- Line 242: `output := make([]float64, a.config.Cols)` - already correct

The REAL issue is a semantic mismatch between GPU (VMM) and CPU (MVM) operations.

### Actual Code Analysis (Verified)

**mvmCPU (array.go lines 437-468):**
```go
// Output has Rows elements
output := make([]float64, a.config.Rows)  // line 438

// Iterate over rows, sum over cols
for i := 0; i < a.config.Rows; i++ {      // line 444
    for j := 0; j < len(input); j++ {     // line 446 - input has Cols elements
        sum += g * vIn                     // Accumulate
    }
    output[i] = a.quantizeADC(normalizedSum)
}
```
Semantics: **MVM** - y = W*x where x has Cols elements, y has Rows elements.

**mvmGPU (array.go lines 201-250):**
```go
// Output has Cols elements
output := make([]float64, a.config.Cols)  // line 242

// Params passed to GPU
params := CrossbarParams{
    Rows: int32(a.config.Rows),           // line 219
    Cols: int32(a.config.Cols),           // line 220
    ...
}
```
Note: mvmGPU returns `a.config.Cols` elements.

**GPU Shader (mvm.comp lines 140-201):**
```glsl
// One thread per column output
uint colIdx = gl_GlobalInvocationID.x;    // line 141
if (colIdx >= cols) return;               // line 144

// Sum over all rows for this column
for (int rowIdx = 0; rowIdx < rows; rowIdx++) {  // line 156
    I_sum += G_varied * V_effective;      // Accumulate
}
I_out[colIdx] = I_quantized;              // line 201
```
Semantics: **VMM** - I_j = Sum_i(V_i * G_ij), voltage on rows, current from columns.

**Caller Expectations (module3-mnist/pkg/training/network.go):**
- Line 181: `hiddenRaw, _ := n.layer1.MVM(input)`
  - `input` is 784 elements (cols)
  - `layer1` is `hiddenSize x 784` (rows x cols)
  - Expected `hiddenRaw` length: `hiddenSize` (rows)
- This confirms callers expect **MVM semantics**

### The Actual Problem

| Component | Operation | Input Size | Output Size |
|-----------|-----------|------------|-------------|
| mvmCPU | MVM: y = W*x | Cols | Rows |
| mvmGPU | VMM: I_j = Sum(V_i * G_ij) | Rows (from shader) | Cols |
| Callers | Expect MVM | Cols | Rows |

**mvmGPU produces Cols output elements, but callers expect Rows elements.**

This is currently masked because:
1. Most tests use square matrices (Rows == Cols)
2. The GPU fallback to CPU works, so functionality isn't broken

### Decision: MVM Semantics Required

The callers (neural network layers) expect MVM semantics:
- Input: Cols elements (voltage on columns/bit lines)
- Output: Rows elements (current from rows/word lines)

The GPU shader correctly models VMM physics (voltage on rows, current from columns), which is valid crossbar physics. The integration layer must transpose the operation.

**Chosen Solution:** Modify `mvmGPU` to transpose the operation:
1. Swap rows/cols when setting params (shader "rows" = our cols)
2. Transpose the conductance matrix before GPU upload
3. Keep input size as `len(input)` = cols
4. Output size becomes `rows` elements

---

## Work Objectives

### Core Objective
GPU-accelerated MVM must produce results identical to CPU MVM within float32 tolerance.

### Deliverables
1. Fixed `mvmGPU` function that transposes operation for MVM semantics
2. GPU-specific tests comparing GPU vs CPU results
3. Verification that neural network inference works with GPU

### Definition of Done
- [ ] All existing tests pass (`go test ./module2-crossbar/...`)
- [ ] New GPU tests pass when Vulkan available, skip gracefully when not
- [ ] GPU MVM output matches CPU MVM output within 1e-4 tolerance
- [ ] No regressions in CPU-only path
- [ ] Neural network inference produces same results with GPU and CPU

---

## Guardrails

### Must Have
- GPU/CPU result parity (within float32 tolerance)
- Graceful fallback to CPU when GPU unavailable
- No changes to public API signatures
- Preserve all existing test coverage

### Must NOT Have
- Breaking changes to Config struct
- Changes to shader physics/algorithm (shader is correct)
- Removal of CPU fallback path
- New dependencies beyond existing Vulkan setup

---

## Task Flow

```
[Task 1: Fix mvmGPU to transpose operation]
         |
         v
[Task 2: Add GPU-vs-CPU parity test]
         |
         v
[Task 3: Verify existing tests pass]
         |
         v
[Task 4: Test neural network inference with GPU]
```

---

## Detailed TODOs

### Task 1: Fix mvmGPU to provide MVM semantics

**File:** `<local-path>`

**Current code (lines 201-250):**
```go
func (a *Array) mvmGPU(input []float64) ([]float64, error) {
    // Convert input to float32
    input32 := make([]float32, len(input))
    for i, v := range input {
        input32[i] = float32(a.quantizeDAC(v))
    }

    // Build conductance matrix from cells
    conductances := make([]float32, a.config.Rows*a.config.Cols)
    for i := 0; i < a.config.Rows; i++ {
        for j := 0; j < a.config.Cols; j++ {
            conductances[i*a.config.Cols+j] = float32(a.cells[i][j].Conductance)
        }
    }

    // Set up GPU parameters
    params := CrossbarParams{
        Rows:           int32(a.config.Rows),
        Cols:           int32(a.config.Cols),
        // ... other params
    }

    // Execute GPU MVM
    outputs32, err := a.gpuAccelerator.MVM(conductances, input32, params)
    // ...

    // Output vector has Cols elements (one per output column)
    output := make([]float64, a.config.Cols)
    // ...
}
```

**Required changes:**

1. **Transpose conductance matrix for GPU:** Build as column-major (G^T)
2. **Swap Rows/Cols in params:** shader "rows" = our cols, shader "cols" = our rows
3. **Adjust input buffer:** Input has `len(input)` = cols elements, which becomes shader's "rows"
4. **Output becomes Rows elements:** After transpose, output is our rows

**New code:**
```go
func (a *Array) mvmGPU(input []float64) ([]float64, error) {
    // For MVM: y = W*x where W is [Rows x Cols], x has Cols elements, y has Rows elements
    // GPU shader does VMM: I_j = Sum_i(V_i * G_ij), input size = shader rows, output size = shader cols
    //
    // To get MVM from VMM shader:
    // - Transpose W: shader sees W^T which is [Cols x Rows]
    // - Swap dimensions: shader rows = our cols, shader cols = our rows
    // - Input (our cols) goes to shader rows
    // - Output (shader cols) = our rows

    // Convert input to float32 (input has Cols elements)
    input32 := make([]float32, len(input))
    for i, v := range input {
        input32[i] = float32(a.quantizeDAC(v))
    }

    // Build TRANSPOSED conductance matrix: G^T[j][i] = G[i][j]
    // Layout: G^T stored as [Cols x Rows] row-major = G stored column-major
    conductancesT := make([]float32, a.config.Rows*a.config.Cols)
    for i := 0; i < a.config.Rows; i++ {
        for j := 0; j < a.config.Cols; j++ {
            // G^T[j][i] at index j*Rows + i
            conductancesT[j*a.config.Rows+i] = float32(a.cells[i][j].Conductance)
        }
    }

    // Set up GPU parameters with SWAPPED dimensions
    // Shader "rows" = our Cols (input size)
    // Shader "cols" = our Rows (output size)
    params := CrossbarParams{
        Rows:           int32(a.config.Cols),  // SWAPPED: shader rows = our cols
        Cols:           int32(a.config.Rows),  // SWAPPED: shader cols = our rows
        NoiseLevel:     float32(a.config.NoiseLevel),
        ADCBits:        int32(a.config.ADCBits),
        DACBits:        int32(a.config.DACBits),
        Time:           0.0,
        WireResistance: 0.0,
        DriftCoeff:     0.0,
    }

    // Execute GPU MVM (shader computes VMM on transposed matrix = MVM on original)
    outputs32, err := a.gpuAccelerator.MVM(conductancesT, input32, params)
    if err != nil {
        // Fall back to CPU on GPU error
        return a.mvmCPU(input)
    }

    // Find max possible current for normalization (same as CPU)
    maxCurrent := float64(len(input))

    // Convert back to float64 and apply normalization/quantization
    // Output vector now has Rows elements (one per output row)
    output := make([]float64, a.config.Rows)
    for i := 0; i < a.config.Rows; i++ {
        normalizedSum := float64(outputs32[i]) / maxCurrent
        output[i] = a.quantizeADC(normalizedSum)
        a.totalReads++
    }

    return output, nil
}
```

**Acceptance Criteria:**
- [ ] `mvmGPU` output length equals `a.config.Rows` (matches CPU)
- [ ] Output values match `mvmCPU` within 1e-4 tolerance
- [ ] Conductance matrix is correctly transposed
- [ ] Params have swapped Rows/Cols

---

### Task 2: Add GPU-specific tests

**File:** `<local-path>` (new file)

```go
package crossbar

import (
    "math"
    "testing"
)

// TestGPUMVMParityWithCPU verifies GPU MVM produces identical results to CPU MVM.
func TestGPUMVMParityWithCPU(t *testing.T) {
    // Use non-square matrix to catch row/col confusion
    cfg := &Config{
        Rows: 4, Cols: 8,  // Non-square: 4 outputs, 8 inputs
        NoiseLevel: 0,      // Zero noise for deterministic comparison
        ADCBits: 8, DACBits: 8,
    }

    arr, err := NewArray(cfg)
    if err != nil {
        t.Fatal(err)
    }
    defer arr.Destroy()

    // Check if GPU is available
    arr.initGPU()
    gpuAvailable := arr.gpuAccelerator != nil && arr.gpuAccelerator.IsAvailable()

    // Program known weights (pattern that reveals transpose errors)
    for i := 0; i < cfg.Rows; i++ {
        for j := 0; j < cfg.Cols; j++ {
            // Weight = row*0.1 + col*0.01 (unique per cell)
            w := float64(i)*0.1 + float64(j)*0.01
            arr.ProgramWeight(i, j, w)
        }
    }

    // Test input (8 elements for 8 columns)
    input := []float64{1.0, 0.5, 0.8, 0.3, 0.6, 0.9, 0.2, 0.7}

    // Get CPU result
    outputCPU, err := arr.mvmCPU(input)
    if err != nil {
        t.Fatalf("CPU MVM failed: %v", err)
    }

    // Verify CPU output has correct length
    if len(outputCPU) != cfg.Rows {
        t.Fatalf("CPU output length wrong: got %d, expected %d", len(outputCPU), cfg.Rows)
    }

    if !gpuAvailable {
        t.Skip("GPU not available, skipping GPU parity test")
    }

    // Get GPU result
    outputGPU, err := arr.mvmGPU(input)
    if err != nil {
        t.Fatalf("GPU MVM failed: %v", err)
    }

    // Verify GPU output has correct length
    if len(outputGPU) != cfg.Rows {
        t.Fatalf("GPU output length wrong: got %d, expected %d", len(outputGPU), cfg.Rows)
    }

    // Compare results
    tolerance := 1e-4  // GPU uses float32, allow some precision loss
    for i := range outputGPU {
        diff := math.Abs(outputGPU[i] - outputCPU[i])
        if diff > tolerance {
            t.Errorf("Output[%d] mismatch: GPU=%f, CPU=%f, diff=%f",
                i, outputGPU[i], outputCPU[i], diff)
        }
    }
}

// TestGPUMVMIdentityMatrix verifies GPU MVM with identity-like weights.
func TestGPUMVMIdentityMatrix(t *testing.T) {
    cfg := &Config{
        Rows: 4, Cols: 4,
        NoiseLevel: 0,
        ADCBits: 8, DACBits: 8,
    }

    arr, err := NewArray(cfg)
    if err != nil {
        t.Fatal(err)
    }
    defer arr.Destroy()

    arr.initGPU()
    if arr.gpuAccelerator == nil || !arr.gpuAccelerator.IsAvailable() {
        t.Skip("GPU not available")
    }

    // Program diagonal weights (identity-like)
    for i := 0; i < cfg.Rows; i++ {
        for j := 0; j < cfg.Cols; j++ {
            if i == j {
                arr.ProgramWeight(i, j, 1.0)
            } else {
                arr.ProgramWeight(i, j, 0.0)
            }
        }
    }

    input := []float64{1.0, 0.5, 0.25, 0.75}

    outputGPU, err := arr.mvmGPU(input)
    if err != nil {
        t.Fatalf("GPU MVM failed: %v", err)
    }

    outputCPU, err := arr.mvmCPU(input)
    if err != nil {
        t.Fatalf("CPU MVM failed: %v", err)
    }

    // With identity matrix, output should approximate input (after quantization)
    tolerance := 1e-4
    for i := range outputGPU {
        diff := math.Abs(outputGPU[i] - outputCPU[i])
        if diff > tolerance {
            t.Errorf("Identity test Output[%d]: GPU=%f, CPU=%f, diff=%f",
                i, outputGPU[i], outputCPU[i], diff)
        }
    }
}

// TestGPUFallbackToCPU verifies graceful fallback when GPU unavailable.
func TestGPUFallbackToCPU(t *testing.T) {
    cfg := &Config{
        Rows: 4, Cols: 4,
        NoiseLevel: 0,
        ADCBits: 8, DACBits: 8,
    }

    arr, err := NewArray(cfg)
    if err != nil {
        t.Fatal(err)
    }
    defer arr.Destroy()

    // Program some weights
    arr.ProgramWeight(0, 0, 1.0)
    arr.ProgramWeight(1, 1, 1.0)

    // MVM should work regardless of GPU availability
    input := []float64{1.0, 0.5, 0.3, 0.8}
    output, err := arr.MVM(input)
    if err != nil {
        t.Fatalf("MVM failed: %v", err)
    }

    if len(output) != cfg.Rows {
        t.Errorf("Output length wrong: got %d, expected %d", len(output), cfg.Rows)
    }
}
```

**Acceptance Criteria:**
- [ ] Test file compiles without errors
- [ ] TestGPUMVMParityWithCPU passes when GPU available (or skips gracefully)
- [ ] TestGPUMVMIdentityMatrix passes when GPU available (or skips gracefully)
- [ ] TestGPUFallbackToCPU always passes

---

### Task 3: Verify all existing tests pass

**Command:**
```bash
go test ./module2-crossbar/pkg/crossbar/... -v
```

**Acceptance Criteria:**
- [ ] All existing tests pass
- [ ] No regressions introduced

---

### Task 4: Test neural network inference with GPU

**File:** `<local-path>` (add to existing)

```go
// TestGPUNeuralNetworkLayer verifies GPU works for typical neural network use.
func TestGPUNeuralNetworkLayer(t *testing.T) {
    // Simulate MNIST first layer: 784 -> 128
    cfg := &Config{
        Rows: 128, Cols: 784,  // 128 outputs (hidden), 784 inputs (pixels)
        NoiseLevel: 0,
        ADCBits: 8, DACBits: 8,
    }

    arr, err := NewArray(cfg)
    if err != nil {
        t.Fatal(err)
    }
    defer arr.Destroy()

    arr.initGPU()
    gpuAvailable := arr.gpuAccelerator != nil && arr.gpuAccelerator.IsAvailable()

    // Program random-ish weights
    for i := 0; i < cfg.Rows; i++ {
        for j := 0; j < cfg.Cols; j++ {
            w := float64((i*cfg.Cols+j)%30) / 29.0  // Deterministic "random"
            arr.ProgramWeight(i, j, w)
        }
    }

    // Simulate 784-pixel input
    input := make([]float64, 784)
    for i := range input {
        input[i] = float64(i%10) / 10.0
    }

    // Get CPU result
    outputCPU, err := arr.mvmCPU(input)
    if err != nil {
        t.Fatalf("CPU MVM failed: %v", err)
    }

    if len(outputCPU) != 128 {
        t.Fatalf("CPU output wrong size: got %d, expected 128", len(outputCPU))
    }

    if !gpuAvailable {
        t.Skip("GPU not available")
    }

    // Get GPU result
    outputGPU, err := arr.mvmGPU(input)
    if err != nil {
        t.Fatalf("GPU MVM failed: %v", err)
    }

    if len(outputGPU) != 128 {
        t.Fatalf("GPU output wrong size: got %d, expected 128", len(outputGPU))
    }

    // Compare
    tolerance := 1e-4
    maxDiff := 0.0
    for i := range outputGPU {
        diff := math.Abs(outputGPU[i] - outputCPU[i])
        if diff > maxDiff {
            maxDiff = diff
        }
        if diff > tolerance {
            t.Errorf("Layer output[%d]: GPU=%f, CPU=%f, diff=%f", i, outputGPU[i], outputCPU[i], diff)
        }
    }
    t.Logf("Max difference: %e", maxDiff)
}
```

**Acceptance Criteria:**
- [ ] Test passes when GPU available
- [ ] Test skips gracefully when GPU unavailable
- [ ] Output length is 128 (Rows, not 784 Cols)

---

## Commit Strategy

### Commit 1: Fix mvmGPU transpose for MVM semantics
```
fix(crossbar): transpose GPU operation to provide MVM semantics

The GPU shader implements VMM physics (voltage on rows, current from cols),
but the MVM API expects MVM semantics (voltage on cols, current from rows).

Fix mvmGPU to:
- Transpose conductance matrix before GPU upload (G -> G^T)
- Swap Rows/Cols in shader parameters
- Output Rows elements instead of Cols elements

This makes GPU MVM output identical to CPU MVM output.
```

### Commit 2: Add GPU parity tests
```
test(crossbar): add GPU vs CPU MVM parity tests

- TestGPUMVMParityWithCPU: non-square matrix comparison
- TestGPUMVMIdentityMatrix: identity matrix verification
- TestGPUFallbackToCPU: graceful degradation
- TestGPUNeuralNetworkLayer: MNIST-scale layer test

Tests skip gracefully when Vulkan unavailable.
```

---

## Success Criteria

| Criterion | Measurement |
|-----------|-------------|
| All existing tests pass | `go test ./module2-crossbar/...` exits 0 |
| GPU/CPU parity | New parity tests pass within 1e-4 tolerance |
| Correct output dimensions | GPU output has Rows elements (not Cols) |
| Neural network compatible | 784->128 layer produces 128 outputs |
| Graceful fallback | Fallback test passes on systems without GPU |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Transpose performance overhead | Low | Low | Matrix transpose is O(n*m), negligible vs GPU transfer |
| Float precision issues | Medium | Low | Use 1e-4 tolerance; GPU uses float32 vs CPU float64 |
| Test environment lacks GPU | High | Low | Tests skip gracefully; CI may be CPU-only |

---

## Notes

- The shader is CORRECT for VMM physics - no shader changes needed
- The fix is entirely in the Go integration layer (mvmGPU function)
- Non-square matrices are critical for catching row/col confusion
- The 1e-4 tolerance accounts for float32 (GPU) vs float64 (CPU) precision difference
