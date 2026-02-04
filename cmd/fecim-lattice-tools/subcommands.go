package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	hysteresiscli "fecim-lattice-tools/module1-hysteresis/cmd/hysteresis"
	crossbarcmd "fecim-lattice-tools/module2-crossbar/cmd/crossbar-gui"
	mnistcli "fecim-lattice-tools/module3-mnist/cmd/mnist"
	mnistgui "fecim-lattice-tools/module3-mnist/cmd/mnist-gui"
	mnisttrain "fecim-lattice-tools/module3-mnist/cmd/train-network"
	mnisttrainptq "fecim-lattice-tools/module3-mnist/cmd/train-ptq"
	mnisttrainsingle "fecim-lattice-tools/module3-mnist/cmd/train-single-layer"
	circuitscli "fecim-lattice-tools/module4-circuits/cmd/circuits"
	circuitsgui "fecim-lattice-tools/module4-circuits/cmd/circuits-gui"
	comparisoncli "fecim-lattice-tools/module5-comparison/cmd/comparison"
	comparisongui "fecim-lattice-tools/module5-comparison/cmd/comparison-gui"
	edacli "fecim-lattice-tools/module6-eda/cmd/eda-cli"
	edagui "fecim-lattice-tools/module6-eda/cmd/eda-gui"
	edahello "fecim-lattice-tools/module6-eda/cmd/hello"
	edalattice "fecim-lattice-tools/module6-eda/cmd/lattice-gen"
)

func maybeDispatchSubcommand(args []string) bool {
	if len(args) == 0 {
		return false
	}

	first := args[0]
	if first == "-h" || first == "--help" || first == "help" {
		printRootUsage(os.Stdout)
		return true
	}

	if strings.HasPrefix(first, "-") {
		return false
	}

	if err := dispatchSubcommand(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		printRootUsage(os.Stderr)
		os.Exit(1)
	}
	return true
}

func dispatchSubcommand(args []string) error {
	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "hysteresis":
		return hysteresiscli.Run(args[1:])
	case "crossbar":
		return runCrossbarSubcommand(args[1:])
	case "mnist":
		return runMNISTSubcommand(args[1:])
	case "circuits":
		return runCircuitsSubcommand(args[1:])
	case "comparison":
		return runComparisonSubcommand(args[1:])
	case "eda":
		return runEDASubcommand(args[1:])
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func runCrossbarSubcommand(args []string) error {
	if len(args) == 0 || args[0] == "gui" {
		return crossbarcmd.RunGUI(args[cmdSkip(args):])
	}

	switch args[0] {
	case "inference":
		return crossbarcmd.RunInference(args[1:])
	default:
		return fmt.Errorf("unknown crossbar subcommand %q", args[0])
	}
}

func runMNISTSubcommand(args []string) error {
	if len(args) == 0 || args[0] == "gui" {
		return mnistgui.Run(args[cmdSkip(args):])
	}

	switch args[0] {
	case "cli":
		return mnistcli.Run(args[1:])
	case "train-network":
		return mnisttrain.Run(args[1:])
	case "train-single-layer":
		return mnisttrainsingle.Run(args[1:])
	case "train-ptq":
		return mnisttrainptq.Run(args[1:])
	default:
		return fmt.Errorf("unknown mnist subcommand %q", args[0])
	}
}

func runCircuitsSubcommand(args []string) error {
	if len(args) == 0 || args[0] == "gui" {
		return circuitsgui.Run(args[cmdSkip(args):])
	}

	switch args[0] {
	case "cli":
		return circuitscli.Run(args[1:])
	default:
		return fmt.Errorf("unknown circuits subcommand %q", args[0])
	}
}

func runComparisonSubcommand(args []string) error {
	if len(args) == 0 || args[0] == "gui" {
		return comparisongui.Run(args[cmdSkip(args):])
	}

	switch args[0] {
	case "cli":
		return comparisoncli.Run(args[1:])
	default:
		return fmt.Errorf("unknown comparison subcommand %q", args[0])
	}
}

func runEDASubcommand(args []string) error {
	if len(args) == 0 || args[0] == "gui" {
		return edagui.Run(args[cmdSkip(args):])
	}

	switch args[0] {
	case "cli":
		return edacli.Run(args[1:])
	case "lattice-gen":
		return edalattice.Run(args[1:])
	case "hello":
		return edahello.Run(args[1:])
	default:
		return fmt.Errorf("unknown eda subcommand %q", args[0])
	}
}

func cmdSkip(args []string) int {
	if len(args) == 0 {
		return 0
	}
	if args[0] == "gui" {
		return 1
	}
	return 0
}

func printRootUsage(w io.Writer) {
	fmt.Fprintln(w, "FeCIM Lattice Tools")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  fecim-lattice-tools [gui flags]")
	fmt.Fprintln(w, "  fecim-lattice-tools <subcommand> [options]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Subcommands:")
	fmt.Fprintln(w, "  hysteresis         Hysteresis demo (GUI/TUI/headless via flags)")
	fmt.Fprintln(w, "  crossbar           Crossbar GUI (default) or inference CLI")
	fmt.Fprintln(w, "  mnist              MNIST GUI (default) or CLI/training tools")
	fmt.Fprintln(w, "  circuits           Circuits GUI (default) or CLI")
	fmt.Fprintln(w, "  comparison         Comparison GUI (default) or CLI")
	fmt.Fprintln(w, "  eda                EDA GUI (default) or CLI tools")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  fecim-lattice-tools")
	fmt.Fprintln(w, "  fecim-lattice-tools crossbar inference -size=64 -show-mvm")
	fmt.Fprintln(w, "  fecim-lattice-tools mnist cli -evaluate")
	fmt.Fprintln(w, "  fecim-lattice-tools circuits cli -all")
	fmt.Fprintln(w, "  fecim-lattice-tools eda cli -mode=compute -rows=128 -cols=128")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "GUI flags:")
	fmt.Fprintln(w, "  --logger [LEVEL]    Enable file logging (debug|info|trace|off)")
	fmt.Fprintln(w, "  --verbosity LEVEL   Logging verbosity (0|off,1|info,2|debug,3|trace)")
	fmt.Fprintln(w, "  --module NAME       Start module: home,hysteresis,crossbar,mnist,circuits,comparison,eda,docs")
	fmt.Fprintln(w, "  --calibrate         Run hysteresis calibration and exit")
	fmt.Fprintln(w, "  --list-materials    List available materials and exit")
}
