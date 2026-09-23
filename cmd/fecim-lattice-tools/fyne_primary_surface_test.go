//go:build cgo

package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultLauncherFooterKeepsSimulationBoundary(t *testing.T) {
	footer := NewResponsiveFooter()
	renderer := footer.CreateRenderer()
	for _, width := range []float32{360, 800, 1400} {
		renderer.Layout(fyne.NewSize(width, 80))
		for _, object := range renderer.Objects() {
			text, ok := object.(*canvas.Text)
			if !ok {
				continue
			}
			for _, claim := range []string{"30 States", "96-98%", "25-100", "10⁹ Cycles", "TRL 4"} {
				if strings.Contains(text.Text, claim) {
					t.Errorf("width %v footer presents unverified claim %q: %q", width, claim, text.Text)
				}
			}
		}
	}
}

func TestDefaultCommandUsesFyneUISurface(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	cmd := exec.Command("go", "list", "-e", "-deps", "./cmd/fecim-lattice-tools")
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		t.Fatalf("go list default command deps failed: %v\n%s", err, out)
	}
	deps := strings.Fields(string(out))
	if !containsPackagePrefix(deps, "fyne.io/fyne/v2") {
		t.Fatalf("default command deps must include Fyne UI surface; deps did not contain fyne.io/fyne/v2")
	}
	if containsPackagePrefix(deps, "fecim-lattice-tools/internal/"+"go"+"gpuapp") {
		t.Fatalf("default command deps must not include retired UI app surface when fully switched back to Fyne")
	}
	if containsPackagePrefix(deps, "github.com/"+"go"+"gpu") {
		t.Fatalf("default command deps must not include external retired UI packages when fully switched back to Fyne")
	}
}

func containsPackagePrefix(pkgs []string, prefix string) bool {
	for _, pkg := range pkgs {
		if pkg == prefix || strings.HasPrefix(pkg, prefix+"/") {
			return true
		}
	}
	return false
}
