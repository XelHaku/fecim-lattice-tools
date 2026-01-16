// Package render provides P-E hysteresis curve plotting utilities.
package render

import (
	"fmt"
	"math"
)

// PlotConfig contains configuration for P-E plot rendering.
type PlotConfig struct {
	// Plot area (normalized coordinates 0-1)
	X, Y          float64
	Width, Height float64

	// Axis configuration
	ShowGrid      bool
	GridDivisions int
	AxisColor     Color
	GridColor     Color

	// Labels
	XLabel string
	YLabel string
	Title  string

	// Animation
	TrailLength int     // Number of trailing points to show
	TrailFade   float32 // Alpha fade for trail (0-1)
}

// DefaultPlotConfig returns default plot configuration.
func DefaultPlotConfig() *PlotConfig {
	return &PlotConfig{
		X:             0.35,
		Y:             0.1,
		Width:         0.6,
		Height:        0.8,
		ShowGrid:      true,
		GridDivisions: 10,
		AxisColor:     Color{0.0, 0.0, 0.0, 1.0},
		GridColor:     Color{0.8, 0.8, 0.8, 0.5},
		XLabel:        "Electric Field E (MV/cm)",
		YLabel:        "Polarization P (μC/cm²)",
		Title:         "Ferroelectric Hysteresis",
		TrailLength:   500,
		TrailFade:     0.3,
	}
}

// PlotVertex represents a vertex for GPU rendering.
type PlotVertex struct {
	Position [2]float32 // X, Y in NDC
	Color    [4]float32 // RGBA
}

// GeneratePlotVertices creates vertices for the entire P-E plot visualization.
func GeneratePlotVertices(hp *HysteresisPlot, cfg *PlotConfig) []PlotVertex {
	var vertices []PlotVertex

	// Generate grid lines
	if cfg.ShowGrid {
		gridVerts := generateGridVertices(hp, cfg)
		vertices = append(vertices, gridVerts...)
	}

	// Generate axes
	axisVerts := generateAxisVertices(hp, cfg)
	vertices = append(vertices, axisVerts...)

	// Generate hysteresis curve
	curveVerts := generateCurveVertices(hp, cfg)
	vertices = append(vertices, curveVerts...)

	// Generate current point marker
	markerVerts := generateMarkerVertices(hp, cfg)
	vertices = append(vertices, markerVerts...)

	return vertices
}

// generateGridVertices creates grid line vertices.
func generateGridVertices(hp *HysteresisPlot, cfg *PlotConfig) []PlotVertex {
	var vertices []PlotVertex

	// Vertical grid lines
	for i := 0; i <= cfg.GridDivisions; i++ {
		t := float64(i) / float64(cfg.GridDivisions)
		x := cfg.X + t*cfg.Width

		vertices = append(vertices,
			PlotVertex{
				Position: [2]float32{float32(x*2 - 1), float32(cfg.Y*2 - 1)},
				Color:    [4]float32{cfg.GridColor.R, cfg.GridColor.G, cfg.GridColor.B, cfg.GridColor.A},
			},
			PlotVertex{
				Position: [2]float32{float32(x*2 - 1), float32((cfg.Y+cfg.Height)*2 - 1)},
				Color:    [4]float32{cfg.GridColor.R, cfg.GridColor.G, cfg.GridColor.B, cfg.GridColor.A},
			},
		)
	}

	// Horizontal grid lines
	for i := 0; i <= cfg.GridDivisions; i++ {
		t := float64(i) / float64(cfg.GridDivisions)
		y := cfg.Y + t*cfg.Height

		vertices = append(vertices,
			PlotVertex{
				Position: [2]float32{float32(cfg.X*2 - 1), float32(y*2 - 1)},
				Color:    [4]float32{cfg.GridColor.R, cfg.GridColor.G, cfg.GridColor.B, cfg.GridColor.A},
			},
			PlotVertex{
				Position: [2]float32{float32((cfg.X+cfg.Width)*2 - 1), float32(y*2 - 1)},
				Color:    [4]float32{cfg.GridColor.R, cfg.GridColor.G, cfg.GridColor.B, cfg.GridColor.A},
			},
		)
	}

	return vertices
}

