//go:build legacy_fyne

// Demo 4 GUI: Peripheral Circuits for Ferroelectric CIM
//
// This demo visualizes the peripheral circuits required for a complete
// ferroelectric compute-in-memory system: DAC, ADC, TIA, and Charge Pump.
// Shows how digital values are converted to/from analog for crossbar operations.
package circuitsgui

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"fecim-lattice-tools/module4-circuits/pkg/gui"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/peripherals"
)

func isVerbosityToken(token string) bool {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case "0", "off", "none", "1", "info", "2", "debug", "3", "trace", "all":
		return true
	default:
		return false
	}
}

func Run(args []string) error {
	fs := flag.NewFlagSet("circuits-gui", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	loggerFlag := fs.Bool("logger", false, "Enable file logging (logs/). Optional shorthand: --logger debug|info|trace|off")
	verbosityFlag := fs.String("verbosity", "info", "Logging verbosity: 0|off, 1|info, 2|debug, 3|trace (only used with --logger)")
	help := fs.Bool("help", false, "Show help")
	helpShort := fs.Bool("h", false, "Show help (shorthand)")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintln(out, "FeCIM Circuits GUI")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  fecim-lattice-tools circuits [gui flags]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "GUI flags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(fs.Output(), "Error:", err)
		fs.Usage()
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if *help || *helpShort {
		fs.Usage()
		return nil
	}

	verbosityProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "verbosity" {
			verbosityProvided = true
		}
	})
	if *loggerFlag && !verbosityProvided {
		if rest := fs.Args(); len(rest) > 0 && isVerbosityToken(rest[0]) {
			*verbosityFlag = rest[0]
		}
	}

	if *loggerFlag {
		logging.EnableFileLogging()
		verbosity := logging.ParseVerbosityFlag(*verbosityFlag)
		logging.SetVerbosity(verbosity)
		log := logging.NewLogger("circuits-gui")
		defer log.Close()
		log.Info("Circuits GUI starting with verbosity=%s", logging.VerbosityString(verbosity))

		peripherals.EnableLogging()
		gui.EnableComputeLog(true)
	}

	app := gui.NewCircuitsApp()
	app.Run()
	return nil
}
