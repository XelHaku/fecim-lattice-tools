package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	crossbarcli "fecim-lattice-tools/module2-crossbar/cmd/crossbar"
	mnistcli "fecim-lattice-tools/module3-mnist/cmd/mnist"
	mnisttrain "fecim-lattice-tools/module3-mnist/cmd/train-network"
	mnisttrainptq "fecim-lattice-tools/module3-mnist/cmd/train-ptq"
	mnisttrainsingle "fecim-lattice-tools/module3-mnist/cmd/train-single-layer"
	circuitscli "fecim-lattice-tools/module4-circuits/cmd/circuits"
	comparisoncli "fecim-lattice-tools/module5-comparison/cmd/comparison"
	edacli "fecim-lattice-tools/module6-eda/cmd/eda-cli"
	edahello "fecim-lattice-tools/module6-eda/cmd/hello"
	edalattice "fecim-lattice-tools/module6-eda/cmd/lattice-gen"
	"fecim-lattice-tools/shared/viewmodel"
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
		return runHysteresisSubcommand(args[1:])
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
	case "research":
		return runResearchSubcommand(args[1:])
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func runHysteresisSubcommand(args []string) error {
	if len(args) == 0 || args[0] == "gui" {
		return runModuleApp(viewmodel.ModuleHysteresis)
	}
	fs := flag.NewFlagSet("hysteresis", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	engine := fs.String("engine", "preisach", "Headless engine: preisach|lk")
	headless := fs.Bool("headless", false, "Run headless hysteresis mode")
	listMaterials := fs.Bool("list-materials", false, "List available materials and exit")
	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintln(out, "FeCIM Hysteresis")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  fecim-lattice-tools hysteresis gui")
		fmt.Fprintln(out, "  fecim-lattice-tools hysteresis --headless --engine lk")
		fmt.Fprintln(out)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if *listMaterials {
		return runRoot([]string{"--list-materials"})
	}
	if *headless {
		return runMode("hysteresis", *engine)
	}
	return runModuleApp(viewmodel.ModuleHysteresis)
}

func runCrossbarSubcommand(args []string) error {
	if len(args) == 0 || args[0] == "gui" {
		return runModuleApp(viewmodel.ModuleCrossbar)
	}
	switch args[0] {
	case "inference":
		return crossbarcli.RunInference(args[1:])
	default:
		return fmt.Errorf("unknown crossbar subcommand %q", args[0])
	}
}

func runMNISTSubcommand(args []string) error {
	if len(args) == 0 || args[0] == "gui" {
		return runModuleApp(viewmodel.ModuleMNIST)
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
	if len(args) == 0 || args[0] == "gui" || strings.HasPrefix(args[0], "-") {
		return runModuleApp(viewmodel.ModuleCircuits)
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
		return runModuleApp(viewmodel.ModuleComparison)
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
		return runModuleApp(viewmodel.ModuleEDA)
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
	fmt.Fprintln(w, "The default GUI is the zero-CGO gogpu/ui shell.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Subcommands:")
	fmt.Fprintln(w, "  hysteresis         gogpu/ui module or headless mode")
	fmt.Fprintln(w, "  crossbar           gogpu/ui module or inference CLI")
	fmt.Fprintln(w, "  mnist              gogpu/ui module or CLI/training tools")
	fmt.Fprintln(w, "  circuits           gogpu/ui module or CLI")
	fmt.Fprintln(w, "  comparison         gogpu/ui module or CLI")
	fmt.Fprintln(w, "  eda                gogpu/ui module or CLI tools")
	fmt.Fprintln(w, "  research           paper acquisition, ingestion, indexing, and evidence search")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  fecim-lattice-tools --module crossbar")
	fmt.Fprintln(w, "  fecim-lattice-tools --mode hysteresis --engine lk")
	fmt.Fprintln(w, "  fecim-lattice-tools crossbar inference -size=64 -show-mvm")
	fmt.Fprintln(w, "  fecim-lattice-tools circuits cli -all")
	fmt.Fprintln(w, "  fecim-lattice-tools eda cli -mode=compute -rows=128 -cols=128")
	fmt.Fprintln(w, "  fecim-lattice-tools research acquire --download")
	fmt.Fprintln(w, "  fecim-lattice-tools research acquire --doi 10.1002/adma.201404531 --download")
	fmt.Fprintln(w, "  fecim-lattice-tools research audit")
	fmt.Fprintln(w, "  fecim-lattice-tools research ingest")
	fmt.Fprintln(w, "  fecim-lattice-tools research search \"HZO coercive field Preisach\"")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "GUI flags:")
	fmt.Fprintln(w, "  --module NAME       Start module: home,hysteresis,crossbar,mnist,circuits,comparison,eda,docs")
	fmt.Fprintln(w, "  --logger [LEVEL]    Enable file logging (debug|info|trace|off)")
	fmt.Fprintln(w, "  --verbosity LEVEL   Logging verbosity (0|off,1|info,2|debug,3|trace)")
	fmt.Fprintln(w, "  --calibrate         Run hysteresis calibration and exit")
	fmt.Fprintln(w, "  --list-materials    List available materials and exit")
}
