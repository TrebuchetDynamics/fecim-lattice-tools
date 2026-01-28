# MNIST Module E2E and Unit Test Plan

**Generated:** 2026-01-28
**Module:** `module3-mnist/`
**Status:** Ready for Implementation

---

## 1. Overview

### Goals
Create comprehensive test coverage for Module 3 MNIST to verify:
1. **Math correctness** - All inference paths produce mathematically valid results
2. **Regression prevention** - Previously fixed bugs (CRIT-001, CRIT-002, HIGH-003) remain fixed
3. **Weight management** - Loading, saving, and accessor functions work correctly
4. **Edge case handling** - Empty slices, invalid inputs, boundary conditions

### Approach
- **Unit tests** for individual functions with clear inputs/outputs
- **Integration tests** for end-to-end workflows
- **Regression tests** for previously fixed bugs
- **Property-based tests** for mathematical invariants

### Current Test Coverage
| File | Lines | Coverage Focus |
|------|-------|----------------|
| `quantize_test.go` | ~250 | Weight quantization math |
| `physics_test.go` | ~700 | Physics accuracy, energy, dual-mode |
| `integration_test.go` | ~460 | E2E workflows, concurrency |
| `helpers_test.go` (gui) | ~200 | GUI helper functions |
| `network_test.go` (training) | ~420 | Training network |
| `loader_test.go` (mnist) | ~varies | MNIST data loading |

---

## 2. Test Matrix by Priority

### P0 CRITICAL - Core Math Functions (NEW TESTS REQUIRED)

| Function | File | Current Test | Gap |
|----------|------|--------------|-----|
| `InferFPOnly()` | network_inference.go:151-162 | None | FP-only path not tested |
| `InferCIMOnly()` | network_inference.go:166-188 | None | CIM-only path not tested |
| `quantizeDAC()` | network_inference.go:294-314 | None | DAC quantization logic |
| `quantizeADC()` | network_inference.go:317-359 | None | ADC quantization logic |
| `softmax()` empty slice | network_inference.go:237-261 | Basic cases tested, no empty slice edge case | CRIT-001 regression |
| `argmax()` empty slice | network_inference.go:265-279 | Basic cases tested, no empty slice edge case | Regression test |
| `Infer()` input validation | network_inference.go:14-16 | None | MED-008 returns nil |

### P1 HIGH - Weight Management Functions (NEW TESTS REQUIRED)

| Function | File | Current Test | Gap |
|----------|------|--------------|-----|
| `LoadWeights()` | network.go:208-328 | None | JSON parsing, errors |
| `LoadWeightsForLevel()` | network.go:333-345 | None | QAT level matching |
| `GetBestMatchingWeightsLevel()` | network.go:188-202 | None | Nearest level selection |
| `GetFPWeights()` | network_quantization.go:85-94 | None | Getter correctness |
| `GetQuantWeights()` | network_quantization.go:97-106 | None | Getter correctness |
| `GetWeightsFilename()` | network.go:176-185 | None | Path generation |

### P2 MEDIUM - Integration Tests (EXPAND EXISTING)

| Test Scenario | Current Coverage | Gap |
|---------------|------------------|-----|
| Full MNIST evaluation | Basic structure | Accuracy benchmarks |
| Agreement rate vs noise | Partial | Statistical significance |
| Energy calculation | Basic | Per-layer PTQ energy |
| SingleLayer mode inference | None | Tour mode path |

### P3 LOW - GUI Tests (PARTIALLY DONE)

| Test | Status | Notes |
|------|--------|-------|
| Helper functions | Done | formatEnergy, weightToColor, clamp |
| Widget state | Missing | Requires Fyne test framework |

---

## 3. New Test Files to Create

### File 1: `module3-mnist/pkg/core/network_inference_test.go` (NEW)

**Purpose:** Test all inference path functions and their edge cases.

```
Test Functions:
- TestInferFPOnly_BasicInference
- TestInferFPOnly_OutputValidation
- TestInferFPOnly_ConsistencyWithInfer
- TestInferCIMOnly_BasicInference
- TestInferCIMOnly_UsesQuantizedWeights
- TestInferCIMOnly_OutputValidation
- TestInferCIMOnly_DACDACEffect
- TestQuantizeDAC_BitLevels
- TestQuantizeDAC_InputClamping
- TestQuantizeDAC_16BitPassthrough
- TestQuantizeADC_BitLevels
- TestQuantizeADC_EmptySlice (CRIT-002 regression)
- TestQuantizeADC_16BitPassthrough
- TestQuantizeADC_RangeNormalization
- TestSoftmax_EmptySlice (CRIT-001 regression)
- TestArgmax_EmptySlice
- TestInfer_InputLengthValidation (MED-008 regression)
- TestInfer_SingleLayerMode
- TestInferFPOnly_SingleLayerNotSupported
```

