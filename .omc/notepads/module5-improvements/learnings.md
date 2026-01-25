# Module 5 UI Improvements - Learnings

## Scroll Indicator Implementation (2026-01-25)

### Problem
Users were not discovering content below the fold in Module 5. The Data Center Calculator and other important sections were hidden below the visible viewport (900px tall window).

### Solution
Added a visual scroll indicator below the hero Energy Race section:
- Text: "▼ Scroll down for Data Center Calculator, Market Analysis, and more ▼"
- Style: Center-aligned, italic, low importance (subtle gray)
- Placement: Between hero section and first row of content

### Location
File: `<local-path>`
Lines: After `heroEnergyRace`, before `row1`

### Pattern
This pattern can be reused in other scrollable content areas where:
1. Important content exists below the fold
2. Users may not realize scrolling is needed
3. A gentle hint improves discoverability without being intrusive

### Fyne Widget Usage
```go
scrollHintLabel := widget.NewLabelWithStyle(
    "▼ Scroll down for more ▼", 
    fyne.TextAlignCenter, 
    fyne.TextStyle{Italic: true}
)
scrollHintLabel.Importance = widget.LowImportance
```

The `LowImportance` setting applies theme-specific subtle coloring (usually gray).
