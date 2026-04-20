<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# Module 5: Technology Comparison Visualizer

## Purpose

Module 5 provides interactive side-by-side technology comparison views showing FeCIM advantages versus SRAM, RRAM, and other architectures. It combines architecture analysis with market opportunity visualization, including energy efficiency comparisons, data center cost models, and competitive positioning.

Key features:
- Hero visualizations: energy race, market opportunity, competitive matrix
- Interactive scenario profiles and fabrication reality analysis
- Data center calculator with evidence-based energy models
- Export capabilities (CSV, JSON, images)

## Key Files

### Core Comparison Engine
- `pkg/comparison/architecture.go` - Technology architecture definitions and parameters
- `pkg/comparison/render.go` - Rendering logic for comparison visuals
- `pkg/comparison/comparison_test.go` - Unit tests for comparison calculations
- `pkg/comparison/advantage_test.go`, `sanity_test.go`, `comparison_additional_test.go` - Extended test coverage

### GUI Components
- `pkg/gui/app.go` - Main ComparisonApp, window setup, animation state management
- `pkg/gui/hero.go` - Hero visualizations (AnimatedEnergyRace, MarketOpportunityChart, CompetitiveMatrix, PhasedStrategyDiagram)
- `pkg/gui/market.go` - MarketOpportunityChart and data center cost calculator
- `pkg/gui/liveslide.go` - Interactive presentation mode with auto-demo phases
- `pkg/gui/widgets.go` - Reusable GUI widgets for charts and comparisons
- `pkg/gui/embedded.go` - EmbeddedComparisonApp interface for unified app integration

### Export & Data
- `pkg/gui/export.go` - CSV and JSON export functionality
- `pkg/gui/evidence_model.go` - Evidence provenance and confidence tracking
- `pkg/gui/scenario_profiles.go` - Scenario profile definitions (efficiency-focused, balanced, performance-focused)
- `pkg/gui/fabrication_reality.go` - Real-world fabrication constraints and yield models

### UI/UX Features
- `pkg/gui/keyboard.go` - Keyboard shortcuts and navigation
- `pkg/gui/calculations_test.go` - Calculation verification tests
- Data confidence interval, sensitivity, provenance, and diff tests (data_*.go)
- UX mode tests: dual-mode, evidence-first, plain-text evidence

## Subdirectories

```
module5-comparison/
├── pkg/
│   ├── comparison/          # Architecture analysis and comparison logic
│   └── gui/                 # Fyne GUI components and application
├── README.md
└── AGENTS.md               # This file
```

## For AI Agents

### Working

**Current State:**
- Comparison engine stable with multiple architecture types (CPU, GPU, FeCIM, RRAM, SRAM)
- Hero visualizations operational with animation support
- Evidence-based data model prevents unsubstantiated claims
- Market opportunity calculator with configurable scenarios
- Export to CSV/JSON for further analysis

**Task Pattern:**
1. Read the energy model sources in `app.go` (constants like `cpuEnergyPJPerMAC`)
2. Verify any new comparison data against `docs/4-research/honesty-audit.md`
3. Test changes with `go test ./module5-comparison/...`
4. Check hero visualization rendering with visual tests

**Key Patterns:**
- Animation state protected by `animMu` (RWMutex) in ComparisonApp
- All energy comparisons sourced from documented references
- Confidence intervals tracked per data point (avoid unverified claims)
- Market calculator uses configurable energy specs with source tracking

### Testing

**Test Files:**
- `pkg/comparison/comparison_test.go` - Core comparison calculations
- `pkg/comparison/advantage_test.go` - Technology advantage logic
- `pkg/comparison/sanity_test.go` - Data validation and bounds checking
- `pkg/gui/market_test.go`, `hero_test.go` - GUI component tests
- `pkg/gui/*_test.go` - Data model and UX tests

**Run Tests:**
```bash
go test ./module5-comparison/...              # All tests
go test -v ./module5-comparison/pkg/comparison  # Comparison engine only
go test -v ./module5-comparison/pkg/gui       # GUI tests only
```

**Coverage Notes:**
- Comparison logic: well-tested with unit tests
- Market calculator: tested with multiple scenarios
- GUI rendering: tested with mock data
- Evidence tracking: validated to prevent claim inflation

### Patterns

**Architecture Comparison:**
- All architectures defined in `architecture.go` with parameters (power, area, bandwidth)
- Energy model follows pattern: `EnergySpec` struct with Name, EnergyFJ, Source, Verified flag
- Comparison uses `ComparisonMetrics` to normalize across scales

**Data Integrity:**
- Evidence model (`evidence_model.go`) prevents unsubstantiated visualizations
- Verified flag distinguishes measured data from model inputs
- Export includes provenance metadata (source, timestamp, author)

**GUI State Management:**
- ComparisonApp holds all UI components and animation state
- Animations protected by mutex for thread safety
- Presentation mode cycles through phases (INTRO, ENERGY, MARKET, COMPETITIVE, STRATEGY, FINISH)

**Export Strategy:**
- CSV format for spreadsheet analysis
- JSON with full metadata for reproducibility
- Image export via canvas capture

## Dependencies

**Internal:**
- `shared/logging` - Logging infrastructure
- `shared/theme` - Fyne theme and styling
- `shared/widgets` - Reusable Fyne widgets
- `shared/export` - Shared export utilities

**External:**
- `fyne.io/fyne/v2` - GUI framework
- Standard Go packages (fmt, sync, time, math)

## MANUAL

### Adding a New Technology Comparison

1. **Define Architecture** in `pkg/comparison/architecture.go`:
   ```go
   const TechMyNewArchitecture = "my-arch"
   // Add to ArchitectureParams
   ```

2. **Add Energy Spec** with source reference:
   ```go
   mySpec := EnergySpec{
       Name:          "MyArch",
       EnergyFJ:      500.0,  // femtojoules/MAC
       Source:        "Journal/Conf Name",
       Verified:      false,  // Only true for peer-reviewed measurements
       SourceDetails: "Full citation or reference",
   }
   ```

3. **Test Comparison** with existing tests
4. **Update Visualizations** (hero.go, market.go) if needed
5. **Document** in README.md with source reference

### Evidence Model Rules

- Every comparison must have a Source field
- Verified=true only for peer-reviewed, reproducible measurements
- Model inputs (simulations) use Verified=false
- Export always includes provenance metadata
- Never claim precision beyond measured data supports

### Running Presentations

- Activate via ComparisonApp.SetPresentationMode()
- Auto-demo cycles through phases with configurable duration
- Manual phase navigation via keyboard shortcuts (see keyboard.go)
- Keyboard shortcut: Space for next phase, P for presentation toggle

### Customizing Scenarios

Edit `pkg/gui/scenario_profiles.go`:
- Define new ScenarioProfile with energy specs, power budget, use-case description
- Three pre-defined scenarios: Efficiency, Balanced, Performance
- Profiles used by market calculator and liveslide visualization

### Debugging Comparison Logic

Enable debug logging via `FECIM_DEBUG=comparison` environment variable:
```bash
FECIM_DEBUG=comparison ./fecim-lattice-tools
```

This enables detailed logging in ComparisonApp initialization and animation state.

---

**Last Updated:** 2026-02-13