// generateAxisVertices creates axis line vertices.
func generateAxisVertices(hp *HysteresisPlot, cfg *PlotConfig) []PlotVertex {
	var vertices []PlotVertex

	// Find where zero crosses in normalized space
	zeroX := -hp.EMin / (hp.EMax - hp.EMin)
	zeroY := -hp.PMin / (hp.PMax - hp.PMin)

	// Clamp to visible area
	zeroX = math.Max(0, math.Min(1, zeroX))
	zeroY = math.Max(0, math.Min(1, zeroY))

	// Convert to plot coordinates
	axisX := cfg.X + zeroX*cfg.Width
	axisY := cfg.Y + zeroY*cfg.Height

	// X-axis (horizontal through P=0)
	vertices = append(vertices,
		PlotVertex{
			Position: [2]float32{float32(cfg.X*2 - 1), float32(axisY*2 - 1)},
			Color:    [4]float32{cfg.AxisColor.R, cfg.AxisColor.G, cfg.AxisColor.B, cfg.AxisColor.A},
		},
		PlotVertex{
			Position: [2]float32{float32((cfg.X+cfg.Width)*2 - 1), float32(axisY*2 - 1)},
			Color:    [4]float32{cfg.AxisColor.R, cfg.AxisColor.G, cfg.AxisColor.B, cfg.AxisColor.A},
		},
	)

	// Y-axis (vertical through E=0)
	vertices = append(vertices,
		PlotVertex{
			Position: [2]float32{float32(axisX*2 - 1), float32(cfg.Y*2 - 1)},
			Color:    [4]float32{cfg.AxisColor.R, cfg.AxisColor.G, cfg.AxisColor.B, cfg.AxisColor.A},
		},
		PlotVertex{
			Position: [2]float32{float32(axisX*2 - 1), float32((cfg.Y+cfg.Height)*2 - 1)},
			Color:    [4]float32{cfg.AxisColor.R, cfg.AxisColor.G, cfg.AxisColor.B, cfg.AxisColor.A},
		},
	)

	return vertices
}

// generateCurveVertices creates hysteresis curve vertices with trail effect.
func generateCurveVertices(hp *HysteresisPlot, cfg *PlotConfig) []PlotVertex {
	var vertices []PlotVertex

	if len(hp.Points) < 2 {
		return vertices
	}

	// Determine which points to render (trail)
	startIdx := 0
	if len(hp.Points) > cfg.TrailLength {
		startIdx = len(hp.Points) - cfg.TrailLength
	}

	points := hp.Points[startIdx:]

	for i := 0; i < len(points)-1; i++ {
		// Calculate alpha fade based on position in trail
		progress := float32(i) / float32(len(points)-1)
		alpha := cfg.TrailFade + progress*(1.0-cfg.TrailFade)

		// Convert data coordinates to normalized plot coordinates
		x1, y1 := hp.NormalizeToScreen(points[i].X, points[i].Y)
		x2, y2 := hp.NormalizeToScreen(points[i+1].X, points[i+1].Y)

		// Convert to plot area coordinates
		px1 := cfg.X + x1*cfg.Width
		py1 := cfg.Y + y1*cfg.Height
		px2 := cfg.X + x2*cfg.Width
		py2 := cfg.Y + y2*cfg.Height

		// Convert to NDC
		ndcX1 := float32(px1*2 - 1)
		ndcY1 := float32(py1*2 - 1)
		ndcX2 := float32(px2*2 - 1)
		ndcY2 := float32(py2*2 - 1)

		vertices = append(vertices,
			PlotVertex{
				Position: [2]float32{ndcX1, ndcY1},
				Color:    [4]float32{hp.LineColor.R, hp.LineColor.G, hp.LineColor.B, alpha},
			},
			PlotVertex{
				Position: [2]float32{ndcX2, ndcY2},
				Color:    [4]float32{hp.LineColor.R, hp.LineColor.G, hp.LineColor.B, alpha},
			},
		)
	}

	return vertices
}

