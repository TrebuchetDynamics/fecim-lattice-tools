package experiment

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"fecim-lattice-tools/workbench/project"
)

func fixtureBundle() project.Bundle {
	return project.Bundle{
		Root: "/tmp/project",
		Project: project.Project{
			SchemaVersion: 1,
			ID:            "hzo-study",
			Hypothesis:    "ADC resolution trades area for energy fidelity.",
			ModelVersion:  "fecim-system-v1",
			Objectives: []project.Objective{
				{Metric: "energy_pj", Direction: project.Minimize},
				{Metric: "latency_ns", Direction: project.Minimize},
			},
		},
		Design: project.Design{
			SchemaVersion: 1,
			Device:        project.Device{Material: "HZO", ConductanceLevels: 30, GMinS: 1e-6, GMaxS: 30e-6},
			Array:         project.Array{Rows: 32, Cols: 32, ReadVoltageV: 0.2},
			Circuit:       project.Circuit{ADCBits: 6, DACBits: 4, TIAGainOhm: 10_000, TechNode: "65nm"},
		},
		Sweep: project.Sweep{
			SchemaVersion: 1,
			Seed:          17,
			MaxPoints:     32,
			Parameters: []project.Parameter{
				{Path: "device.conductance_levels", Values: []float64{16, 30}},
				{Path: "array.rows", Range: &project.LinearRange{Start: 16, Stop: 32, Count: 2}},
				{Path: "circuit.adc_bits", Values: []float64{4, 6}},
			},
		},
	}
}

func TestExpandDeterministicCartesianOrder(t *testing.T) {
	points, err := Expand(fixtureBundle())
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 8 {
		t.Fatalf("points=%d want 8", len(points))
	}
	got := make([]string, 0, len(points))
	for _, point := range points {
		got = append(got, fmt.Sprintf("%d/%d/%d/%d", point.Design.Device.ConductanceLevels, point.Design.Array.Rows, point.Design.Circuit.ADCBits, point.Seed))
	}
	want := []string{"16/16/4/17", "16/16/6/18", "16/32/4/19", "16/32/6/20", "30/16/4/21", "30/16/6/22", "30/32/4/23", "30/32/6/24"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestExpandRejectsInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*project.Bundle)
		want   string
	}{
		{"duplicate", func(b *project.Bundle) { b.Sweep.Parameters[1].Path = b.Sweep.Parameters[0].Path }, "duplicate"},
		{"unknown", func(b *project.Bundle) { b.Sweep.Parameters[0].Path = "device.unknown" }, "unsupported sweep path"},
		{"integer", func(b *project.Bundle) { b.Sweep.Parameters[0].Values = []float64{3.5} }, "integer"},
		{"finite", func(b *project.Bundle) { b.Sweep.Parameters[0].Values = []float64{math.NaN()} }, "finite"},
		{"both", func(b *project.Bundle) {
			b.Sweep.Parameters[0].Range = &project.LinearRange{Start: 1, Stop: 2, Count: 2}
		}, "exactly one"},
		{"limit", func(b *project.Bundle) { b.Sweep.MaxPoints = 7 }, "max_points"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := fixtureBundle()
			test.mutate(&bundle)
			_, err := Expand(bundle)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("err=%v want %q", err, test.want)
			}
		})
	}
}
