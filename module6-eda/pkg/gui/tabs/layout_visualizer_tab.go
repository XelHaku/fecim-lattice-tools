package tabs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

// MakeLayoutVisualizerTab provides an SVG-backed layout summary with layer toggles.
//
// Fyne does not have a native SVG renderer, so this tab offers two views:
//   - "Summary" (default): structured text showing element counts per layer
//   - "SVG Source": raw SVG XML for copy-paste into a browser or Inkscape
//
// A "Save SVG" button writes the current SVG to disk for external viewing.
func MakeLayoutVisualizerTab(cfg *config.ArrayConfig, window fyne.Window) fyne.CanvasObject {
	if cfg == nil {
		cfg = &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}
	}

	content := widget.NewMultiLineEntry()
	content.Wrapping = fyne.TextWrapOff
	content.TextStyle.Monospace = true
	content.Disable()

	status := widget.NewLabel("Ready")

	showWL := widget.NewCheck("WL", nil)
	showBL := widget.NewCheck("BL", nil)
	showSL := widget.NewCheck("SL", nil)
	showCells := widget.NewCheck("Cells", nil)
	showGrid := widget.NewCheck("Grid", nil)
	showLegend := widget.NewCheck("Legend", nil)
	for _, c := range []*widget.Check{showWL, showBL, showSL, showCells, showGrid, showLegend} {
		c.SetChecked(true)
	}

	// showSource toggles between structured summary and raw SVG XML
	showSource := false
	var svgData string
	var svgSourceBtn *widget.Button

	updateContent := func() {
		var src string
		svgData, src = loadLayoutSVGContent(cfg)
		status.SetText("Source: " + src)

		if showSource {
			content.SetText(svgData)
		} else {
			content.SetText(buildLayerSummary(svgData, layerFilter{
				WL:     showWL.Checked,
				BL:     showBL.Checked,
				SL:     showSL.Checked,
				Cells:  showCells.Checked,
				Grid:   showGrid.Checked,
				Legend: showLegend.Checked,
			}))
		}
	}

	svgSourceBtn = widget.NewButton("View SVG Source", func() {
		showSource = !showSource
		if showSource {
			svgSourceBtn.SetText("View Summary")
		} else {
			svgSourceBtn.SetText("View SVG Source")
		}
		updateContent()
	})

	saveSVGBtn := widget.NewButton("Save SVG…", func() {
		design := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
		dlg := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
			if err != nil || w == nil {
				return
			}
			defer w.Close()
			if _, werr := w.Write([]byte(svgData)); werr != nil {
				dialog.ShowError(werr, window)
			}
		}, window)
		dlg.SetFileName(design + ".svg")
		dlg.Show()
	})

	showWL.OnChanged = func(bool) { updateContent() }
	showBL.OnChanged = func(bool) { updateContent() }
	showSL.OnChanged = func(bool) { updateContent() }
	showCells.OnChanged = func(bool) { updateContent() }
	showGrid.OnChanged = func(bool) { updateContent() }
	showLegend.OnChanged = func(bool) { updateContent() }

	hintLabel := widget.NewLabel("Tip: Save SVG → open in browser or Inkscape for visual rendering")
	hintLabel.TextStyle.Italic = true

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Layers:"),
			showWL, showBL, showSL, showCells, showGrid, showLegend,
			widget.NewButton("Refresh", updateContent),
			widget.NewSeparator(),
			svgSourceBtn,
			saveSVGBtn,
		),
		container.NewHBox(status, widget.NewSeparator(), hintLabel),
	)

	updateContent()
	return container.NewBorder(header, nil, nil, nil, container.NewScroll(content))
}

func loadLayoutSVGContent(cfg *config.ArrayConfig) (string, string) {
	p := filepath.Join("data", fmt.Sprintf("fecim_crossbar_%dx%d.svg", cfg.Rows, cfg.Cols))
	if b, err := os.ReadFile(p); err == nil {
		return string(b), p
	}
	return export.GenerateLayoutSVGWithDefaults(*cfg), "generated (in-memory)"
}

type layerFilter struct {
	WL, BL, SL bool
	Cells      bool
	Grid       bool
	Legend     bool
}

func buildLayerSummary(svg string, f layerFilter) string {
	var sb strings.Builder
	sb.WriteString("Layout Visualizer — SVG Layer Summary\n")
	sb.WriteString("(Fyne has no native SVG renderer; use 'View SVG Source' + Save SVG to view in browser)\n")
	sb.WriteString("─────────────────────────────────────────\n\n")

	totalSize := len(svg)
	sb.WriteString(fmt.Sprintf("SVG size: %d bytes\n\n", totalSize))

	sb.WriteString("Layer element counts:\n")
	if f.WL {
		n := strings.Count(svg, "class=\"wire-wl\"")
		sb.WriteString(fmt.Sprintf("  WL wires:               %d\n", n))
	}
	if f.BL {
		n := strings.Count(svg, "class=\"wire-bl\"")
		sb.WriteString(fmt.Sprintf("  BL wires:               %d\n", n))
	}
	if f.SL {
		n := strings.Count(svg, "class=\"wire-sl\"")
		sb.WriteString(fmt.Sprintf("  SL wires:               %d\n", n))
	}
	if f.Cells {
		passive := strings.Count(svg, "class=\"cell-passive\"")
		t1r := strings.Count(svg, "class=\"cell-1t1r\"")
		xistors := strings.Count(svg, "class=\"cell-transistor\"")
		sb.WriteString(fmt.Sprintf("  Passive cells:          %d\n", passive))
		sb.WriteString(fmt.Sprintf("  1T1R cells:             %d\n", t1r))
		sb.WriteString(fmt.Sprintf("  Transistor visuals:     %d\n", xistors))
	}
	if f.Grid {
		if strings.Contains(svg, "<!-- Grid -->") {
			sb.WriteString("  Grid:                   enabled\n")
		} else {
			sb.WriteString("  Grid:                   not present\n")
		}
	}
	if f.Legend {
		if strings.Contains(svg, "<!-- Legend -->") {
			sb.WriteString("  Legend:                 present\n")
		} else {
			sb.WriteString("  Legend:                 not present\n")
		}
	}

	sb.WriteString("\nKey SVG attributes:\n")
	for _, marker := range []string{"viewBox=", "width=", "height=", "FeCIM", "WL[", "BL[", "SL["} {
		if i := strings.Index(svg, marker); i >= 0 {
			start := i
			end := i + 80
			if end > len(svg) {
				end = len(svg)
			}
			snippet := strings.ReplaceAll(svg[start:end], "\n", " ")
			if len(snippet) > 72 {
				snippet = snippet[:72] + "…"
			}
			sb.WriteString("  " + snippet + "\n")
		}
	}

	return sb.String()
}
