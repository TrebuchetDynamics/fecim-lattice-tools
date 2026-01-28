# Hysteresis Plot Spike Fix Plan

## Summary

Fix vertical spikes appearing in the P-E hysteresis plot at transition corners when temperature changes. The root cause is a mismatch between temperature-corrected physics values (Ec(T), Pr(T)) and the plot bounds/markers which continue to use base material values.

## Root Cause Analysis

### Physics Background
- **Ec(T) = Ec0 * (1 - T/Tc)^beta** where Tc=723K (Curie temp), beta=0.5
- At 400K: Ec(T) = 0.67 * Ec0 (significant 33% reduction)
- At 600K: Ec(T) = 0.40 * Ec0 (60% reduction)

### Current Code Issues

1. **controls.go:148** - Material selection uses base Ec for bounds:
   ```go
   a.plot.SetBounds(a.material.Ec*1.5, a.material.Ps*1.2)
   ```
   Should use: `a.preisach.GetEffectiveEc()*1.5`

2. **gui.go:458-460** - Plot creation uses base material values:
   ```go
   a.plot = widgets.NewPEPlot(a.material.Ec*1.5, a.material.Ps*1.2, ...)
   a.plot.SetMaterialParams(a.material.Ec, a.material.Pr)
   ```
   Should use temperature-corrected values from preisach model.

3. **simulation.go:335** - Temperature change handler updates markers but NOT bounds:
   ```go
   // Update plot markers (outside lock, uses fyne.Do internally)
   a.plot.SetMaterialParams(effEc, effPr)
   ```
   Missing: `a.plot.SetBounds(effEc*1.5, effPr*1.2)` call

4. **History data mismatch** - History points collected at different temperatures have different E-field ranges. When temperature changes and bounds are updated, old history data may appear as spikes because the scale changed.

## Acceptance Criteria

- [ ] **AC1**: Plot bounds update dynamically when temperature slider changes
- [ ] **AC2**: Ec markers (+Ec, -Ec vertical lines) move to correct temperature-corrected positions
- [ ] **AC3**: Pr markers (+Pr, -Pr horizontal lines) move to correct temperature-corrected positions
- [ ] **AC4**: No vertical spikes appear at plot edges during temperature transitions
- [ ] **AC5**: History trail is cleared when temperature changes by more than 25K to prevent scale artifacts
- [ ] **AC6**: Axis tick labels update to reflect new temperature-corrected bounds
- [ ] **AC7**: All existing waveform modes (Manual, Sine, Triangle, Write/Read Demo, Time-Resolved) work correctly after fix

## Implementation Steps

### Task 1: Update Temperature Change Handler (simulation.go)
**File:** `<local-path>`
**Lines:** 326-336

**Current code:**
```go
go func() {
    a.mu.Lock()
    a.onTemperatureChanged(v)
    // Get plot markers with temperature-corrected Ec and Pr
    effEc := a.preisach.GetEffectiveEc()
    effPr := a.preisach.GetEffectivePr()
    a.mu.Unlock()

    // Update plot markers (outside lock, uses fyne.Do internally)
    a.plot.SetMaterialParams(effEc, effPr)
}()
```

**Changes needed:**
1. Add `SetBounds()` call after `SetMaterialParams()` using temperature-corrected values
2. Clear history when temperature changes significantly (>25K) to prevent scale artifacts
3. Store previous temperature to detect significant changes

**New code:**
```go
go func() {
    a.mu.Lock()
    previousTemp := a.calibrationTemp
    a.onTemperatureChanged(v)
    // Get plot markers with temperature-corrected Ec and Pr
    effEc := a.preisach.GetEffectiveEc()
    effPr := a.preisach.GetEffectivePr()
    currentTemp := a.preisach.Temperature

    // Clear history if temperature changed significantly to prevent scale artifacts
    if math.Abs(currentTemp - previousTemp) > 25 {
        a.eHistory = a.eHistory[:0]
        a.pHistory = a.pHistory[:0]
    }
    a.mu.Unlock()

    // Update plot bounds AND markers with temperature-corrected values
    a.plot.SetBounds(effEc*1.5, effPr*1.2)
    a.plot.SetMaterialParams(effEc, effPr)
}()
```

### Task 2: Update Material Selection Handler (controls.go)
**File:** `<local-path>`
**Lines:** 145-149

**Current code:**
```go
a.eHistory = a.eHistory[:0]
a.pHistory = a.pHistory[:0]
a.plot.SetBounds(a.material.Ec*1.5, a.material.Ps*1.2)
a.plot.SetMaterialParams(a.material.Ec, a.material.Pr)
```

**Changes needed:**
Use temperature-corrected values from the newly created Preisach model.

**New code:**
```go
a.eHistory = a.eHistory[:0]
a.pHistory = a.pHistory[:0]
// Use temperature-corrected values for plot bounds and markers
effEc := a.preisach.GetEffectiveEc()
effPr := a.preisach.GetEffectivePr()
a.plot.SetBounds(effEc*1.5, effPr*1.2)
a.plot.SetMaterialParams(effEc, effPr)
```

### Task 3: Update Initial Plot Creation (gui.go)
**File:** `<local-path>`
**Lines:** 458-460

