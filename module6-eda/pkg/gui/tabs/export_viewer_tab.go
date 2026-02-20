package tabs

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

var exportFormats = []string{"LEF", "Liberty", "Verilog", "DEF", "SPICE"}

// MakeExportViewerTab creates a read-only export preview tab for LEF/Liberty/Verilog/DEF/SPICE.
func MakeExportViewerTab(cfg *config.ArrayConfig, window fyne.Window) fyne.CanvasObject {
	if cfg == nil {
		cfg = &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}
	}

	formatSelect := widget.NewSelect(exportFormats, nil)
	formatSelect.SetSelected("LEF")

	status := widget.NewLabel("Ready")
	preview := widget.NewMultiLineEntry()
	preview.Wrapping = fyne.TextWrapOff
	preview.TextStyle.Monospace = true
	preview.Disable()

	refresh := func() {
		content, source := loadExportPreviewContent(formatSelect.Selected, cfg)
		preview.SetText(content)
		status.SetText("Source: " + source)
	}

	formatSelect.OnChanged = func(string) { refresh() }

	refreshBtn := widget.NewButton("Refresh", refresh)

	saveBtn := widget.NewButton("Save to File…", func() {
		ext := formatExtension(formatSelect.Selected)
		design := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
		defaultName := design + ext

		dlg := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
			if err != nil || w == nil {
				return
			}
			defer w.Close()
			if _, werr := w.Write([]byte(preview.Text)); werr != nil {
				dialog.ShowError(werr, window)
			}
		}, window)
		dlg.SetFileName(defaultName)
		dlg.Show()
	})

	header := container.NewHBox(
		widget.NewLabel("Format:"),
		formatSelect,
		refreshBtn,
		saveBtn,
		widget.NewSeparator(),
		status,
	)

	refresh()

	return container.NewBorder(header, nil, nil, nil, container.NewScroll(preview))
}

// formatExtension returns the canonical file extension for the given format name.
func formatExtension(format string) string {
	switch format {
	case "LEF":
		return ".lef"
	case "Liberty":
		return ".lib"
	case "Verilog":
		return ".v"
	case "DEF":
		return ".def"
	case "SPICE":
		return ".sp"
	default:
		return ".txt"
	}
}

func loadExportPreviewContent(format string, cfg *config.ArrayConfig) (content string, source string) {
	tech := cfg.Technology
	if tech == "" {
		tech = "sky130"
	}
	cellCfg := config.CellConfig{
		Name:         "fecim_bitcell",
		Width:        cfg.CellWidth,
		Height:       cfg.CellHeight,
		CellType:     cfg.Architecture,
		Technology:   tech,
		RiseTime:     10.0,
		FallTime:     10.0,
		InputCap:     0.015,
		LeakagePower: 0.0003,
	}

	design := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	dataDir := "data"

	tryRead := func(path string) (string, bool) {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", false
		}
		return string(b), true
	}

	switch format {
	case "LEF":
		paths := []string{
			filepath.Join("cells", "fecim_bitcell", "fecim_bitcell.lef"),
			filepath.Join("cells", "fecim_1t1r_bitcell", "fecim_1t1r_bitcell.lef"),
			filepath.Join("cells", "fecim_2t1r_bitcell", "fecim_2t1r_bitcell.lef"),
		}
		for _, p := range paths {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		return export.GenerateLEF(cellCfg), "generated (in-memory)"
	case "Liberty":
		paths := []string{
			filepath.Join("cells", "fecim_bitcell", "fecim_bitcell.lib"),
			filepath.Join("cells", "fecim_1t1r_bitcell", "fecim_1t1r_bitcell.lib"),
			filepath.Join("cells", "fecim_2t1r_bitcell", "fecim_2t1r_bitcell.lib"),
		}
		for _, p := range paths {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		return export.GenerateLiberty(cellCfg), "generated (in-memory)"
	case "Verilog":
		p := filepath.Join(dataDir, design+".v")
		if s, ok := tryRead(p); ok {
			return s, p
		}
		return export.GenerateArrayVerilog(*cfg), "generated (in-memory)"
	case "DEF":
		paths := []string{
			filepath.Join(dataDir, design+".def"),
			filepath.Join("output", design+".def"),
		}
		for _, p := range paths {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		return "# DEF requires array compilation.\n# Generate via CLI:\n#   fecim-lattice-tools eda cli --def\n", "(not generated)"
	case "SPICE":
		paths := []string{
			filepath.Join(dataDir, design+".sp"),
			filepath.Join(dataDir, "fecim_array.sp"),
		}
		for _, p := range paths {
			if s, ok := tryRead(p); ok {
				return s, p
			}
		}
		return "* SPICE export not found on disk\n* Expected: data/" + design + ".sp\n* CLI export supports SPICE output.", "not found"
	default:
		return "", "unknown format"
	}
}
