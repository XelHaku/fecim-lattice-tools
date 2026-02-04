// Demo 4 GUI: Peripheral Circuits for Ferroelectric CIM
//
// This demo visualizes the peripheral circuits required for a complete
// ferroelectric compute-in-memory system: DAC, ADC, TIA, and Charge Pump.
// Shows how digital values are converted to/from analog for crossbar operations.
package circuitsgui

import (
	"fecim-lattice-tools/module4-circuits/pkg/gui"
)

func Run(args []string) error {
	app := gui.NewCircuitsApp()
	app.Run()
	return nil
}
