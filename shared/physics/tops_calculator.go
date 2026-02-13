package physics

// ComputeMetrics captures throughput and efficiency inputs for an MVM engine.
type ComputeMetrics struct {
	ArrayRows, ArrayCols int
	Frequency            float64 // Hz
	DACBits, ADCBits     int
	EnergyPerMVM         float64 // J
	LatencyPerMVM        float64 // s
}

// BaselineSystem is a literature baseline for comparison.
type BaselineSystem struct {
	Name     string
	PowerW   float64
	TOPSPerW float64
}

func (b BaselineSystem) TOPS() float64 {
	return b.PowerW * b.TOPSPerW
}

var (
	BaselineCPUXeon = BaselineSystem{Name: "CPU (Intel Xeon)", PowerW: 150.0, TOPSPerW: 0.5}
	BaselineGPUA100 = BaselineSystem{Name: "GPU (NVIDIA A100 INT8)", PowerW: 300.0, TOPSPerW: 5.0}
	BaselineTPUv4   = BaselineSystem{Name: "TPU (Google v4)", PowerW: 170.0, TOPSPerW: 2.7}
)

// OpsPerSecond returns 2×N×M×f MAC operations/s.
func (m ComputeMetrics) OpsPerSecond() float64 {
	if m.ArrayRows <= 0 || m.ArrayCols <= 0 || m.Frequency <= 0 {
		return 0
	}
	return 2.0 * float64(m.ArrayRows*m.ArrayCols) * m.Frequency
}

// TOPS returns throughput in tera-ops/s.
func (m ComputeMetrics) TOPS() float64 {
	return m.OpsPerSecond() / 1e12
}

// PowerW returns dynamic power from per-MVM energy and MVM rate.
func (m ComputeMetrics) PowerW() float64 {
	if m.EnergyPerMVM <= 0 || m.Frequency <= 0 {
		return 0
	}
	return m.EnergyPerMVM * m.Frequency
}

// TOPSPerW returns energy efficiency in TOPS/W.
func (m ComputeMetrics) TOPSPerW() float64 {
	pw := m.PowerW()
	if pw <= 0 {
		return 0
	}
	return m.TOPS() / pw
}

// LatencyNs returns per-MVM latency in ns.
func (m ComputeMetrics) LatencyNs() float64 {
	if m.LatencyPerMVM <= 0 {
		return 0
	}
	return m.LatencyPerMVM * 1e9
}
