# Proposed Hysteresis Module Improvements from Open-Source Tools Analysis

**A comprehensive analysis of what FeCIM Lattice Tools can learn from the open-source hysteresis ecosystem**

*Generated: January 2026*

---

## Executive Summary

After analyzing 70+ open-source tools for hysteresis modeling, phase-field simulation, and ferroelectric device characterization, this document proposes **31 specific improvements** to the FeCIM Lattice Tools hysteresis module (module1-hysteresis), organized into:

- **Physics Enhancements** (9 proposals) - More accurate models
- **UI/UX Improvements** (10 proposals) - Better visualization and interaction
- **Integration & Export** (7 proposals) - Interoperability with external tools
- **Appendix** (5 additional proposals) - Advanced/stretch goals

### Current State Assessment

| Aspect | Current Implementation | Rating |
|--------|----------------------|--------|
| Preisach model | Mayergoyz formulation ✅ | Excellent |
| Temperature dependence | Ec(T), Pr(T) scaling ✅ | Good |
| Material library | 8 materials including FeCIM, cryogenic ✅ | Excellent |
| Wake-up/fatigue | Basic modeling ✅ | Good |
| 30-level quantization | Linear discretization ✅ | Good |
| Real-time visualization | 60 FPS Fyne GUI ✅ | Excellent |
| Minor loops | Implicit via hysteron states ✅ | Good |
| **Kinetic models** | KAI implemented but unused ⚠️ | Underutilized |
| **Preisach plane viz** | Data available, not displayed ⚠️ | Missing |
| **Frequency dependence** | Not implemented ❌ | Gap |
| **FORC analysis** | Not implemented ❌ | Gap |
| **Data import/export** | Limited formats ❌ | Gap |

---

## Part 1: Physics Enhancements

### P1. Add Landau-Khalatnikov (LK) Dynamics Model

**Learned from:** FerroX, FERRET, Q-POP-Thermo, landau_khalatnikov_circuit_model_2001.pdf

**What it is:**
The Landau-Khalatnikov equation provides time-dependent polarization evolution based on free energy minimization:

```
η · dP/dt = -∂F/∂P + E
```

Where F is the Landau free energy: `F = α(T)P² + βP⁴ + γP⁶`

**Why it matters:**
- More physics-based than phenomenological Preisach
- Captures frequency-dependent hysteresis loop thinning
- Natural temperature dependence via α(T) = a(T - Tc)
- Used in all serious phase-field simulators

**Implementation sketch:**

```go
// pkg/ferroelectric/landau_khalatnikov.go

type LandauKhalatnikov struct {
    // Landau coefficients (from DFT or literature)
    Alpha  float64 // Linear: α = a(T - Tc)
    Beta   float64 // Cubic term
    Gamma  float64 // Quintic term
    Eta    float64 // Viscosity/damping (Pa·s)

    // State
    Polarization float64
    Temperature  float64
    CurieTemp    float64
}

func (lk *LandauKhalatnikov) FreeEnergyGradient(P float64) float64 {
    // dF/dP = 2αP + 4βP³ + 6γP⁵
    return 2*lk.Alpha*P + 4*lk.Beta*P*P*P + 6*lk.Gamma*P*P*P*P*P
}

func (lk *LandauKhalatnikov) Update(E float64, dt float64) float64 {
    // Euler integration: dP/dt = (1/η)(-dF/dP + E)
    dPdt := (1/lk.Eta) * (-lk.FreeEnergyGradient(lk.Polarization) + E)
    lk.Polarization += dPdt * dt
    return lk.Polarization
}
```

**Expected benefit:**
- Frequency-dependent loops (high f → thinner loops)
- More accurate HfO₂ switching dynamics
- Alternative physics model for comparison

**Complexity:** Medium
**Priority:** High

---

### P2. Implement Nucleation-Limited Switching (NLS) Model for HfO₂

**Learned from:** NLS academic papers (AIP APL 2018), hysteresis-modeling-tools.md Section 3.2

**What it is:**
The NLS model is specifically developed for hafnia-based ferroelectrics and captures:
- Incubation time (dead time before switching starts)
- Stochastic nucleation events
- Field-dependent switching probability

```
t_switch(E) = t_inc(E) + t_growth(E)
t_inc ∝ exp(Ea / (E - Ec))  // Arrhenius-type
```

**Why it matters:**
- HfO₂-specific (our target material)
- More accurate than generic KAI for write latency prediction
- Captures the "nucleation delay" seen in experiments
- Critical for write pulse optimization

**Implementation sketch:**

```go
type NLSModel struct {
    Ec          float64 // Coercive field
    Psat        float64 // Saturation polarization
    TauNucl     float64 // Nucleation time constant
    TauGrowth   float64 // Growth time constant
    ActivationE float64 // Activation energy
}

func (nls *NLSModel) IncubationTime(E float64) float64 {
    if math.Abs(E) <= nls.Ec {
        return math.Inf(1) // No switching below Ec
    }
    Eeff := math.Abs(E) - nls.Ec
    return nls.TauNucl * math.Exp(nls.ActivationE / Eeff)
}

func (nls *NLSModel) SwitchingTime(E float64) float64 {
    return nls.IncubationTime(E) + nls.TauGrowth
}
```

