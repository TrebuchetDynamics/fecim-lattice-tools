// pe-reference performs offline conversion of source-qualified P-E reference data.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"fecim-lattice-tools/validation/peloop"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "pe-reference:", err)
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer) error {
	if len(args) > 0 && args[0] == "compare" {
		return runCompare(args[1:], stderr)
	}
	if len(args) == 0 || args[0] != "digitize" {
		return fmt.Errorf("usage: pe-reference digitize --profile PATH --pixels PATH --out PATH; pe-reference compare --profile PATH --data PATH --out-dir PATH")
	}
	flags := flag.NewFlagSet("digitize", flag.ContinueOnError)
	flags.SetOutput(stderr)
	profilePath := flags.String("profile", "", "local source-profile JSON")
	pixelsPath := flags.String("pixels", "", "local original-raster pixel CSV (at most 2MiB and 10000 points)")
	outPath := flags.String("out", "", "new physical-unit CSV path (must not exist)")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 || *profilePath == "" || *pixelsPath == "" || *outPath == "" {
		return fmt.Errorf("digitize requires --profile, --pixels and --out paths, with no positional arguments")
	}
	profileFile, err := os.Open(*profilePath)
	if err != nil {
		return fmt.Errorf("profile: %w", err)
	}
	defer profileFile.Close()
	profile, err := peloop.ReadSourceProfile(profileFile)
	if err != nil {
		return err
	}
	pixelsFile, err := os.Open(*pixelsPath)
	if err != nil {
		return fmt.Errorf("pixels: %w", err)
	}
	defer pixelsFile.Close()
	data, err := peloop.Digitize(profile, pixelsFile)
	if err != nil {
		return err
	}
	// Validate everything before creating output. O_EXCL protects existing files
	// (including inputs and symlinks), without a check-then-create race.
	out, err := os.OpenFile(*outPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("output: %w", err)
	}
	_, writeErr := out.Write(data)
	closeErr := out.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(*outPath)
		if writeErr != nil {
			return fmt.Errorf("output: %w", writeErr)
		}
		return fmt.Errorf("output: %w", closeErr)
	}
	return nil
}
