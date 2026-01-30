# Module 1: Hysteresis - Features

P-E Curve Simulator for Ferroelectric Memory

---

## Features

- **Interactive P-E Loop Visualization** - Real-time hysteresis curves with animated polarization switching
- **Multi-Level Analog Memory Demo** - Program/read discrete FeCIM states with ISPP write verification (level count per material)
- **Multiple Run Modes** - Fyne GUI, TUI, headless ASCII, Vulkan graphics
- **8 Material Library** - HZO variants, AlScN, cryogenic, FTJ (140 states)
- **Waveform Control** - Sine, triangle, square, manual slider
- **Educational Slides** - Built-in ferroelectric physics explanations
- **Calibration System** - Temperature-aware multi-level calibration (233-423 K)
- **Phase State Machine** - 5-phase indicator (RESET | SETTLE | WRITE | READ | VERIFY)
- **Stability Indicator** - Color-coded level stability warnings

## Physics Models

| Model | Description |
|-------|-------------|
| **Preisach (Basic)** | Hyperbolic tangent switching with history-dependent minor loops |
| **Mayergoyz Preisach** | Full 40×40 hysteron grid, bivariate Gaussian distribution |
| **KAI Switching** | Kolmogorov-Avrami-Ishibashi time-resolved domain switching |
| **Temperature Effects** | Curie-Weiss law for Ec(T), Pr(T), Arrhenius for τ(T) |
| **Fatigue/Wake-up** | Stretched exponential endurance degradation |

## Key Parameters

| Parameter | Value | Notes |
|-----------|-------|-------|
| FeCIM Levels | 8-140 | Material-dependent (8-140 states) |
| Pr (RT) | 15-34 µC/cm² | Material-dependent |
| Pr (4K) | 75 µC/cm² | Cryogenic enhanced |
| Ec | 0.6-5.0 MV/cm | Material-dependent |
| Switching τ | 1-20 ns | Temperature-dependent |
| Endurance | 10⁸-10¹² cycles | Material-dependent |
| Retention | 10-100 years @ 85°C | Arrhenius model |

## Materials Available

| Material | Levels | Notes |
|----------|--------|-------|
| HZO (Si-doped) | 30 | Baseline |
| FeCIM HZO | 30 | Dr. Tour specs |
| Literature Superlattice | 64 | Cheema 2020 |
| Cryogenic HZO | 30 | 75 µC/cm² at 4K |
| HZO Standard 32 | 32 | Oh 2017 |
| HZO FTJ 140 | 140 | Song 2024 |
| AlScN | 8-16 | High Pr (120 µC/cm²) |