**Expected benefit:**
- Accurate write latency prediction for HfO₂ devices
- Better write pulse optimization
- More realistic cycling simulations

**Complexity:** Low-Medium
**Priority:** High (HfO₂-specific)

---

### P3. Expose KAI Switching Dynamics in Visualization

**Learned from:** hysteresis.research.md Section 4.2, current code analysis

**Current state:**
`SimulateDomainSwitching()` in `preisach_advanced.go` already implements KAI dynamics but is **never called** in the GUI visualization loop.

**What to do:**
Add a new visualization mode "Time-Resolved Switching" that:
1. Shows P(t) during a write pulse
2. Displays domain switching fraction over time
3. Animates hysteron flipping

**Implementation:**

```go
// In gui.go, add new waveform mode
const WaveformTimeResolved WaveformType = 4

// In simulationLoop, when WaveformTimeResolved:
if a.waveform == WaveformTimeResolved {
    Eapplied := 2.0 * a.material.Ec  // Write pulse
    times, pols, switched := a.preisach.SimulateDomainSwitching(
        Eapplied,
        100e-9,  // 100 ns duration
        100,     // 100 time steps
    )
    // Update plot with time-domain data
    a.plot.SetTimeSeriesData(times, pols, switched)
}
```

**Expected benefit:**
- Visualize switching dynamics already computed
- Educational: shows KAI stretched exponential
- Demonstrates nanosecond timescales

**Complexity:** Low (code exists, just needs GUI hookup)
**Priority:** High

---

### P4. Add Preisach Plane Visualization

**Learned from:** python-preisach, Preisachmodel, hysteresis.physics.md

**Current state:**
`GetPreisachPlane()` returns hysteron α, β, and states but this data is **not visualized**.

**What to add:**
A 2D heatmap showing:
- α (vertical) vs β (horizontal) plane
- Color = hysteron state (+1 red, -1 blue)
- Triangle region (valid: α > β)
- "Staircase line" showing current switching boundary

**Implementation:**

```go
// New widget: widgets/preisach_plane.go

type PreisachPlaneWidget struct {
    widget.BaseWidget
    alphas  []float64
    betas   []float64
    states  []int
    weights []float64
}

func (w *PreisachPlaneWidget) Update(p *MayergoyzPreisach) {
    w.alphas, w.betas, w.states = p.GetPreisachPlane()
    w.weights = p.GetDistribution()
    w.Refresh()
}

func (w *PreisachPlaneWidget) CreateRenderer() fyne.WidgetRenderer {
    // Draw 2D scatter plot:
    // - Each hysteron at (β, α)
    // - Color by state: +1=red, -1=blue
    // - Size/opacity by weight
}
```

**Expected benefit:**
- Educational: See how Preisach works internally
- Debug minor loops: Staircase path visible
- Understand distribution shape effects

**Complexity:** Medium
**Priority:** Medium-High

---

### P5. Add Frequency-Dependent Hysteresis

**Learned from:** Landau-Khalatnikov theory, FerroX, PFECAP

**What it is:**
At higher frequencies, hysteresis loops:
- Become thinner (less energy dissipation)
- Have reduced coercivity
- Show phase lag between E and P

**Physics:**
```
Effective Ec(f) ≈ Ec0 × (1 + f/f0)^0.1
Loop area ∝ 1/f (energy dissipation decreases)
```

**Implementation:**

```go
func (m *MayergoyzPreisach) SetFrequency(freq float64) {
    m.frequency = freq
    // Adjust effective Ec based on frequency
    f0 := 1e3 // Reference frequency (1 kHz)
    m.frequencyFactor = 1.0 + math.Pow(freq/f0, 0.1)
    m.initializeHysterons() // Recalculate with scaled Ec
}
```

**GUI addition:**
- Frequency slider (1 Hz to 1 MHz)
- Display loop area vs frequency plot
- Show "AC Hysteresis" mode

**Expected benefit:**
- Realistic behavior at different operating frequencies
- Understand AC vs DC loop differences
- Important for high-speed memory applications

**Complexity:** Medium
**Priority:** Medium

---

### P6. Implement Jiles-Atherton Alternative Model

**Learned from:** JAmodel (MATLAB), pyjam (Python)

**What it is:**
An alternative to Preisach based on differential equations:

```
dM/dH = (M_an - M) / (k × δ - α(M_an - M))
```

Adapted for ferroelectrics (M→P, H→E).

**Why useful:**
- Fewer parameters than Preisach (5 vs 400+ hysterons)
- Faster computation
- Better for control applications
- More physical interpretation of parameters

**Implementation sketch:**

```go
type JilesAtherton struct {
    Ps    float64 // Saturation polarization
    a     float64 // Shape parameter (Langevin)
    k     float64 // Pinning coefficient
    c     float64 // Reversibility
    alpha float64 // Interdomain coupling
}

func (ja *JilesAtherton) AnhystereticP(E float64) float64 {
    // Langevin function: P_an = Ps × coth(E/a) - Ps×a/E
    if math.Abs(E) < 1e-10 {
        return 0
    }
    x := E / ja.a
    return ja.Ps * (1/math.Tanh(x) - 1/x)
}
```

