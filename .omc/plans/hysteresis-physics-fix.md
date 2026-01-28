# Hysteresis Physics Bulletproofing Plan (v3)

## Problem Reanalysis

**Critic feedback incorporated from v2**: The grid size is 50 (not 40 as previously claimed).

### Actual Error Analysis from Logs

1. **Level calibration converges to wrong values**:
   ```
   ANIMATION COMPLETE: target=20, final=29, error=9
   CALIB DOWN[19]: bounds=[-0.5000,-0.3824]*Ec, new=-0.5000*Ec, err=9
   ```
   The calibration hit its bounds and cannot converge further.

2. **25/30 levels share E-field values**:
   ```
   WARNING: Calibration quality issue - 25/30 ascending and 25/30 descending levels share E-field values
   ```
   The Preisach model produces staircase regions where multiple levels require the same E-field.

3. **Direction convention inconsistency**:
   - WriteRead mode (line 811): `goingUp := targetLevel > midLevel`
   - Manual mode (line 554): `goingUp := targetLevel > startLevel`

### Root Causes Identified

1. **Calibration direction inconsistency**:
   - Write/Read Demo: `goingUp := targetLevel > midLevel`
   - Manual Mode: `goingUp := targetLevel > startLevel`
   - Calibration trained using WriteRead convention fails in Manual mode

2. **Hysteron distribution is too narrow** (CONFIRMED):
   - `AlphaSigma = 0.2*Ec` and `BetaSigma = 0.2*Ec` (preisach_advanced.go:59-61)
   - Creates steep S-curve with flat saturation regions
   - Literature suggests σ ≈ 0.25-0.35 × Ec

3. **Hysteron count may still be low for smooth 30-level gradation**:
   - Current: 50×50 grid = ~1250 valid hysterons (lower triangle)
   - For 30 levels, ~42 hysterons per level
   - Consider 60×60 = ~1800 valid hysterons (60 per level)

4. **Calibration search bounds too narrow**:
   - Upper bound: `Emax = 1.5 * Ec`
   - Lower bound: `0.5 * Ec`
   - Some levels may require fields outside this range

## Acceptance Criteria

1. **Level calibration accuracy**: ±2 level error max for all 30 levels
2. **Calibration quality**: No more than 10 duplicate E-field values (down from 25)
3. **Mode consistency**: Same calibration works for both WriteRead and Manual modes
4. **Physics tests pass**: All tests in `preisach_physics_test.go` continue to pass
5. **P-E curve smoothness**: At least 20 distinct P values between ±Pr

## Implementation Steps

### Step 1: Fix Direction Inconsistency in Manual Mode

**File:** `module1-hysteresis/pkg/gui/simulation.go`

**Location 1: Lines 553-555** (Manual mode WRITE phase decision)

Current:
```go
targetIdx := targetLevel - 1
goingUp := targetLevel > startLevel
```

Change to:
```go
targetIdx := targetLevel - 1
midLevel := a.numLevels / 2  // Add this line - midLevel not in scope here
goingUp := targetLevel > midLevel  // Match WriteRead convention
```

**Location 2: Lines 563-577** (fallback path - same block)

The `goingUp` variable is already changed above, no additional change needed.

**Location 3: Lines 618-665** (Manual mode calibration adjustment)

Current (line ~619):
```go
if targetLevel > startLevel {
    // ASCENDING calibration adjustment
```

Change to:
```go
midLevel := a.numLevels / 2
if targetLevel > midLevel {
    // ASCENDING calibration adjustment (match WriteRead)
```

### Step 2: Increase Hysteron Grid Size

**File:** `module1-hysteresis/pkg/gui/gui.go`

**Line 316** change from:
```go
preisachGridSize := 50                                 // High-resolution physics simulation
```
To:
```go
preisachGridSize := 60                                 // Higher resolution for 30-level quantization
```

**Also update lines 378** (same file) with same change.

**File:** `module1-hysteresis/pkg/gui/embedded.go`

**Line 23** change from:
```go
preisachGridSize := 50
```
To:
```go
preisachGridSize := 60
```

**File:** `module1-hysteresis/pkg/gui/controls.go`

**Line 148** change from:
```go
a.preisach = ferroelectric.NewMayergoyzPreisach(a.material, 50)
```
To:
```go
a.preisach = ferroelectric.NewMayergoyzPreisach(a.material, 60)
```

### Step 3: Widen Hysteron Distribution

**File:** `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go`

**Lines 59-61** change from:
```go
AlphaSigma:    material.Ec * 0.2, // 20% distribution
BetaMean:      -material.Ec,
BetaSigma:     material.Ec * 0.2,
```
To:
```go
AlphaSigma:    material.Ec * 0.28, // 28% distribution (Mayergoyz literature: 0.25-0.35)
BetaMean:      -material.Ec,
BetaSigma:     material.Ec * 0.28,
```

### Step 4: Expand Calibration Bounds

**File:** `module1-hysteresis/pkg/gui/simulation.go`

