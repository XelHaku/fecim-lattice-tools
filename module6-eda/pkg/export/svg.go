// pkg/export/svg.go
// SVG layout visualization generator for FeCIM crossbar arrays
// Generates visual representation of cell placement from DEF data

package export

import (
	"fmt"
	"strings"

	"fecim-lattice-tools/module6-eda/pkg/config"
	"fecim-lattice-tools/shared/logging"
)

var logSVG = logging.NewLogger("eda-export-svg")

// SVGConfig holds configuration for SVG generation
type SVGConfig struct {
	CellWidth   float64 // Cell width in pixels
	CellHeight  float64 // Cell height in pixels
	Margin      float64 // Margin around array
	ShowGrid    bool    // Show grid lines
	ShowLabels  bool    // Show WL/BL/SL labels
	ShowCellIDs bool    // Show cell row,col labels
	ColorScheme string  // "default", "1t1r", "thermal"
}

// DefaultSVGConfig returns standard SVG generation parameters
func DefaultSVGConfig() SVGConfig {
	return SVGConfig{
		CellWidth:   40,
		CellHeight:  60,
		Margin:      80,
		ShowGrid:    true,
		ShowLabels:  true,
		ShowCellIDs: false,
		ColorScheme: "default",
	}
}

// GenerateLayoutSVG creates an SVG visualization of the FeCIM crossbar array
func GenerateLayoutSVG(cfg config.ArrayConfig, svgCfg SVGConfig) string {
	logSVG.Input("GenerateLayoutSVG", map[string]interface{}{
		"rows": cfg.Rows, "cols": cfg.Cols, "arch": cfg.Architecture,
	})

	var sb strings.Builder

	is1T1R := cfg.Architecture == "1t1r"
	is2T1R := cfg.Architecture == "2t1r"

	// Calculate dimensions
	arrayWidth := float64(cfg.Cols) * svgCfg.CellWidth
	arrayHeight := float64(cfg.Rows) * svgCfg.CellHeight
	totalWidth := arrayWidth + 2*svgCfg.Margin
	totalHeight := arrayHeight + 2*svgCfg.Margin

	// Extra vertical space for signal labels below the array
	if is2T1R {
		totalHeight += 50 // SL labels at +50, CSL labels at +65, legend below
	} else if is1T1R {
		totalHeight += 30 // SL labels at +50
	}

	// SVG header
	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.0f %.0f" width="%.0f" height="%.0f">
  <defs>
    <style>
      .cell { stroke: #0088cc; stroke-width: 1; }
      .cell-passive { fill: #1a3a4a; }
      .cell-1t1r { fill: #2a4a3a; }
      .cell-2t1r { fill: #3a2a5a; }
      .cell-transistor { fill: #3a5a4a; }
      .cell-transistor2 { fill: #6a5a2a; }
      .wire-csl { stroke: #ff8800; stroke-width: 1.5; }
      .wire-wl { stroke: #ff6600; stroke-width: 2; }
      .wire-bl { stroke: #00cc66; stroke-width: 2; }
      .wire-sl { stroke: #cc66ff; stroke-width: 2; }
      .label { font-family: monospace; font-size: 14px; fill: #00ccff; }
      .label-small { font-family: monospace; font-size: 14px; fill: #66aacc; }
      .title { font-family: sans-serif; font-size: 14px; fill: #ffffff; font-weight: bold; }
      .grid { stroke: #334455; stroke-width: 0.5; stroke-dasharray: 2,2; }
      .pin { fill: #ffcc00; }
      .pin-sl { fill: #cc66ff; }
    </style>
  </defs>

  <!-- Background -->
  <rect width="100%%" height="100%%" fill="#0a1520"/>

`, totalWidth, totalHeight, totalWidth, totalHeight))

	// Title
	archLabel := "Passive"
	if is1T1R {
		archLabel = "1T1R"
	} else if is2T1R {
		archLabel = "2T1R"
	}
	sb.WriteString(fmt.Sprintf(`  <!-- Title -->
  <text x="%.0f" y="25" class="title" text-anchor="middle">FeCIM %dx%d Crossbar (%s)</text>

`, totalWidth/2, cfg.Rows, cfg.Cols, archLabel))

	// Draw grid lines if enabled
	if svgCfg.ShowGrid {
		sb.WriteString("  <!-- Grid -->\n")
		// Vertical lines
		for col := 0; col <= cfg.Cols; col++ {
			x := svgCfg.Margin + float64(col)*svgCfg.CellWidth
			sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="grid"/>
`, x, svgCfg.Margin, x, svgCfg.Margin+arrayHeight))
		}
		// Horizontal lines
		for row := 0; row <= cfg.Rows; row++ {
			y := svgCfg.Margin + float64(row)*svgCfg.CellHeight
			sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="grid"/>
`, svgCfg.Margin, y, svgCfg.Margin+arrayWidth, y))
		}
		sb.WriteString("\n")
	}

	// Draw Word Lines (horizontal)
	sb.WriteString("  <!-- Word Lines -->\n")
	for row := 0; row < cfg.Rows; row++ {
		y := svgCfg.Margin + float64(row)*svgCfg.CellHeight + svgCfg.CellHeight/2
		// WL wire
		sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="wire-wl"/>
`, svgCfg.Margin-20, y, svgCfg.Margin+arrayWidth, y))
		// WL pin
		sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="4" class="pin"/>
`, svgCfg.Margin-20, y))
		// WL label
		if svgCfg.ShowLabels {
			sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="label" text-anchor="end">WL[%d]</text>
`, svgCfg.Margin-30, y+4, row))
		}
	}
	sb.WriteString("\n")

	// Draw Bit Lines (vertical)
	sb.WriteString("  <!-- Bit Lines -->\n")
	for col := 0; col < cfg.Cols; col++ {
		x := svgCfg.Margin + float64(col)*svgCfg.CellWidth + svgCfg.CellWidth/2
		// BL wire
		sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="wire-bl"/>
`, x, svgCfg.Margin, x, svgCfg.Margin+arrayHeight+20))
		// BL pin
		sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="4" class="pin"/>
`, x, svgCfg.Margin+arrayHeight+20))
		// BL label
		if svgCfg.ShowLabels {
			sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="label" text-anchor="middle">BL[%d]</text>
`, x, svgCfg.Margin+arrayHeight+35, col))
		}
	}
	sb.WriteString("\n")

	// Draw Source Lines for 1T1R and 2T1R (vertical, separate from BL)
	if is1T1R || is2T1R {
		sb.WriteString("  <!-- Source Lines -->\n")
		for col := 0; col < cfg.Cols; col++ {
			x := svgCfg.Margin + float64(col)*svgCfg.CellWidth + svgCfg.CellWidth/2 + 8
			// SL wire (offset right of BL center)
			sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="wire-sl" stroke-dasharray="4,2"/>
`, x, svgCfg.Margin, x, svgCfg.Margin+arrayHeight+20))
			// SL pin
			sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="4" class="pin-sl"/>
`, x, svgCfg.Margin+arrayHeight+20))
			// SL label
			if svgCfg.ShowLabels {
				sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="label" text-anchor="middle" fill="#cc66ff">SL[%d]</text>
`, x, svgCfg.Margin+arrayHeight+50, col))
			}
		}
		sb.WriteString("\n")
	}

	// Draw Column Select Lines for 2T1R (vertical, one per column like BL/SL)
	// CSL[col] gates the column transistor for every cell in column col.
	// Matches the Verilog port: input wire [numCols-1:0] CSL.
	if is2T1R {
		sb.WriteString("  <!-- Column Select Lines (2T1R) -->\n")
		for col := 0; col < cfg.Cols; col++ {
			x := svgCfg.Margin + float64(col)*svgCfg.CellWidth + svgCfg.CellWidth/2 - 8
			// CSL wire (vertical, offset left of BL)
			sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="wire-csl" stroke-dasharray="4,2"/>
`, x, svgCfg.Margin, x, svgCfg.Margin+arrayHeight+20))
			// CSL pin at bottom
			sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="3" fill="#ff8800"/>
`, x, svgCfg.Margin+arrayHeight+20))
			if svgCfg.ShowLabels {
				sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="label" text-anchor="middle" fill="#ff8800">CSL[%d]</text>
`, x, svgCfg.Margin+arrayHeight+65, col))
			}
		}
		sb.WriteString("\n")
	}

	// Draw cells
	sb.WriteString("  <!-- Cells -->\n")
	cellClass := "cell-passive"
	if is1T1R {
		cellClass = "cell-1t1r"
	} else if is2T1R {
		cellClass = "cell-2t1r"
	}

	for row := 0; row < cfg.Rows; row++ {
		for col := 0; col < cfg.Cols; col++ {
			x := svgCfg.Margin + float64(col)*svgCfg.CellWidth
			y := svgCfg.Margin + float64(row)*svgCfg.CellHeight

			if is2T1R {
				// 2T1R cell: row transistor + column transistor + FeFET stack
				thirdH := (svgCfg.CellHeight - 14) / 3
				// Row transistor (top third, green)
				sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" class="cell cell-transistor" rx="2"/>
`, x+4, y+4, svgCfg.CellWidth-8, thirdH))
				// Column transistor (middle third, amber)
				sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" class="cell cell-transistor2" rx="2"/>
`, x+4, y+4+thirdH+2, svgCfg.CellWidth-8, thirdH))
				// FeFET (bottom third, purple)
				sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" class="cell %s" rx="2"/>
`, x+4, y+4+2*thirdH+4, svgCfg.CellWidth-8, thirdH, cellClass))
				cx := x + svgCfg.CellWidth/2
				sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#00ccff" stroke-width="1.5"/>