**Expected benefit:**
- Compare two physics approaches in same GUI
- Faster simulation for real-time control
- Educational: Different modeling philosophies

**Complexity:** Medium
**Priority:** Low-Medium

---

### P7. Add First-Order Reversal Curve (FORC) Analysis

**Learned from:** python-preisach, pyhist, Ferro package, hysteresis package

**What it is:**
FORC diagrams reveal the Preisach distribution function μ(α,β) from measured minor loops:

1. Saturate positive
2. Decrease to reversal field Hr
3. Measure P while increasing field H back to saturation
4. Repeat for many Hr values
5. Compute FORC distribution: ρ(Hr, H) = -½ ∂²P/∂Hr∂H

**Why useful:**
- Extract actual Preisach distribution from experiments
- Compare simulated vs measured distributions
- Understand material heterogeneity

**Implementation:**

```go
// Generate FORC dataset
func (m *MayergoyzPreisach) GenerateFORC(HrValues []float64, points int) FORCData {
    forc := FORCData{
        Hr: HrValues,
        H:  make([][]float64, len(HrValues)),
        P:  make([][]float64, len(HrValues)),
    }

    Hsat := 2.0 * m.temperatureCorrectedEc()

    for i, Hr := range HrValues {
        m.Reset()
        // Saturate positive
        for j := 0; j <= 20; j++ {
            m.Update(Hsat * float64(j) / 20)
        }
        // Go to reversal field Hr
        for j := 0; j <= 20; j++ {
            m.Update(Hsat - (Hsat-Hr)*float64(j)/20)
        }
        // Measure ascending branch
        forc.H[i] = make([]float64, points)
        forc.P[i] = make([]float64, points)
        for j := 0; j < points; j++ {
            H := Hr + (Hsat-Hr)*float64(j)/float64(points-1)
            forc.H[i][j] = H
            forc.P[i][j] = m.Update(H)
        }
    }
    return forc
}
```

**GUI addition:**
- "FORC Analysis" mode
- 2D contour plot of ρ(Hr, H)
- Overlay with theoretical Gaussian distribution

**Complexity:** Medium-High
**Priority:** Medium

---

### P8. Improve Fatigue Modeling with Stretched Exponential

**Learned from:** HZO_Wakeup_Fatigue_Mechanisms_arXiv.pdf, material.go

**Current state:**
Basic linear fatigue: `P *= (1 - fatigueRate × cycles)`

**Improvement:**
Literature uses stretched exponential (Kohlrausch-Williams-Watts):

```
Pr(N) = Pr0 × exp(-(N/N0)^β)
```

Where β ≈ 0.3 for HZO (already in material.go EnduranceAtCycles but not used in Preisach).

**Implementation:**

```go
func (m *MayergoyzPreisach) applyFatigue() float64 {
    // Stretched exponential fatigue
    beta := 0.3
    N0 := m.material.EnduranceCycles
    N := float64(m.cycleCount)

    fatigueFactor := math.Exp(-math.Pow(N/N0, beta))
    return fatigueFactor
}
```

**Expected benefit:**
- Match experimental fatigue curves
- Realistic endurance simulation to 10^9+ cycles
- Better retention/degradation predictions

**Complexity:** Low
**Priority:** Medium

---

### P9. Add Imprint Field Modeling

**Learned from:** Ferro package, PFECAP, material.go (ImrintField defined but unused)

**What it is:**
Imprint is the shift of the P-E loop along the E-axis after prolonged DC bias or time:

```
Ec+ ≠ |Ec-|  (asymmetric coercivity)
```

**Implementation:**

```go
type ImrintState struct {
    ImrintField float64  // Accumulated shift (V/m)
    ImrintRate  float64  // Shift rate per second under bias
}

func (m *MayergoyzPreisach) ApplyImrint(biasField float64, duration float64) {
    // Imprint accumulates logarithmically
    m.imprint.ImrintField += m.imprint.ImrintRate *
        math.Log10(1 + duration) * math.Sign(biasField)
}

func (m *MayergoyzPreisach) Update(E float64) float64 {
    // Shift effective field by imprint
    Eeff := E - m.imprint.ImrintField
    // ... rest of hysteron update with Eeff
}
```

**Expected benefit:**
- Understand retention degradation mechanisms
- Model long-term reliability
- Important for analog memory applications

**Complexity:** Low
**Priority:** Low

---

## Part 2: UI/UX Improvements

### U1. Add Temperature Slider with Live Ec/Pr Display

**Learned from:** All phase-field tools have temperature controls

**Current state:**
Temperature is settable via `SetTemperature()` but no GUI control exists.

**Implementation:**

```go
// In createControlsPanel()
tempSlider := widget.NewSlider(4, 700) // 4K to 700K (below Curie)
tempSlider.Value = 300 // Default room temp
tempSlider.OnChanged = func(T float64) {
    a.mu.Lock()
    a.preisach.SetTemperature(T)
    a.mu.Unlock()

    // Update display labels
    fyne.Do(func() {
        a.tempLabel.SetText(fmt.Sprintf("T: %.0f K", T))
        Ec := a.preisach.GetEffectiveEc()
        a.ecLabel.SetText(fmt.Sprintf("Ec(T): %.2f MV/cm", Ec/1e8))
    })
}
```

