package simulation

import (
	"sync"
	"testing"
	"time"

	"ironlattice-vis/demo1-hysteresis/pkg/ferroelectric"
)

// TestEngineStartStop verifies thread-safe start/stop operations
func TestEngineStartStop(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	if engine.IsRunning() {
		t.Error("Engine should not be running initially")
	}

	engine.Start()
	if !engine.IsRunning() {
		t.Error("Engine should be running after Start()")
	}

	engine.Stop()
	if engine.IsRunning() {
		t.Error("Engine should not be running after Stop()")
	}
}

// TestEnginePause verifies thread-safe pause operations
func TestEnginePause(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	engine.Start()

	if engine.IsPaused() {
		t.Error("Engine should not be paused after Start()")
	}

	engine.Pause()
	if !engine.IsPaused() {
		t.Error("Engine should be paused after Pause()")
	}

	engine.Pause()
	if engine.IsPaused() {
		t.Error("Engine should not be paused after second Pause()")
	}
}

// TestEngineConcurrentAccess verifies no data races under concurrent access
func TestEngineConcurrentAccess(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	var wg sync.WaitGroup
	const goroutines = 10
	const iterations = 100

	// Start the engine
	engine.Start()

	// Spawn multiple goroutines doing concurrent operations
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				engine.IsRunning()
				engine.IsPaused()
				engine.Step()
			}
		}()
	}

	// Also do pause/unpause in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < iterations; j++ {
			engine.Pause()
			time.Sleep(time.Microsecond)
		}
	}()

	wg.Wait()
	engine.Stop()

	// If we get here without race detector complaints, we're good
}

// TestEngineStep verifies simulation advances
func TestEngineStep(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	engine.Start()
	initialTime := engine.State().Time

	for i := 0; i < 10; i++ {
		engine.Step()
	}

	if engine.State().Time <= initialTime {
		t.Error("Simulation time should advance after steps")
	}
}

// TestEngineReset verifies reset clears state
func TestEngineReset(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	engine.Start()
	for i := 0; i < 100; i++ {
		engine.Step()
	}

	engine.Reset()

	if engine.State().Time != 0 {
		t.Errorf("Time should be 0 after reset, got %v", engine.State().Time)
	}
}
