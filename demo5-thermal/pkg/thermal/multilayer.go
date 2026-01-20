package thermal

import (
	"math"
)

// MultiLayerSim manages thermal coupling between multiple stacked layers.
// Models vertical heat transfer through inter-layer thermal resistance.
type MultiLayerSim struct {
	Layers        []*ThermalSim // Stack of thermal layers
	Coupling      []float64     // Inter-layer thermal coupling coefficients
	HeatSinkTemp  float64       // Heat sink temperature at bottom
	HeatSinkCoeff float64       // Heat sink thermal coupling coefficient
}

// NewMultiLayerSim creates a multi-layer thermal simulation.
func NewMultiLayerSim(numLayers, width, height int) *MultiLayerSim {
	layers := make([]*ThermalSim, numLayers)
	coupling := make([]float64, numLayers-1)

	for i := 0; i < numLayers; i++ {
		layers[i] = NewThermalSim(width, height)
		layers[i].Reset()
	}

	// Default coupling between layers (higher = better thermal contact)
	for i := 0; i < numLayers-1; i++ {
		coupling[i] = 0.1 // Moderate coupling
	}

	return &MultiLayerSim{
		Layers:        layers,
		Coupling:      coupling,
		HeatSinkTemp:  25.0, // Heat sink at ambient
		HeatSinkCoeff: 0.5,  // Good thermal contact with heat sink
	}
}

// DefaultMultiLayerSim creates a 3-layer simulation matching FeCIM architecture.
func DefaultMultiLayerSim() *MultiLayerSim {
	sim := NewMultiLayerSim(3, 32, 32)

	// Configure layers with different thermal properties
	// Layer 0: Bottom (closest to heat sink)
	sim.Layers[0].Conductivity = 150.0 // Silicon substrate

	// Layer 1: Middle (crossbar array)
	sim.Layers[1].Conductivity = 50.0 // HZO/metal has lower conductivity

	// Layer 2: Top (BEOL interconnects)
	sim.Layers[2].Conductivity = 100.0 // Metal interconnects

	return sim
}

// Step advances all layers and handles inter-layer coupling.
func (m *MultiLayerSim) Step(dt float64) {
	// First, step each layer individually (in-plane diffusion)
	for _, layer := range m.Layers {
		layer.Step(dt)
	}

	// Then handle inter-layer coupling (vertical heat transfer)
	m.stepVerticalCoupling(dt)
}

// stepVerticalCoupling handles heat transfer between layers.
func (m *MultiLayerSim) stepVerticalCoupling(dt float64) {
	numLayers := len(m.Layers)
	if numLayers < 2 {
		return
	}

	// Process from top to bottom
	for i := numLayers - 1; i >= 0; i-- {
		layer := m.Layers[i]
		layer.mu.Lock()

		for y := 0; y < layer.Height; y++ {
			for x := 0; x < layer.Width; x++ {
				T := layer.Grid[y][x]
				heatExchange := 0.0

				// Heat transfer to layer above (if exists)
				if i < numLayers-1 {
					m.Layers[i+1].mu.RLock()
					Tabove := m.Layers[i+1].Grid[y][x]
					m.Layers[i+1].mu.RUnlock()
					heatExchange -= m.Coupling[i] * (T - Tabove) * dt
				}

				// Heat transfer to layer below (if exists)
				if i > 0 {
					m.Layers[i-1].mu.RLock()
					Tbelow := m.Layers[i-1].Grid[y][x]
					m.Layers[i-1].mu.RUnlock()
					heatExchange -= m.Coupling[i-1] * (T - Tbelow) * dt
				}

				// Bottom layer connects to heat sink
				if i == 0 {
					heatExchange -= m.HeatSinkCoeff * (T - m.HeatSinkTemp) * dt
				}

				layer.Grid[y][x] = T + heatExchange
			}
		}

		layer.mu.Unlock()
	}
}

// StepMultiple runs multiple time steps.
func (m *MultiLayerSim) StepMultiple(steps int, dt float64) {
	for i := 0; i < steps; i++ {
		m.Step(dt)
	}
}

// Reset resets all layers to ambient temperature.
func (m *MultiLayerSim) Reset() {
	for _, layer := range m.Layers {
		layer.Reset()
	}
}

// SetLayerPower sets the power map for a specific layer.
func (m *MultiLayerSim) SetLayerPower(layerIndex int, powerMap [][]float64) {
	if layerIndex >= 0 && layerIndex < len(m.Layers) {
		m.Layers[layerIndex].SetPowerMap(powerMap)
	}
}