**Additional features:**
- Show Curie temperature marker on slider
- Display Ec(T) and Pr(T) curves in side panel
- Highlight "Cryogenic" and "Room Temp" presets

**Complexity:** Low
**Priority:** High

---

### U2. Add Multi-Material Overlay Comparison Mode

**Learned from:** Tool comparison matrices in opensource docs

**What it is:**
Display 2-4 P-E loops simultaneously for different materials.

**Implementation:**

```go
// Overlay mode in PEPlot widget
type PEPlot struct {
    // ... existing fields
    overlayLoops []OverlayLoop
    overlayMode  bool
}

type OverlayLoop struct {
    Material *HZOMaterial
    EData    []float64
    PData    []float64
    Color    color.Color
}

func (p *PEPlot) AddOverlay(mat *HZOMaterial, E, P []float64, col color.Color) {
    p.overlayLoops = append(p.overlayLoops, OverlayLoop{mat, E, P, col})
}
```

**GUI:**
- "Compare Materials" button opens overlay mode
- Select up to 4 materials from dropdown
- Color-coded loops with legend
- Show table: Ec, Pr, NumLevels for each

**Expected benefit:**
- Easy comparison of HZO variants
- Educational: See how parameters affect loop shape
- Marketing: Show FeCIM advantage over competitors

**Complexity:** Medium
**Priority:** High

---

### U3. Add Interactive Minor Loop Drawing

**Learned from:** python-preisach educational mode, Ferro package

**What it is:**
Let user draw arbitrary field paths and see resulting minor loops.

**Implementation:**

```go
// Track mouse drag as field path
func (p *PEPlot) onDragged(event *fyne.DragEvent) {
    // Convert screen position to E value
    E := p.screenToField(event.Position.X)

    // Update physics
    P := a.preisach.Update(E)

    // Draw current position on plot
    p.currentPoint = fyne.Position{event.Position.X, p.fieldToScreen(P)}
    p.Refresh()
}
```

**Educational features:**
- "Try drawing a minor loop!" prompt
- Highlight wiping-out property when it happens
- Show turning points on plot

**Complexity:** Medium
**Priority:** Medium

---

### U4. Add Real-Time Metrics Dashboard

**Learned from:** NeuroSim, AIHWKIT dashboards

**What to display:**

| Metric | Source | Update Rate |
|--------|--------|-------------|
| Loop Area (J/m³) | Integrate P dE | Per loop |
| Effective Ec | From zero-crossings | Real-time |
| Effective Pr | From E=0 intercepts | Real-time |
| Squareness (Pr/Ps) | Ratio | Real-time |
| Bits per cell | log₂(NumLevels) | Static |
| Cycle count | From model | Per cycle |
| Wake-up factor | From model | Per cycle |
| Degradation % | From fatigue | Per cycle |

**Implementation:**

```go
type MetricsPanel struct {
    loopArea    *widget.Label
    effectiveEc *widget.Label
    effectivePr *widget.Label
    squareness  *widget.Label
    bitsPerCell *widget.Label
    cycleCount  *widget.Label
    wakeupLabel *widget.Label
    degradation *widget.Label
}

func (m *MetricsPanel) Update(preisach *MayergoyzPreisach, E, P []float64) {
    // Calculate loop area (numerical integration)
    area := 0.0
    for i := 1; i < len(E); i++ {
        area += (P[i] + P[i-1]) / 2 * (E[i] - E[i-1])
    }
    m.loopArea.SetText(fmt.Sprintf("%.2e J/m³", math.Abs(area)))

    // Get model metrics
    cycles, degradation, wakeup := preisach.GetFatigueState()
    m.cycleCount.SetText(fmt.Sprintf("%d", cycles))
    m.degradation.SetText(fmt.Sprintf("%.2f%%", degradation*100))
    m.wakeupLabel.SetText(fmt.Sprintf("%.1f%%", wakeup*100))
}
```

**Complexity:** Low-Medium
**Priority:** Medium-High

---

### U5. Add Domain Wall Visualization (Simplified)

**Learned from:** FerroX, FerroSim domain structure plots

**What it is:**
Show a 1D or 2D representation of domain structure:
- Color by polarization direction (+P red, -P blue)
- Animate switching during write operations

**Simplified 1D approach:**

```go
type DomainStrip struct {
    widget.BaseWidget
    domains []int // +1 or -1 for each segment
    width   int   // Number of domain segments (e.g., 50)
}

func (d *DomainStrip) UpdateFromPreisach(p *MayergoyzPreisach) {
    // Map hysteron states to visual domains
    _, _, states := p.GetPreisachPlane()
    d.domains = make([]int, d.width)

    // Sample hysteron states into domain strip
    step := len(states) / d.width
    for i := 0; i < d.width; i++ {
        d.domains[i] = states[i*step]
    }
}

func (d *DomainStrip) CreateRenderer() fyne.WidgetRenderer {
    // Draw horizontal strip of colored rectangles
    // Red for +1, Blue for -1
    // Animate transitions
}
```

**Expected benefit:**
- Visual intuition for domain switching
- Educational: See partial switching during minor loops
- Engaging visualization

**Complexity:** Medium
**Priority:** Medium

---

### U6. Add Export Screenshot / Animation

