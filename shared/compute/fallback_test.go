package compute

import (
	"strings"
	"testing"
)

// calculateDispatchSizeCPUReference is a CPU-only reference for dispatch sizing.
func calculateDispatchSizeCPUReference(totalElements uint32, cfg WorkgroupConfig) (uint32, uint32, uint32) {
	if totalElements == 0 {
		return 0, 0, 0
	}
	return (totalElements + cfg.LocalSizeX - 1) / cfg.LocalSizeX, 1, 1
}

func TestComputeFallback_UnavailableContextReportsInformativeErrors(t *testing.T) {
	ctx := &VulkanContext{available: false}

	_, err := ctx.CreateFence()
	if err == nil {
		t.Fatal("expected CreateFence to fail on unavailable context")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not available") {
		t.Fatalf("CreateFence error should mention availability, got: %q", err.Error())
	}

	err = ctx.WaitForFence(nil, 0)
	if err == nil {
		t.Fatal("expected WaitForFence to fail on unavailable context")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not available") {
		t.Fatalf("WaitForFence error should mention availability, got: %q", err.Error())
	}

	err = ctx.ResetFence(nil)
	if err == nil {
		t.Fatal("expected ResetFence to fail on unavailable context")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not available") {
		t.Fatalf("ResetFence error should mention availability, got: %q", err.Error())
	}
}

func TestComputeFallback_ContextInitFailurePathIsGraceful(t *testing.T) {
	ctx, err := NewVulkanContext()
	if err != nil {
		t.Fatalf("NewVulkanContext returned unexpected error: %v", err)
	}
	if ctx == nil {
		t.Fatal("NewVulkanContext returned nil context")
	}

	// No assumptions about host GPU presence. We only assert no crashes and stable API behavior.
	ctx.Destroy()
	if ctx.IsAvailable() {
		t.Fatal("context should report unavailable after Destroy")
	}
}

func TestComputeFallback_CPUReferenceMatchesDispatchComputation(t *testing.T) {
	cfg := WorkgroupConfig{LocalSizeX: 256, LocalSizeY: 1, LocalSizeZ: 1}

	cases := []uint32{0, 1, 17, 255, 256, 257, 1024, 1000000}
	for _, total := range cases {
		wantX, wantY, wantZ := calculateDispatchSizeCPUReference(total, cfg)
		gotX, gotY, gotZ := CalculateDispatchSize(total, cfg)
		if gotX != wantX || gotY != wantY || gotZ != wantZ {
			t.Fatalf("dispatch mismatch for total=%d: got (%d,%d,%d), want (%d,%d,%d)",
				total, gotX, gotY, gotZ, wantX, wantY, wantZ)
		}
	}
}

func TestComputeFallback_NoPanicsWhenGPUUnavailable(t *testing.T) {
	ctx := &VulkanContext{available: false}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic in unavailable GPU path: %v", r)
		}
	}()

	for i := 0; i < 50; i++ {
		_, _ = ctx.CreateFence()
		_ = ctx.WaitForFence(nil, 0)
		_ = ctx.ResetFence(nil)
		ctx.DestroyFence(nil) // should be a no-op
	}
}
