package legacycommand

import (
	"fmt"
	"io"
	"os"
)

func Error(name, replacement, legacy string) error {
	return fmt.Errorf("%s is a legacy Fyne entrypoint and is not available from this non-legacy command; use %q or %q", name, replacement, legacy)
}

func Exit(name, replacement, legacy string) {
	fmt.Fprintln(os.Stderr, Error(name, replacement, legacy))
	os.Exit(1)
}

func WriteUsage(w io.Writer, name, replacement, legacy string) {
	fmt.Fprintln(w, Error(name, replacement, legacy))
}