**Learned from:** ParaView, matplotlib export features

**What to add:**
- "Export PNG" button for current P-E plot
- "Record GIF" for animated loop sequence
- "Export SVG" for publication-quality vectors

**Implementation:**

```go
func (a *App) exportScreenshot(filename string) error {
    // Capture plot widget to image
    img := a.plot.Snapshot()

    // Save as PNG
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()

    return png.Encode(f, img)
}

func (a *App) recordAnimation(filename string, duration time.Duration) {
    // Capture frames at 30 FPS
    frames := make([]*image.Paletted, 0)
    ticker := time.NewTicker(33 * time.Millisecond)
    defer ticker.Stop()

    start := time.Now()
    for time.Since(start) < duration {
        <-ticker.C
        frame := a.plot.Snapshot()
        frames = append(frames, frame)
    }

    // Encode as GIF
    gif.EncodeAll(f, &gif.GIF{Image: frames, Delay: delays})
}
```

**Complexity:** Medium
**Priority:** Medium

---

### U7. Add Keyboard Shortcuts Panel

**Learned from:** Professional GUI applications

**Current state:**
Some shortcuts exist in `keyboard.go` but not documented in UI.

**What to add:**
- "?" key opens shortcuts overlay
- Show all available shortcuts in semi-transparent panel
- Add more shortcuts for common actions

**Proposed shortcuts:**

| Key | Action |
|-----|--------|
| Space | Pause/Resume |
| R | Reset to initial state |
| 1-5 | Select waveform mode |
| M | Cycle materials |
| T | Toggle temperature slider |
| P | Toggle Preisach plane |
| S | Screenshot |
| ? | Show this help |
| Esc | Close overlays |

**Complexity:** Low
**Priority:** Low-Medium

---

### U8. Add Touch/Mobile Gesture Support

**Learned from:** Modern cross-platform apps, Fyne mobile support

**What to add:**
- Pinch to zoom on P-E plot
- Swipe to change materials
- Long-press for context menu
- Two-finger drag for pan

**Implementation:**
Fyne supports gestures via `fyne.Draggable`, `fyne.Scrollable`, and custom gesture detection.

**Complexity:** Medium
**Priority:** Low (desktop-focused currently)

---

### U9. Add Dark/Light Theme Toggle

**Learned from:** Professional applications, accessibility guidelines

**Current state:**
Custom FeCIM theme exists (`theme.go`) but no toggle.

**What to add:**
- Theme toggle button in corner
- Respect system preference
- Persist preference

**Implementation:**

```go
func (a *App) toggleTheme() {
    if a.darkMode {
        a.fyneApp.Settings().SetTheme(&feCIMLightTheme{})
    } else {
        a.fyneApp.Settings().SetTheme(&feCIMDarkTheme{})
    }
    a.darkMode = !a.darkMode
}
```

**Complexity:** Low
**Priority:** Low

---

### U10. Add Guided Tutorial Mode

**Learned from:** Educational software, onboarding UX

**What it is:**
Step-by-step walkthrough for new users:

1. "This is the P-E hysteresis loop"
2. "Drag the slider to change electric field"
3. "Watch how polarization follows with memory"
4. "Try creating a minor loop by reversing early"
5. "Change materials to see different characteristics"

**Implementation:**

```go
type TutorialOverlay struct {
    steps []TutorialStep
    currentStep int
}

type TutorialStep struct {
    Title       string
    Description string
    Highlight   fyne.CanvasObject // Widget to highlight
    NextTrigger func() bool       // Condition to advance
}

func (t *TutorialOverlay) ShowStep(step int) {
    // Dim everything except highlighted widget
    // Show instruction text
    // Add "Next" / "Skip Tutorial" buttons
}
```

**Complexity:** High
**Priority:** Low (nice-to-have for education)

---

## Part 3: Integration & Export

### I1. Add JSON/CSV Export for P-E Data

**Learned from:** Ferro package, hysteresis package, CrossSim

**What to export:**

```json
{
  "format": "fecim_hysteresis_v1",
  "model": "mayergoyz_preisach",
  "material": "FeCIM HZO",
  "parameters": {
    "Ec": 1.0e8,
    "Ps": 35e-2,
    "Pr": 30e-2,
    "temperature": 300,
    "num_hysterons": 2500,
    "sigma": 0.2
  },
  "loop_data": {
    "E_field_V_m": [0, 1e7, 2e7, ...],
    "polarization_C_m2": [0, 0.05, 0.15, ...]
  },
  "metadata": {
    "generated": "2026-01-28T12:00:00Z",
    "software": "FeCIM Lattice Tools v1.0"
  }
}
```

**CSV format:**
```csv
E_field_V_m,P_C_m2,Level
0,0,15
1e7,0.05,16
...
```

**Implementation:**

```go
func (a *App) ExportJSON(filename string) error {
    data := ExportData{
        Format:     "fecim_hysteresis_v1",
        Model:      "mayergoyz_preisach",
        Material:   a.material.Name,
        Parameters: a.getParametersMap(),
        LoopData: LoopData{
            EField:       a.eHistory,
            Polarization: a.pHistory,
        },
    }
    bytes, _ := json.MarshalIndent(data, "", "  ")
    return os.WriteFile(filename, bytes, 0644)
}
```

