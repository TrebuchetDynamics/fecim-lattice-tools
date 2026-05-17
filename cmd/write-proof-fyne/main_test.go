//go:build legacy_fyne

package main

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
)

func TestSavePNGWritesFile(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "proof.png")
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := savePNG(p, img); err != nil {
		t.Fatalf("savePNG failed: %v", err)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if st.Size() == 0 {
		t.Fatal("expected non-empty png")
	}
}
