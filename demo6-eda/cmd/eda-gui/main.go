// cmd/eda-gui/main.go
package main

import (
	"fyne.io/fyne/v2/app"

	"demo6-eda/pkg/gui"
)

func main() {
	a := app.New()
	w := gui.CreateMainWindow(a)
	w.ShowAndRun()
}