### File 2: `module3-mnist/pkg/core/network_weights_test.go` (NEW)

**Purpose:** Test weight loading, saving, and accessor functions.

```
Test Functions:
- TestLoadWeights_ValidJSON
- TestLoadWeights_InvalidJSON
- TestLoadWeights_MissingFile
- TestLoadWeights_WithScaleOffset
- TestLoadWeights_WithoutScaleOffset
- TestLoadWeights_SingleLayerWeights
- TestLoadWeights_PerLayerQuantLevels
- TestLoadWeights_LegacyQuantLevels
- TestLoadWeightsForLevel_ExactMatch
- TestLoadWeightsForLevel_FallbackToNearest
- TestLoadWeightsForLevel_DefaultFallback
- TestGetBestMatchingWeightsLevel_ExactMatch
- TestGetBestMatchingWeightsLevel_NearestLevel
- TestGetBestMatchingWeightsLevel_AllCases
- TestGetWeightsFilename_AvailableLevels
- TestGetWeightsFilename_Default30
- TestGetFPWeights_ReturnsCopy
- TestGetFPWeights_NotAffectedByModification
- TestGetQuantWeights_ReturnsCopy
- TestGetQuantWeights_MatchesRequantized
```

### File 3: Expand `module3-mnist/pkg/core/integration_test.go`

**Add test functions:**
```
- TestSingleLayerModeE2E
- TestPerLayerQuantizationE2E
- TestAgreementRateStatistics (N runs, confidence interval)
- TestEnergyCalculationVerification
```

---

## 4. Detailed Test Implementations

### 4.1 P0 Tests - Inference Functions

#### TestInferFPOnly_BasicInference
```go
// Verify FP-only inference produces valid predictions
// Setup: Network with known weights
// Input: Valid 784-element input
// Assert: prediction in [0,9], confidence in [0,1], probs sum to 1
```

#### TestInferFPOnly_InputLengthMismatch
```go
// Verify behavior with wrong input length
// Setup: Network initialized
// Input: 100-element slice (wrong size, should be 784)
// Assert: Returns valid result (relies on forwardFP's implicit handling)
// Note: These functions don't validate input - they rely on forwardFP
```

#### TestInferCIMOnly_UsesQuantizedWeights
```go
// Verify CIM path uses QuantWeights not FPWeights (HIGH-003 regression)
// Setup: Network with FPWeights != QuantWeights
// Input: Valid input
// Assert: Output matches expected from quantized weights
```

#### TestInferCIMOnly_InputLengthMismatch
```go
// Verify behavior with wrong input length
// Setup: Network initialized
// Input: 100-element slice (wrong size, should be 784)
// Assert: Returns valid result (relies on forwardFP's implicit handling)
// Note: These functions don't validate input - they rely on forwardFP
```

#### TestQuantizeDAC_BitLevels
```go
// Verify DAC quantization produces correct number of levels
// Test cases: 3-bit (8 levels), 8-bit (256 levels), 16-bit (passthrough)
// Assert: Output values match expected quantization grid
```

#### TestQuantizeADC_EmptySlice
```go
// CRIT-002 regression test
// Input: Empty slice
// Assert: Returns empty slice without panic
```

#### TestSoftmax_EmptySlice
```go
// CRIT-001 regression test
// Input: Empty slice
// Assert: Returns nil without panic
```

#### TestInfer_InputLengthValidation
```go
// MED-008 regression test
// Input: Wrong length slice (not 784)
// Assert: Returns nil, no panic
```

#### TestInferFPOnly_SingleLayerNotSupported
```go
// Verify InferFPOnly behavior with SingleLayer mode
// Design decision: InferFPOnly is a fast path that doesn't support SingleLayer mode
// Setup: Network with SingleLayer=true
// Input: Valid 784-element input
// Assert: Function still produces valid output using two-layer weights (ignores SingleLayer config)
// Note: This is expected behavior - fast path always uses two-layer architecture
```

### 4.2 P1 Tests - Weight Management

