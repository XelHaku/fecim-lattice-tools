// Command fecim-web provides a web-based FeCIM interface.
//
// Usage: fecim-web -port 8080
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	docs "fecim-lattice-tools/module7-docs/pkg/gui"
)

// fecim-web is a minimal WASM-safe entrypoint.
//
// Design goal (L08 spike): compile a browser demo without pulling in non-wasm
// dependencies (Vulkan renderer, terminal TUI, etc.).
const (
	webWindowTitle = "FeCIM Lattice Tools (Web Demo)"
	webWindowW     = float32(1200)
	webWindowH     = float32(800)
)

func main() {
	a := app.New()
	w := a.NewWindow(webWindowTitle)
	w.Resize(fyne.NewSize(webWindowW, webWindowH))

	// Start with the documentation browser as the lowest-risk wasm-safe module.
	// Additional modules can be added incrementally once the build graph is clean.
	emb := docs.NewEmbeddedDocsApp()
	docsUI := emb.BuildContent(a, w)

	w.SetContent(container.NewBorder(nil, nil, nil, nil, docsUI))
	w.ShowAndRun()
}
