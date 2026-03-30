// Package export — SVG vector export for publication-quality P-E loop figures.
// Uses only fmt.Sprintf for SVG XML generation (no external dependencies).
package export

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// SVGPlotConfig configures the appearance of an SVG P-E loop plot.
type SVGPlotConfig struct {
	Width    float64 // Total SVG width in pixels (default 640)
	Height   float64 // Total SVG height in pixels (default 480)
	MarginL  float64 // Left margin (default 80)
	MarginR  float64 // Right margin (default 30)
	MarginT  float64 // Top margin (default 50)
	MarginB  float64 // Bottom margin (default 70)
	Title    string  // Plot title
	XLabel   string  // X-axis label (default "E [MV/cm]")
	YLabel   string  // Y-axis label (default "P [µC/cm²]")
	Citation string  // Citation footer (material, date, commit)
}

// DefaultSVGPlotConfig returns sensible defaults for a P-E loop figure.
func DefaultSVGPlotConfig() SVGPlotConfig {
	return SVGPlotConfig{
		Width:   640,
		Height:  480,
		MarginL: 80,
		MarginR: 30,
		MarginT: 50,
		MarginB: 70,
		Title:   "P-E Hysteresis Loop",
		XLabel:  "E [MV/cm]",
		YLabel:  "P [\u00b5C/cm\u00b2]",
	}
}

// niceStep picks a human-friendly tick spacing for the given data range
// so that roughly 4-8 ticks appear.
func niceStep(dataRange float64) float64 {
	if dataRange <= 0 {
		return 1
	}
	raw := dataRange / 5.0
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	norm := raw / mag
	var nice float64
	switch {
	case norm < 1.5:
		nice = 1
	case norm < 3.5:
		nice = 2
	case norm < 7.5:
		nice = 5
	default:
		nice = 10
	}
	return nice * mag
}

