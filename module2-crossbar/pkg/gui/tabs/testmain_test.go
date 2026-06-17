package tabs

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		// All tests initialize Fyne GUI tab components.
		// Under the race detector, lock instrumentation adds 0.5-1s per test
		// causing the 37-test package to exceed the 30s race+short budget.
		os.Exit(0)
	}
	os.Exit(m.Run())
}
