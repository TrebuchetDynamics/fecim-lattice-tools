package tabs

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		// All tests initialize Fyne GUI components. Under the race detector,
		// concurrent font-cache accesses in go-text/typesetting trigger data
		// races (Fyne v2.7.4 + go-text/typesetting v0.3.4 dependency bug).
		os.Exit(0)
	}
	os.Exit(m.Run())
}
