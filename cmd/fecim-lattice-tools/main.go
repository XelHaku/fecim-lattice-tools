package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fecim-lattice-tools/internal/gogpuapp"
	hysheadless "fecim-lattice-tools/module1-hysteresis/pkg/headless"
	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/viewmodel"
)

func main() {
	os.Exit(runMain(os.Args[1:], os.Stdout, os.Stderr))
}

func runMain(args []string, stdout, stderr io.Writer) int {
	if handled, code := runSubcommandDispatch(args, stdout, stderr); handled {
		return code
	}
	if err := runRoot(args, stdout, stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(stderr, "Error: %v\n", err)
		var coded exitCodeError
		if errors.As(err, &coded) {
			return coded.code
		}
		return 1
	}
	return 0
}

type exitCodeError struct {
	code int
	err  error
}

func (e exitCodeError) Error() string { return e.err.Error() }
func (e exitCodeError) Unwrap() error { return e.err }

func runRoot(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("fecim-lattice-tools", flag.ContinueOnError)
	fs.SetOutput(stderr)
	loggerFlag := fs.Bool("logger", false, "Enable file logging (logs/). Optional shorthand: --logger debug|info|trace|off")
	verbosityFlag := fs.String("verbosity", "info", "Logging verbosity: 0|off, 1|info, 2|debug, 3|trace (only used with --logger)")
	calibrateFlag := fs.Bool("calibrate", false, "Run hysteresis calibration and exit")
	materialFlag := fs.String("material", "all", "Material to calibrate (use 'all' for all materials, or specify name)")
	forceFlag := fs.Bool("force", false, "Force recalibration even if calibration file exists")
	verifyFlag := fs.Bool("verify", false, "Verify calibration accuracy after calibrating")
	listMaterialsFlag := fs.Bool("list-materials", false, "List available materials and exit")
	materialsFlag := fs.Bool("materials", false, "Alias for --list-materials")
	modeFlag := fs.String("mode", "", "Run a headless mode (e.g., hysteresis) and exit")
	engineFlag := fs.String("engine", "", "Headless hysteresis engine for --mode hysteresis: preisach|lk (default: preisach)")
	moduleFlag := fs.String("module", "home", "Start module: home, hysteresis, crossbar, mnist, circuits, comparison, eda, docs")
	_ = fs.String("screenshot-dir", filepath.Clean("screenshots"), "Reserved for gogpu/ui screenshot capture")
	_ = fs.String("recording-dir", filepath.Clean("recordings"), "Reserved for gogpu/ui recording capture")

	fs.Usage = func() { printRootUsage(fs.Output()) }
	if err := fs.Parse(args); err != nil {
		return exitCodeError{code: 2, err: err}
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
	if *materialsFlag {
		*listMaterialsFlag = true
	}
	if *listMaterialsFlag {
		fmt.Fprintln(stdout, "Available materials:")
		for _, name := range hysheadless.ListMaterials() {
			fmt.Fprintf(stdout, "  - %s\n", name)
		}
		return nil
	}
	if *calibrateFlag {
		opts := hysheadless.CLICalibrationOptions{
			MaterialName: *materialFlag,
			Temperature:  300,
			Force:        *forceFlag,
			Verbose:      true,
			Verify:       *verifyFlag,
		}
		if err := hysheadless.RunCLICalibration(opts); err != nil {
			return fmt.Errorf("calibration error: %w", err)
		}
		fmt.Fprintln(stdout, "Calibration complete.")
		return nil
	}

	if *loggerFlag {
		logging.EnableFileLogging()
		verbosity := logging.ParseVerbosityFlag(*verbosityFlag)
		logging.SetVerbosity(verbosity)
		log := logging.NewLogger("fecim-lattice-tools")
		defer log.Close()
		log.Info("FeCIM Lattice Tools starting with verbosity=%s", logging.VerbosityString(verbosity))
	}

	modeName := strings.TrimSpace(*modeFlag)
	if modeName == "preisach" || modeName == "lk" {
		if strings.TrimSpace(*engineFlag) == "" {
			*engineFlag = modeName
		}
		modeName = "hysteresis"
	}
	if modeName != "" {
		return runMode(modeName, *engineFlag)
	}

	return runModuleApp(moduleIDFromFlag(*moduleFlag))
}

func runModuleApp(module viewmodel.ModuleID) error {
	return gogpuapp.Run(gogpuapp.Options{ActiveModuleID: module})
}

func moduleIDFromFlag(raw string) viewmodel.ModuleID {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "home":
		return ""
	case "hysteresis", "module1", "m1":
		return viewmodel.ModuleHysteresis
	case "crossbar", "module2", "m2":
		return viewmodel.ModuleCrossbar
	case "mnist", "module3", "m3":
		return viewmodel.ModuleMNIST
	case "circuits", "module4", "m4":
		return viewmodel.ModuleCircuits
	case "comparison", "module5", "m5":
		return viewmodel.ModuleComparison
	case "eda", "module6", "m6":
		return viewmodel.ModuleEDA
	case "docs", "documentation", "module7", "m7":
		return viewmodel.ModuleDocs
	default:
		return ""
	}
}

func isVerbosityToken(token string) bool {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case "0", "off", "none", "1", "info", "2", "debug", "3", "trace", "all":
		return true
	default:
		return false
	}
}
