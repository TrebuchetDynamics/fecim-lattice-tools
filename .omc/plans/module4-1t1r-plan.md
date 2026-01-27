# Module 4: 1T1R Architecture Toggle Implementation Plan

## Overview

Add 1T1R (1 Transistor-1 Resistor) architecture toggle to Module 4 (Peripheral Circuits), mirroring the existing implementation in Module 2 (Crossbar). This toggle affects how rows are selected during Read, Write, and Compute operations.

## Physics Background

### What 1T1R Does

**1T1R Architecture:**
- Each memory cell has a transistor acting as a "gate" for row selection
- **Word Line (WL)** controls the transistor gate (not the FeFET directly)
- When WL=HIGH: transistor ON, current flows through selected row
- When WL=LOW: transistor OFF, row isolated (~1000:1 on/off ratio)

**0T1R (Passive) Architecture:**
- FeFET connects directly between WL and BL
- No isolation - current can "sneak" through unselected paths
- Sneak path errors: 5-20% of signal

### Operation Mode Behavior

| Mode | 0T1R (Passive) | 1T1R (Gated) |
|------|----------------|--------------|
| **WRITE** | All rows see partial voltage | Only selected row gets full write pulse |
| **READ** | Sneak paths corrupt sense current | Clean isolated read of single cell |
| **COMPUTE** | All rows active (intentional) | All transistors ON for full MVM |

**Key Insight**: In COMPUTE mode, 1T1R transistors are ALL turned ON to enable full matrix-vector multiplication. The toggle affects:
- WRITE: Single row selection (transistor gates one row)
- READ: Single row selection (transistor gates one row)
- COMPUTE: All rows ON (all transistors conducting)

## Visual Design

### Toggle Button Placement

Add to OPERATIONS view, below the mode selector (WRITE/READ/COMPUTE):

```
┌─────────────────────────────────────────────────────────────┐
│  View: [OPERATIONS ▼]                                       │
├─────────────────────────────────────────────────────────────┤
│  Mode: ◉ WRITE  ○ READ  ○ COMPUTE        Status: Ready      │
│  Architecture: [PASSIVE] [1T1R GATE]     ← NEW TOGGLE       │
│  WRITE: Program cells using DAC voltage pulses...           │
├───────────────────────────┬─────────────────────────────────┤
│                           │                                 │
│   CROSSBAR ARRAY          │   Mode-specific config panel    │
│   (with row indicators)   │                                 │
│                           │                                 │
└───────────────────────────┴─────────────────────────────────┘
```

### Array Visualization Changes

When 1T1R is selected, draw transistor symbols on the LEFT side of each row:

```
PASSIVE (0T1R):                1T1R GATE:
                               ┌─T─┐
  x0   x1   x2   x3              │  x0   x1   x2   x3
   │    │    │    │            ├─○──●────●────●────●─→ y0
   ●────●────●────●─→ y0       ├─○──●────●────●────●─→ y1
   ●────●────●────●─→ y1       ├─○──●────●────●────●─→ y2
   ●────●────●────●─→ y2       ├─○──●────●────●────●─→ y3
   ●────●────●────●─→ y3       └─T─┘
                               (○ = transistor gate)
```

### Row Selection Indicators

- **WRITE mode (1T1R)**: Highlight ONLY the selected row's transistor as "ON"
- **READ mode (1T1R)**: Highlight ONLY the selected row's transistor as "ON"
- **COMPUTE mode (1T1R)**: Show ALL transistors as "ON" (bright green)
- **All modes (0T1R)**: No transistor indicators shown

## Implementation Tasks

### Phase 1: Add State and Toggle UI

**File: `module4-circuits/pkg/gui/app.go`**

1. Add architecture field to `CircuitsApp` struct:
```go
// Add near line 60 (after inputVector, outputVector)
architecture string // "1T1R" or "0T1R" (default: 0T1R for educational comparison)
```

2. Add toggle button fields:
```go
// Add near line 188 (after opsComputeButtons)
archPassiveBtn  *widget.Button
arch1T1RBtn     *widget.Button
archToggle      *fyne.Container
```

