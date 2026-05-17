package gogpucommand

import (
	"fmt"
	"os"
)

func Error(name, replacement string) error {
	return fmt.Errorf("%s is served by the canonical gogpu/ui command; use %q", name, replacement)
}

func Exit(name, replacement string) {
	fmt.Fprintln(os.Stderr, Error(name, replacement))
	os.Exit(1)
}
