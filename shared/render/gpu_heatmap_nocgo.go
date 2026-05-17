//go:build !js && !cgo

package render

import "image"

// GPUHeatmapRenderer is a no-op renderer when Vulkan is not in the build.
type GPUHeatmapRenderer struct {
	available bool
}

// NewGPUHeatmapRenderer returns a renderer that reports unavailable.
func NewGPUHeatmapRenderer() *GPUHeatmapRenderer { return &GPUHeatmapRenderer{} }

// Available reports whether GPU rendering is functional.
func (r *GPUHeatmapRenderer) Available() bool { return r.available }

// RenderHeatmap returns nil when GPU rendering is unavailable.
func (r *GPUHeatmapRenderer) RenderHeatmap(values []float64, rows, cols, pixW, pixH int) image.Image {
	return nil
}

// Destroy releases renderer resources.
func (r *GPUHeatmapRenderer) Destroy() {}
