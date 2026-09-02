package experiment

import (
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/project"
)

func TestIDStableAndBehaviorSensitive(t *testing.T) {
	point := DesignPoint{Index: 0, Design: fixtureBundle().Design, Seed: 17}
	inputs := []project.ResolvedInput{{Path: "inputs/a.csv", SHA256: strings.Repeat("a", 64), Citation: "park2015_advmat_hzo", Evidence: "literature-backed"}}
	first, err := ID(point, "fecim-system-v1", inputs)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ID(point, "fecim-system-v1", append([]project.ResolvedInput(nil), inputs...))
	if err != nil {
		t.Fatal(err)
	}
	if first != second || len(first) != 64 {
		t.Fatalf("ids=%q/%q", first, second)
	}
	point.Design.Circuit.ADCBits++
	changed, err := ID(point, "fecim-system-v1", inputs)
	if err != nil {
		t.Fatal(err)
	}
	if changed == first {
		t.Fatal("ADC change did not change identity")
	}
}

func TestIDSortsInputsByPath(t *testing.T) {
	point := DesignPoint{Design: fixtureBundle().Design, Seed: 17}
	a := project.ResolvedInput{Path: "inputs/a", SHA256: strings.Repeat("a", 64)}
	b := project.ResolvedInput{Path: "inputs/b", SHA256: strings.Repeat("b", 64)}
	first, err := ID(point, "v1", []project.ResolvedInput{b, a})
	if err != nil {
		t.Fatal(err)
	}
	second, err := ID(point, "v1", []project.ResolvedInput{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("input order changed identity: %s != %s", first, second)
	}
}
