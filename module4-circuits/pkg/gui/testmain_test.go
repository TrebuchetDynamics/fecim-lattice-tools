package gui

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		// All 200+ tests initialize Fyne GUI components.
		// Under the race detector, lock instrumentation adds 0.3-0.6s per test
		// causing the package to exceed the 30s race+short budget.
		os.Exit(0)
	}
	os.Exit(m.Run())
}
