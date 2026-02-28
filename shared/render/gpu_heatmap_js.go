//go:build js
// +build js

// Package render provides GPU-accelerated rendering backends for crossbar
// array visualization. This file provides the WASM/JS stub which always
// reports GPU as unavailable.
package render

import "image"

// GPUHeatmapRenderer stub for platforms without Vulkan (WASM builds).
type GPUHeatmapRenderer struct{}

// NewGPUHeatmapRenderer returns a no-op renderer on WASM platforms.
func NewGPUHeatmapRenderer() *GPUHeatmapRenderer { return &GPUHeatmapRenderer{} }

// Available always returns false in WASM builds.
func (r *GPUHeatmapRenderer) Available() bool { return false }

// RenderHeatmap is a no-op that returns nil in WASM builds.
func (r *GPUHeatmapRenderer) RenderHeatmap(values []float64, rows, cols, pixW, pixH int) image.Image {
	return nil
}

// Destroy is a no-op in WASM builds.
func (r *GPUHeatmapRenderer) Destroy() {}
