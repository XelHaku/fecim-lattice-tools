// cmd/eda-gui/main.go
package edagui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"

	"fecim-lattice-tools/module6-eda/pkg/gui"
	"fecim-lattice-tools/shared/logging"
)

func Run(args []string) error {
	// Initialize logger with graceful fallback
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir unavailable
		homeDir = "."
	}
	
	logPath := filepath.Join(homeDir, ".fecim", "logs", "module6-eda.log")
	
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Log initialization failure is not fatal - continue without file logging
		fmt.Fprintf(os.Stderr, "Warning: Could not create log directory %s: %v\n", logDir, err)
		fmt.Fprintf(os.Stderr, "Continuing without file logging\n")
	} else if err := logging.Init("module6-eda-gui", logPath); err != nil {
		// Log initialization failure is not fatal - continue without file logging
		fmt.Fprintf(os.Stderr, "Warning: Could not initialize logging to %s: %v\n", logPath, err)
		fmt.Fprintf(os.Stderr, "Continuing without file logging\n")
	} else {
		defer logging.CloseGlobal()
	}

	// Enable logging by default
	logging.SetVerbosity(logging.VerbosityInfo)

	logging.GlobalInfo("Starting EDA GUI application")

	a := app.New()
	w := gui.CreateMainWindow(a)
	w.ShowAndRun()
	return nil
}
