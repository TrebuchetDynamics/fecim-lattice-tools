package export

import (
	"fmt"
	"strings"

	"fecim-lattice-tools/shared/physics"
)

const vacuumPermittivity = 8.8541878128e-12 // F/m

// FECapParams contains parameters for Landau-Khalatnikov ferroelectric capacitor SPICE export.
type FECapParams struct {
	Name     string  // e.g. "FECAP_HZO_10nm"
	Alpha    float64 // Landau coefficient (C^-2 m^2 N)
	Beta     float64 // C^-4 m^6 N
	Gamma    float64 // C^-6 m^10 N
	Rho      float64 // damping/viscosity (Ohm*m)
	EpsR     float64 // relative permittivity
	Area_m2  float64 // device area
	Thick_m  float64 // film thickness
}

// GenerateFECapSubcircuit returns an ngspice-compatible Landau-Khalatnikov FeCap subcircuit.
func GenerateFECapSubcircuit(params FECapParams) string {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		name = "FECAP"
	}

	cFe := vacuumPermittivity * params.EpsR * params.Area_m2 / params.Thick_m
	rVisc := params.Rho * params.Thick_m / params.Area_m2

	// Polarization approximation from electrostatics: Q = C*V, P = Q/A.
	pExpr := fmt.Sprintf("((V(pos,mid)*%.12e)/%.12e)", cFe, params.Area_m2)
	vLandauExpr := fmt.Sprintf("-(2*%.12e*(%s) + 4*%.12e*pow((%s),3) + 6*%.12e*pow((%s),5))*%.12e",
		params.Alpha, pExpr, params.Beta, pExpr, params.Gamma, pExpr, params.Thick_m)

	return fmt.Sprintf(`.subckt %s pos neg
* Ferroelectric capacitor — Landau-Khalatnikov model
* Reference: Sivasubramanian & Widom, IEEE (2003); Materlik et al., J. Appl. Phys. 117, 134109 (2015)
C_fe pos mid %.12e
R_visc mid neg %.12e
B_landau mid neg V = %s
.ends %s
`, name, cFe, rVisc, vLandauExpr, name)
}

// Generate1T1RSubcircuit returns an ngspice 1T1R wrapper with one NMOS selector and one FeCap.
func Generate1T1RSubcircuit(fecap FECapParams, mosfetModel string) string {
	if strings.TrimSpace(fecap.Name) == "" {
		fecap.Name = "FECAP"
	}

	sel := physics.SKY130NMOS()
	modelName := strings.TrimSpace(mosfetModel)
	modelCard := ""
	if modelName == "" {
		modelName = "SKY130NMOS"
		modelCard = fmt.Sprintf(".model %s NMOS (LEVEL=1 VTO=%.6g KP=120e-6 LAMBDA=0.03)\n", modelName, sel.Vth)
	}

	return fmt.Sprintf(`%s
.subckt %s_1T1R bl wl sl
* 1T1R wrapper: NMOS selector in series with Landau-Khalatnikov FeCap
M_sel nmid wl sl sl %s W=%.12e L=%.12e
X_fecap bl nmid %s
.ends %s_1T1R
`, modelCard, fecap.Name, modelName, sel.W, sel.L, fecap.Name, fecap.Name)
}
