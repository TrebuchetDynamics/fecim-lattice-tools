package utils

import (
	"sync"
	"testing"
	"time"
)

func TestSafeGo_NormalExecution(t *testing.T) {
	var wg sync.WaitGroup
	var executed bool

	wg.Add(1)
	SafeGo("test-normal", func() {
		executed = true
		wg.Done()
	})

	wg.Wait()

	if !executed {
		t.Error("SafeGo should execute the provided function")
	}
}

func TestSafeGo_PanicRecovery(t *testing.T) {
	// This test verifies that SafeGo recovers from panics without crashing
	done := make(chan bool, 1)

	SafeGo("test-panic", func() {
		defer func() {
			done <- true
		}()
		panic("intentional panic for testing")
	})

	select {
	case <-done:
		// Success - the goroutine completed (after panic recovery)
	case <-time.After(time.Second):
		t.Error("SafeGo should recover from panic and complete")
	}
}

func TestSafeGo_PanicDoesNotCrash(t *testing.T) {
	// Start multiple goroutines, some of which panic
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		shouldPanic := i%2 == 0
		SafeGo("test-mixed", func() {
			defer wg.Done()
			if shouldPanic {
				panic("test panic")
			}
			mu.Lock()
			successCount++
			mu.Unlock()
		})
	}

	wg.Wait()

	// Half of the goroutines should have succeeded
	if successCount != 5 {
		t.Errorf("expected 5 successful executions, got %d", successCount)
	}
}

func TestSafeGo_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	counter := 0
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		SafeGo("test-concurrent", func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		})
	}

	wg.Wait()

	if counter != 100 {
		t.Errorf("expected counter=100, got %d", counter)
	}
}

func TestSafeGo_NilPanic(t *testing.T) {
	done := make(chan bool, 1)

	SafeGo("test-nil-panic", func() {
		defer func() {
			done <- true
		}()
		var ptr *int
		_ = *ptr // This will panic with nil pointer dereference
	})

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("SafeGo should recover from nil pointer panic")
	}
}

func TestSafeGo_StringPanic(t *testing.T) {
	done := make(chan bool, 1)

	SafeGo("test-string-panic", func() {
		defer func() {
			done <- true
		}()
		panic("string panic message")
	})

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("SafeGo should recover from string panic")
	}
}

func TestSafeGo_ErrorPanic(t *testing.T) {
	done := make(chan bool, 1)

	SafeGo("test-error-panic", func() {
		defer func() {
			done <- true
		}()
		panic(struct{ msg string }{"structured panic"})
	})

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("SafeGo should recover from struct panic")
	}
}
