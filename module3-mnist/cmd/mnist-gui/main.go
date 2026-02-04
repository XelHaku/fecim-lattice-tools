// Demo 3 GUI: MNIST Neural Network Visualization with Fyne
//
// This provides an interactive GUI for visualizing MNIST digit classification
// using ferroelectric crossbar arrays with 30 discrete analog states.
//
// Features:
// - Interactive digit drawing canvas (28x28)
// - Real-time neural network inference
// - Layer activation visualization (input, hidden, output)
// - Confusion matrix with per-class metrics
// - Precision, recall, F1 score display
//
// Run with: go run ./cmd/fecim-lattice-tools mnist
package mnistgui

import (
	"fecim-lattice-tools/module3-mnist/pkg/gui"
)

func Run(args []string) error {
	app := gui.NewMNISTApp()
	app.Run()
	return nil
}