**Complexity:** Low
**Priority:** High

---

### I2. Add Import from Experimental P-E Data

**Learned from:** Ferro package, hysteresis package

**What to import:**
- CSV files from P-E tracers (Radiant, Keithley)
- Match format: E (MV/cm), P (µC/cm²)
- Optional metadata: material, temperature, frequency

**Implementation:**

```go
func ImportPEData(filename string) (*ImportedLoop, error) {
    // Read CSV
    f, _ := os.Open(filename)
    reader := csv.NewReader(f)
    records, _ := reader.ReadAll()

    loop := &ImportedLoop{
        EField: make([]float64, len(records)-1),
        P:      make([]float64, len(records)-1),
    }

    for i, record := range records[1:] { // Skip header
        loop.EField[i], _ = strconv.ParseFloat(record[0], 64)
        loop.P[i], _ = strconv.ParseFloat(record[1], 64)
    }

    return loop, nil
}
```

**GUI addition:**
- "Import Data" button
- Overlay imported data on simulated loop
- Auto-fit Preisach parameters to match

**Complexity:** Medium
**Priority:** Medium

---

### I3. Add SPICE Netlist Export

**Learned from:** PFECAP, ngspice, OpenVAF documentation

**What to export:**
Generate a SPICE-compatible ferroelectric capacitor model:

```spice
* FeCIM Ferroelectric Capacitor
* Exported from FeCIM Lattice Tools
* Material: FeCIM HZO

.subckt fecap_fecim p1 p2
  * Preisach-based behavioral model
  * Parameters extracted from GUI
  Bfecap p1 p2 V=Pr*tanh((V(p1,p2)-Imprint)/delta) + eps0*epsR*A/d*V(p1,p2)

  .param Pr=30e-6
  .param delta=0.3e8
  .param epsR=32
  .param A=2.025e-15
  .param d=10e-9
  .param Imprint=0
.ends
```

**Full Verilog-A export (advanced):**

```verilog
`include "disciplines.vams"

module fecap_fecim(p1, p2);
  inout p1, p2;
  electrical p1, p2;

  parameter real Pr = 30e-6;
  parameter real Ps = 35e-6;
  parameter real Ec = 1.0e8;
  parameter real tau = 10e-9;

  real P;  // Polarization state

  analog begin
    P = Pr * tanh((V(p1,p2) * 1e9 / 10) / (Ec * 1e-9));
    I(p1, p2) <+ ddt(P * 100e-12);  // C = P × A
  end
endmodule
```

**Complexity:** Medium
**Priority:** Medium

---

### I4. Add NeuroSim-Compatible Weight Export

**Learned from:** NeuroSim, CrossSim, AIHWKIT

**What to export:**
Conductance matrix for crossbar simulation:

```json
{
  "format": "neurosim_weight_v1",
  "array_size": [32, 32],
  "conductance_min_S": 1e-6,
  "conductance_max_S": 100e-6,
  "num_levels": 30,
  "weights": [
    [15, 22, 8, ...],  // Row 0: Level indices
    [10, 30, 1, ...],  // Row 1
    ...
  ]
}
```

**Implementation:**

```go
func ExportWeightMatrix(weights [][]int, filename string) error {
    export := NeurosimExport{
        Format:   "neurosim_weight_v1",
        ArraySize: [2]int{len(weights), len(weights[0])},
        GMin:     1e-6,
        GMax:     100e-6,
        NumLevels: 30,
        Weights:  weights,
    }
    bytes, _ := json.MarshalIndent(export, "", "  ")
    return os.WriteFile(filename, bytes, 0644)
}
```

**Complexity:** Low
**Priority:** Low-Medium

---

### I5. Add FerroX Material Parameter Export

**Learned from:** FerroX input file format

**What to export:**
FerroX-compatible input deck:

```
# Generated by FeCIM Lattice Tools
# Material: FeCIM HZO

material.eps_r = 32
material.Pr = 30e-6
material.Ec = 1.0e8
material.tau = 10e-9

# Landau coefficients (derived from P-E fit)
material.alpha1 = -1.72e8
material.alpha11 = 7.3e8
material.alpha111 = 2.6e9
```

**Complexity:** Low
**Priority:** Low

---

### I6. Add REST API for External Control

**Learned from:** Modern web services, AIHWKIT remote training

**What to provide:**
Simple HTTP API for:
- GET /status - Current E, P, level
- POST /field - Set electric field
- POST /material - Change material
- GET /loop - Get full P-E loop data
- WebSocket for real-time streaming

**Implementation sketch:**

```go
func (a *App) startAPIServer(port int) {
    http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
        a.mu.RLock()
        status := map[string]interface{}{
            "E": a.electricField,
            "P": a.polarization,
            "level": a.discreteLevel,
        }
        a.mu.RUnlock()
        json.NewEncoder(w).Encode(status)
    })

    http.HandleFunc("/api/field", func(w http.ResponseWriter, r *http.Request) {
        var req struct { E float64 `json:"e"` }
        json.NewDecoder(r.Body).Decode(&req)
        a.mu.Lock()
        a.electricField = req.E
        a.mu.Unlock()
        w.WriteHeader(http.StatusOK)
    })

    go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