3. Initialize in `NewCircuitsApp()`:
```go
// Add after line 210 (before compArraySize initialization)
architecture: sharedwidgets.Architecture0T1R, // Default to passive for educational demo
```

### Phase 2: Create Toggle UI Component

**File: `module4-circuits/pkg/gui/tab_operations.go`**

1. Add architecture toggle creation function (after `createModeSelector`):
```go
// createArchitectureToggle creates the 0T1R/1T1R toggle buttons
func (ca *CircuitsApp) createArchitectureToggle() fyne.CanvasObject {
    // Create toggle buttons (same pattern as Module 2)
    ca.archPassiveBtn = widget.NewButton("PASSIVE", nil)
    ca.arch1T1RBtn = widget.NewButton("1T1R GATE", nil)

    // Helper to update button styles
    updateArchButtons := func() {
        if ca.architecture == sharedwidgets.Architecture0T1R {
            ca.archPassiveBtn.Importance = widget.HighImportance
            ca.arch1T1RBtn.Importance = widget.LowImportance
        } else {
            ca.archPassiveBtn.Importance = widget.LowImportance
            ca.arch1T1RBtn.Importance = widget.HighImportance
        }
        ca.archPassiveBtn.Refresh()
        ca.arch1T1RBtn.Refresh()
    }

    updateArchButtons()

    // Wire callbacks
    ca.archPassiveBtn.OnTapped = func() {
        if ca.architecture == sharedwidgets.Architecture0T1R {
            return
        }
        ca.mu.Lock()
        ca.architecture = sharedwidgets.Architecture0T1R
        ca.mu.Unlock()
        updateArchButtons()
        ca.refreshSharedArray()
        ca.updateArchitectureHelp()
    }

    ca.arch1T1RBtn.OnTapped = func() {
        if ca.architecture == sharedwidgets.Architecture1T1R {
            return
        }
        ca.mu.Lock()
        ca.architecture = sharedwidgets.Architecture1T1R
        ca.mu.Unlock()
        updateArchButtons()
        ca.refreshSharedArray()
        ca.updateArchitectureHelp()
    }

    ca.archToggle = container.NewGridWithColumns(2, ca.archPassiveBtn, ca.arch1T1RBtn)

    archLabel := widget.NewLabel("Array:")
    return container.NewHBox(archLabel, ca.archToggle)
}
```

2. Update `createModeSelector` to include architecture toggle:
```go
// In createModeSelector(), add architecture toggle after mode radio
archToggle := ca.createArchitectureToggle()

return container.NewVBox(
    container.NewHBox(
        widget.NewLabel("Mode:"),
        modeRadio,
        layout.NewSpacer(),
        archToggle,  // ADD THIS
        layout.NewSpacer(),
        ca.operationsStatusLabel,
    ),
    modeHelp,
    widget.NewSeparator(),
)
```

### Phase 3: Visual Integration - Array Drawing

**File: `module4-circuits/pkg/gui/tab_operations.go`**

Update `drawSharedArray` to show transistor indicators when 1T1R is selected:

1. After reading state (line ~250), add architecture read:
```go
arch := ca.architecture
```

2. Add transistor drawing section (after line ~326, before "Draw cells"):
```go
// Draw 1T1R transistor indicators on left side of rows
if arch == sharedwidgets.Architecture1T1R {
    transistorWidth := 20
    transistorX := offsetX - transistorWidth - 5

    for r := 0; r < rows; r++ {
        transistorY := offsetY + r*cellSize + cellSize/2

        // Determine transistor state based on mode
        var transistorOn bool
        switch mode {
        case ModeWrite, ModeRead:
            // Only selected row is ON
            transistorOn = (r == selectedRow)
        case ModeCompute:
            // All rows ON for MVM
            transistorOn = true
        }

        // Draw transistor symbol (simplified: circle with line)
        var tColor color.RGBA
        if transistorOn {
            tColor = color.RGBA{100, 255, 100, 255} // Bright green = ON
        } else {
            tColor = color.RGBA{80, 80, 80, 255}    // Gray = OFF
        }

        // Draw gate circle
        radius := 5
        cx := transistorX + transistorWidth/2
        cy := transistorY
        for dy := -radius; dy <= radius; dy++ {
            for dx := -radius; dx <= radius; dx++ {
                if dx*dx+dy*dy <= radius*radius {
                    px, py := cx+dx, cy+dy
                    if px >= 0 && px < w && py >= 0 && py < h {
                        img.Set(px, py, tColor)
                    }
                }
            }
        }

        // Draw connection line to row
        for x := cx + radius; x < offsetX; x++ {
            if x >= 0 && x < w {
                img.Set(x, cy, tColor)
            }
        }
    }
}
```

