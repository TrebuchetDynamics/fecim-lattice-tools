# Module 1: Hysteresis - Features

P-E curve simulator for ferroelectric memory physics.

---

## Physics Engines

- **Preisach (quasi-static)** - Mayergoyz-based hysteresis with memory and discrete states.
- **Landau-Khalatnikov (dynamic)** - Time-resolved switching engine for educational visualization.

> Note: The Landau engine is intended for interactive learning, not calibrated device modeling.

---

## Features

- **Interactive P-E Loop** - Real-time hysteresis curve with polarization animation
- **Discrete Level Programming** - Write/Read/Verify state machine with ISPP-style calibration
- **Material Presets** - HZO baseline, FeCIM baseline, superlattice, cryogenic HZO, 32-level HZO, 140-level FTJ, AlScN (all presets are illustrative)
- **Temperature Control** - Temperature slider with calibration cache (range configurable in code)
- **Waveform Modes** - Manual, sine, triangle, write/read demo, time-resolved switching
- **Multi-Mode UI** - Fyne GUI, TUI, headless ASCII

---

## Materials (From `shared/physics`)

| Material | Levels | Notes |
|---|---:|---|
| HZO (Si-doped) | 30 | Baseline demo preset (configurable) |
| FeCIM HZO | 30 | Simulation baseline (configurable) |
| Literature Superlattice | 64 | Preset (illustrative) |
| Cryogenic HZO | 30 | Preset (illustrative) |
| HZO Standard 32 | 32 | Preset (illustrative) |
| HZO FTJ 140 | 140 | Preset (illustrative) |
| AlScN | 8-16 | Preset (illustrative) |

---

## GUI Components

- **P-E Hysteresis Plot** - Polarization vs field
- **Level Indicator** - Discrete level gauge
- **Phase Indicator** - RESET -> SETTLE -> WRITE -> READ -> VERIFY
- **Material Picker** - Searchable list with property tables
- **Calibration Status** - Per-temperature calibration state and interpolation

---

## Export

- JSON export with metadata (material, temperature, parameters)
- CSV export for data analysis
- Debug logs for calibration and write/verify steps