```

**Complexity:** Medium
**Priority:** Low (advanced use case)

---

### I7. Add Preisach Distribution Import/Export

**Learned from:** python-preisach, Preisachmodel FORC analysis

**What to export:**
The μ(α, β) distribution:

```csv
alpha,beta,weight
1.2e8,-1.2e8,0.0015
1.15e8,-1.15e8,0.0018
...
```

**What to import:**
Load pre-computed or experimentally-derived distributions:

```go
func (m *MayergoyzPreisach) ImportDistribution(filename string) error {
    // Read CSV
    records := readCSV(filename)

    // Rebuild hysterons with imported weights
    m.hysterons = make([]Hysteron, len(records))
    m.distribution = make([][]float64, len(records))

    for i, r := range records {
        m.hysterons[i] = Hysteron{
            Alpha: r.Alpha,
            Beta:  r.Beta,
            State: -1,
        }
        m.distribution[i] = []float64{r.Weight}
    }

    return nil
}
```

**Expected benefit:**
- Load distributions from FORC measurements
- Share/compare distributions between researchers
- Reproduce published results exactly

**Complexity:** Low-Medium
**Priority:** Medium

---

## Priority Summary

### Immediate (Next Sprint)

| ID | Improvement | Type | Complexity | Impact |
|----|-------------|------|------------|--------|
| U1 | Temperature slider | UI | Low | High |
| P3 | Expose KAI dynamics | Physics | Low | High |
| I1 | JSON/CSV export | Integration | Low | High |
| U2 | Material overlay comparison | UI | Medium | High |
| U4 | Metrics dashboard | UI | Low-Medium | High |

### Short-Term (1-2 Months)

| ID | Improvement | Type | Complexity | Impact |
|----|-------------|------|------------|--------|
| P1 | Landau-Khalatnikov model | Physics | Medium | High |
| P2 | NLS model for HfO₂ | Physics | Medium | High |
| P4 | Preisach plane visualization | Physics/UI | Medium | Medium |
| I2 | Import experimental data | Integration | Medium | Medium |
| U3 | Interactive minor loops | UI | Medium | Medium |

### Medium-Term (3-6 Months)

| ID | Improvement | Type | Complexity | Impact |
|----|-------------|------|------------|--------|
| P5 | Frequency-dependent hysteresis | Physics | Medium | Medium |
| P7 | FORC analysis | Physics | Medium-High | Medium |
| U5 | Domain wall visualization | UI | Medium | Medium |
| I3 | SPICE export | Integration | Medium | Medium |
| U6 | Export screenshots/animation | UI | Medium | Medium |

### Long-Term / Nice-to-Have

| ID | Improvement | Type | Complexity | Impact |
|----|-------------|------|------------|--------|
| P6 | Jiles-Atherton model | Physics | Medium | Low |
| P8 | Stretched exponential fatigue | Physics | Low | Medium |
| P9 | Imprint modeling | Physics | Low | Low |
| U7 | Keyboard shortcuts panel | UI | Low | Low |
| U10 | Guided tutorial | UI | High | Medium |
| I6 | REST API | Integration | Medium | Low |

---

## References

### Primary Sources for This Analysis

1. **hysteresis-modeling-tools.md** - Comprehensive tool catalog (1400+ lines)
2. **hysteresis.opensource.md** - Shorter overview of available tools
3. **ferroelectric-simulation-tools.md** - Phase-field and circuit tools
4. **hysteresis.research.md** - Meta-study of 50+ papers
5. **hysteresis.physics.md** - Physics fundamentals

### Key Open-Source Projects Studied

| Project | Learning | Applied to |
|---------|----------|------------|
| python-preisach | FORC analysis, education | P7, U3 |
| Preisachmodel | Inverse model, fitting | I2, P4 |
| FerroX | GPU phase-field, HfO₂ params | P1, I5 |
| FERRET | Landau-Devonshire | P1 |
| Ferro | Data analysis, PUND | I1, I2 |
| hysteresis (pkg) | Loop metrics | U4 |
| PFECAP | SPICE models | I3 |
| NeuroSim | Array export | I4 |
| AIHWKIT | Hardware-aware | I6 |
| CrossSim | Crossbar integration | I4 |

---

## Conclusion

The FeCIM Lattice Tools hysteresis module already implements **state-of-the-art Preisach modeling** with physics accuracy comparable to academic simulators. However, analysis of 70+ open-source tools reveals opportunities to:

1. **Deepen physics accuracy** - Add Landau-Khalatnikov and NLS models for frequency-dependent and HfO₂-specific behavior
2. **Enhance visualization** - Expose existing capabilities (KAI, Preisach plane) and add new ones (domain walls, FORC)
3. **Improve interoperability** - Enable data exchange with the broader simulation ecosystem

The highest-impact improvements are:
- **Temperature slider** (trivial to add, high educational value)
- **KAI dynamics visualization** (code exists, just needs GUI)
- **Material comparison overlay** (unique selling point for education)
- **JSON/CSV export** (enables ecosystem integration)

These improvements would transform the module from an excellent standalone visualizer into a **bridge between education and professional simulation tools**.

---

## Appendix: Additional Proposals from Deep Analysis

After deeper review of the 1400+ line `hysteresis-modeling-tools.md`, here are additional proposals:

### A1. Add Inverse Preisach for Write Pulse Optimization

**Learned from:** Preisachmodel (fddf22), newton_secant_preisach_control_2024.pdf

**What it is:**
Given a target polarization P_target, compute the required E field sequence:

```go
func (m *MayergoyzPreisach) InverseSolve(targetP float64) float64 {
    // Newton-Raphson iteration to find E that gives targetP
    E := 0.0
    for i := 0; i < 20; i++ {
        P := m.Update(E)
        dP := (m.Update(E+1e4) - P) / 1e4  // Numerical derivative
        E = E - (P-targetP)/dP
    }
    return E
}
```

**Use case:** Optimal write pulse calculation for 30-level programming.

**Complexity:** Medium | **Priority:** Medium

---

### A2. Add PUND Measurement Support

**Learned from:** Ferro package, experimental protocols

**What it is:**
Positive-Up Negative-Down protocol for distinguishing switching vs non-switching charge:

```
P1: +Emax → 0 (switching + capacitive)
U:  +Emax → 0 (capacitive only, already saturated)
N:  -Emax → 0 (switching + capacitive)
D:  -Emax → 0 (capacitive only, already saturated)

