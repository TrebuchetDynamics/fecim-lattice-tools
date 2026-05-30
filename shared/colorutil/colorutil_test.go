// Package colorutil provides shared color manipulation utilities.

package colorutil

import (
	"image/color"
	"testing"
)

func TestWithAlpha(t *testing.T) {
	c := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	result := WithAlpha(c, 128)
	rc, gc, bc, ac := result.RGBA()
	if rc>>8 != 100 || gc>>8 != 150 || bc>>8 != 200 || ac>>8 != 128 {
		t.Errorf("WithAlpha(100,150,200,255, 128) = (%d,%d,%d,%d), want (100,150,200,128)",
			rc>>8, gc>>8, bc>>8, ac>>8)
	}
}

func TestWithAlpha_NonRGBA(t *testing.T) {
	// Test with a non-RGBA color (e.g., NRGBA)
	c := color.NRGBA{R: 50, G: 100, B: 150, A: 200}
	result := WithAlpha(c, 64)
	_, _, _, a := result.RGBA()
	if a>>8 != 64 {
		t.Errorf("WithAlpha on NRGBA got alpha %d, want 64", a>>8)
	}
}

func TestLuminance(t *testing.T) {
	tests := []struct {
		name  string
		c     color.Color
		want  float64
		eps   float64
		below bool // if true, want is upper bound
	}{
		{"white", color.RGBA{255, 255, 255, 255}, 1.0, 0.01, false},
		{"black", color.RGBA{0, 0, 0, 255}, 0.0, 0.01, false},
		{"mid-gray", color.RGBA{128, 128, 128, 255}, 0.5, 0.05, false},
		{"red", color.RGBA{255, 0, 0, 255}, 0.299, 0.01, false},
		{"green", color.RGBA{0, 255, 0, 255}, 0.587, 0.01, false},
		{"blue", color.RGBA{0, 0, 255, 255}, 0.114, 0.01, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Luminance(tt.c)
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.eps {
				t.Errorf("Luminance(%s) = %.6f, want %.6f ±%.6f", tt.name, got, tt.want, tt.eps)
			}
		})
	}
}

func TestGetContrastColor(t *testing.T) {
	tests := []struct {
		name string
		bg   color.Color
		want color.Color
	}{
		{"dark bg returns white", color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}},
		{"light bg returns black", color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 255}},
		{"dark blue bg returns white", color.RGBA{0, 50, 100, 255}, color.RGBA{255, 255, 255, 255}},
		{"light blue bg returns black", color.RGBA{200, 220, 240, 255}, color.RGBA{0, 0, 0, 255}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetContrastColor(tt.bg)
			gotR, gotG, gotB, _ := got.RGBA()
			wantR, wantG, wantB, _ := tt.want.RGBA()
			if gotR != wantR || gotG != wantG || gotB != wantB {
				t.Errorf("GetContrastColor(%s) = (%d,%d,%d), want (%d,%d,%d)",
					tt.name, gotR>>8, gotG>>8, gotB>>8, wantR>>8, wantG>>8, wantB>>8)
			}
		})
	}
}