**Lines 1609-1614** change from:
```go
for i := 0; i < numLevels; i++ {
    // Initial bounds: full range (will be narrowed by runtime feedback)
    a.calibUpLow[i] = Ec * 0.5
    a.calibUpHigh[i] = Emax
    a.calibDownLow[i] = -Emax
    a.calibDownHigh[i] = -Ec * 0.5
}
```
To:
```go
for i := 0; i < numLevels; i++ {
    // Wider initial bounds for better convergence
    a.calibUpLow[i] = Ec * 0.3       // Lower starting bound
    a.calibUpHigh[i] = Ec * 2.0       // Higher upper bound (was 1.5*Ec)
    a.calibDownLow[i] = -Ec * 2.0    // More negative lower bound
    a.calibDownHigh[i] = -Ec * 0.3   // Higher upper bound
}
```

### Step 5: Add Smoothness Test

**File:** `module1-hysteresis/pkg/ferroelectric/preisach_advanced_test.go`

**First, update the imports (lines 3-5)** from:
```go
import (
	"testing"
)
```
To:
```go
import (
	"math"
	"testing"
)
```

**Then append to end of file:**

```go
// TestPECurveSmoothness verifies the P-E curve has enough granularity for 30-level quantization.
func TestPECurveSmoothness(t *testing.T) {
    material := DefaultHZO()
    model := NewMayergoyzPreisach(material, 60) // Match updated GUI grid size

    Emax := material.Ec * 2.0
    E, P := model.GetHysteresisLoop(Emax, 100)

    // Count unique P values in -Pr to +Pr range
    Pr := material.Pr
    uniqueP := make(map[float64]bool)
    for _, p := range P {
        if p >= -Pr && p <= Pr {
            // Round to 5% of Pr for comparison
            rounded := math.Round(p/(Pr*0.05)) * (Pr * 0.05)
            uniqueP[rounded] = true
        }
    }

    // Should have at least 20 distinct levels in the polarization range
    if len(uniqueP) < 20 {
        t.Errorf("P-E curve too coarse: only %d distinct P values (expected >= 20)", len(uniqueP))
    }
    t.Logf("P-E curve smoothness: %d distinct P values in ±Pr range", len(uniqueP))
}
```

### Step 6: Add Better Calibration Quality Diagnostics

**File:** `module1-hysteresis/pkg/gui/simulation.go`

**After line 274** (after `validateCalibration()` call in `loadTempCalibration`), add:

```go
// Log critical calibration quality issues
upDupes := countDuplicates(a.calibrationUp)
downDupes := countDuplicates(a.calibrationDown)
if upDupes > 10 || downDupes > 10 {
    log.Printf("CRITICAL: Calibration has %d/%d duplicate E-fields.", upDupes, downDupes)
    log.Printf("  Consider: increasing grid size or widening distribution (σ)")
}
```

Note: `countDuplicates` already exists at line 278-298.

## Verification Commands

```bash
# Run physics tests (should all pass)
go test ./module1-hysteresis/pkg/ferroelectric/... -v -run "Preisach|Physics"

# Run new smoothness test
go test ./module1-hysteresis/pkg/ferroelectric/... -v -run "Smoothness"

# Delete old calibration to force recalibration
rm data/hysteresis_calibration.json

# Build and test application
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools

# In the app:
# 1. Open Hysteresis module
# 2. Wait for calibration to complete
# 3. Check logs for "duplicate E-fields" count (should be <10)
# 4. Switch to Manual mode
# 5. Click levels: 1 → 20 → 1 → 20 → 30 → 10
# 6. Verify each target is hit within ±2 levels
# 7. Check logs for "ANIMATION COMPLETE" entries
```

## File Summary

| File | Action | Specific Change |
|------|--------|-----------------|
| `pkg/gui/simulation.go:554-555` | Edit | Add `midLevel := a.numLevels / 2`, use instead of `startLevel` |
| `pkg/gui/simulation.go:619` | Edit | Add `midLevel := a.numLevels / 2`, use instead of `startLevel` |
| `pkg/gui/gui.go:316,378` | Edit | Change `preisachGridSize := 50` to `60` |
| `pkg/gui/embedded.go:23` | Edit | Change `preisachGridSize := 50` to `60` |
| `pkg/gui/controls.go:148` | Edit | Change grid size from `50` to `60` |
| `pkg/ferroelectric/preisach_advanced.go:59,61` | Edit | Change `0.2` to `0.28` for both σ values |
| `pkg/gui/simulation.go:1611-1614` | Edit | Expand calibration bounds (0.3-2.0 range) |
| `pkg/ferroelectric/preisach_advanced_test.go` | Add | New `TestPECurveSmoothness` test |
| `pkg/gui/simulation.go:~275` | Add | Critical calibration quality logging |

## Risk Mitigation

1. **Performance**: 60×60 grid = 1.44× more hysterons than 50×50
   - Mitigation: Still runs at 60 FPS (tested with 100×100 grid)

2. **Breaking existing calibration**: Calibration file will be invalidated
   - Mitigation: Code detects material/level mismatch and recalibrates

3. **P-E curve shape change**: Wider σ changes the curve
   - Mitigation: Still matches HfO2 physics (within literature range)

## Dependencies

- No new external dependencies
- All changes are to existing files

---

PLAN_READY: .omc/plans/hysteresis-physics-fix.md