// GeneratePELoopSVG produces a complete SVG string of a P-E hysteresis loop.
// eField and polarization must have equal length and at least 2 points.
func GeneratePELoopSVG(eField, polarization []float64, cfg SVGPlotConfig) (string, error) {
	if len(eField) != len(polarization) {
		return "", fmt.Errorf("eField length %d != polarization length %d", len(eField), len(polarization))
	}
	if len(eField) < 2 {
		return "", fmt.Errorf("need at least 2 data points, got %d", len(eField))
	}

	// Apply defaults for zero-valued fields.
	def := func(v *float64, d float64) {
		if *v == 0 {
			*v = d
		}
	}
	def(&cfg.Width, 640)
	def(&cfg.Height, 480)
	def(&cfg.MarginL, 80)
	def(&cfg.MarginR, 30)
	def(&cfg.MarginT, 50)
	def(&cfg.MarginB, 70)

	// Data bounds.
	xMin, xMax := eField[0], eField[0]
	yMin, yMax := polarization[0], polarization[0]
	for i := 1; i < len(eField); i++ {
		xMin, xMax = math.Min(xMin, eField[i]), math.Max(xMax, eField[i])
		yMin, yMax = math.Min(yMin, polarization[i]), math.Max(yMax, polarization[i])
	}

	// Add 10% padding to data range.
	padded := func(lo, hi float64) (float64, float64) {
		p := (hi - lo) * 0.10
		if p == 0 {
			p = 1
		}
		return lo - p, hi + p
	}
	xMin, xMax = padded(xMin, xMax)
	yMin, yMax = padded(yMin, yMax)

	plotW := cfg.Width - cfg.MarginL - cfg.MarginR
	plotH := cfg.Height - cfg.MarginT - cfg.MarginB

	// Mapping functions: data -> SVG pixel coordinates.
	mapX := func(v float64) float64 { return cfg.MarginL + (v-xMin)/(xMax-xMin)*plotW }
	mapY := func(v float64) float64 { return cfg.MarginT + plotH - (v-yMin)/(yMax-yMin)*plotH }

	var sb strings.Builder

	// SVG header.
	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.0f %.0f" width="%.0f" height="%.0f"
     font-family="'Helvetica Neue', Helvetica, Arial, sans-serif">
  <defs>
    <style>
      .axis { stroke: #333; stroke-width: 1.2; fill: none; }
      .tick { stroke: #333; stroke-width: 0.8; }
      .tick-label { font-size: 11px; fill: #333; }
      .axis-label { font-size: 13px; fill: #222; font-weight: 600; }
      .title { font-size: 15px; fill: #111; font-weight: 700; text-anchor: middle; }
      .loop { fill: none; stroke: #1a6fcc; stroke-width: 1.8; stroke-linejoin: round; }
      .citation { font-size: 9px; fill: #888; }
      .grid { stroke: #e0e0e0; stroke-width: 0.5; }
    </style>
  </defs>

  <!-- Background -->
  <rect width="100%%" height="100%%" fill="#ffffff"/>

`, cfg.Width, cfg.Height, cfg.Width, cfg.Height))

	// Title.
	if cfg.Title != "" {
		sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="title">%s</text>
`, cfg.MarginL+plotW/2, cfg.MarginT-18, xmlEscape(cfg.Title)))
	}

	// Grid lines and ticks — X axis.
	axisBottom := cfg.MarginT + plotH
	for v := math.Ceil(xMin/niceStep(xMax-xMin)) * niceStep(xMax-xMin); v <= xMax; v += niceStep(xMax - xMin) {
		px := mapX(v)
		sb.WriteString(fmt.Sprintf("  <line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"grid\"/>\n", px, cfg.MarginT, px, axisBottom))
		sb.WriteString(fmt.Sprintf("  <line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"tick\"/>\n", px, axisBottom, px, axisBottom+5))
		sb.WriteString(fmt.Sprintf("  <text x=\"%.1f\" y=\"%.1f\" class=\"tick-label\" text-anchor=\"middle\">%g</text>\n", px, axisBottom+18, roundTick(v)))
	}

	// Grid lines and ticks — Y axis.
	for v := math.Ceil(yMin/niceStep(yMax-yMin)) * niceStep(yMax-yMin); v <= yMax; v += niceStep(yMax - yMin) {
		py := mapY(v)
		sb.WriteString(fmt.Sprintf("  <line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"grid\"/>\n", cfg.MarginL, py, cfg.MarginL+plotW, py))
		sb.WriteString(fmt.Sprintf("  <line x1=\"%.1f\" y1=\"%.1f\" x2=\"%.1f\" y2=\"%.1f\" class=\"tick\"/>\n", cfg.MarginL-5, py, cfg.MarginL, py))
		sb.WriteString(fmt.Sprintf("  <text x=\"%.1f\" y=\"%.1f\" class=\"tick-label\" text-anchor=\"end\">%g</text>\n", cfg.MarginL-8, py+4, roundTick(v)))
	}

	// Axis lines (drawn on top of grid).
	sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="axis"/>
`, cfg.MarginL, cfg.MarginT, cfg.MarginL, axisBottom))
	sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="axis"/>
`, cfg.MarginL, axisBottom, cfg.MarginL+plotW, axisBottom))

	// Axis labels.
	xLabel := cfg.XLabel
	if xLabel == "" {
		xLabel = "E [MV/cm]"
	}
	yLabel := cfg.YLabel
	if yLabel == "" {
		yLabel = "P [\u00b5C/cm\u00b2]"
	}
	sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="axis-label" text-anchor="middle">%s</text>
`, cfg.MarginL+plotW/2, axisBottom+40, xmlEscape(xLabel)))
	sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="axis-label" text-anchor="middle" transform="rotate(-90,%.1f,%.1f)">%s</text>
`, cfg.MarginL-50, cfg.MarginT+plotH/2, cfg.MarginL-50, cfg.MarginT+plotH/2, xmlEscape(yLabel)))

	// P-E loop polyline path.
	sb.WriteString(`  <polyline class="loop" points="`)
	for i := range eField {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%.2f,%.2f", mapX(eField[i]), mapY(polarization[i])))
	}
	sb.WriteString(`"/>
`)

	// Legend box.
	lx := cfg.MarginL + plotW - 120
	ly := cfg.MarginT + 10
	sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="115" height="24" rx="3" fill="#fff" stroke="#ccc" stroke-width="0.8"/>
`, lx, ly))
	sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#1a6fcc" stroke-width="2"/>
`, lx+6, ly+12, lx+26, ly+12))
	sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="tick-label">P-E Loop</text>
`, lx+30, ly+16))

	// Citation footer.
	if cfg.Citation != "" {
		sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="citation">%s</text>
`, cfg.MarginL, cfg.Height-8, xmlEscape(cfg.Citation)))
	}

	sb.WriteString("</svg>\n")
	return sb.String(), nil
}

// xmlEscape escapes text for safe embedding in SVG/XML attributes and content.
func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&apos;")
	return r.Replace(s)
}

// roundTick rounds tiny floating-point residuals (e.g. 1.0000000000000002 -> 1).
func roundTick(v float64) float64 {
	if v == 0 {
		return 0 // avoid -0
	}
	return math.Round(v*1e9) / 1e9
}

// ExportSVG writes raw SVG content to a file, following the Exporter pattern.
func (e *Exporter) ExportSVG(svgContent string) *ExportResult {
	result := &ExportResult{Format: FormatSVG}

	if err := e.ensureOutputDir(); err != nil {
		result.Error = fmt.Errorf("failed to create output directory: %w", err)
		return result
	}

	result.FilePath = e.generateFilename("svg")
	payload := []byte(svgContent)
	if err := os.WriteFile(result.FilePath, payload, 0644); err != nil {
		result.Error = fmt.Errorf("failed to write SVG file: %w", err)
		return result
	}
	result.BytesWritten = int64(len(payload))
	return result
}
