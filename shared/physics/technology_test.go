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
		})
	}
}

func TestVDDDecreasesWithSmallerNodes(t *testing.T) {
	if SKY130().VDD <= TSMC14().VDD {
		t.Fatalf("expected SKY130 VDD > TSMC14 VDD, got %.2f <= %.2f", SKY130().VDD, TSMC14().VDD)
	}
	if math.Abs(SKY130().VDD-1.8) > 1e-9 {
		t.Fatalf("expected SKY130 VDD ~1.8V, got %.3f", SKY130().VDD)
	}
	if math.Abs(TSMC14().VDD-0.8) > 1e-9 {
		t.Fatalf("expected TSMC14 VDD ~0.8V, got %.3f", TSMC14().VDD)
	}
}

func TestMetalPitchDecreasesWithNode(t *testing.T) {
	n130 := SKY130()
	n28 := TSMC28()
	n14 := TSMC14()
	if !(n130.MetalPitch > n28.MetalPitch && n28.MetalPitch > n14.MetalPitch) {
		t.Fatalf("expected pitch trend SKY130 > TSMC28 > TSMC14, got %.3gnm > %.3gnm > %.3gnm",
			n130.MetalPitch*1e9, n28.MetalPitch*1e9, n14.MetalPitch*1e9)
	}
}

func TestFeatureSizeMatchesName(t *testing.T) {
	expected := map[string]float64{
		"SKY130":  130e-9,
		"GF180MCU": 180e-9,
		"TSMC28":  28e-9,
		"TSMC14":  14e-9,
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
		{name: "unknown", want: "SKY130"},
		{name: "N28", want: "TSMC28"},
		{name: "14nm", want: "TSMC14"},
	} {
		t.Run(fmt.Sprintf("alias_%s", tc.name), func(t *testing.T) {
			if got := TechnologyNodeFromName(tc.name).Name; got != tc.want {
				t.Fatalf("TechnologyNodeFromName(%q) = %s, want %s", tc.name, got, tc.want)
			}
		})
	}
}
