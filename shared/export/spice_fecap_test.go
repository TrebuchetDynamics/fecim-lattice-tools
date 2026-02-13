package export

import (
	"fmt"
	"strings"
	"testing"
)

func testFECapParams() FECapParams {
	return FECapParams{
		Name:    "FECAP_HZO_10nm",
		Alpha:   -1.2e8,
		Beta:    2.5e9,
		Gamma:   1.0e10,
		Rho:     2.0e-3,
		EpsR:    30,
		Area_m2: 1.0e-12,
		Thick_m: 10e-9,
	}
}

func TestFECapSubcircuit_ValidSyntax(t *testing.T) {
	params := testFECapParams()
	netlist := GenerateFECapSubcircuit(params)

	if !strings.Contains(netlist, ".subckt FECAP_HZO_10nm pos neg") {
		t.Fatalf("missing/invalid .subckt header: %s", netlist)
	}
	if !strings.Contains(netlist, "C_fe pos mid") {
		t.Fatal("missing capacitor instance with expected nodes")
	}
	if !strings.Contains(netlist, "R_visc mid neg") {
		t.Fatal("missing resistor instance with expected nodes")
	}
	if !strings.Contains(netlist, "B_landau mid neg V =") {
		t.Fatal("missing behavioral source with expected nodes")
	}
	if !strings.Contains(netlist, ".ends FECAP_HZO_10nm") {
		t.Fatal("missing .ends statement")
	}
}

func TestFECapSubcircuit_CorrectValues(t *testing.T) {
	params := testFECapParams()
	netlist := GenerateFECapSubcircuit(params)

	expectedC := vacuumPermittivity * params.EpsR * params.Area_m2 / params.Thick_m
	expectedR := params.Rho * params.Thick_m / params.Area_m2

	cLine := fmt.Sprintf("C_fe pos mid %.12e", expectedC)
	rLine := fmt.Sprintf("R_visc mid neg %.12e", expectedR)

	if !strings.Contains(netlist, cLine) {
		t.Fatalf("expected C_fe line %q not found", cLine)
	}
	if !strings.Contains(netlist, rLine) {
		t.Fatalf("expected R_visc line %q not found", rLine)
	}
}

func Test1T1RSubcircuit_IncludesMOSFET(t *testing.T) {
	params := testFECapParams()
	netlist := Generate1T1RSubcircuit(params, "")

	if !strings.Contains(netlist, ".subckt FECAP_HZO_10nm_1T1R bl wl sl") {
		t.Fatal("missing 1T1R subcircuit header")
	}
	if !strings.Contains(netlist, "M_sel nmid wl sl sl") {
		t.Fatal("missing MOSFET selector instance")
	}
	if !strings.Contains(netlist, "X_fecap bl nmid FECAP_HZO_10nm") {
		t.Fatal("missing FeCap instance inside 1T1R")
	}
	if !strings.Contains(netlist, ".model SKY130NMOS NMOS") {
		t.Fatal("missing default SKY130 NMOS model card")
	}
}
