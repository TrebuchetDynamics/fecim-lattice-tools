// Package constants provides fundamental physical constants for FeCIM simulations.
//
// All values use CODATA 2018 exact values where available:
//   - Boltzmann constant (k): 1.380649e-23 J/K (exact SI 2019 definition)
//   - Elementary charge (e): 1.602176634e-19 C (exact SI 2019 definition)
//   - Vacuum permittivity (ε₀): 8.8541878128e-12 F/m (derived from c=299792458, μ₀=4π×10⁻⁷)
//   - k in eV/K: 8.617333262145e-05 eV/K (derived from exact values)
//
// Use these constants instead of redefining them locally to avoid
// precision drift across the codebase.
package constants

// Physical constants (CODATA 2018 exact SI values).

const (
	// BoltzmannConstantJPerK is the Boltzmann constant in J/K (exact SI definition).
	BoltzmannConstantJPerK = 1.380649e-23

	// BoltzmannConstanteVPerK is the Boltzmann constant in eV/K.
	BoltzmannConstanteVPerK = 8.617333262145e-05

	// ElectronChargeC is the elementary charge in C (exact SI definition).
	ElectronChargeC = 1.602176634e-19

	// VacuumPermittivityFPerM is the vacuum permittivity ε₀ in F/m.
	// Derived from the exact c=299792458 m/s and μ₀=4π×10⁻⁷ H/m.
	VacuumPermittivityFPerM = 8.8541878128e-12
)