// generateMarkerVertices creates current point marker (as a small quad/circle).
func generateMarkerVertices(hp *HysteresisPlot, cfg *PlotConfig) []PlotVertex {
	var vertices []PlotVertex

	// Normalize current position
	x, y := hp.NormalizeToScreen(hp.CurrentE, hp.CurrentP)

	// Convert to plot coordinates
	px := cfg.X + x*cfg.Width
	py := cfg.Y + y*cfg.Height

	// Convert to NDC
	ndcX := float32(px*2 - 1)
	ndcY := float32(py*2 - 1)

	// Create a small diamond shape for the marker
	size := float32(0.015) // Marker size in NDC

	// Diamond vertices (4 triangles forming a diamond)
	color := [4]float32{hp.MarkerColor.R, hp.MarkerColor.G, hp.MarkerColor.B, hp.MarkerColor.A}

	// Top triangle
	vertices = append(vertices,
		PlotVertex{Position: [2]float32{ndcX, ndcY + size}, Color: color},
		PlotVertex{Position: [2]float32{ndcX - size, ndcY}, Color: color},
		PlotVertex{Position: [2]float32{ndcX + size, ndcY}, Color: color},
	)

	// Bottom triangle
	vertices = append(vertices,
		PlotVertex{Position: [2]float32{ndcX, ndcY - size}, Color: color},
		PlotVertex{Position: [2]float32{ndcX + size, ndcY}, Color: color},
		PlotVertex{Position: [2]float32{ndcX - size, ndcY}, Color: color},
	)

	return vertices
}

// GenerateCellVertices creates vertices for the ferroelectric cell display.
func GenerateCellVertices(cell *CellDisplay) []PlotVertex {
	var vertices []PlotVertex

	color := cell.GetColor()

	// Convert cell bounds to NDC
	x1 := float32(cell.X*2 - 1)
	y1 := float32(cell.Y*2 - 1)
	x2 := float32((cell.X+cell.Width)*2 - 1)
	y2 := float32((cell.Y+cell.Height)*2 - 1)

	colorArr := [4]float32{color.R, color.G, color.B, color.A}

	// Two triangles forming a quad
	vertices = append(vertices,
		// First triangle
		PlotVertex{Position: [2]float32{x1, y1}, Color: colorArr},
		PlotVertex{Position: [2]float32{x2, y1}, Color: colorArr},
		PlotVertex{Position: [2]float32{x1, y2}, Color: colorArr},
		// Second triangle
		PlotVertex{Position: [2]float32{x2, y1}, Color: colorArr},
		PlotVertex{Position: [2]float32{x2, y2}, Color: colorArr},
		PlotVertex{Position: [2]float32{x1, y2}, Color: colorArr},
	)

	// Add border
	borderColor := [4]float32{0.2, 0.2, 0.2, 1.0}
	vertices = append(vertices,
		// Top edge
		PlotVertex{Position: [2]float32{x1, y2}, Color: borderColor},
		PlotVertex{Position: [2]float32{x2, y2}, Color: borderColor},
		// Right edge
		PlotVertex{Position: [2]float32{x2, y2}, Color: borderColor},
		PlotVertex{Position: [2]float32{x2, y1}, Color: borderColor},
		// Bottom edge
		PlotVertex{Position: [2]float32{x2, y1}, Color: borderColor},
		PlotVertex{Position: [2]float32{x1, y1}, Color: borderColor},
		// Left edge
		PlotVertex{Position: [2]float32{x1, y1}, Color: borderColor},
		PlotVertex{Position: [2]float32{x1, y2}, Color: borderColor},
	)

	return vertices
}

// FormatAxisLabel generates a formatted axis label string.
func FormatAxisLabel(value, min, max float64, label string) string {
	return fmt.Sprintf("%s: %.2f", label, value)
}

// CalculateTickPositions returns tick mark positions for an axis.
func CalculateTickPositions(min, max float64, numTicks int) []float64 {
	ticks := make([]float64, numTicks+1)
	step := (max - min) / float64(numTicks)
	for i := 0; i <= numTicks; i++ {
		ticks[i] = min + float64(i)*step
	}
	return ticks
}

// FormatTickLabel formats a tick value for display.
func FormatTickLabel(value float64) string {
	if math.Abs(value) < 1e-10 {
		return "0"
	}
	if math.Abs(value) >= 1000 || math.Abs(value) < 0.01 {
		return fmt.Sprintf("%.1e", value)
	}
	return fmt.Sprintf("%.2f", value)
}
