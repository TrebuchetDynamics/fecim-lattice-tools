# Work Plan: WRD Hit Targets Fix

## Context

### Original Request
Fix Write/Read Demo to hit targets: Currently the demo changes targets even when it fails to hit the current target. The demo MUST keep retrying the SAME target until it successfully hits it (within tolerance), not give up after max retries and move to a new target.

### Problem Analysis
In `module1-hysteresis/pkg/gui/simulation.go`, the WRD (Write/Read Demo) phase 4 (READ) has this logic structure:

```go
if !success && a.wrdRetryCount < a.wrdMaxRetries {
    // RETRY logic - loops back to phase 0
} else {
    // Either success OR max retries reached - proceed to DISPLAY (phase 5)
    // This picks a new target even if we failed!
}
```

The problem: When `wrdRetryCount >= wrdMaxRetries`, the code proceeds to DISPLAY phase and picks a new target, even though the current target was never successfully hit.

### Current Behavior (Problematic)
1. Attempt to write target level N
2. Read back level - not matching
3. Retry up to 3 times (wrdMaxRetries=3)
4. After 3 retries, **give up and move to new target**
5. Success rate drops because failures are counted but never corrected

### Required Behavior
1. Attempt to write target level N
2. Read back level - not matching
3. Update calibration based on error
4. Retry (no limit) until success
5. Only after success, move to new target
6. 100% target hit rate guaranteed

---

## Work Objectives

### Core Objective
Modify WRD phase 4 (READ) logic to retry indefinitely until target is successfully hit.

### Deliverables
1. Modified phase 4 logic in `simulation.go`
2. Remove wrdMaxRetries field usage (or set to infinite)
3. Updated logging to reflect unlimited retry behavior

### Definition of Done
- WRD demo never gives up on a target
- Every target is eventually hit (within ±1 level tolerance)
- Success rate approaches 100% over time
- Calibration converges to correct values through repeated attempts

---

## Guardrails

### MUST Have
- Keep retrying same target until success (±1 level tolerance)
- Update calibration on each failed attempt
- Only pick new target after successful hit
- Maintain existing phase structure (0-5)
- Preserve energy tracking and logging

### MUST NOT Have
- Max retry limit that causes target abandonment
- Moving to new target on failure
- Counting failed targets as "completed"
- Infinite loops without calibration updates (must converge)

---

## Task Flow

```
Task 1: Modify Phase 4 Logic
    |
    v
Task 2: Update/Remove wrdMaxRetries Usage
    |
    v
Task 3: Update Logging Messages
    |
    v
Task 4: Verify Behavior
```

---

## Detailed TODOs

### Task 1: Modify Phase 4 Logic
**File:** `module1-hysteresis/pkg/gui/simulation.go`
**Location:** Lines 1034-1146 (case 4 in WRD phase switch)

**Change Required:**
Replace the current conditional structure:
```go
// CURRENT (problematic)
if !success && a.wrdRetryCount < a.wrdMaxRetries {
    // retry
} else {
    // proceed (even on failure!)
}
```

With:
```go
// NEW (correct)
if success {
    // Success! Proceed to DISPLAY phase, pick new target
    // Reset retry count
    // Update metrics
} else {
    // Failed - update calibration and retry
    // NO max retry limit
    // Loop back to phase 0
}
```

**Specific Code Change:**

In lines 1034-1056, change from:
```go
if !success && a.wrdRetryCount < a.wrdMaxRetries {
    a.wrdRetryCount++
    log.Printf("WRD VERIFY FAIL: L_read=%d L_target=%d err=%+d | RETRY %d/%d",
        a.wrdReadLevel, a.wrdTargetLevel, levelError, a.wrdRetryCount, a.wrdMaxRetries)
    // ... calibration update ...
    // Loop back to RESET with SAME target
    a.wrdResetStartP = a.polarization * 100
    a.wrdPhase = 0
    a.wrdPhaseTimer = 0
} else {
    // Either success OR max retries reached - proceed to DISPLAY
```

