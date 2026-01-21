// Package visualization provides crossbar array visualization utilities.
package visualization

import (
	"fmt"
	"math"
)

// HeatmapConfig contains configuration for crossbar heatmap rendering.
type HeatmapConfig struct {
	Width, Height int     // Dimensions in pixels
	CellPadding   float64 // Padding between cells (0-1)
	ShowLabels    bool    // Show row/column labels
	ColorScheme   string  // "viridis", "plasma", "coolwarm"
}

// DefaultHeatmapConfig returns default configuration.
func DefaultHeatmapConfig() *HeatmapConfig {
	return &HeatmapConfig{
		Width:       512,
		Height:      512,
		CellPadding: 0.05,
		ShowLabels:  true,
		ColorScheme: "viridis",
	}
}

// RGB represents a color.
type RGB struct {
	R, G, B float64
}

// HeatmapVertex represents a vertex for GPU rendering.
type HeatmapVertex struct {
	Position [2]float32
	Color    [4]float32
}

// CrossbarHeatmap generates visualization data for a crossbar array.
type CrossbarHeatmap struct {
	config     *HeatmapConfig
	rows, cols int
	data       [][]float64
	minVal     float64
	maxVal     float64
}

// NewCrossbarHeatmap creates a new heatmap visualization.
func NewCrossbarHeatmap(rows, cols int, config *HeatmapConfig) *CrossbarHeatmap {
	if config == nil {
		config = DefaultHeatmapConfig()
	}

	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
	}

	return &CrossbarHeatmap{
		config: config,
		rows:   rows,
		cols:   cols,
		data:   data,
		minVal: 0,
		maxVal: 1,
	}
}

// SetData updates the heatmap data.
func (h *CrossbarHeatmap) SetData(data [][]float64) {
	h.minVal = math.Inf(1)
	h.maxVal = math.Inf(-1)

	for i := 0; i < h.rows && i < len(data); i++ {
		for j := 0; j < h.cols && j < len(data[i]); j++ {
			h.data[i][j] = data[i][j]
			if data[i][j] < h.minVal {
				h.minVal = data[i][j]
			}
			if data[i][j] > h.maxVal {
				h.maxVal = data[i][j]
			}
		}
	}

	// Prevent division by zero
	if h.maxVal == h.minVal {
		h.maxVal = h.minVal + 1
	}
}

// SetRange sets the color mapping range manually.
func (h *CrossbarHeatmap) SetRange(min, max float64) {
	h.minVal = min
	h.maxVal = max
}

// GenerateVertices creates GPU-ready vertex data for the heatmap.
func (h *CrossbarHeatmap) GenerateVertices() []HeatmapVertex {
	var vertices []HeatmapVertex

	cellW := 2.0 / float64(h.cols) // NDC width per cell
	cellH := 2.0 / float64(h.rows) // NDC height per cell

	padding := h.config.CellPadding

	for i := 0; i < h.rows; i++ {
		for j := 0; j < h.cols; j++ {
			// Normalize value
			normVal := (h.data[i][j] - h.minVal) / (h.maxVal - h.minVal)

			// Get color
			color := h.valueToColor(normVal)
			colorArr := [4]float32{float32(color.R), float32(color.G), float32(color.B), 1.0}

			// Calculate cell bounds in NDC (-1 to 1)
			x0 := -1.0 + float64(j)*cellW + padding*cellW/2
			y0 := 1.0 - float64(i+1)*cellH + padding*cellH/2
			x1 := -1.0 + float64(j+1)*cellW - padding*cellW/2
			y1 := 1.0 - float64(i)*cellH - padding*cellH/2

			// Two triangles per cell
			vertices = append(vertices,
				// First triangle
				HeatmapVertex{Position: [2]float32{float32(x0), float32(y0)}, Color: colorArr},
				HeatmapVertex{Position: [2]float32{float32(x1), float32(y0)}, Color: colorArr},
				HeatmapVertex{Position: [2]float32{float32(x0), float32(y1)}, Color: colorArr},
				// Second triangle
				HeatmapVertex{Position: [2]float32{float32(x1), float32(y0)}, Color: colorArr},
				HeatmapVertex{Position: [2]float32{float32(x1), float32(y1)}, Color: colorArr},
				HeatmapVertex{Position: [2]float32{float32(x0), float32(y1)}, Color: colorArr},
			)
		}
	}

	return vertices
}

// valueToColor converts a normalized value (0-1) to a color.
func (h *CrossbarHeatmap) valueToColor(t float64) RGB {
	// Clamp to [0, 1]
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	switch h.config.ColorScheme {
	case "viridis":
		return viridis(t)
	case "plasma":
		return plasma(t)
	case "coolwarm":
		return coolwarm(t)
	default:
		return viridis(t)
	}
}

