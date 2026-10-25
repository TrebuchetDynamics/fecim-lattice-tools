package peloop_test

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"fecim-lattice-tools/validation/peloop"
)

func TestDigitizeCalibratedPixels(t *testing.T) {
	f, err := os.Open("../literature/data/experimental/kim2020_fig3a_ref_digitized/source.provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	profile, err := peloop.ReadSourceProfile(f)
	if err != nil {
		t.Fatal(err)
	}
	// Synthetic tick positions, not extracted experimental samples:
	// E=(x-100)/100-3; P=45-(y-100)/10, in original raster coordinates.
	profile.Calibration.X.Anchors = []peloop.AxisAnchor{{Pixel: 100, Value: -3}, {Pixel: 700, Value: 3}}
	profile.Calibration.Y.Anchors = []peloop.AxisAnchor{{Pixel: 1000, Value: -45}, {Pixel: 100, Value: 45}}
	const header = "branch,point_index,x_px,y_px\n"
	const pixels = header + "descending,1,400,550\nascending,2,700,100\nascending,0,100,1000\ndescending,0,700,325\nascending,1,250,775\ndescending,2,100,550\n"
	const want = "branch,point_index,field_MV_cm,polarization_uC_cm2\nascending,0,-3,-45\nascending,1,-1.5,-22.5\nascending,2,3,45\ndescending,0,3,22.5\ndescending,1,0,0\ndescending,2,-3,0\n"
	for i := 0; i < 2; i++ {
		got, err := peloop.Digitize(profile, strings.NewReader(pixels))
		if err != nil || string(got) != want {
			t.Fatalf("conversion = %q, %v; want %q", got, err, want)
		}
	}
	for _, tc := range []struct{ name, input string }{
		{"empty", ""},
		{"no points", header},
		{"wrong header", "branch,point_index,field_MV_cm,polarization_uC_cm2\nascending,0,100,550\n"},
		{"short row", header + "ascending,0,100\n"},
		{"extra column", header + "ascending,0,100,550,extra\n"},
		{"invalid pixel", header + "ascending,0,word,550\n"},
		{"nonfinite x", header + "ascending,0,NaN,550\n"},
		{"nonfinite y", header + "ascending,0,100,+Inf\n"},
		{"overflow", header + "ascending,0,1e999,550\n"},
		{"outside raster", header + "ascending,0,-1,550\n"},
		{"outside x ticks", header + "ascending,0,701,550\n"},
		{"outside y ticks", header + "ascending,0,100,99\n"},
		{"branch label", header + "up,0,100,550\n"},
		{"missing index", header + "ascending,,100,550\n"},
		{"fractional index", header + "ascending,0.5,100,550\n"},
		{"negative index", header + "ascending,-1,100,550\n"},
		{"missing zero", header + "ascending,1,100,550\n"},
		{"index gap", header + "ascending,0,100,550\nascending,2,200,550\n"},
		{"repeated index", header + "descending,0,100,550\ndescending,0,200,550\n"},
		{"duplicate field", header + "ascending,0,100,550\nascending,1,100.0,325\n"},
		{"malformed CSV", header + "ascending,0,\"100,550\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, err := peloop.Digitize(profile, strings.NewReader(tc.input)); err == nil || len(got) != 0 {
				t.Fatalf("invalid input returned %q, %v", got, err)
			}
		})
	}
	for _, tc := range []struct {
		name string
		axis peloop.AxisCalibration
	}{
		{"missing anchors", peloop.AxisCalibration{Unit: "MV/cm"}},
		{"degenerate pixels", peloop.AxisCalibration{Unit: "MV/cm", Anchors: []peloop.AxisAnchor{{Pixel: 100, Value: -3}, {Pixel: 100, Value: 3}}}},
		{"degenerate values", peloop.AxisCalibration{Unit: "MV/cm", Anchors: []peloop.AxisAnchor{{Pixel: 100, Value: -3}, {Pixel: 700, Value: -3}}}},
		{"nonfinite anchor", peloop.AxisCalibration{Unit: "MV/cm", Anchors: []peloop.AxisAnchor{{Pixel: math.NaN(), Value: -3}, {Pixel: 700, Value: 3}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := profile
			p.Calibration.X = tc.axis
			if got, err := peloop.Digitize(p, strings.NewReader(pixels)); err == nil || len(got) != 0 {
				t.Fatalf("invalid calibration returned %q, %v", got, err)
			}
		})
	}
	t.Run("row bound", func(t *testing.T) {
		var input strings.Builder
		input.WriteString(header)
		for i := 0; i < 10000; i++ {
			index := i % 5000
			branch, x := "ascending", 100+float64(index)/100
			if i >= 5000 {
				branch, x = "descending", 700-float64(index)/100
			}
			fmt.Fprintf(&input, "%s,%d,%.8f,550\n", branch, index, x)
		}
		if _, err := peloop.Digitize(profile, strings.NewReader(input.String())); err != nil {
			t.Fatalf("10000 points: %v", err)
		}
		input.WriteString("descending,5000,100,550\n")
		if got, err := peloop.Digitize(profile, strings.NewReader(input.String())); err == nil || len(got) != 0 {
			t.Fatalf("10001 points returned %d bytes, %v", len(got), err)
		}
	})
	t.Run("byte bound", func(t *testing.T) {
		input := pixels + strings.Repeat("\n", (2<<20)-len(pixels))
		if _, err := peloop.Digitize(profile, strings.NewReader(input)); err != nil {
			t.Fatalf("2MiB: %v", err)
		}
		if got, err := peloop.Digitize(profile, strings.NewReader(input+"\n")); err == nil || len(got) != 0 {
			t.Fatalf("over 2MiB returned %d bytes, %v", len(got), err)
		}
	})
}
