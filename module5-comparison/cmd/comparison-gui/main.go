// Demo 8: Architecture Comparison GUI
//
// This demo provides an interactive comparison of FeCIM vs traditional
// CPU and GPU architectures for neural network inference.
//
// IMPORTANT: FeCIM specifications are model inputs and are NOT
// independently verified. CPU/GPU specs are from published datasheets.
// This is TRL 4 technology - lab validation only.
package comparisongui

import (
	"fecim-lattice-tools/module5-comparison/pkg/gui"
)

func Run(args []string) error {
	app := gui.NewComparisonApp()
	app.Run()
	return nil
}
