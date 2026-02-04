// cmd/eda-gui/main.go
package edagui

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"

	"fecim-lattice-tools/module6-eda/pkg/gui"
	"fecim-lattice-tools/shared/logging"
)

func Run(args []string) error {
	// Initialize logger
	homeDir, _ := os.UserHomeDir()
	logPath := filepath.Join(homeDir, ".fecim", "logs", "module6-eda.log")
	if err := logging.Init("module6-eda-gui", logPath); err != nil {
		panic(err)
	}
	defer logging.CloseGlobal()

	// Enable logging by default
	logging.SetVerbosity(logging.VerbosityInfo)

	logging.GlobalInfo("Starting EDA GUI application")

	a := app.New()
	w := gui.CreateMainWindow(a)
	w.ShowAndRun()
	return nil
}
