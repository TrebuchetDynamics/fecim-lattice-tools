package render

import (
	"errors"
	"testing"
	"time"
)

func TestRendererConfigValidationErrors(t *testing.T) {
	var nilCfg *Config
	if err := nilCfg.Validate(); err == nil {
		t.Fatalf("expected nil config error")
	}

	cfg := DefaultConfig()
	cfg.Height = 0
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid viewport error")
	}

	cfg = DefaultConfig()
	cfg.TargetFPS = -1
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid fps error")
	}
}

func TestRendererInitializeNilAndRecreateState(t *testing.T) {
	var nilRenderer *Renderer
	if err := nilRenderer.Initialize(); err == nil {
		t.Fatalf("expected nil renderer initialize error")
	}

	r := NewRenderer(DefaultConfig())
	r.cell = nil
	r.levels = nil
	if err := r.Initialize(); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	if r.cell == nil || r.levels == nil {
		t.Fatalf("initialize should recreate nil render subcomponents")
	}
}

func TestRendererSettersAndCleanup(t *testing.T) {
	r := NewRenderer(DefaultConfig())
	hp := NewHysteresisPlot(2, 3)
	r.SetHysteresisPlot(hp)
	if r.plot != hp {
		t.Fatalf("plot was not set")
	}

	r.UpdatePolarization(1.0)
	if got := r.GetCurrentLevel(); got != FeCIMLevels-1 {
		t.Fatalf("current level = %d, want %d", got, FeCIMLevels-1)
	}
	if r.GetLevelIndicator() == nil {
		t.Fatalf("level indicator should not be nil")
	}
	if color := r.cell.GetColor(); color.R < 0.79 || color.G > 0.11 || color.B > 0.11 {
		t.Fatalf("cell color not mapped to positive polarization: %+v", color)
	}

	r.running.Store(true)
	r.Cleanup()
	if r.running.Load() || r.initialized {
		t.Fatalf("cleanup should reset runtime flags")
	}
	if r.plot != nil || r.cell != nil || r.levels != nil || r.onUpdate != nil {
		t.Fatalf("cleanup should clear render state")
	}
}

func TestRendererRunErrorPaths(t *testing.T) {
	var nilRenderer *Renderer
	if err := nilRenderer.Run(); err == nil {
		t.Fatalf("expected nil renderer run error")
	}

	r := NewRenderer(&Config{Width: 1, Height: 1, TargetFPS: 0})
	r.initialized = true
	if err := r.Run(); err == nil {
		t.Fatalf("expected config validation error")
	}

	r = NewRenderer(DefaultConfig())
	if err := r.Run(); !errors.Is(err, ErrRendererNotInitialized) {
		t.Fatalf("expected not initialized error, got %v", err)
	}

	r = NewRenderer(DefaultConfig())
	if err := r.Initialize(); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}
	r.running.Store(true)
	if err := r.Run(); !errors.Is(err, ErrRendererAlreadyRunning) {
		t.Fatalf("expected already running error, got %v", err)
	}
}

func TestRendererHeadlessRunLoopNoCallback(t *testing.T) {
	r := NewRenderer(DefaultConfig())
	r.config.TargetFPS = 120
	if err := r.Initialize(); err != nil {
		t.Fatalf("initialize failed: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- r.Run()
	}()

	time.Sleep(20 * time.Millisecond)
	r.Stop()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("run loop did not stop in time")
	}
}

func TestLevelIndicatorVerticesAndBounds(t *testing.T) {
	li := NewLevelIndicator()
	li.SetFromPolarization(99)
	if li.CurrentLevel != FeCIMLevels-1 {
		t.Fatalf("high polarization should clamp to top level")
	}
	li.SetFromPolarization(-99)
	if li.CurrentLevel != 0 {
		t.Fatalf("low polarization should clamp to bottom level")
	}

	li.CurrentLevel = 3
	vertices := li.GetLevelVertices()
	expected := FeCIMLevels*6 + 8 // bars + border line segments
	if len(vertices) != expected {
		t.Fatalf("vertex count = %d, want %d", len(vertices), expected)
	}

	active := [4]float32{li.ActiveColor.R, li.ActiveColor.G, li.ActiveColor.B, li.ActiveColor.A}
	activeCount := 0
	for _, v := range vertices {
		if v.Color == active {
			activeCount++
		}
	}
	if activeCount != 6 {
		t.Fatalf("expected exactly 6 active-color vertices for one active bar, got %d", activeCount)
	}
}
