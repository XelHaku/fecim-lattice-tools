//go:build cgo

package export

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// ExportCanvasAsPNG captures a Fyne canvas and exports it as PNG.
func (e *Exporter) ExportCanvasAsPNG(canvas fyne.Canvas) *ExportResult {
	img := canvas.Capture()
	return e.ExportPNG(img)
}

// ShowExportDialog shows a file save dialog for export.
func ShowExportDialog(window fyne.Window, title string, filter []string, callback func(path string, err error)) {
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			callback("", err)
			return
		}
		if writer == nil {
			callback("", nil)
			return
		}
		path := writer.URI().Path()
		writer.Close()
		callback(path, nil)
	}, window)

	saveDialog.SetFileName(fmt.Sprintf("fecim_export_%s", time.Now().Format("2006-01-02_15-04-05")))
	saveDialog.Show()
}
