//go:build cgo

// Package export provides unified data export utilities for FeCIM tools.
// This file provides Fyne widgets for export UI integration.
package export

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ExportDataProvider is an interface that modules implement to provide exportable data
type ExportDataProvider interface {
	// GetCSVData returns the simulation data for CSV export
	GetCSVData() (*CSVData, error)
	// GetJSONConfig returns the configuration for JSON export
	GetJSONConfig() (interface{}, error)
	// GetVisualization returns the current visualization as an image
	GetVisualization() (image.Image, error)
}

// ExportButtonsConfig configures which export buttons to show
type ExportButtonsConfig struct {
	EnableCSV  bool
	EnableJSON bool
	EnablePNG  bool
	OutputDir  string
	FilePrefix string
	ModuleName string
}

// DefaultExportConfig returns a default export configuration
func DefaultExportConfig(moduleName string) *ExportButtonsConfig {
	return &ExportButtonsConfig{
		EnableCSV:  true,
		EnableJSON: true,
		EnablePNG:  true,
		OutputDir:  "exports",
		FilePrefix: moduleName,
		ModuleName: moduleName,
	}
}

// ExportButtons provides a container with export buttons for a module
type ExportButtons struct {
	widget.BaseWidget

	config   *ExportButtonsConfig
	provider ExportDataProvider
	window   fyne.Window

	csvBtn  *widget.Button
	jsonBtn *widget.Button
	pngBtn  *widget.Button

	statusLabel *widget.Label
	container   *fyne.Container
}

// NewExportButtons creates a new export buttons widget
func NewExportButtons(config *ExportButtonsConfig, provider ExportDataProvider, window fyne.Window) *ExportButtons {
	eb := &ExportButtons{
		config:   config,
		provider: provider,
		window:   window,
	}

	eb.ExtendBaseWidget(eb)
	eb.buildUI()

	return eb
}

// buildUI constructs the export buttons UI
func (eb *ExportButtons) buildUI() {
	eb.statusLabel = widget.NewLabel("")
	eb.statusLabel.Wrapping = fyne.TextTruncate

	var buttons []fyne.CanvasObject

	if eb.config.EnableCSV {
		eb.csvBtn = widget.NewButtonWithIcon("Export CSV", theme.DocumentSaveIcon(), eb.exportCSV)
		buttons = append(buttons, eb.csvBtn)
	}

	if eb.config.EnableJSON {
		eb.jsonBtn = widget.NewButtonWithIcon("Export Config", theme.FileIcon(), eb.exportJSON)
		buttons = append(buttons, eb.jsonBtn)
	}

	if eb.config.EnablePNG {
		eb.pngBtn = widget.NewButtonWithIcon("Save Image", theme.MediaPhotoIcon(), eb.exportPNG)
		buttons = append(buttons, eb.pngBtn)
	}

	buttonsRow := container.NewHBox(buttons...)
	eb.container = container.NewVBox(
		buttonsRow,
		eb.statusLabel,
	)
}

// CreateRenderer implements fyne.Widget
func (eb *ExportButtons) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(eb.container)
}

// exportCSV handles CSV export button click
func (eb *ExportButtons) exportCSV() {
	if eb.provider == nil {
		eb.showError("No data provider configured")
		return
	}

	data, err := eb.provider.GetCSVData()
	if err != nil {
		eb.showError(fmt.Sprintf("Failed to get CSV data: %v", err))
		return
	}

	if data == nil || len(data.Rows) == 0 {
		eb.showError("No data to export")
		return
	}

	exporter := NewExporter(eb.config.OutputDir, eb.config.FilePrefix+"-data")
	result := exporter.ExportCSV(data.Headers, data.Rows)

	if result.Error != nil {
		eb.showError(fmt.Sprintf("Export failed: %v", result.Error))
		return
	}

	eb.showSuccess(fmt.Sprintf("CSV exported: %s (%d rows)", result.FilePath, len(data.Rows)))
}

// exportJSON handles JSON export button click
func (eb *ExportButtons) exportJSON() {
	if eb.provider == nil {
		eb.showError("No data provider configured")
		return
	}

	config, err := eb.provider.GetJSONConfig()
	if err != nil {
		eb.showError(fmt.Sprintf("Failed to get config: %v", err))
		return
	}

	exporter := NewExporter(eb.config.OutputDir, eb.config.FilePrefix+"-config")
	result := exporter.ExportJSON(config)

	if result.Error != nil {
		eb.showError(fmt.Sprintf("Export failed: %v", result.Error))
		return
	}

	eb.showSuccess(fmt.Sprintf("Config exported: %s", result.FilePath))
}