#### TestLoadWeights_ValidJSON
```go
// Test loading well-formed weights file
// Setup: Create temp JSON with valid weights
// Assert: Weights loaded correctly, dimensions match
```

#### TestLoadWeights_InvalidJSON
```go
// Test error handling for malformed JSON
// Input: Invalid JSON content
// Assert: Returns error, network not corrupted
```

#### TestGetBestMatchingWeightsLevel_AllCases
```go
// AvailableQATLevels = []int{10, 20, 29, 30, 31}
// Test cases:
//   - 10 -> 10 (exact)
//   - 15 -> 10 or 20 (nearest)
//   - 30 -> 30 (exact, default)
//   - 5 -> 10 (closest)
//   - 35 -> 31 (closest)
```

#### TestGetFPWeights_ReturnsCopy
```go
// Verify getter returns a copy, not original reference
// Get weights, modify returned slice
// Assert: Original network weights unchanged
```

### 4.3 P2 Tests - Integration

#### TestSingleLayerModeE2E
```go
// Test Calibration Mode (784->10 single layer)
// Setup: Enable SingleLayer mode
// Assert: Inference works, energy calculation correct
```

#### TestAgreementRateStatistics
```go
// Statistical test of FP/CIM agreement
// Run 100 inferences with noise
// Assert: Agreement rate within expected bounds for noise level
```

---

## 5. Implementation Tasks

### Task 1: Create `network_inference_test.go`
**File:** `module3-mnist/pkg/core/network_inference_test.go`
**Lines:** ~430
**Tests:** 20 test functions

| # | Test Function | Priority | Est. Lines |
|---|---------------|----------|------------|
| 1 | TestInferFPOnly_BasicInference | P0 | 30 |
| 2 | TestInferFPOnly_OutputValidation | P0 | 25 |
| 3 | TestInferFPOnly_ConsistencyWithInfer | P0 | 35 |
| 4 | TestInferFPOnly_InputLengthMismatch | P0 | 15 |
| 5 | TestInferCIMOnly_BasicInference | P0 | 30 |
| 6 | TestInferCIMOnly_UsesQuantizedWeights | P0 | 40 |
| 7 | TestInferCIMOnly_OutputValidation | P0 | 25 |
| 8 | TestInferCIMOnly_DACDACEffect | P0 | 35 |
| 9 | TestInferCIMOnly_InputLengthMismatch | P0 | 15 |
| 10 | TestQuantizeDAC_BitLevels | P0 | 40 |
| 11 | TestQuantizeDAC_InputClamping | P0 | 20 |
| 12 | TestQuantizeDAC_16BitPassthrough | P0 | 15 |
| 13 | TestQuantizeADC_BitLevels | P0 | 40 |
| 14 | TestQuantizeADC_EmptySlice | P0 | 10 |
| 15 | TestQuantizeADC_16BitPassthrough | P0 | 15 |
| 16 | TestQuantizeADC_RangeNormalization | P0 | 25 |
| 17 | TestSoftmax_EmptySlice | P0 | 10 |
| 18 | TestArgmax_EmptySlice | P0 | 10 |
| 19 | TestInfer_InputLengthValidation | P0 | 15 |
| 20 | TestInfer_SingleLayerMode | P0 | 30 |

### Task 2: Create `network_weights_test.go`
**File:** `module3-mnist/pkg/core/network_weights_test.go`
**Lines:** ~500
**Tests:** 20 test functions

| # | Test Function | Priority | Est. Lines |
|---|---------------|----------|------------|
| 1 | TestLoadWeights_ValidJSON | P1 | 45 |
| 2 | TestLoadWeights_InvalidJSON | P1 | 20 |
| 3 | TestLoadWeights_MissingFile | P1 | 15 |
| 4 | TestLoadWeights_WithScaleOffset | P1 | 35 |
| 5 | TestLoadWeights_WithoutScaleOffset | P1 | 30 |
| 6 | TestLoadWeights_SingleLayerWeights | P1 | 40 |
| 7 | TestLoadWeights_PerLayerQuantLevels | P1 | 30 |
| 8 | TestLoadWeights_LegacyQuantLevels | P1 | 25 |
| 9 | TestLoadWeightsForLevel_ExactMatch | P1 | 25 |
| 10 | TestLoadWeightsForLevel_FallbackToNearest | P1 | 30 |
| 11 | TestLoadWeightsForLevel_DefaultFallback | P1 | 20 |
| 12 | TestGetBestMatchingWeightsLevel_ExactMatch | P1 | 20 |
| 13 | TestGetBestMatchingWeightsLevel_NearestLevel | P1 | 25 |
| 14 | TestGetBestMatchingWeightsLevel_AllCases | P1 | 35 |
| 15 | TestGetWeightsFilename_AvailableLevels | P1 | 25 |
| 16 | TestGetWeightsFilename_Default30 | P1 | 15 |
| 17 | TestGetFPWeights_ReturnsCopy | P1 | 25 |
| 18 | TestGetFPWeights_NotAffectedByModification | P1 | 20 |
| 19 | TestGetQuantWeights_ReturnsCopy | P1 | 25 |
| 20 | TestGetQuantWeights_MatchesRequantized | P1 | 30 |

