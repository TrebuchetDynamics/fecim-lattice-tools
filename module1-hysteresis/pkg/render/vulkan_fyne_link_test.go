//go:build !js && cgo

package render_test

import (
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/render"
	_ "fyne.io/fyne/v2/app" // Exercise linking Fyne and the Vulkan renderer in one binary.
)

func TestVulkanRendererLinksWithFyne(t *testing.T) {
	// Construction does not initialize GLFW or require a GPU/display.
	if renderer := render.NewVulkanRenderer(render.DefaultConfig()); renderer == nil {
		t.Fatal("expected a renderer without initializing a graphics context")
	}
}