**Current code:**
```go
a.plot = widgets.NewPEPlot(a.material.Ec*1.5, a.material.Ps*1.2, ColorBackground, ColorGrid, ColorAxis, ColorPositive, ColorNegative, ColorWarning)
a.plot.SetMinSize(fyne.NewSize(400, 350))
a.plot.SetMaterialParams(a.material.Ec, a.material.Pr)
```

**Changes needed:**
Use temperature-corrected values from preisach model for initial plot setup.

**New code:**
```go
// Use temperature-corrected values for initial plot setup
effEc := a.preisach.GetEffectiveEc()
effPr := a.preisach.GetEffectivePr()
a.plot = widgets.NewPEPlot(effEc*1.5, effPr*1.2, ColorBackground, ColorGrid, ColorAxis, ColorPositive, ColorNegative, ColorWarning)
a.plot.SetMinSize(fyne.NewSize(400, 350))
a.plot.SetMaterialParams(effEc, effPr)
```

### Task 4: Add Import for math Package (controls.go)
**File:** `<local-path>`
**Line:** 5 (import block)

**Changes needed:**
Add "math" to import block if not already present (needed for `math.Abs` in temperature comparison).

Check current imports and add if missing:
```go
import (
    "fmt"
    "math"
    "math/rand"
    "strconv"
    ...
)
```

Note: `math` is already imported in controls.go (line 5), so no change needed for this file.

### Task 5: Verify Preisach Temperature State During Material Change
**File:** `<local-path>`
**Lines:** 145-198 (material selection handler)

**Verification needed:**
Ensure the Preisach model's temperature is preserved when material changes. Current code creates a new Preisach model (line 145) which defaults to 300K. Need to check if current temperature is applied.

**Current code:**
```go
a.preisach = ferroelectric.NewMayergoyzPreisach(a.material, 50)
```

**Potential issue:** New Preisach model is created with default temperature (300K), not current slider temperature.

**Fix approach:** After creating new Preisach model, set temperature to current slider value. This happens in the background goroutine (line 186-197) via `calibrateLevelsAtTemperature(currentTemp)` which calls `a.preisach.SetTemperature(tempK)`. However, the bounds are set BEFORE this happens (line 148-149).

**Revised Task 2 - Complete fix for controls.go:131-199:**
Move the SetBounds/SetMaterialParams calls to AFTER the Preisach temperature is set, or set temperature immediately after creating the model.

### Task 6: Add Verification Tests
**File:** `<local-path>` (may need to create)

**Tests to add:**
1. Test that plot bounds update correctly when temperature changes
2. Test that history is cleared on significant temperature change
3. Test that Ec/Pr markers are positioned correctly at different temperatures

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Regression in waveform modes | Medium | High | Run all existing tests, manual verification of each mode |
| Thread safety issues with plot updates | Low | Medium | Use fyne.Do() for all UI updates, maintain existing locking patterns |
| Performance impact from frequent bounds updates | Low | Low | SetBounds() is lightweight, only triggers on temperature change |
| Calibration state corruption | Low | High | Temperature check happens inside existing lock, calibration logic unchanged |

## Verification Steps

### Manual Verification Checklist
1. [ ] Launch app: `go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools`
2. [ ] Select Hysteresis module
3. [ ] Set temperature to 300K (default) - verify normal hysteresis loop
4. [ ] Increase temperature to 400K - verify:
   - [ ] No vertical spikes at edges
   - [ ] Ec markers moved inward (lower field)
   - [ ] History trail cleared
   - [ ] Loop shape appropriate for lower Ec
5. [ ] Increase temperature to 500K - verify same behavior
6. [ ] Decrease temperature back to 300K - verify recovery
7. [ ] Test with different waveforms:
   - [ ] Sine Wave
   - [ ] Triangle Wave
   - [ ] Write/Read Demo
   - [ ] Time-Resolved Switching
   - [ ] Manual mode
8. [ ] Change materials while at non-300K temperature - verify bounds update correctly

### Automated Tests
```bash
# Run all module1 tests
go test ./module1-hysteresis/...

# Run with verbose output
go test -v ./module1-hysteresis/pkg/gui/...
```

## Dependencies

- Task 2 depends on Task 1 pattern
- Task 3 is independent
- Task 4 is required for Task 1
- Task 5 is a refinement of Task 2
- Task 6 should be done last

## Commit Strategy

**Single commit** with message:
```
fix(hysteresis): use temperature-corrected Ec/Pr for plot bounds and markers

- Update plot bounds dynamically when temperature changes
- Clear history trail on significant temperature changes (>25K) to prevent scale artifacts
- Use GetEffectiveEc()/GetEffectivePr() instead of base material values
- Apply temperature correction during material selection and initial plot creation

Fixes vertical spikes appearing at plot edges during temperature transitions.
```

## Estimated Effort

| Task | Complexity | Time Estimate |
|------|------------|---------------|
| Task 1 | Medium | 15 min |
| Task 2 | Low | 10 min |
| Task 3 | Low | 5 min |
| Task 4 | Trivial | 2 min |
| Task 5 | Medium | 10 min |
| Task 6 | Medium | 20 min |
| **Testing & Verification** | - | 20 min |
| **Total** | - | ~80 min |
