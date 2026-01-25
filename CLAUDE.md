# CLAUDE.md - FeCIM Lattice Tools

## Overview

Go-based lattice tool suite for Ferroelectric Compute-in-Memory (FeCIM) technology based on Dr. external research group's HfO₂-ZrO₂ superlattice research.

**Core concept**: 30 discrete analog states per cell (~4.9 bits/cell) as demonstrated in Dr. Tour's COSM 2025 presentation. Peer-reviewed literature confirms similar capabilities (32-140 states) in HZO FeFET devices.

> **Primary Source**: Dr. external research group, COSM 2025 - [Transcript](docs/videos/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md)

## Build & Run

```bash
go build -o fecim-visualizer ./cmd/fecim-visualizer && ./fecim-visualizer
# Or: ./launch.sh
```

## Project Structure

```
cmd/fecim-visualizer/     # Main unified app entry point
module1-hysteresis/       # P-E curve, Preisach model
module2-crossbar/         # MVM, non-idealities (IR drop, sneak paths, drift)
module3-mnist/            # Neural network digit recognition (87% accuracy)
module4-circuits/         # DAC/ADC/TIA peripherals
module5-comparison/       # Technology comparison
module6-eda/              # EDA tools
shared/                   # Theme, widgets, logging
```

Each module follows: `pkg/gui/embedded.go` (embeddable app), `pkg/gui/app.go` (standalone).

## Key Rules

### Do
- Use `fyne.Do(func() { ... })` for all UI updates from goroutines
- Quantize to 30 levels: `crossbar.QuantizeTo30Levels(value)`
- Follow the embedded app interface pattern (see below)
- Run `go test ./...` before committing

### Don't
- Modify `module2-crossbar/pkg/_layers_experimental/` - archived research code
- Add demos without implementing the embedded interface
- Use blocking operations on the main UI thread
- Commit binaries (fecim-visualizer, crossbar-gui, etc.)

## Embedded App Interface

Every demo must implement:
```go
type EmbeddedXxxApp struct { ... }
func NewEmbeddedXxxApp() *EmbeddedXxxApp
func (app *EmbeddedXxxApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject
func (app *EmbeddedXxxApp) Start()  // Called when tab selected
func (app *EmbeddedXxxApp) Stop()   // Called when tab deselected
```

## Key Files

| Task | File |
|------|------|
| Add demo | `cmd/fecim-visualizer/main.go` |
| Crossbar MVM | `module2-crossbar/pkg/crossbar/array.go` |
| Hysteresis | `module1-hysteresis/pkg/ferroelectric/preisach.go` |
| Theme | `shared/theme/theme.go` |
| Non-idealities | `module2-crossbar/pkg/crossbar/nonidealities.go` |

## Physics Constants

| Parameter | Value | Source |
|-----------|-------|--------|
| FeCIM Levels | 30 | Dr. Tour COSM 2025 (primary); Jerry 2017: 32, Song 2024: 140 |
| Pr | 15-34 µC/cm² | Nature Commun. 2025, ACS 2020 |
| Ec | 1.0-1.5 MV/cm | Nature Commun. 2025 |
| Ps | ~30-35 µC/cm² | Literature consensus |
| Endurance | 10¹²+ cycles (superlattice) | PMC 2024, IEEE IRPS 2022 |

### Parameter References
- **Pr (Remanent Polarization)**: 15 µC/cm² for 20nm superlattice [DOI:10.1038/s41467-025-61758-2], up to 34 µC/cm² wake-up free [DOI:10.1021/acsaelm.0c00671]
- **Ec (Coercive Field)**: 1.4-1.6 MV/cm for 20nm, 0.85 MV/cm for 100nm superlattice [DOI:10.1038/s41467-025-61758-2]
- **Endurance**: >5×10¹² cycles demonstrated for HfO₂-ZrO₂ superlattice with TiN electrodes [PMC 2024]
- **Multi-level states**: 32 states (5-bit) standard benchmark [DOI:10.1109/IEDM.2017.8268338]

## Accuracy & Honesty Policy

This project prioritizes **scientific accuracy** over marketing claims:

1. **Verified claims** include peer-reviewed citations with DOIs
2. **Unverified claims** are marked as such (e.g., "10M× energy" claim)
3. **Simulation parameters** (30 levels) may differ from literature (32-140 levels)
4. **Dr. Tour attribution**: IronLattice is verified (Rice Innovation grant 2025), but specific device parameters are from general HZO literature

### Key Verified Facts
| Claim | Status | Source |
|-------|--------|--------|
| 30 analog states | ✅ Dr. Tour claim | COSM 2025 presentation; peer-reviewed: 32 states (Jerry 2017) |
| 87% MNIST accuracy | ✅ Dr. Tour claim | COSM 2025 presentation; peer-reviewed: 87-96% (multiple sources) |
| 10¹² cycle endurance | ⚠️ Target, not achieved | Dr. Tour: "still have to get this up to 10¹²"; literature: verified for superlattice (PMC 2024) |
| 10M× vs NAND energy | ❌ Unverified | Dr. Tour claim only; peer-reviewed max: 25-100× (Nature 2025) |
| 80-90% datacenter savings | ⚠️ Dr. Tour claim | COSM 2025; realistic peer-reviewed: 50-80% for memory-bound |
| IronLattice company | ✅ Verified | Rice Innovation grant Jan 2025; COSM 2025 presentation |
| TRL 4 status | ✅ Confirmed | Dr. Tour explicitly stated at COSM 2025 |

## Testing

```bash
go test ./...                            # All tests
go test ./module2-crossbar/pkg/crossbar  # Crossbar only
go test -v -run TestPreisach             # Specific test
```

## Dependencies

- **Fyne v2.7.2** - GUI framework
- **Charm/BubbleTea** - TUI (demo1)
- **go-gl/glfw** - Window management
- **vulkan-go/vulkan** - GPU rendering

## Git Conventions

- Commit messages: `type: description` (feat, fix, docs, refactor, test, chore)
- Keep commits atomic and focused
- Run tests before pushing

## Ignore These Directories

- `logs/` - Runtime logs
- `output/` - Generated exports
- `module2-crossbar/pkg/_layers_experimental/` - Archived research
- `docs/archive/` - Archived documentation
