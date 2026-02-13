package physics

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestAllTechnologyNodesPositiveValues(t *testing.T) {
	for _, n := range AllTechnologyNodes() {
		t.Run(n.Name, func(t *testing.T) {
			if n.Name == "" {
				t.Fatal("name must be non-empty")
			}
			if n.FeatureSize <= 0 || n.MetalPitch <= 0 || n.MetalWidth <= 0 || n.MetalThickness <= 0 ||
				n.MetalResistivity <= 0 || n.VDD <= 0 || n.CellPitchX <= 0 || n.CellRowHeight <= 0 {
				t.Fatalf("all physical parameters must be positive: %+v", n)
			}
			if n.Transistor.NMOSVth <= 0 || n.Transistor.PMOSVth >= 0 || n.Transistor.NMOSIon <= 0 || n.Transistor.PMOSIon <= 0 || n.Transistor.GateCapF <= 0 {
				t.Fatalf("transistor model must be populated: %+v", n.Transistor)
			}
		})
	}
}

func TestVDDDecreasesWithSmallerNodes(t *testing.T) {
	if Node130nm().VDD <= Node14nm().VDD {
		t.Fatalf("expected 130nm VDD > 14nm VDD, got %.2f <= %.2f", Node130nm().VDD, Node14nm().VDD)
	}
	if math.Abs(Node130nm().VDD-1.8) > 1e-9 {
		t.Fatalf("expected 130nm VDD ~1.8V, got %.3f", Node130nm().VDD)
	}
	if math.Abs(Node14nm().VDD-0.8) > 1e-9 {
		t.Fatalf("expected 14nm VDD ~0.8V, got %.3f", Node14nm().VDD)
	}
}

func TestMetalPitchDecreasesWithNode(t *testing.T) {
	n130 := Node130nm()
	n65 := Node65nm()
	n28 := Node28nm()
	n14 := Node14nm()
	if !(n130.MetalPitch > n65.MetalPitch && n65.MetalPitch > n28.MetalPitch && n28.MetalPitch > n14.MetalPitch) {
		t.Fatalf("expected pitch trend 130nm > 65nm > 28nm > 14nm")
	}
}

func TestFeatureSizeMatchesName(t *testing.T) {
	expected := map[string]float64{
		"130nm": 130e-9,
		"65nm":  65e-9,
		"28nm":  28e-9,
		"14nm":  14e-9,
	}
	for _, n := range AllTechnologyNodes() {
		want, ok := expected[n.Name]
		if !ok {
			t.Fatalf("unexpected technology node in list: %s", n.Name)
		}
		if math.Abs(n.FeatureSize-want) > 1e-18 {
			t.Fatalf("%s feature size mismatch: got %.3gnm want %.3gnm", n.Name, n.FeatureSize*1e9, want*1e9)
		}
		fromName := TechnologyNodeFromName(strings.ToLower(n.Name))
		if math.Abs(fromName.FeatureSize-n.FeatureSize) > 1e-18 {
			t.Fatalf("TechnologyNodeFromName(%q) mismatch: got %.3gnm want %.3gnm", strings.ToLower(n.Name), fromName.FeatureSize*1e9, n.FeatureSize*1e9)
		}
	}

	for _, tc := range []struct {
		name string
		want string
	}{
		{name: "unknown", want: "130nm"},
		{name: "N28", want: "28nm"},
		{name: "14nm", want: "14nm"},
		{name: "sky130", want: "130nm"},
		{name: "65", want: "65nm"},
	} {
		t.Run(fmt.Sprintf("alias_%s", tc.name), func(t *testing.T) {
			if got := TechnologyNodeFromName(tc.name).Name; got != tc.want {
				t.Fatalf("TechnologyNodeFromName(%q) = %s, want %s", tc.name, got, tc.want)
			}
		})
	}
}
