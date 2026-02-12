// cmd/lattice-gen/main.go
// CLI tool for generating FeCIM lattice Verilog and DEF files

package edalattice

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"fecim-lattice-tools/module6-eda/pkg/export"
	"fecim-lattice-tools/shared/logging"
)

func Run(args []string) error {
	// Initialize logger
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve user home dir for logging: %w", err)
	}
	logPath := filepath.Join(homeDir, ".fecim", "logs", "module6-eda-lattice-gen.log")
	if err := logging.Init("module6-eda-lattice-gen", logPath); err != nil {
		// Fallback to standard error if logger init fails
		os.Stderr.WriteString("Failed to initialize logging: " + err.Error() + "\n")
		return err
	}
	defer logging.CloseGlobal()

	// Enable logging by default
	logging.SetVerbosity(logging.VerbosityInfo)

	fs := flag.NewFlagSet("lattice-gen", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	rows := fs.Int("rows", 4, "Number of rows")
	cols := fs.Int("cols", 4, "Number of columns")
	outputDir := fs.String("output", "output/lattices", "Output directory")
	help := fs.Bool("help", false, "Show help")
	helpShort := fs.Bool("h", false, "Show help (shorthand)")

	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintln(out, "FeCIM Lattice Generator")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  fecim-lattice-tools eda lattice-gen [options]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Options:")
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

	if err := export.ValidateLatticeDimensions(*rows, *cols); err != nil {
		return fmt.Errorf("invalid lattice dimensions: %w", err)
	}

	// Create output directory if needed
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logging.GlobalError("Error creating output dir: %v\n", err)
		return err
	}

	// Generate files
	verilogPath, err := export.WriteLatticeVerilog(*rows, *cols, *outputDir)
	if err != nil {
		logging.GlobalError("Error writing Verilog: %v\n", err)
		return err
	}
	logging.Printf("Generated: %s\n", verilogPath)

	defPath, err := export.WriteLatticeDEF(*rows, *cols, *outputDir)
	if err != nil {
		logging.GlobalError("Error writing DEF: %v\n", err)
		return err
	}
	logging.Printf("Generated: %s\n", defPath)

	logging.Printf("\nLattice %dx%d generated successfully (%d cells)\n", *rows, *cols, (*rows)*(*cols))
	return nil
}
