//go:build !js && !ci
// +build !js,!ci

package render

import (
	"testing"
)

func TestGPUHeatmapRenderer_Available(t *testing.T) {
	r := NewGPUHeatmapRenderer()
	defer r.Destroy()
	t.Logf("GPU heatmap renderer available: %v", r.Available())
}

func TestGPUHeatmapRenderer_SmallArray(t *testing.T) {
	r := NewGPUHeatmapRenderer()
	defer r.Destroy()
	if !r.Available() {
		t.Skip("Vulkan not available")
	}

	values := make([]float64, 16) // 4x4
	for i := range values {
		values[i] = float64(i) / 15.0
	}

	img := r.RenderHeatmap(values, 4, 4, 200, 200)
	if img == nil {
		t.Fatal("RenderHeatmap returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 200 {
		t.Errorf("wrong size: %dx%d, want 200x200", bounds.Dx(), bounds.Dy())
	}
}

func TestGPUHeatmapRenderer_LargerArray(t *testing.T) {
	r := NewGPUHeatmapRenderer()
	defer r.Destroy()
	if !r.Available() {
		t.Skip("Vulkan not available")
	}

	const rows, cols = 32, 32
	values := make([]float64, rows*cols)
	for i := range values {
		values[i] = float64(i) / float64(len(values)-1)
	}

	img := r.RenderHeatmap(values, rows, cols, 400, 400)
	if img == nil {
		t.Fatal("RenderHeatmap returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 400 {
		t.Errorf("wrong size: %dx%d, want 400x400", bounds.Dx(), bounds.Dy())
	}
}

func TestGPUHeatmapRenderer_ResolutionChange(t *testing.T) {
	r := NewGPUHeatmapRenderer()
	defer r.Destroy()
	if !r.Available() {
		t.Skip("Vulkan not available")
	}

	values := make([]float64, 4)
	for i := range values {
		values[i] = float64(i) / 3.0
	}

	// Render at 100x100.
	img1 := r.RenderHeatmap(values, 2, 2, 100, 100)
	if img1 == nil {
		t.Fatal("first RenderHeatmap returned nil")
	}
	if b := img1.Bounds(); b.Dx() != 100 || b.Dy() != 100 {
		t.Errorf("first render size: %dx%d, want 100x100", b.Dx(), b.Dy())
	}

	// Render at 300x200 (triggers framebuffer re-creation).
	img2 := r.RenderHeatmap(values, 2, 2, 300, 200)
	if img2 == nil {
		t.Fatal("second RenderHeatmap returned nil")
	}
	if b := img2.Bounds(); b.Dx() != 300 || b.Dy() != 200 {
		t.Errorf("second render size: %dx%d, want 300x200", b.Dx(), b.Dy())
	}
}

func TestGPUHeatmapRenderer_InvalidInputs(t *testing.T) {
	r := NewGPUHeatmapRenderer()
	defer r.Destroy()
	if !r.Available() {
		t.Skip("Vulkan not available")
	}

	// Zero dimensions.
	if img := r.RenderHeatmap(nil, 0, 0, 100, 100); img != nil {
		t.Error("expected nil for zero dimensions")
	}

	// Insufficient values.
	values := make([]float64, 2) // need 4 for 2x2
	if img := r.RenderHeatmap(values, 2, 2, 100, 100); img != nil {
		t.Error("expected nil for insufficient values")
	}

	// Negative pixel dimensions.
	values = make([]float64, 4)
	if img := r.RenderHeatmap(values, 2, 2, -1, 100); img != nil {
		t.Error("expected nil for negative pixW")
	}
}

func TestGPUHeatmapRenderer_UnavailableReturnsNil(t *testing.T) {
	// A renderer with available=false should return nil without crashing.
	r := &GPUHeatmapRenderer{available: false}
	values := make([]float64, 4)
	img := r.RenderHeatmap(values, 2, 2, 100, 100)
	if img != nil {
		t.Error("expected nil from unavailable renderer")
	}
}

func TestGPUHeatmapRenderer_DestroyIdempotent(t *testing.T) {
	r := NewGPUHeatmapRenderer()
	r.Destroy()
	r.Destroy() // Should not panic.
}

func BenchmarkGPUHeatmap_64x64(b *testing.B) {
	r := NewGPUHeatmapRenderer()
	defer r.Destroy()
	if !r.Available() {
		b.Skip("Vulkan not available")
	}

	const n = 64 * 64
	values := make([]float64, n)
	for i := range values {
		values[i] = float64(i) / float64(n-1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.RenderHeatmap(values, 64, 64, 400, 400)
	}
}

func BenchmarkGPUHeatmap_256x256(b *testing.B) {
	r := NewGPUHeatmapRenderer()
	defer r.Destroy()
	if !r.Available() {
		b.Skip("Vulkan not available")
	}

	const n = 256 * 256
	values := make([]float64, n)
	for i := range values {
		values[i] = float64(i) / float64(n-1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.RenderHeatmap(values, 256, 256, 800, 800)
	}
}
