//go:build !cgo

package gogpuscreenshot

import (
	"image"
	"image/png"
	"os"
	"testing"

	"fecim-lattice-tools/internal/gogpuapp"
	"fecim-lattice-tools/shared/viewmodel"
)

func TestGenerateCapturesRealGogpuAppFrame(t *testing.T) {
	opts := DefaultOptions()
	opts.OutputDir = t.TempDir()
	opts.Only = "docs"
	opts.Width = 420
	opts.Height = 260

	if err := Generate(opts); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	got := readPNG(t, opts.OutputPath("docs-overview.png"))
	want, err := gogpuapp.CaptureFrameImage(viewmodel.ModuleDocs, opts.Width, opts.Height)
	if err != nil {
		t.Fatalf("CaptureFrameImage error: %v", err)
	}

	if !imagesEqual(got, want) {
		t.Fatal("generated screenshot did not match the real gogpu app frame")
	}
}

func readPNG(t *testing.T, path string) image.Image {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return img
}

func imagesEqual(a, b image.Image) bool {
	if !a.Bounds().Eq(b.Bounds()) {
		return false
	}
	bounds := a.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}
	return true
}