// viridis colormap (matplotlib-style)
func viridis(t float64) RGB {
	// Simplified viridis approximation
	r := 0.267004 + t*(0.282327-0.267004+t*(0.748818-0.282327))
	g := 0.004874 + t*(0.140926-0.004874+t*(0.983868-0.140926))
	b := 0.329415 + t*(0.457517-0.329415+t*(0.144061-0.457517))

	return RGB{
		R: math.Max(0, math.Min(1, r)),
		G: math.Max(0, math.Min(1, g)),
		B: math.Max(0, math.Min(1, b)),
	}
}

// plasma colormap
func plasma(t float64) RGB {
	r := 0.050383 + t*(0.940015-0.050383)
	g := 0.029803 + t*(0.975158-0.029803)
	b := 0.527975 + t*(0.131326-0.527975)

	return RGB{
		R: math.Max(0, math.Min(1, r)),
		G: math.Max(0, math.Min(1, g)),
		B: math.Max(0, math.Min(1, b)),
	}
}

// coolwarm colormap (diverging)
func coolwarm(t float64) RGB {
	// Blue (0) -> White (0.5) -> Red (1)
	if t < 0.5 {
		s := t * 2 // 0 to 1 for blue to white
		return RGB{
			R: s,
			G: s,
			B: 1.0,
		}
	} else {
		s := (t - 0.5) * 2 // 0 to 1 for white to red
		return RGB{
			R: 1.0,
			G: 1.0 - s,
			B: 1.0 - s,
		}
	}
}

// PrintASCII prints an ASCII representation of the heatmap.
func (h *CrossbarHeatmap) PrintASCII() {
	chars := []rune{' ', '░', '▒', '▓', '█'}

	fmt.Printf("Crossbar Heatmap (%dx%d)\n", h.rows, h.cols)
	fmt.Printf("Range: [%.3f, %.3f]\n", h.minVal, h.maxVal)

	for i := 0; i < h.rows; i++ {
		for j := 0; j < h.cols; j++ {
			normVal := (h.data[i][j] - h.minVal) / (h.maxVal - h.minVal)
			idx := int(normVal * float64(len(chars)-1))
			if idx >= len(chars) {
				idx = len(chars) - 1
			}
			fmt.Printf("%c", chars[idx])
		}
		fmt.Println()
	}
}

// GetStats returns statistics about the heatmap data.
func (h *CrossbarHeatmap) GetStats() (min, max, mean, std float64) {
	min = h.minVal
	max = h.maxVal

	var sum float64
	count := float64(h.rows * h.cols)

	for i := 0; i < h.rows; i++ {
		for j := 0; j < h.cols; j++ {
			sum += h.data[i][j]
		}
	}
	mean = sum / count

	var variance float64
	for i := 0; i < h.rows; i++ {
		for j := 0; j < h.cols; j++ {
			diff := h.data[i][j] - mean
			variance += diff * diff
		}
	}
	std = math.Sqrt(variance / count)

	return
}

// HighlightCell adds a highlight marker for a specific cell.
func (h *CrossbarHeatmap) HighlightCell(row, col int) []HeatmapVertex {
	if row < 0 || row >= h.rows || col < 0 || col >= h.cols {
		return nil
	}

	cellW := 2.0 / float64(h.cols)
	cellH := 2.0 / float64(h.rows)

	// Calculate cell center
	cx := -1.0 + (float64(col)+0.5)*cellW
	cy := 1.0 - (float64(row)+0.5)*cellH

	// Create highlight ring
	var vertices []HeatmapVertex
	highlightColor := [4]float32{1.0, 1.0, 0.0, 1.0} // Yellow

	size := math.Min(cellW, cellH) * 0.4
	segments := 16

	for i := 0; i < segments; i++ {
		angle1 := float64(i) * 2 * math.Pi / float64(segments)
		angle2 := float64(i+1) * 2 * math.Pi / float64(segments)

		x1 := cx + size*math.Cos(angle1)
		y1 := cy + size*math.Sin(angle1)
		x2 := cx + size*math.Cos(angle2)
		y2 := cy + size*math.Sin(angle2)

		vertices = append(vertices,
			HeatmapVertex{Position: [2]float32{float32(cx), float32(cy)}, Color: highlightColor},
			HeatmapVertex{Position: [2]float32{float32(x1), float32(y1)}, Color: highlightColor},
			HeatmapVertex{Position: [2]float32{float32(x2), float32(y2)}, Color: highlightColor},
		)
	}

	return vertices
}
