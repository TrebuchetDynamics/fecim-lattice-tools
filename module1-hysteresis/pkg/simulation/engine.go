// Package simulation provides the time-stepping simulation engine.
package simulation

import (
	"math"
	"sync"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/logging"
)

// Package-level logger
var log *logging.Logger

// Engine default constants
const (
	// defaultDt is the default simulation timestep (1 ns).
	defaultDt = 1e-9

	// defaultFrequencyHz is the default waveform frequency (1 MHz).
	defaultFrequencyHz = 1e6

	// defaultMaxHistory is the number of simulation steps retained for plotting.
	defaultMaxHistory = 1000

	// amplitudeEcMultiplier is the factor applied to coercive voltage for the
	// default waveform amplitude (2x ensures full saturation on each half-cycle).
	amplitudeEcMultiplier = 2.0
)

func init() {
	log = logging.NewLogger("hysteresis-sim")
}

// State represents the current simulation state.
type State struct {
	Time          float64 // Simulation time (s)
	Voltage       float64 // Applied voltage (V)
	ElectricField float64 // Electric field (V/m)
	Polarization  float64 // Current polarization (C/m²)
	NormPol       float64 // Normalized polarization (-1 to +1)

	// History for plotting
	VoltageHistory []float64
	PolHistory     []float64
	MaxHistory     int
}

// Engine manages the ferroelectric simulation.
type Engine struct {
	model    *ferroelectric.PreisachModel
	material *ferroelectric.HZOMaterial
	state    *State

	// Simulation parameters
	dt float64 // Time step (s)

	// Thread-safe state (protected by mu)
	mu      sync.RWMutex
	running bool
	paused  bool

	// Waveform generation
	waveform  WaveformType
	frequency float64 // Hz
	amplitude float64 // V
}

// WaveformType defines the input voltage waveform.
type WaveformType int

const (
	WaveformSine WaveformType = iota
	WaveformTriangle
	WaveformSquare
	WaveformManual
)

// NewEngine creates a new simulation engine.
func NewEngine(material *ferroelectric.HZOMaterial) *Engine {
	e := newInertEngine()
	materialSnapshot := snapshotMaterial(material)
	if materialSnapshot == nil || !isValidMaterialThickness(materialSnapshot.Thickness) {
		return e
	}

	model := ferroelectric.NewPreisachModel(materialSnapshot)
	if model == nil {
		return e
	}

	e.model = model
	e.material = materialSnapshot
	e.amplitude = materialSnapshot.CoerciveVoltage() * amplitudeEcMultiplier

	log.Debug("NewEngine: material=%s, dt=%.0f ns, freq=%.0f MHz, amplitude=%.2f V",
		materialSnapshot.Name, e.dt*1e9, e.frequency/1e6, e.amplitude)

	return e
}

func newInertEngine() *Engine {
	return &Engine{
		state:     newState(defaultMaxHistory),
		dt:        defaultDt,
		waveform:  WaveformSine,
		frequency: defaultFrequencyHz,
	}
}

func snapshotMaterial(material *ferroelectric.HZOMaterial) *ferroelectric.HZOMaterial {
	if material == nil {
		return nil
	}
	snapshot := *material
	return &snapshot
}

func newState(maxHistory int) *State {
	return &State{
		VoltageHistory: make([]float64, 0, maxHistory),
		PolHistory:     make([]float64, 0, maxHistory),
		MaxHistory:     maxHistory,
	}
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func isValidMaterialThickness(thickness float64) bool {
	return thickness > 0 && isFinite(thickness)
}

func isValidWaveformFrequency(frequency float64) bool {
	return frequency >= 0 && isFinite(frequency) && isFinite(2*math.Pi*frequency)
}

func isValidWaveformAmplitude(amplitude float64) bool {
	return amplitude >= 0 && isFinite(amplitude)
}

func isRepresentableField(voltage, thickness float64) bool {
	return isFinite(voltage) && isValidMaterialThickness(thickness) && isFinite(voltage/thickness)
}

func isValidWaveformType(waveform WaveformType) bool {
	switch waveform {
	case WaveformSine, WaveformTriangle, WaveformSquare, WaveformManual:
		return true
	default:
		return false
	}
}

func realtimeFrameInterval(targetFPS int) time.Duration {
	if targetFPS <= 0 {
		return 0
	}
	interval := time.Second / time.Duration(targetFPS)
	if interval <= 0 {
		return 0
	}
	return interval
}

// Start begins the simulation loop.
func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running = true
	e.paused = false
	log.Info("Simulation started")
}

// Stop halts the simulation.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running = false
	log.Info("Simulation stopped")
}

// Pause toggles the paused state.
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.paused = !e.paused
	if e.paused {
		log.Debug("Simulation paused")
	} else {
		log.Debug("Simulation resumed")
	}
}

// IsPaused returns true if simulation is paused.
func (e *Engine) IsPaused() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.paused
}

