//go:build !cgo

package render

import (
	"fmt"
	"os"

	"github.com/gogpu/gg"
)

// CaptureContext saves a gg.Context pixmap as a PNG file.
func CaptureContext(cc *gg.Context, filename string) error {
	if cc == nil {
		return fmt.Errorf("render: nil context")
	}
	return cc.SavePNG(filename)
}

// CaptureAndSave renders to a standalone context and saves as PNG.
// width/height define the output resolution.
func CaptureAndSave(width, height int, draw func(*gg.Context), filename string) error {
	dc := gg.NewContext(width, height)
	defer dc.Close()
	if draw != nil {
		draw(dc)
	}
	if err := os.MkdirAll("screenshots", 0755); err != nil {
		return fmt.Errorf("render: mkdir screenshots: %w", err)
	}
	path := "screenshots/" + filename
	return dc.SavePNG(path)
}