Psw = (P1 - U + N - D) / 2
```

**Implementation:**

```go
func (m *MayergoyzPreisach) PUNDMeasurement(Emax float64) PUNDResult {
    m.Reset()
    // P1: saturate positive
    P1 := m.Update(Emax)
    m.Update(0)
    // U: repeat (should be same, no switching)
    U := m.Update(Emax)
    m.Update(0)
    // ... continue for N, D
    return PUNDResult{P1: P1, U: U, N: N, D: D}
}
```

**Use case:** Experimental validation, industry-standard characterization.

**Complexity:** Low | **Priority:** Medium

---

### A3. Add ML Surrogate Model for Speed

**Learned from:** pyjam, arXiv:2511.09976, general ML trends

**What it is:**
Train a neural network to approximate Preisach for 100x faster inference:

```go
type PreisachSurrogate struct {
    weights [][]float64  // Pre-trained from Python TensorFlow
}

func (s *PreisachSurrogate) FastUpdate(E float64, lastP float64) float64 {
    // Simple MLP: input=[E, lastP, dE/dt] → output=P
    input := []float64{E / material.Ec, lastP / material.Ps}
    return s.forward(input) * material.Ps
}
```

**Use case:** Real-time crossbar simulation with thousands of cells.

**Complexity:** High | **Priority:** Low (future enhancement)

---

### A4. Add Negative Capacitance Visualization Mode

**Learned from:** negativec, NC-FET research

**What it is:**
Show the S-curve region where dP/dE < 0 (negative capacitance):

```
Standard FE:        NC region:
    P                  P
    │  /─              │   ╭─╮
    │ /                │  /   ╲  ← S-curve
    │/                 │ /     ↘
    └────E             └────────E
```

**Implementation:** Already captured in Landau free energy - just needs visualization highlighting.

**Use case:** Understanding steep-slope transistors, advanced physics education.

**Complexity:** Medium | **Priority:** Low

---

### A5. Add Phase Diagram Generation

**Learned from:** Q-POP-Thermo, Landau-Devonshire theory

**What it is:**
Generate T-E and T-σ phase diagrams showing ferroelectric/paraelectric boundaries:

```go
func GeneratePhaseDiagram(material *HZOMaterial) PhaseDiagram {
    temps := linspace(100, 700, 50)
    fields := linspace(0, 2*material.Ec, 50)

    diagram := make([][]float64, len(temps))
    for i, T := range temps {
        diagram[i] = make([]float64, len(fields))
        for j, E := range fields {
            model := NewMayergoyzPreisach(material, 30)
            model.SetTemperature(T)
            diagram[i][j] = model.Update(E)
        }
    }
    return diagram
}
```

**Use case:** Understanding operational temperature limits, cryogenic applications.

**Complexity:** Medium | **Priority:** Low

---

## Final Tool Coverage Summary

| Tool Analyzed | Proposals Derived |
|---------------|-------------------|
| python-preisach | P4, P7, U3 |
| Preisachmodel | A1, I2, I7 |
| pyhist | P7 |
| JAmodel/pyjam | P6, A3 |
| FerroX | P1, I5 |
| FERRET | P1 |
| FerroSim | U5 |
| hysteresis (pkg) | U4, I1 |
| Ferro | A2, I2 |
| PFECAP | I3 |
| ngspice/OpenVAF | I3 |
| Heracles | I3 |
| Q-POP-Thermo | A5 |
| negativec | A4 |
| NLS papers | P2 |
| ML Potential papers | A3 |
| KAI model | P3 |
| NeuroSim/CrossSim | I4 |
| AIHWKIT | I6 |

**Total proposals: 31** (26 main + 5 appendix)

---

*Document generated from comprehensive analysis of:*
- `<local-path>` (1420 lines)
- `<local-path>` (1222 lines)
- `<local-path>` (663 lines)
- `<local-path>` (553 lines)
- `<local-path>` (480 lines)
- `<local-path>` (implementation)
- `<local-path>` (UI)