// SetCellPower sets power at a specific cell in a specific layer.
func (m *MultiLayerSim) SetCellPower(layerIndex, x, y int, power float64) {
	if layerIndex >= 0 && layerIndex < len(m.Layers) {
		m.Layers[layerIndex].SetPower(x, y, power)
	}
}

// GetLayerMaxTemp returns maximum temperature for a specific layer.
func (m *MultiLayerSim) GetLayerMaxTemp(layerIndex int) float64 {
	if layerIndex >= 0 && layerIndex < len(m.Layers) {
		return m.Layers[layerIndex].GetMaxTemperature()
	}
	return 0
}

// GetGlobalMaxTemp returns maximum temperature across all layers.
func (m *MultiLayerSim) GetGlobalMaxTemp() float64 {
	maxTemp := -math.MaxFloat64
	for _, layer := range m.Layers {
		layerMax := layer.GetMaxTemperature()
		if layerMax > maxTemp {
			maxTemp = layerMax
		}
	}
	return maxTemp
}

// GetGlobalMinTemp returns minimum temperature across all layers.
func (m *MultiLayerSim) GetGlobalMinTemp() float64 {
	minTemp := math.MaxFloat64
	for _, layer := range m.Layers {
		layerMin := layer.GetMinTemperature()
		if layerMin < minTemp {
			minTemp = layerMin
		}
	}
	return minTemp
}

// GetStackAverageTemp returns average temperature across all layers.
func (m *MultiLayerSim) GetStackAverageTemp() float64 {
	sum := 0.0
	count := 0
	for _, layer := range m.Layers {
		sum += layer.GetAverageTemperature()
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// TotalHeatGeneration returns total heat generated across all layers.
func (m *MultiLayerSim) TotalHeatGeneration() float64 {
	total := 0.0
	for _, layer := range m.Layers {
		total += layer.TotalHeatGeneration()
	}
	return total
}

// FindAllHotspots finds hotspots across all layers.
func (m *MultiLayerSim) FindAllHotspots(threshold float64) map[int][]HotspotInfo {
	hotspots := make(map[int][]HotspotInfo)
	for i, layer := range m.Layers {
		layerHotspots := layer.FindHotspots(threshold)
		if len(layerHotspots) > 0 {
			hotspots[i] = layerHotspots
		}
	}
	return hotspots
}

// CheckStackWarning checks thermal warning across all layers.
func (m *MultiLayerSim) CheckStackWarning() *ThermalWarning {
	// Find worst warning across layers
	var worstWarning *ThermalWarning
	worstLevel := 0

	for _, layer := range m.Layers {
		warning := layer.CheckThermalWarning()
		if warning != nil && warning.Level > worstLevel {
			worstLevel = warning.Level
			worstWarning = warning
		}
	}

	if worstWarning != nil {
		// Update with stack-wide stats
		worstWarning.MaxTemp = m.GetGlobalMaxTemp()
		worstWarning.AverageTemp = m.GetStackAverageTemp()
	}

	return worstWarning
}

// VerticalTemperatureProfile returns temperature along a vertical line through all layers.
func (m *MultiLayerSim) VerticalTemperatureProfile(x, y int) []float64 {
	profile := make([]float64, len(m.Layers))
	for i, layer := range m.Layers {
		profile[i] = layer.GetTemperature(x, y)
	}
	return profile
}

// HeatFlowBetweenLayers calculates heat flux between adjacent layers.
func (m *MultiLayerSim) HeatFlowBetweenLayers(layer1, layer2 int) float64 {
	if layer1 < 0 || layer2 < 0 || layer1 >= len(m.Layers) || layer2 >= len(m.Layers) {
		return 0
	}
	if layer1 > layer2 {
		layer1, layer2 = layer2, layer1
	}
	if layer2-layer1 != 1 {
		return 0 // Not adjacent
	}

	// Calculate total heat flux Q = coupling * ΔT * Area
	totalFlux := 0.0
	l1, l2 := m.Layers[layer1], m.Layers[layer2]
	coupling := m.Coupling[layer1]

	l1.mu.RLock()
	l2.mu.RLock()
	defer l1.mu.RUnlock()
	defer l2.mu.RUnlock()

	for y := 0; y < l1.Height && y < l2.Height; y++ {
		for x := 0; x < l1.Width && x < l2.Width; x++ {
			dT := l1.Grid[y][x] - l2.Grid[y][x]
			totalFlux += coupling * dT
		}
	}

	return totalFlux
}
