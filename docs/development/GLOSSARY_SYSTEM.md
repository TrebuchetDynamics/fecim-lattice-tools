# Glossary System Documentation

## Overview

The global glossary system provides a reusable, searchable technical glossary widget that can be embedded in all modules. It includes 23+ technical terms organized into 4 categories and 9+ key scientific references.

**Location**: `shared/widgets/glossary.go`

## Features

### GlossaryWidget

Searchable, categorized glossary with:
- **24 technical terms** covering FeCIM physics, architecture, circuits, and metrics
- **4 categories**: Physics, Architecture, Circuits, Metrics
- **Search functionality**: Real-time filtering by term or definition
- **Expandable definitions**: Click term to see full explanation
- **Compact/expanded modes**: Sidebar widget or full-screen dialog

### ReferencesWidget

Key scientific papers with clickable DOI links:
- Nature Communications 2023 (96.6% MNIST)
- ScienceDirect 2025 (98.24% MNIST)
- Nano Letters 2024 (10¹² endurance)
- CEA-Leti 2024 (22nm BEOL)
- Links to HONESTY_AUDIT.md for verification status

## Usage Examples

### Method 1: Help Menu Integration

Add standardized Help menu to any module:

```go
import "fecim-lattice-tools/shared/widgets"

func (g *GUI) BuildContent() fyne.CanvasObject {
    helpMenu := fyne.NewMenu("Help",
        widgets.CreateHelpMenuItems(g.window)...,
    )
    mainMenu := fyne.NewMainMenu(
        // ... other menus ...
        helpMenu,
    )
    g.window.SetMainMenu(mainMenu)

    // ... rest of content ...
}
```

This adds:
- Technical Glossary
- Key References
- Separator
- About dialog

### Method 2: Toolbar Buttons

Add glossary buttons to toolbar:

```go
toolbar := container.NewHBox(
    // ... existing buttons ...
    widget.NewSeparator(),
    widgets.CreateGlossaryButton(window),
    widgets.CreateReferencesButton(window),
)
```

### Method 3: Embedded Widget

Embed full glossary in layout:

```go
glossary := widgets.NewGlossaryWidget()

content := container.NewBorder(
    topBar,
    bottomBar,
    glossary, // Left sidebar
    nil,
    mainContent,
)
```

### Method 4: Popup Specific Term

Show definition for specific term:

```go
helpBtn := widget.NewButton("What is FeCIM?", func() {
    widgets.ShowGlossary("FeCIM", window)
})
```

### Method 5: Programmatic Lookup

Access definitions programmatically:

```go
def := widgets.QuickTermLookup("Ec")
if def != "" {
    fmt.Printf("Ec: %s\n", def)
}

// Get all terms in category
physicsTerms := widgets.GetTermsByCategory("Physics")
```

## API Reference

### Widget Functions

```go
// Create glossary widget
NewGlossaryWidget() *GlossaryWidget

// Create references widget
NewReferencesWidget() *ReferencesWidget

// Show dialogs
ShowFullGlossary(parent fyne.Window)
ShowReferences(parent fyne.Window)
ShowGlossary(term string, parent fyne.Window)

// Create toolbar buttons
CreateGlossaryButton(parent fyne.Window) *widget.Button
CreateReferencesButton(parent fyne.Window) *widget.Button

// Create help menu items
CreateHelpMenuItems(parent fyne.Window) []*fyne.MenuItem
```

### Data Access Functions

```go
// Lookup term (case-insensitive)
QuickTermLookup(term string) string

// Get terms by category
GetTermsByCategory(category string) []GlossaryEntry

// Get all categories
GetCategories() []string
```

### Data Structures

```go
type GlossaryEntry struct {
    Term       string  // e.g., "FeCIM"
    Definition string  // Full explanation
    Category   string  // Physics/Architecture/Circuits/Metrics
}

type ReferenceEntry struct {
    Title    string
    Citation string
    DOI      string  // Empty if not applicable
    URL      string  // DOI link or local path
}
```

## Terms Included

### Physics (7 terms)
- **FeCIM**: Ferroelectric Compute-in-Memory
- **Ec**: Coercive Field (0.6-1.5 MV/cm)
- **Pr**: Remnant Polarization (15-34 µC/cm² RT, 75 µC/cm² at 4K)
- **HZO**: Hafnium Zirconium Oxide superlattice
- **Hysteresis Loop**: P-E curve behavior
- **Preisach Model**: Mathematical hysteresis model
- **Endurance**: 10⁹-10¹² write/erase cycles

