//go:build legacy_fyne

package tabs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/module6-eda/pkg/export"
)

// MakeLayoutVisualizerTab provides an SVG-backed layout summary with layer toggles.
// If direct SVG rendering is not available, this presents a structured text view.
func MakeLayoutVisualizerTab(cfg *config.ArrayConfig, _ fyne.Window) fyne.CanvasObject {
	if cfg == nil {
		cfg = &config.ArrayConfig{Rows: 4, Cols: 4, Mode: "storage", Architecture: "passive", CellWidth: 0.46, CellHeight: 2.72}
	}

	summary := widget.NewMultiLineEntry()
	summary.Wrapping = fyne.TextWrapWord
	summary.TextStyle.Monospace = true
	summary.Disable()

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

	update := func() {
		svgContent, source := loadLayoutSVGContent(cfg)
		status.SetText("Source: " + source)
		summary.SetText(buildLayerSummary(svgContent, layerFilter{
			WL:     showWL.Checked,
			BL:     showBL.Checked,
			SL:     showSL.Checked,
			Cells:  showCells.Checked,
			Grid:   showGrid.Checked,
			Legend: showLegend.Checked,
		}))
	}

	showWL.OnChanged = func(bool) { update() }
	showBL.OnChanged = func(bool) { update() }
	showSL.OnChanged = func(bool) { update() }
	showCells.OnChanged = func(bool) { update() }
	showGrid.OnChanged = func(bool) { update() }
	showLegend.OnChanged = func(bool) { update() }

	header := container.NewHBox(
		widget.NewLabel("Layers:"),
		showWL, showBL, showSL, showCells, showGrid, showLegend,
		widget.NewButton("Refresh", update),
		widget.NewSeparator(),
		status,
	)

	update()
	return container.NewBorder(header, nil, nil, nil, container.NewScroll(summary))
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
	sb.WriteString("Layout Visualizer (structured SVG summary)\n")
	sb.WriteString("----------------------------------------\n")

	if f.WL {
		sb.WriteString(fmt.Sprintf("WL wires: %d\n", strings.Count(svg, "class=\"wire-wl\"")))
	}
	if f.BL {
		sb.WriteString(fmt.Sprintf("BL wires: %d\n", strings.Count(svg, "class=\"wire-bl\"")))
	}
	if f.SL {
		sb.WriteString(fmt.Sprintf("SL wires: %d\n", strings.Count(svg, "class=\"wire-sl\"")))
	}
	if f.Cells {
		cells := strings.Count(svg, "class=\"cell-passive\"") + strings.Count(svg, "class=\"cell-1t1r\"")
		sb.WriteString(fmt.Sprintf("Cells: %d\n", cells))
		sb.WriteString(fmt.Sprintf("Transistors (1T1R visuals): %d\n", strings.Count(svg, "class=\"cell-transistor\"")))
	}
	if f.Grid {
		if strings.Contains(svg, "<!-- Grid -->") {
			sb.WriteString("Grid: enabled in SVG\n")
		} else {
			sb.WriteString("Grid: not present\n")
		}
	}
	if f.Legend {
		if strings.Contains(svg, "<!-- Legend -->") {
			sb.WriteString("Legend: present\n")
		} else {
			sb.WriteString("Legend: not present\n")
		}
	}

	sb.WriteString("\nSVG metadata snippets:\n")
	for _, marker := range []string{"viewBox=", "FeCIM", "WL[", "BL[", "SL["} {
		if i := strings.Index(svg, marker); i >= 0 {
			start := i - 24
			if start < 0 {
				start = 0
			}
			end := i + 64
			if end > len(svg) {
				end = len(svg)
			}
			sb.WriteString("- ..." + strings.ReplaceAll(svg[start:end], "\n", " ") + "...\n")
		}
	}

	return sb.String()
}
