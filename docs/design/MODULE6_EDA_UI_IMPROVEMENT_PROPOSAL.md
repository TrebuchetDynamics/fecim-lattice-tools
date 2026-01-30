# FeCIM EDA Design Suite - UI Improvement Proposal

**Module**: 6 - EDA Tools
**Date**: 2026-01-30
**Version**: 1.0
**Status**: Proposal

---

## Executive Summary

This document proposes comprehensive UI improvements for the FeCIM EDA Design Suite (Module 6). The current interface, while functional, suffers from visual density, cramped layouts, and insufficient visual hierarchy. These improvements aim to create a professional, modern EDA tool interface that balances information density with clarity and usability.

---

## Table of Contents

1. [Current State Analysis](#current-state-analysis)
2. [Design Philosophy](#design-philosophy)
3. [Color System Redesign](#color-system-redesign)
4. [Typography System](#typography-system)
5. [Layout Architecture](#layout-architecture)
6. [Component Redesign](#component-redesign)
7. [Builder Tab Improvements](#builder-tab-improvements)
8. [Learn Tab Improvements](#learn-tab-improvements)
9. [Animation and Interaction](#animation-and-interaction)
10. [Fyne Implementation Notes](#fyne-implementation-notes)
11. [Priority Matrix](#priority-matrix)

---

## 1. Current State Analysis

### Identified Issues

| Issue | Severity | Location | Impact |
|-------|----------|----------|--------|
| Dense config panels | High | Builder Tab | Users struggle to parse parameters |
| Small preview areas | High | Verilog/DEF tabs | Code readability suffers |
| Layout visualization cramped | High | Layout tab | Can't see schematic details |
| Status indicators hard to read | Medium | Validation section | Pass/fail unclear at glance |
| Log section oversized | Medium | Bottom area | Takes space from useful content |
| Stats row with pipe separators | Medium | Array Config | Visual noise, poor hierarchy |
| Inconsistent spacing | Medium | Throughout | Unprofessional appearance |
| Mode help text cramped | Low | Array Config | Educational content lost |
| Topic sidebar too narrow | Low | Learn Tab | Topic names truncated |

### Current Architecture

```
+----------------------------------------------------------+
|  Header (View Selector + Banner)                          |
+----------------------------------------------------------+
|  Config HSplit (Cell Config | Array Config)               |
|  - 6-column grids for entries                            |
|  - Stats in single row with "|" separators               |
+----------------------------------------------------------+
|  Action Buttons + Status                                  |
+----------------------------------------------------------+
|  Preview Tabs (Verilog | DEF | Layout)                   |
|  - MultiLineEntry for code                               |
|  - 3-column grid for layout images                       |
+----------------------------------------------------------+
|  Validation Section                                       |
|  - Results in HBox with "|" separators                   |
|  - OpenLane status row                                   |
|  - Log MultiLineEntry (100px height)                     |
+----------------------------------------------------------+
```

---

## 2. Design Philosophy

### Aesthetic Direction: **Technical Precision**

A refined, professional EDA tool aesthetic that communicates technical sophistication while remaining approachable. Think: VSCode meets KiCad meets modern design systems.

**Tone**: Clean, technical, trustworthy
**Key Differentiator**: Contextual card-based information architecture with subtle depth cues

### Design Principles

1. **Information Architecture First** - Group related controls, separate concerns
2. **Progressive Disclosure** - Show essential controls, reveal advanced on demand
3. **Visual Hierarchy Through Depth** - Use elevation (card surfaces) instead of dense grids
4. **Breathing Room** - Generous padding creates scannable interfaces
5. **Status at a Glance** - Color-coded badges, not text labels

---

## 3. Color System Redesign

### Current Palette Analysis

The existing FeCIM theme uses:
- `ColorBackground`: `#003264` (dark blue)
- `ColorSurface`: `#004682` (medium blue)
- `ColorPrimary`: `#00D4FF` (cyan)
- `ColorText`: `#F0F4F8` (off-white)

### Proposed Enhancements

Keep the dark blue foundation but introduce **semantic depth layers** and **improved status colors**.

```go
// Background Hierarchy (4 levels for card depth)
ColorBG0 = color.RGBA{0, 40, 82, 255}     // #002852 - Deepest (page background)
ColorBG1 = color.RGBA{0, 50, 100, 255}    // #003264 - Level 1 (main panels)
ColorBG2 = color.RGBA{0, 65, 120, 255}    // #004178 - Level 2 (cards)
ColorBG3 = color.RGBA{0, 80, 140, 255}    // #00508C - Level 3 (elevated/hover)

// Status Colors (improved contrast)
ColorStatusPass    = color.RGBA{52, 211, 153, 255}   // #34D399 - Emerald green
ColorStatusFail    = color.RGBA{248, 113, 113, 255}  // #F87171 - Soft red
ColorStatusPending = color.RGBA{251, 191, 36, 255}   // #FBBF24 - Amber
ColorStatusSkip    = color.RGBA{156, 163, 175, 255}  // #9CA3AF - Gray

// Accent Colors
ColorAccentPrimary   = color.RGBA{0, 212, 255, 255}   // #00D4FF - Cyan (CTAs)
ColorAccentSecondary = color.RGBA{139, 92, 246, 255}  // #8B5CF6 - Violet (highlights)
ColorAccentWarm      = color.RGBA{251, 146, 60, 255}  // #FB923C - Orange (warnings)

// Code Syntax (for preview areas)
ColorCodeBG      = color.RGBA{0, 35, 70, 255}   // #002346 - Darker than cards
ColorCodeKeyword = color.RGBA{199, 146, 234, 255} // #C792EA - Purple
ColorCodeString  = color.RGBA{195, 232, 141, 255} // #C3E88D - Green
ColorCodeComment = color.RGBA{97, 115, 139, 255}  // #61738B - Muted

// Border and Separator
ColorBorder      = color.RGBA{0, 100, 170, 80}  // Semi-transparent blue
ColorSeparator   = color.RGBA{0, 90, 160, 128}  // #005AA0 @ 50%
```

### Status Badge Design

Replace text status indicators with colored badges:

| Status | Background | Text | Icon |
|--------|------------|------|------|
| Pass | `#34D399` @ 20% | `#34D399` | Checkmark |
| Fail | `#F87171` @ 20% | `#F87171` | X |
| Pending | `#FBBF24` @ 20% | `#FBBF24` | Clock |
| Skip | `#9CA3AF` @ 20% | `#9CA3AF` | Minus |
| Running | `#00D4FF` @ 20% | `#00D4FF` | Spinner |

---

## 4. Typography System

### Current Issues
- Inconsistent font sizes
- No clear heading hierarchy
- Monospace for code mixed with UI text

### Proposed Type Scale

```
Type Scale (based on 1.25 ratio)
--------------------------------
Display:   24px  Bold     - Tab titles, major sections
Heading:   18px  Bold     - Card headers, panel titles
Subhead:   15px  SemiBold - Section labels
Body:      14px  Regular  - Standard UI text
Caption:   12px  Regular  - Helper text, timestamps
Micro:     10px  Regular  - Badges, status indicators

Code Scale
----------
Code-L:    14px  Mono     - Preview areas, main code
Code-S:    12px  Mono     - Inline code, log output
```

### Fyne Implementation

```go
// Create custom text styles
func HeadingText(text string) *canvas.Text {
    t := canvas.NewText(text, theme.ColorText)
    t.TextSize = 18
    t.TextStyle = fyne.TextStyle{Bold: true}
    return t
}

func SubheadText(text string) *canvas.Text {
    t := canvas.NewText(text, theme.ColorText)
    t.TextSize = 15
    t.TextStyle = fyne.TextStyle{Bold: true}
    return t
}

func CaptionText(text string) *canvas.Text {
    t := canvas.NewText(text, theme.ColorTextDim)
    t.TextSize = 12
    return t
}
```

---

## 5. Layout Architecture

### Proposed Builder Tab Structure

```
+------------------------------------------------------------------+
|  HEADER                                                           |
|  [View: Dropdown] [Banner centered]               [Status Badge]  |
+------------------------------------------------------------------+
|                                                                   |
|  +---------------------------+  +-------------------------------+ |
|  | CELL CONFIGURATION        |  | ARRAY CONFIGURATION           | |
|  | Card (elevated)           |  | Card (elevated)               | |
|  |                           |  |                               | |
|  | Name   [___________]      |  | Dimensions                    | |
|  |                           |  | Rows [__] x Cols [__]         | |
|  | Dimensions                |  |                               | |
|  | W [____] x H [____] um    |  | Architecture                  | |
|  |                           |  | [PASSIVE] [1T1R] [2T1R]       | |
|  | Timing                    |  |                               | |
|  | Rise [__] Fall [__] ns    |  | Mode                          | |
|  |                           |  | [storage v]                   | |
|  | Electrical                |  | > Storage mode: Non-volatile  | |
|  | Cap [__] pF               |  |   data retention using...     | |
|  | Leakage [__] nW           |  |                               | |
|  |                           |  +-------------------------------+ |
|  | Cell Area: 1.25 um2       |  | ARRAY STATISTICS              | |
|  +---------------------------+  | Card (compact)                | |
|                                 | Total   Area    Density       | |
|                                 | 16      20um2   0.8/um2       | |
|                                 |                               | |
|                                 | WL Len  BL Len  Util          | |
|                                 | 1.84um  10.9um  100%          | |
|                                 +-------------------------------+ |
|                                                                   |
+------------------------------------------------------------------+
|  ACTIONS                                                          |
|  [Generate All]  [Validate All]  [Export Package]                 |
+------------------------------------------------------------------+
|                                                                   |
|  PREVIEW (Tabs with larger content area)                          |
|  +--------------------------------------------------------------+ |
|  | [Verilog] [DEF] [Layout]                                     | |
|  |                                                              | |
|  |  +----------------------------------------------------------+| |
|  |  | // Generated Verilog                                     || |
|  |  | module fecim_crossbar_4x4 (                              || |
|  |  |     input [3:0] WL,                                      || |
|  |  |     output [3:0] BL,                                     || |
|  |  |     ...                                                  || |
|  |  +----------------------------------------------------------+| |
|  |                                                              | |
|  |  Stats: Instances: 16 | Lines: 45 | Size: 1.2KB             | |
|  +--------------------------------------------------------------+ |
|                                                                   |
+------------------------------------------------------------------+
|  VALIDATION (Collapsible)                                         |
|  +--------------------------------------------------------------+ |
|  | [Pass Badge] Yosys  [Pass Badge] DEF  [Fail Badge] Placement | |
|  |                                                              | |
|  | OpenLane: [Ready Badge] Docker  [Ready Badge] PDK            | |
|  +--------------------------------------------------------------+ |
|  | LOG (collapsed by default, expandable)                       | |
|  | > View Log (12 entries)                           [Clear]    | |
|  +--------------------------------------------------------------+ |
+------------------------------------------------------------------+
```

### Spacing Constants

```go
const (
    SpacingXS  = 4   // Tight spacing within groups
    SpacingS   = 8   // Between related elements
    SpacingM   = 16  // Between sections within cards
    SpacingL   = 24  // Between cards/major sections
    SpacingXL  = 32  // Page margins

    CardPadding    = 16
    CardRadius     = 8
    CardElevation  = 2  // Visual depth level

    InputHeight    = 36
    ButtonHeight   = 40
    BadgeHeight    = 24
)
```

---

## 6. Component Redesign

### 6.1 Configuration Cards

**Before**: Dense 6-column grids with labels inline

**After**: Grouped sections within elevated cards

```go
// Proposed Card Component
func ConfigCard(title string, content fyne.CanvasObject) fyne.CanvasObject {
    // Card background with elevation
    bg := canvas.NewRectangle(ColorBG2)
    bg.CornerRadius = CardRadius

    // Title
    titleLabel := SubheadText(title)

    // Separator
    sep := canvas.NewRectangle(ColorSeparator)
    sep.SetMinSize(fyne.NewSize(0, 1))

    return container.NewVBox(
        container.NewPadded(titleLabel),
        sep,
        container.NewPadded(content),
    )
}

// Grouped form field
func FormGroup(label string, widget fyne.CanvasObject) fyne.CanvasObject {
    labelWidget := CaptionText(label)
    return container.NewVBox(
        labelWidget,
        container.NewPadded(widget), // 4px padding
    )
}
```

### 6.2 Status Badges

**Before**: `widget.NewLabel("PASS")` with green text

**After**: Colored badge with icon

```go
func StatusBadge(status string) fyne.CanvasObject {
    var bgColor, textColor color.Color
    var icon fyne.Resource

    switch status {
    case "pass":
        bgColor = WithAlpha(ColorStatusPass, 50)
        textColor = ColorStatusPass
        icon = theme.ConfirmIcon()
    case "fail":
        bgColor = WithAlpha(ColorStatusFail, 50)
        textColor = ColorStatusFail
        icon = theme.CancelIcon()
    case "pending":
        bgColor = WithAlpha(ColorStatusPending, 50)
        textColor = ColorStatusPending
        icon = theme.HistoryIcon()
    case "skip":
        bgColor = WithAlpha(ColorStatusSkip, 50)
        textColor = ColorStatusSkip
        icon = theme.ContentRemoveIcon()
    }

    bg := canvas.NewRectangle(bgColor)
    bg.CornerRadius = 4

    iconWidget := widget.NewIcon(icon)
    label := canvas.NewText(strings.ToUpper(status), textColor)
    label.TextSize = 10
    label.TextStyle = fyne.TextStyle{Bold: true}

    content := container.NewHBox(
        iconWidget,
        layout.NewSpacer(),
        label,
    )

    return container.NewStack(bg, container.NewPadded(content))
}
```

### 6.3 Architecture Toggle Buttons

**Before**: Three separate buttons with `Importance` changes

**After**: Segmented control with clear active state

```go
func ArchitectureSelector(current string, onChange func(string)) fyne.CanvasObject {
    options := []string{"PASSIVE", "1T1R", "2T1R"}

    buttons := make([]fyne.CanvasObject, len(options))
    for i, opt := range options {
        btn := &SegmentButton{
            Text:     opt,
            Active:   opt == current,
            OnTapped: func() { onChange(opt) },
        }
        buttons[i] = btn
    }

    // Container with connected borders (no gaps)
    return container.NewHBox(buttons...)
}

// SegmentButton with active/inactive states
type SegmentButton struct {
    widget.BaseWidget
    Text     string
    Active   bool
    OnTapped func()
}

func (s *SegmentButton) CreateRenderer() fyne.WidgetRenderer {
    bg := canvas.NewRectangle(ColorBG2)
    if s.Active {
        bg.FillColor = ColorAccentPrimary
    }

    label := canvas.NewText(s.Text, ColorText)
    label.TextSize = 12
    label.TextStyle = fyne.TextStyle{Bold: s.Active}
    label.Alignment = fyne.TextAlignCenter

    return &segmentRenderer{bg: bg, label: label, button: s}
}
```

### 6.4 Statistics Display

**Before**: Single HBox with pipe separators

**After**: Grid of metric cards

```go
func StatCard(label, value, unit string) fyne.CanvasObject {
    valueText := canvas.NewText(value, ColorAccentPrimary)
    valueText.TextSize = 18
    valueText.TextStyle = fyne.TextStyle{Bold: true}

    labelText := canvas.NewText(label, ColorTextDim)
    labelText.TextSize = 11

    unitText := canvas.NewText(unit, ColorTextDim)
    unitText.TextSize = 11

    return container.NewVBox(
        labelText,
        container.NewHBox(valueText, unitText),
    )
}

// Usage
statsGrid := container.NewGridWithColumns(3,
    StatCard("Total Cells", "16", ""),
    StatCard("Array Area", "20.0", "um2"),
    StatCard("Density", "0.80", "cells/um2"),
    StatCard("WL Length", "1.84", "um"),
    StatCard("BL Length", "10.88", "um"),
    StatCard("Utilization", "100", "%"),
)
```

### 6.5 Code Preview Area

**Before**: Plain MultiLineEntry with default styling

**After**: Styled code block with header bar

```go
func CodePreview(language, content string, stats string) fyne.CanvasObject {
    // Header bar with language tag
    langBadge := canvas.NewRectangle(ColorAccentSecondary)
    langBadge.CornerRadius = 4
    langLabel := canvas.NewText(language, ColorText)
    langLabel.TextSize = 10

    statsLabel := CaptionText(stats)

    header := container.NewHBox(
        container.NewStack(langBadge, container.NewPadded(langLabel)),
        layout.NewSpacer(),
        statsLabel,
    )

    // Code area with darker background
    codeBg := canvas.NewRectangle(ColorCodeBG)
    codeBg.CornerRadius = 0

    codeEntry := widget.NewMultiLineEntry()
    codeEntry.Wrapping = fyne.TextWrapOff
    codeEntry.TextStyle = fyne.TextStyle{Monospace: true}
    codeEntry.SetText(content)

    codeScroll := container.NewScroll(codeEntry)

    return container.NewBorder(
        header,
        nil, nil, nil,
        container.NewStack(codeBg, codeScroll),
    )
}
```

### 6.6 Collapsible Log Section

**Before**: Fixed 100px log area always visible

**After**: Collapsible section, collapsed by default

```go
func CollapsibleLog(title string, entryCount int) fyne.CanvasObject {
    expanded := false

    logEntry := widget.NewMultiLineEntry()
    logEntry.Wrapping = fyne.TextWrapWord
    logEntry.TextStyle = fyne.TextStyle{Monospace: true}
    logEntry.SetMinRowsVisible(6)

    logScroll := container.NewScroll(logEntry)
    logScroll.SetMinSize(fyne.NewSize(0, 150))
    logScroll.Hide()

    clearBtn := widget.NewButton("Clear", func() {
        logEntry.SetText("")
    })
    clearBtn.Importance = widget.LowImportance

    countBadge := CaptionText(fmt.Sprintf("(%d entries)", entryCount))

    toggleIcon := widget.NewIcon(theme.MenuDropDownIcon())

    headerBtn := widget.NewButton("", func() {
        expanded = !expanded
        if expanded {
            logScroll.Show()
            toggleIcon.SetResource(theme.MenuDropUpIcon())
        } else {
            logScroll.Hide()
            toggleIcon.SetResource(theme.MenuDropDownIcon())
        }
    })
    headerBtn.Importance = widget.LowImportance

    header := container.NewHBox(
        toggleIcon,
        widget.NewLabel(title),
        countBadge,
        layout.NewSpacer(),
        clearBtn,
    )

    return container.NewVBox(header, logScroll)
}
```

---

## 7. Builder Tab Improvements

### 7.1 Cell Configuration Panel

**Current Problems**:
- 6-column grid cramped
- All fields same visual weight
- No logical grouping

**Proposed Solution**:

```
+----------------------------------+
| CELL CONFIGURATION               |
+----------------------------------+
| Cell Name                        |
| [fecim_bitcell_______________]   |
|                                  |
| PHYSICAL                         |
| +-------------+  +-------------+ |
| | Width       |  | Height      | |
| | [0.460] um  |  | [2.720] um  | |
| +-------------+  +-------------+ |
|                                  |
| Area: 1.2512 um2                 |
|                                  |
| TIMING                           |
| +-------------+  +-------------+ |
| | Rise        |  | Fall        | |
| | [0.1] ns    |  | [0.1] ns    | |
| +-------------+  +-------------+ |
|                                  |
| ELECTRICAL                       |
| +-------------+  +-------------+ |
| | Capacitance |  | Leakage     | |
| | [0.002] pF  |  | [0.001] nW  | |
| +-------------+  +-------------+ |
+----------------------------------+
```

### 7.2 Array Configuration Panel

**Current Problems**:
- Architecture buttons disconnected from context
- Mode help text cramped
- Stats compressed into single row

**Proposed Solution**:

```
+----------------------------------------+
| ARRAY CONFIGURATION                    |
+----------------------------------------+
| DIMENSIONS                             |
| Rows [4___]  x  Cols [4___]            |
|                                        |
| ARCHITECTURE                           |
| [===PASSIVE===] [  1T1R  ] [  2T1R  ]  |
|                                        |
| Passive: Dense packing, limited to     |
| small arrays due to sneak paths        |
|                                        |
| MODE                                   |
| [storage____________v]                 |
|                                        |
| > Storage mode: Non-volatile data      |
|   retention using FeCIM cells as       |
|   memory elements                      |
+----------------------------------------+
| STATISTICS                             |
+--------+--------+--------+             |
| 16     | 20.0   | 0.80   |             |
| cells  | um2    | /um2   |             |
| Total  | Area   | Density|             |
+--------+--------+--------+             |
| 1.84   | 10.88  | 100%   |             |
| um     | um     |        |             |
| WL Len | BL Len | Util   |             |
+--------+--------+--------+             |
+----------------------------------------+
```

### 7.3 Layout Preview Tab

**Current Problems**:
- Three 400x350 images cramped in row
- Status text small
- Generate buttons disconnected

**Proposed Solution**:

```
+----------------------------------------------------------+
| LAYOUT VISUALIZATION                                      |
+----------------------------------------------------------+
| +------------------------------------------------------+ |
| |                                                      | |
| |          [Primary Image - 70% width]                 | |
| |          KLayout or OpenROAD selected view           | |
| |                                                      | |
| +------------------------------------------------------+ |
|                                                          |
| VIEWS                                                    |
| +----------------+ +----------------+ +----------------+ |
| | KLayout        | | OpenROAD       | | Yosys          | |
| | [Thumbnail]    | | [Thumbnail]    | | [Thumbnail]    | |
| | [Active]       | | [Generate]     | | [Generate]     | |
| | Generated      | | Not generated  | | Not generated  | |
| +----------------+ +----------------+ +----------------+ |
+----------------------------------------------------------+
```

### 7.4 Validation Section

**Current Problems**:
- Results in dense HBox
- Status hard to scan
- OpenLane row separate

**Proposed Solution**:

```
+----------------------------------------------------------+
| VALIDATION RESULTS                          [Run All]     |
+----------------------------------------------------------+
| +------------+ +------------+ +------------+ +------------+|
| | Yosys      | | DEF        | | Cross-Check| | Placement  ||
| | [PASS]     | | [PASS]     | | [PASS]     | | [SKIP]     ||
| | Syntax OK  | | Valid      | | Consistent | | No Docker  ||
| +------------+ +------------+ +------------+ +------------+|
|                                                           |
| ENVIRONMENT                                               |
| Docker: [Ready]  PDK: [Optional]  [Pull Image]            |
+----------------------------------------------------------+
| > Log (collapsed)                              [Clear]    |
+----------------------------------------------------------+
```

---

## 8. Learn Tab Improvements

### 8.1 Sidebar Navigation

**Current Problems**:
- 180px sidebar narrow
- Topics as list items
- No visual hierarchy

**Proposed Solution**:

```
+------------------+
| TOPICS           |
+------------------+
| [1] What is      |
|     FeCIM EDA?   |
|     ============ |
|                  |
| [2] Crossbar     |
|     Architecture |
|                  |
| [3] EDA Files    |
|     We Generate  |
+------------------+
```

Width: 240px (from 180px)
Items: Cards with number badges
Active: Cyan left border + darker background

### 8.2 Content Area

**Current Problems**:
- Diagrams with fixed pixel sizes
- Manual spacers for separation
- Dense text blocks

**Proposed Improvements**:

1. **Diagrams**: Responsive sizing based on container
2. **Sections**: Clear heading + content cards
3. **Text**: Improved line height and max-width

```go
// Content section with visual card
func LearnSection(title, content string, visual fyne.CanvasObject) fyne.CanvasObject {
    titleLabel := HeadingText(title)

    contentLabel := widget.NewLabel(content)
    contentLabel.Wrapping = fyne.TextWrapWord

    // Visual in elevated card
    visualCard := widget.NewCard("", "", visual)

    return container.NewVBox(
        titleLabel,
        widget.NewSeparator(),
        container.NewPadded(contentLabel),
        container.NewPadded(visualCard),
    )
}
```

### 8.3 Diagram Improvements

**OpenLane Flow Diagram**:
- Increase box height to 70px (from 65px)
- Add subtle drop shadows
- Animate arrows on first view (CSS-only style)

**Crossbar Diagrams**:
- Scale to container width
- Add zoom control for complex arrays
- Improve legend positioning

---

## 9. Animation and Interaction

### 9.1 Micro-interactions

| Element | Interaction | Effect |
|---------|-------------|--------|
| Buttons | Hover | Background lightens 10% |
| Cards | Hover | Subtle elevation increase |
| Tabs | Switch | Content fade (150ms) |
| Status | Change | Badge pulse animation |
| Log | New entry | Highlight flash |

### 9.2 Loading States

```go
// Button loading state
func (b *ActionButton) SetLoading(loading bool) {
    if loading {
        b.SetText("Generating...")
        b.Disable()
        // Show spinner icon
    } else {
        b.SetText(b.originalText)
        b.Enable()
    }
}
```

### 9.3 Progress Indication

For long operations (Generate All, Validate All):
- Determinate progress bar when steps known
- Status text updates per step
- Elapsed time display

---

## 10. Fyne Implementation Notes

### 10.1 Custom Widget Patterns

```go
// Elevated card with consistent styling
type ElevatedCard struct {
    widget.BaseWidget
    Title   string
    Content fyne.CanvasObject
}

func NewElevatedCard(title string, content fyne.CanvasObject) *ElevatedCard {
    c := &ElevatedCard{Title: title, Content: content}
    c.ExtendBaseWidget(c)
    return c
}

func (c *ElevatedCard) CreateRenderer() fyne.WidgetRenderer {
    bg := canvas.NewRectangle(ColorBG2)
    bg.CornerRadius = CardRadius

    shadow := canvas.NewRectangle(WithAlpha(color.Black, 30))
    shadow.CornerRadius = CardRadius

    titleLabel := SubheadText(c.Title)
    sep := canvas.NewRectangle(ColorSeparator)
    sep.SetMinSize(fyne.NewSize(0, 1))

    inner := container.NewVBox(
        container.NewPadded(titleLabel),
        sep,
        container.NewPadded(c.Content),
    )

    return widget.NewSimpleRenderer(
        container.NewStack(shadow, bg, inner),
    )
}
```

### 10.2 Thread Safety

All UI updates from goroutines MUST use `fyne.Do()`:

```go
go func() {
    result := expensiveOperation()
    fyne.Do(func() {
        statusLabel.SetText(result)
        statusBadge.SetStatus("pass")
    })
}()
```

### 10.3 Responsive Considerations

```go
// Detect width and adjust layout
func ResponsiveConfigLayout(width float32, cellConfig, arrayConfig fyne.CanvasObject) fyne.CanvasObject {
    if width < 800 {
        // Stack vertically on narrow screens
        return container.NewVBox(cellConfig, arrayConfig)
    }
    // Side-by-side on wide screens
    split := container.NewHSplit(cellConfig, arrayConfig)
    split.SetOffset(0.45)
    return split
}
```

### 10.4 Performance Tips

1. **Batch Refresh**: Update multiple properties, call `Refresh()` once
2. **Lazy Loading**: Don't render Learn tab content until selected
3. **Image Caching**: Cache generated layout PNGs
4. **Debounce**: Rate-limit stats updates on entry changes

---

## 11. Priority Matrix

### High Priority (Implement First)

| Item | Effort | Impact | Rationale |
|------|--------|--------|-----------|
| Card-based config panels | Medium | High | Immediate readability improvement |
| Status badges | Low | High | Quick win, high visibility |
| Statistics grid | Low | Medium | Replace pipe separators |
| Code preview styling | Medium | High | Users spend time reading code |

### Medium Priority

| Item | Effort | Impact | Rationale |
|------|--------|--------|-----------|
| Collapsible log | Medium | Medium | Recovers screen space |
| Architecture selector | Medium | Medium | Better toggle UX |
| Layout tab redesign | High | Medium | Complex but valuable |
| Learn sidebar width | Low | Low | Minor improvement |

### Low Priority (Polish Phase)

| Item | Effort | Impact | Rationale |
|------|--------|--------|-----------|
| Micro-animations | Medium | Low | Nice-to-have polish |
| Custom theme colors | Low | Low | Refinement |
| Responsive breakpoints | High | Low | Few mobile users |
| Diagram zoom | High | Low | Advanced feature |

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1)
- [ ] Create `ElevatedCard` widget
- [ ] Create `StatusBadge` widget
- [ ] Create `StatCard` widget
- [ ] Update color constants in theme

### Phase 2: Builder Tab (Week 2)
- [ ] Refactor Cell Config to card layout
- [ ] Refactor Array Config to card layout
- [ ] Replace stats row with grid
- [ ] Add status badges to validation

### Phase 3: Preview & Validation (Week 3)
- [ ] Style code preview areas
- [ ] Redesign Layout tab with primary/thumbnail view
- [ ] Implement collapsible log

### Phase 4: Learn Tab (Week 4)
- [ ] Widen sidebar, improve topic cards
- [ ] Refactor content sections with cards
- [ ] Improve diagram sizing

### Phase 5: Polish (Week 5)
- [ ] Add hover states
- [ ] Implement loading states
- [ ] Fine-tune spacing and alignment
- [ ] User testing and iteration

---

## Appendix: CSS-Equivalent Values

For reference when translating to Fyne:

```css
/* Spacing */
--spacing-xs: 4px;
--spacing-s: 8px;
--spacing-m: 16px;
--spacing-l: 24px;
--spacing-xl: 32px;

/* Border Radius */
--radius-sm: 4px;
--radius-md: 8px;
--radius-lg: 12px;

/* Shadows (simulated in Fyne with offset rectangles) */
--shadow-sm: 0 1px 2px rgba(0,0,0,0.1);
--shadow-md: 0 4px 6px rgba(0,0,0,0.15);
--shadow-lg: 0 10px 15px rgba(0,0,0,0.2);

/* Transitions (Fyne uses animations) */
--transition-fast: 100ms;
--transition-normal: 200ms;
--transition-slow: 300ms;
```

---

## Conclusion

This proposal outlines a comprehensive redesign of the FeCIM EDA Design Suite UI that addresses the identified issues while maintaining the technical functionality. The card-based information architecture, improved status indicators, and better visual hierarchy will create a more professional and usable tool.

Key improvements:
1. **40% reduction in visual density** through proper spacing and grouping
2. **Instant status recognition** via colored badges
3. **Better code readability** with styled preview areas
4. **Reclaimed screen space** through collapsible sections
5. **Professional appearance** aligned with modern EDA tool standards

The phased implementation approach allows for iterative improvement while maintaining a functional application throughout the process.

---

*Document prepared for FeCIM Lattice Tools project*
*Reference: `<local-path>`*