### Task 3: Expand `integration_test.go`
**File:** `module3-mnist/pkg/core/integration_test.go`
**Add:** ~150 lines
**Tests:** 4 new test functions

| # | Test Function | Priority | Est. Lines |
|---|---------------|----------|------------|
| 1 | TestSingleLayerModeE2E | P2 | 40 |
| 2 | TestPerLayerQuantizationE2E | P2 | 35 |
| 3 | TestAgreementRateStatistics | P2 | 45 |
| 4 | TestEnergyCalculationVerification | P2 | 30 |

### Task 4: Run and Verify
**Commands:**
```bash
go test ./module3-mnist/pkg/core/... -v
go test ./module3-mnist/pkg/core/... -race
go test ./module3-mnist/pkg/core/... -cover
```

---

## 6. Acceptance Criteria

### Functional Criteria
- [ ] All 44 new tests pass
- [ ] No regressions in existing 117 tests
- [ ] Race detector passes: `go test -race`
- [ ] Code coverage for `pkg/core/` increases by 15%+

### Specific Verification
- [ ] CRIT-001 (softmax empty): Test confirms `softmax(nil)` returns `nil`
- [ ] CRIT-002 (quantizeADC empty): Test confirms `quantizeADC(nil, 8)` returns empty slice
- [ ] HIGH-003 (CIM uses quantized): Test verifies `InferCIMOnly` uses `QuantWeights`
- [ ] MED-008 (input validation): Test confirms `Infer(wrongLength)` returns `nil`

### Quality Criteria
- [ ] Each test has clear name describing what it tests
- [ ] Each test has assertion with meaningful error message
- [ ] No test dependencies on external files (use synthetic data)
- [ ] Tests complete in < 30 seconds total

---

## 7. File Structure After Implementation

```
module3-mnist/pkg/core/
  network.go
  network_config.go
  network_inference.go
  network_quantization.go
  quantize.go
  quantize_test.go          (existing, ~250 lines)
  physics_test.go           (existing, ~700 lines)
  integration_test.go       (expanded, +150 lines)
  network_inference_test.go (NEW, ~400 lines)
  network_weights_test.go   (NEW, ~500 lines)
```

---

## 8. Commit Strategy

### Commit 1: Create P0 inference tests
```
test(mnist): add P0 inference function tests

- Add network_inference_test.go with 18 tests
- Cover InferFPOnly, InferCIMOnly, quantizeDAC, quantizeADC
- Add regression tests for CRIT-001, CRIT-002, MED-008
```

### Commit 2: Create P1 weight management tests
```
test(mnist): add P1 weight management tests

- Add network_weights_test.go with 20 tests
- Cover LoadWeights, GetBestMatchingWeightsLevel
- Cover GetFPWeights, GetQuantWeights getters
```

### Commit 3: Expand integration tests
```
test(mnist): expand integration tests with P2 scenarios

- Add SingleLayerMode E2E test
- Add PerLayerQuantization E2E test
- Add statistical agreement rate test
- Add energy calculation verification
```

---

## 9. Success Metrics

| Metric | Before | Target |
|--------|--------|--------|
| Test count (core) | 37 | 81 |
| Code coverage | ~60% | ~80% |
| P0 functions tested | 0/7 | 7/7 |
| P1 functions tested | 0/6 | 6/6 |
| Regression tests | 0 | 4 |

---

## 10. References

- `docs/neural-network/mnist.fixes.todo.md` - Bug tracking
- `docs/neural-network/mnist.architecture.md` - Module architecture
- `module3-mnist/pkg/core/*.go` - Source files under test
