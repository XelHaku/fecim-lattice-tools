// Command demo-frames generates animated frame sequences for hysteresis visualization.
//
// Usage: demo-frames -output frames/
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	docs "fecim-lattice-tools/module7-docs/pkg/gui"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// demo-frames captures a small, reproducible set of UI frames for documentation/demo purposes.
// It is intentionally minimal: it avoids app main(), avoids long-running goroutines, and uses
// fyne/test to create an offscreen window.
func main() {
	outDir := flag.String("out", "docs/demo/frames", "output directory")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatalf("mkdir %s: %v", *outDir, err)
	}

	// Module 7 (Docs) frame.
	app := test.NewApp()
	w := app.NewWindow("Docs")
	w.Resize(fyne.NewSize(1200, 800))

	emb := docs.NewEmbeddedDocsApp()
	content := emb.BuildContent(app, w)
	w.SetContent(content)
	w.Show()

	img := w.Canvas().Capture()
	outPath := filepath.Join(*outDir, "frame_007_docs.png")
	f, err := os.Create(outPath)
	if err != nil {
		fatalf("create %s: %v", outPath, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fatalf("encode png: %v", err)
	}

	fmt.Printf("wrote %s\n", outPath)
}

func fatalf(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