// exportPNG handles PNG export button click
func (eb *ExportButtons) exportPNG() {
	if eb.provider == nil {
		eb.showError("No data provider configured")
		return
	}

	img, err := eb.provider.GetVisualization()
	if err != nil {
		eb.showError(fmt.Sprintf("Failed to capture visualization: %v", err))
		return
	}

	exporter := NewExporter(eb.config.OutputDir, eb.config.FilePrefix+"-viz")
	result := exporter.ExportPNG(img)

	if result.Error != nil {
		eb.showError(fmt.Sprintf("Export failed: %v", result.Error))
		return
	}

	eb.showSuccess(fmt.Sprintf("Image saved: %s", result.FilePath))
}

// showError displays an error message
func (eb *ExportButtons) showError(msg string) {
	eb.statusLabel.SetText("❌ " + msg)
	if eb.window != nil {
		dialog.ShowError(fmt.Errorf("%s", msg), eb.window)
	}
}

// showSuccess displays a success message
func (eb *ExportButtons) showSuccess(msg string) {
	eb.statusLabel.SetText("✓ " + msg)
	if eb.window != nil {
		dialog.ShowInformation("Export Complete", msg, eb.window)
	}
}

// SetProvider updates the data provider
func (eb *ExportButtons) SetProvider(provider ExportDataProvider) {
	eb.provider = provider
}

// SetEnabled enables or disables all export buttons
func (eb *ExportButtons) SetEnabled(enabled bool) {
	if eb.csvBtn != nil {
		if enabled {
			eb.csvBtn.Enable()
		} else {
			eb.csvBtn.Disable()
		}
	}
	if eb.jsonBtn != nil {
		if enabled {
			eb.jsonBtn.Enable()
		} else {
			eb.jsonBtn.Disable()
		}
	}
	if eb.pngBtn != nil {
		if enabled {
			eb.pngBtn.Enable()
		} else {
			eb.pngBtn.Disable()
		}
	}
}

// ClearStatus clears the status message
func (eb *ExportButtons) ClearStatus() {
	eb.statusLabel.SetText("")
}

// CreateExportMenu creates a menu with export options
func CreateExportMenu(config *ExportButtonsConfig, provider ExportDataProvider, window fyne.Window) *fyne.Menu {
	items := make([]*fyne.MenuItem, 0)

	if config.EnableCSV {
		items = append(items, fyne.NewMenuItem("Export CSV...", func() {
			eb := &ExportButtons{config: config, provider: provider, window: window}
			eb.exportCSV()
		}))
	}

	if config.EnableJSON {
		items = append(items, fyne.NewMenuItem("Export Config...", func() {
			eb := &ExportButtons{config: config, provider: provider, window: window}
			eb.exportJSON()
		}))
	}

	if config.EnablePNG {
		items = append(items, fyne.NewMenuItem("Save Image...", func() {
			eb := &ExportButtons{config: config, provider: provider, window: window}
			eb.exportPNG()
		}))
	}

	return fyne.NewMenu("Export", items...)
}

// CreateCompactExportButton creates a single button with dropdown for all export options
func CreateCompactExportButton(config *ExportButtonsConfig, provider ExportDataProvider, window fyne.Window) *widget.Button {
	return widget.NewButtonWithIcon("Export", theme.DocumentSaveIcon(), func() {
		items := make([]*fyne.MenuItem, 0)

		eb := &ExportButtons{config: config, provider: provider, window: window}

		if config.EnableCSV {
			items = append(items, fyne.NewMenuItem("Export Data (CSV)", eb.exportCSV))
		}
		if config.EnableJSON {
			items = append(items, fyne.NewMenuItem("Export Config (JSON)", eb.exportJSON))
		}
		if config.EnablePNG {
			items = append(items, fyne.NewMenuItem("Save Visualization (PNG)", eb.exportPNG))
		}

		menu := fyne.NewMenu("", items...)
		popup := widget.NewPopUpMenu(menu, window.Canvas())
		popup.ShowAtPosition(fyne.CurrentApp().Driver().AbsolutePositionForObject(window.Canvas().Content()))
	})
}
