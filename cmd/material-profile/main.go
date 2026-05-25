// Command material-profile exports material physics parameters as JSON.
//
// Usage: material-profile -material fecim_hzo -output profile.json
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"fecim-lattice-tools/shared/physics"
)

func run(out, errOut io.Writer, profile, mode, sep string) int {
	switch mode {
	case "version":
		fmt.Fprintln(out, physics.MaterialProfileVersion)
		return 0
	case "list":
		mats, err := physics.RequiredMaterialsForProfile(physics.MaterialProfileName(profile))
		if err != nil {
			fmt.Fprintln(errOut, err)
			return 2
		}
		fmt.Fprint(out, strings.Join(mats, sep))
		return 0
	default:
		fmt.Fprintf(errOut, "unknown mode %q\n", mode)
		return 2
	}
}

func main() {
	os.Exit(runMaterialProfile(os.Args[1:], os.Stdout, os.Stderr))
}

func runMaterialProfile(args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("material-profile", flag.ContinueOnError)
	flags.SetOutput(errOut)
	profile := flags.String("profile", "pr", "material profile: pr|nightly")
	mode := flags.String("mode", "list", "mode: list|version")
	sep := flags.String("sep", "\n", "separator for list output")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	return run(out, errOut, *profile, *mode, *sep)
}