To:
```go
if success {
    // SUCCESS: proceed to DISPLAY phase
    a.wrdPhase = 5
    a.wrdPhaseTimer = 0

    // Track metrics
    a.wrdTotalWrites++
    a.wrdSuccessWrites++
    successRate := float64(a.wrdSuccessWrites) / float64(a.wrdTotalWrites) * 100

    if a.wrdRetryCount > 0 {
        log.Printf("WRD PHASE 4→5: VERIFY OK after %d retries | L_read=%d L_target=%d | rate=%.1f%%",
            a.wrdRetryCount, a.wrdReadLevel, a.wrdTargetLevel, successRate)
    } else {
        log.Printf("WRD PHASE 4→5: VERIFY OK (1st try) | L_read=%d L_target=%d | rate=%.1f%%",
            a.wrdReadLevel, a.wrdTargetLevel, successRate)
    }

    // Reset retry count for next target
    a.wrdRetryCount = 0

    // ... rest of success path (energy tracking, debug log, etc.)
} else {
    // FAILED: Update calibration and RETRY (no limit)
    a.wrdRetryCount++
    log.Printf("WRD VERIFY FAIL: L_read=%d L_target=%d err=%+d | RETRY #%d",
        a.wrdReadLevel, a.wrdTargetLevel, levelError, a.wrdRetryCount)

    // Update calibration BEFORE retry
    targetIdx := a.wrdTargetLevel - 1
    if a.calibrated && targetIdx >= 0 && targetIdx < len(a.calibrationUp) {
        midLevel := a.numLevels / 2
        goingUp := a.wrdTargetLevel > midLevel
        if goingUp {
            a.updateCalibrationUp(targetIdx, levelError, Ec)
        } else {
            a.updateCalibrationDown(targetIdx, levelError, Ec)
        }
    }

    // Loop back to RESET with SAME target
    a.wrdResetStartP = a.polarization * 100
    a.wrdPhase = 0
    a.wrdPhaseTimer = 0
}
```

**Acceptance Criteria:**
- [ ] Success branch only triggers when `success == true`
- [ ] Failure branch always retries (no max limit check)
- [ ] Calibration updated on every failure
- [ ] Retry count incremented and logged
- [ ] Phase 0 loop-back happens on every failure

### Task 2: Update wrdMaxRetries Usage
**File:** `module1-hysteresis/pkg/gui/gui.go`
**Location:** Lines 88-89, 349, 415

**Option A (Recommended):** Remove wrdMaxRetries field entirely
- Delete field declaration (line 89)
- Remove initialization (lines 349, 415)
- Remove from log messages

**Option B:** Keep field but don't use it as a limit
- Keep for potential future use (configurable retry limit)
- Just remove the `< a.wrdMaxRetries` check

**Acceptance Criteria:**
- [ ] No code path checks wrdMaxRetries as a limit
- [ ] Code compiles without errors

### Task 3: Update Logging Messages
**File:** `module1-hysteresis/pkg/gui/simulation.go`

**Changes:**
1. Remove "RETRY x/y" format (no max to compare against)
2. Use "RETRY #x" format instead
3. Remove "FAIL" outcome logging from success path (since we no longer have failed completions)

**Acceptance Criteria:**
- [ ] Log messages reflect unlimited retry behavior
- [ ] No references to max retries in logs

### Task 4: Verify Behavior
**Manual verification steps:**
1. Run the application: `go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools`
2. Switch to "Write/Read Demo" mode
3. Observe that:
   - Demo keeps retrying same target until hit
   - Retry count accumulates in logs
   - Success rate stays at 100%
   - Eventually all targets are hit

**Acceptance Criteria:**
- [ ] WRD demo achieves 100% success rate
- [ ] No targets are abandoned
- [ ] Calibration converges (retry count decreases over time)

---

## Commit Strategy

### Single Commit
```
fix(hysteresis): WRD demo retries indefinitely until target hit

Previously the Write/Read Demo would give up after 3 retries and move
to a new target, even if the current target was never successfully hit.
This caused the success rate to drop below 100%.

Changes:
- Remove max retry limit in phase 4 (READ) logic
- Restructure conditional: success path vs failure/retry path
- Update calibration on every failure to converge on correct field
- Only pick new target after successful hit

This ensures 100% target hit rate by never abandoning a target.
```

---

## Success Criteria

1. **Functional:** WRD demo never abandons a target
2. **Convergent:** Retry count decreases over time as calibration improves
3. **Metrics:** Success rate is 100% (all attempted targets eventually hit)
4. **Observable:** Logs show retry attempts and eventual success

---

## Risk Assessment

### Low Risk
- Logic change is localized to one case statement
- Existing calibration mechanism ensures convergence
- No structural changes to phase machine

### Mitigation
- If calibration doesn't converge (infinite retries), the existing binary search bounds will eventually find the correct value
- Worst case: user can switch modes to escape if truly stuck (unlikely given physics model)