### Architecture (5 terms)
- **1T1R**: One Transistor, One Resistor
- **MVM**: Matrix-Vector Multiplication
- **MAC**: Multiply-Accumulate operation
- **BEOL**: Back-End-Of-Line integration
- **Sneak Path**: Unintended current flow
- **IR Drop**: Voltage loss in large arrays

### Circuits (4 terms)
- **DAC**: Digital-to-Analog Converter
- **ADC**: Analog-to-Digital Converter
- **TIA**: Transimpedance Amplifier
- **Sense Amplifier**: High-gain differential amplifier

### Metrics (6 terms)
- **TRL**: Technology Readiness Level (1-9 scale)
- **TOPS/W**: Tera-Operations Per Second per Watt
- **Bits per Cell**: 4.9 bits (30-level baseline; simulation baseline), up to 7.1 bits (140 states)
- **MNIST Accuracy**: 96.6%-98.24% (vs 99.7% software)
- **Retention Time**: >10 years at 85°C
- **Write Energy**: 1-10 fJ/bit (vs 100 fJ/bit NAND)

## References Included

1. **Nature Communications 2023** - 96.6% MNIST accuracy
2. **ScienceDirect 2025** - 98.24% MNIST with FTJ reservoir
3. **Nano Letters 2024** - 10¹² cycle endurance (V:HfO₂)
4. **Nature 2025** - 512-layer 3D FeFET
5. **CEA-Leti 2024** - 22nm BEOL demonstration
6. **Fraunhofer IPMS 2024** - Automotive grade (AEC-Q100)
7. **IEEE Transactions 2024** - Cryogenic operation (5K-300K)
8. **HONESTY_AUDIT.md** - Scientific verification status
9. **Dr. Tour COSM 2025** - Conference transcript (unverified)

## Testing

Run glossary tests:

```bash
go test ./shared/widgets/ -v -run="Glossary|References|Terms"
```

**Test coverage**:
- ✅ 24 terms with complete definitions
- ✅ All 4 categories populated
- ✅ Case-insensitive term lookup
- ✅ Critical terms present (FeCIM, Ec, Pr, MAC, etc.)
- ✅ Reference structure validation
- ⏭️ Widget creation (requires Fyne app context)

## Adding New Terms

To add a new term:

1. Edit `shared/widgets/glossary.go`
2. Add to `TermsData` slice:

```go
{
    Term:       "NewTerm",
    Definition: "Clear, concise definition with context and typical values.",
    Category:   "Physics", // or Architecture, Circuits, Metrics
},
```

3. Run tests: `go test ./shared/widgets/ -v`
4. Verify no duplicates or missing fields

## Integration Checklist

When adding glossary to a new module:

- [ ] Add Help menu with `CreateHelpMenuItems()`
- [ ] Or add toolbar buttons with `CreateGlossaryButton()` / `CreateReferencesButton()`
- [ ] Test popup functionality (`ShowGlossary("FeCIM", window)`)
- [ ] Verify window parent is passed correctly
- [ ] Test on module startup (no crashes)

## Design Principles

1. **Centralized**: Single source of truth for all technical terms
2. **Searchable**: Real-time filtering for quick access
3. **Categorized**: Logical grouping (Physics, Architecture, Circuits, Metrics)
4. **Reusable**: Drop-in widget for any module
5. **Verifiable**: Links to HONESTY_AUDIT.md for scientific rigor
6. **Accessible**: Multiple access methods (menu, toolbar, programmatic)

## File Structure

```
shared/widgets/
├── glossary.go              # Main widget implementation
├── glossary_test.go         # Comprehensive tests (15 test cases)
├── glossary_example.go      # Integration examples
├── color_legend.go          # Existing widget (reference)
└── color_legend_test.go     # Existing tests
```

## Future Enhancements

Potential additions (not yet implemented):

- [ ] Export glossary to PDF/Markdown
- [ ] Add images/diagrams to definitions
- [ ] Cross-reference terms (clickable links within definitions)
- [ ] Add equations to physics terms (e.g., Ohm's law for MAC)
- [ ] Multi-language support
- [ ] History/recently viewed terms
- [ ] Favorites/bookmarks
- [ ] Integration with online documentation

## See Also

- `CLAUDE.md` - Project instructions with key physics constants
- `docs/cim/HONESTY_AUDIT.md` - Scientific verification status
- `docs/cim/physics.md` - Detailed physics documentation
- `docs/development/scriptReference.md` - Function reference guide
