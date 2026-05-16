//go:build cgo

package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CreateExportButton creates a consistently-styled export button.
func CreateExportButton(label string, action func(), window fyne.Window) *widget.Button {
	_ = window // reserved for future dialog standardization hooks
	return widget.NewButtonWithIcon(label, iconForExportLabel(label), action)
}

func iconForExportLabel(label string) fyne.Resource {
	lower := strings.ToLower(label)
	switch {
	case strings.Contains(lower, "image") || strings.Contains(lower, "png"):
		return theme.MediaPhotoIcon()
	case strings.Contains(lower, "repro"):
		return theme.StorageIcon()
	default:
		return theme.DocumentSaveIcon()
	}
}

// ExportLogger is a minimal logger interface for export status messages.
// Compatible with *log.Logger, *logging.Logger, and any Printf-style logger.
type ExportLogger interface {
	Printf(format string, v ...interface{})
}

// ShowExportError shows an error dialog and optionally logs the message.
// Pass nil for logger if no logging is needed.
func ShowExportError(window fyne.Window, logger ExportLogger, msg string) {
	if window != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("%s", msg), window)
		})
	}
	if logger != nil {
		logger.Printf("Export error: %s", msg)
	}
}

// ShowExportSuccess shows a success dialog and optionally logs the message.
// Pass nil for logger if no logging is needed.
func ShowExportSuccess(window fyne.Window, logger ExportLogger, msg string) {
	if window != nil {
		fyne.Do(func() {
			dialog.ShowInformation("Export Complete", msg, window)
		})
	}
	if logger != nil {
		logger.Printf("Export complete: %s", msg)
	}
}

// ExportVisualization captures the window canvas and exports it as PNG.
// moduleName is used in the output directory and filename prefix.
func ExportVisualization(window fyne.Window, moduleName string, logger ExportLogger) {
	if window == nil {
		return
	}

	dataDir := filepath.Join("exports", moduleName)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		ShowExportError(window, logger, fmt.Sprintf("Cannot create exports folder: %v", err))
		return
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	exporter := NewExporter(dataDir, fmt.Sprintf("%s-viz_%s", moduleName, timestamp))
	cnv := window.Canvas()
	if cnv == nil {
		ShowExportError(window, logger, "Canvas not available")
		return
	}

	img := cnv.Capture()
	result := exporter.ExportPNG(img)

	if result.Error != nil {
		ShowExportError(window, logger, fmt.Sprintf("Image export failed: %v", result.Error))
		return
	}

	ShowExportSuccess(window, logger, fmt.Sprintf("Image saved:\n• %s", result.FilePath))
}