`, cx, y+4+thirdH, cx, y+4+thirdH+2))
				sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#00ccff" stroke-width="1.5"/>
`, cx, y+4+2*thirdH+2, cx, y+4+2*thirdH+4))
			} else if is1T1R {
				// 1T1R cell: transistor + FeFET stack
				// Transistor (top half)
				sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" class="cell cell-transistor" rx="2"/>
`, x+4, y+4, svgCfg.CellWidth-8, svgCfg.CellHeight/2-6))
				// FeFET (bottom half)
				sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" class="cell %s" rx="2"/>
`, x+4, y+svgCfg.CellHeight/2+2, svgCfg.CellWidth-8, svgCfg.CellHeight/2-6, cellClass))
				// Connection between transistor and FeFET
				cx := x + svgCfg.CellWidth/2
				sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#00ccff" stroke-width="2"/>
`, cx, y+svgCfg.CellHeight/2-2, cx, y+svgCfg.CellHeight/2+2))
			} else {
				// Passive cell: single FeFET
				sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" class="cell %s" rx="3"/>
`, x+4, y+4, svgCfg.CellWidth-8, svgCfg.CellHeight-8, cellClass))
			}

			// Cell ID label
			if svgCfg.ShowCellIDs {
				cx := x + svgCfg.CellWidth/2
				cy := y + svgCfg.CellHeight/2
				sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="label-small" text-anchor="middle">%d,%d</text>
`, cx, cy+3, row, col))
			}
		}
	}
	sb.WriteString("\n")

	// Legend
	legendY := totalHeight - 20
	if is1T1R || is2T1R {
		legendY = totalHeight - 35
	}
	sb.WriteString(fmt.Sprintf(`  <!-- Legend -->
  <g transform="translate(%.0f, %.0f)">
    <line x1="0" y1="0" x2="20" y2="0" class="wire-wl"/>
    <text x="25" y="4" class="label-small">WL (Word Line)</text>
    <line x1="100" y1="0" x2="120" y2="0" class="wire-bl"/>
    <text x="125" y="4" class="label-small">BL (Bit Line)</text>
`, svgCfg.Margin, legendY))

	if is1T1R {
		sb.WriteString(`    <line x1="200" y1="0" x2="220" y2="0" class="wire-sl" stroke-dasharray="4,2"/>
    <text x="225" y="4" class="label-small">SL (Source Line)</text>
`)
	}
	if is2T1R {
		sb.WriteString(`    <line x1="200" y1="0" x2="220" y2="0" class="wire-sl" stroke-dasharray="4,2"/>
    <text x="225" y="4" class="label-small">SL (Source Line)</text>
    <line x1="320" y1="0" x2="340" y2="0" class="wire-csl" stroke-dasharray="4,2"/>
    <text x="345" y="4" class="label-small" fill="#ff8800">CSL (Col Select)</text>
`)
	}
	sb.WriteString("  </g>\n\n")

	// Close SVG
	sb.WriteString("</svg>\n")

	return sb.String()
}

// GenerateLayoutSVGWithDefaults uses default SVG configuration
func GenerateLayoutSVGWithDefaults(cfg config.ArrayConfig) string {
	return GenerateLayoutSVG(cfg, DefaultSVGConfig())
}
