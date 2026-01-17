// Package physics provides TDGL phase-field physics models for ferroelectrics.
package physics

// HZOMaterial contains Landau coefficients for HfO2-ZrO2 ferroelectrics.
// Based on first-principles calculations and experimental data.
type HZOMaterial struct {
	// Landau coefficients
	Alpha float64 // α (Vm/C) - quadratic term
	Beta  float64 // β (Vm⁵/C³) - quartic term
	Gamma float64 // γ (Vm⁹/C⁵) - sixth-order term

	// Gradient coefficient
	Kappa float64 // κ (Vm³/C) - domain wall energy

	// Kinetic coefficient
	L float64 // L (m³/VsC) - relaxation rate

	// Temperature dependence (Curie-Weiss)
	Tc float64 // Curie temperature (K)
	T0 float64 // Curie-Weiss temperature (K)

	// Physical dimensions
	LatticeCellSize float64 // a (m) - discretization length
}

// DefaultHZO returns standard HZO Landau parameters.
// Values from literature on orthorhombic Pca21 phase.
func DefaultHZO() *HZOMaterial {
	return &HZOMaterial{
		Alpha:           1.72e6,  // Vm/C
		Beta:            -2.5e9,  // Vm⁵/C³ (negative for double-well)
		Gamma:           1.5e11,  // Vm⁹/C⁵
		Kappa:           1e-9,    // Vm³/C
		L:               1e-3,    // m³/VsC
		Tc:              723,     // K (450°C)
		T0:              673,     // K
		LatticeCellSize: 0.5e-9,  // 0.5 nm
	}
}

// AlphaTemperature returns temperature-dependent α using Curie-Weiss law.
// α(T) = α₀ * (T - T₀) / Tc
func (m *HZOMaterial) AlphaTemperature(T float64) float64 {
	return m.Alpha * (T - m.T0) / m.Tc
}

// SpontaneousPolarization returns the equilibrium polarization from Landau theory.
// At equilibrium, dF/dP = 0 → P² = (-β ± sqrt(β² - 4αγ)) / (2γ)
func (m *HZOMaterial) SpontaneousPolarization(T float64) float64 {
	alpha := m.AlphaTemperature(T)

	// For T < Tc, we have spontaneous polarization
	if T >= m.Tc {
		return 0
	}

	// Simplified: P_s ≈ sqrt(-α/β) for small γ contributions
	if m.Beta >= 0 {
		return 0 // No double-well potential
	}

	// P_s² = -α/β (simplified, ignoring sixth-order term)
	PsSq := -alpha / m.Beta
	if PsSq <= 0 {
		return 0
	}

	return sqrt(PsSq)
}

// DomainWallWidth estimates the characteristic domain wall width.
// δ ≈ sqrt(κ / |α|)
func (m *HZOMaterial) DomainWallWidth(T float64) float64 {
	alpha := m.AlphaTemperature(T)
	if alpha == 0 {
		return 0
	}
	return sqrt(m.Kappa / abs(alpha))
}

// sqrt is a helper for square root of float64.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// Newton-Raphson approximation
	guess := x
	for i := 0; i < 10; i++ {
		guess = 0.5 * (guess + x/guess)
	}
	return guess
}

// abs returns absolute value.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