### Phase 4: Mode Help Text Update

Add architecture context to mode help:

```go
// updateArchitectureHelp shows how 1T1R affects current mode
func (ca *CircuitsApp) updateArchitectureHelp() {
    // Called after architecture toggle change
    // Mode help already shows mode-specific info
    // This could update a tooltip or secondary label
}
```

Update `updateModeHelp` to include architecture context:

```go
func (ca *CircuitsApp) updateModeHelp() {
    ca.mu.RLock()
    mode := ca.currentMode
    arch := ca.architecture
    ca.mu.RUnlock()

    var helpText string
    is1T1R := arch == sharedwidgets.Architecture1T1R

    switch mode {
    case ModeWrite:
        if is1T1R {
            helpText = "WRITE: Transistor gates ONLY selected row. Full write pulse to target cell."
        } else {
            helpText = "WRITE: Passive array - partial voltages affect unselected rows (sneak paths)."
        }
    case ModeRead:
        if is1T1R {
            helpText = "READ: Transistor isolates selected row. Clean sense current from target cell."
        } else {
            helpText = "READ: Passive array - sneak currents add ~5-20% noise to sense signal."
        }
    case ModeCompute:
        if is1T1R {
            helpText = "COMPUTE: ALL transistors ON for full MVM. Sneak-free parallel computation."
        } else {
            helpText = "COMPUTE: Passive MVM - sneak paths cause ~5-20% output error."
        }
    }

    fyne.Do(func() {
        ca.operationsModeHelp.SetText(helpText)
    })
}
```

### Phase 5: Legend Update

Update the legend label in `createSharedArraySection`:

```go
// After architecture toggle is implemented, update legend dynamically
func (ca *CircuitsApp) updateLegend() {
    ca.mu.RLock()
    arch := ca.architecture
    ca.mu.RUnlock()

    var legendText string
    if arch == sharedwidgets.Architecture1T1R {
        legendText = "Level: Low→High | Yellow=Selected | Green○=Transistor ON | Gray○=OFF"
    } else {
        legendText = "Level: Low (blue)→High (red) | Yellow=Selected | Click to select"
    }

    // Update legend label
}
```

## Files to Modify

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/app.go` | Add `architecture` field, toggle button fields, initialization |
| `module4-circuits/pkg/gui/tab_operations.go` | Add toggle creation, update `drawSharedArray`, update mode help |

## Testing Checklist

- [ ] Toggle switches between PASSIVE and 1T1R GATE visually
- [ ] WRITE mode: Only selected row shows green transistor (1T1R)
- [ ] READ mode: Only selected row shows green transistor (1T1R)
- [ ] COMPUTE mode: ALL rows show green transistors (1T1R)
- [ ] Mode help text updates based on architecture selection
- [ ] Array visualization shows transistor symbols on left (1T1R only)
- [ ] Clicking different cells updates which row shows "ON" transistor
- [ ] Toggle state persists across mode changes

## Future Enhancements (Not in Scope)

1. **Sneak path visualization**: Show parasitic current paths in 0T1R mode
2. **Error metrics**: Display computed vs ideal output difference per architecture
3. **TIA integration**: Adjust TIA saturation thresholds based on architecture
4. **Animation**: Show transistor switching during write/read operations
