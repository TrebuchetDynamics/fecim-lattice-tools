package testutil

import (
	"runtime"
	"time"
)

type goroutineLeakTB interface {
	Helper()
	Fatalf(format string, args ...any)
}

// AssertNoGoroutineLeak runs exercise and waits for the process goroutine count
// to return to the pre-exercise baseline plus maxExtra. It is intended for
// lifecycle tests that start and stop short-lived module workers/tickers.
func AssertNoGoroutineLeak(t goroutineLeakTB, maxExtra int, exercise func()) {
	t.Helper()
	baseline := settledGoroutineCount()

	exercise()

	deadline := time.Now().Add(2 * time.Second)
	last := runtime.NumGoroutine()
	for time.Now().Before(deadline) {
		last = settledGoroutineCount()
		if last <= baseline+maxExtra {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("goroutine leak suspected: baseline=%d maxExtra=%d current=%d", baseline, maxExtra, last)
}

func settledGoroutineCount() int {
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	runtime.GC()
	return runtime.NumGoroutine()
}