// IsRunning returns true if simulation is running.
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// Step advances the simulation by one time step.
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) Step() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running || e.paused {
		return
	}
	if e.material == nil || e.model == nil || !isValidMaterialThickness(e.material.Thickness) {
		e.state.ElectricField = 0
		return
	}

	// Generate voltage based on waveform
	e.state.Voltage = e.generateVoltage(e.state.Time)

	// Convert voltage to electric field.
	e.state.ElectricField = e.state.Voltage / e.material.Thickness

	// Update polarization via Preisach model using the engine timestep.
	e.state.Polarization = e.model.TimeStep(e.state.ElectricField, e.dt)
	e.state.NormPol = e.model.NormalizedPolarization()

	log.Calculation("Step", map[string]interface{}{
		"time":    e.state.Time,
		"voltage": e.state.Voltage,
		"E_field": e.state.ElectricField,
	}, map[string]interface{}{
		"polarization": e.state.Polarization,
		"normPol":      e.state.NormPol,
	})

	// Record history
	e.recordHistory()

	// Advance time
	e.state.Time += e.dt
}

// generateVoltage produces the input voltage at time t.
func (e *Engine) generateVoltage(t float64) float64 {
	if e.waveform == WaveformManual {
		return e.state.Voltage // Use manually set value
	}

	omega := 2 * math.Pi * e.frequency
	phase := omega * t

	switch e.waveform {
	case WaveformSine:
		return e.amplitude * math.Sin(phase)

	case WaveformTriangle:
		// Triangle wave from -A to +A
		p := math.Mod(phase, 2*math.Pi) / (2 * math.Pi)
		if p < 0.25 {
			return e.amplitude * (4 * p)
		} else if p < 0.75 {
			return e.amplitude * (2 - 4*p)
		} else {
			return e.amplitude * (4*p - 4)
		}

	case WaveformSquare:
		if math.Sin(phase) >= 0 {
			return e.amplitude
		}
		return -e.amplitude

	default:
		return 0
	}
}

// recordHistory saves current state for plotting.
func (e *Engine) recordHistory() {
	s := e.state

	// Add new values
	s.VoltageHistory = append(s.VoltageHistory, s.Voltage)
	s.PolHistory = append(s.PolHistory, s.NormPol)

	// Trim if too long
	if len(s.VoltageHistory) > s.MaxHistory {
		s.VoltageHistory = s.VoltageHistory[1:]
		s.PolHistory = s.PolHistory[1:]
	}
}

// SetVoltage manually sets the voltage (for WaveformManual mode).
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) SetVoltage(v float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !isFinite(v) || (e.material != nil && !isRepresentableField(v, e.material.Thickness)) {
		return
	}
	e.state.Voltage = v
}

// SetWaveform changes the voltage waveform type.
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) SetWaveform(w WaveformType) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !isValidWaveformType(w) {
		return
	}
	e.waveform = w
	log.Debug("SetWaveform: %v", w)
}

// SetFrequency changes the waveform frequency.
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) SetFrequency(f float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !isValidWaveformFrequency(f) {
		return
	}
	e.frequency = f
	log.Debug("SetFrequency: %.2f Hz", f)
}

// SetAmplitude changes the waveform amplitude.
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) SetAmplitude(a float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !isValidWaveformAmplitude(a) || (e.material != nil && !isRepresentableField(a, e.material.Thickness)) {
		return
	}
	e.amplitude = a
	log.Debug("SetAmplitude: %.2f V", a)
}

// State returns a copy of the current simulation state.
// Thread-safe: returns a copy to prevent data races.
func (e *Engine) State() State {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// Return a copy of the state to prevent race conditions
	stateCopy := *e.state
	// Deep copy the history slices
	stateCopy.VoltageHistory = make([]float64, len(e.state.VoltageHistory))
	copy(stateCopy.VoltageHistory, e.state.VoltageHistory)
	stateCopy.PolHistory = make([]float64, len(e.state.PolHistory))
	copy(stateCopy.PolHistory, e.state.PolHistory)
	return stateCopy
}

// Reset clears the simulation state.
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.model != nil {
		e.model.Reset()
	}
	maxHistory := defaultMaxHistory
	if e.state != nil {
		maxHistory = e.state.MaxHistory
	}
	e.state = newState(maxHistory)
	log.Debug("Engine reset: state cleared, maxHistory=%d", e.state.MaxHistory)
}

// GetHysteresisData returns P-E data for plotting the hysteresis loop.
func (e *Engine) GetHysteresisData() ([]float64, []float64) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.material == nil || e.model == nil || !isValidMaterialThickness(e.material.Thickness) {
		return nil, nil
	}
	Emax := e.amplitude / e.material.Thickness
	return e.model.GetHysteresisLoop(Emax, 100)
}

// RunRealtime runs the simulation in real-time with the given callback.
// The callback receives a copy of the state to ensure thread safety.
func (e *Engine) RunRealtime(updateCallback func(State), targetFPS int) {
	frameInterval := realtimeFrameInterval(targetFPS)
	if frameInterval == 0 {
		return
	}

	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	stepsPerFrame := int(1.0 / (e.dt * float64(targetFPS)))

	for range ticker.C {
		if !e.IsRunning() {
			break
		}

		if !e.IsPaused() {
			// Run multiple physics steps per frame
			for i := 0; i < stepsPerFrame; i++ {
				e.Step()
			}
		}

		// Call the update callback with a copy of state
		if updateCallback != nil {
			state := e.State() // Returns a safe copy
			updateCallback(state)
		}
	}
}
