package thermal

import (
	"fmt"
	"math"
	"strings"
)

// HeatMapRenderer renders thermal data as ASCII art.
type HeatMapRenderer struct {
	MinTemp   float64 // Minimum temperature for scale
	MaxTemp   float64 // Maximum temperature for scale
	UseColor  bool    // Enable ANSI color output
	ShowScale bool    // Show temperature scale
	Width     int     // Rendered width (0 = auto)
	Height    int     // Rendered height (0 = auto)
}

// DefaultRenderer returns a renderer with typical settings.
func DefaultRenderer() *HeatMapRenderer {
	return &HeatMapRenderer{
		MinTemp:   25.0,
		MaxTemp:   85.0,
		UseColor:  true,
		ShowScale: true,
		Width:     0,
		Height:    0,
	}
}

// heatChars defines ASCII characters for heat intensity (low to high).
var heatChars = []rune{'░', '▒', '▓', '█'}

// Render converts a thermal grid to an ASCII heat map string.
func (r *HeatMapRenderer) Render(sim *ThermalSim) string {
	grid := sim.GetGridCopy()
	height := len(grid)
	if height == 0 {
		return ""
	}
	width := len(grid[0])

	// Auto-scale if not set
	minT, maxT := r.MinTemp, r.MaxTemp
	if maxT <= minT {
		minT = sim.GetMinTemperature()
		maxT = sim.GetMaxTemperature()
		if maxT <= minT {
			maxT = minT + 1
		}
	}

	var sb strings.Builder

	// Render grid
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			temp := grid[y][x]
			char, color := r.tempToChar(temp, minT, maxT)

			if r.UseColor && color != "" {
				sb.WriteString(color)
				sb.WriteRune(char)
				sb.WriteString("\033[0m")
			} else {
				sb.WriteRune(char)
			}
		}
		sb.WriteString("\n")
	}

	// Add scale if requested
	if r.ShowScale {
		sb.WriteString("\n")
		sb.WriteString(r.renderScale(minT, maxT))
	}

	return sb.String()
}

// tempToChar converts temperature to display character and color.
func (r *HeatMapRenderer) tempToChar(temp, minT, maxT float64) (rune, string) {
	// Normalize to 0-1 range
	normalized := (temp - minT) / (maxT - minT)
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}

	// Select character based on intensity
	charIndex := int(normalized * float64(len(heatChars)-1))
	if charIndex >= len(heatChars) {
		charIndex = len(heatChars) - 1
	}

	// Select color based on temperature
	var color string
	if normalized < 0.25 {
		color = "\033[34m" // Blue (cool)
	} else if normalized < 0.5 {
		color = "\033[32m" // Green (warm)
	} else if normalized < 0.75 {
		color = "\033[33m" // Yellow (hot)
	} else {
		color = "\033[31m" // Red (critical)
	}

	return heatChars[charIndex], color
}

// renderScale creates a temperature scale legend.
func (r *HeatMapRenderer) renderScale(minT, maxT float64) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%.0f°C ", minT))
	for i, char := range heatChars {
		temp := minT + (maxT-minT)*float64(i)/float64(len(heatChars)-1)
		_, color := r.tempToChar(temp, minT, maxT)
		if r.UseColor && color != "" {
			sb.WriteString(color)
			sb.WriteRune(char)
			sb.WriteString("\033[0m")
		} else {
			sb.WriteRune(char)
		}
	}
	sb.WriteString(fmt.Sprintf(" %.0f°C", maxT))

	return sb.String()
}

// RenderSideView creates a side view showing heat through layers.
func (r *HeatMapRenderer) RenderSideView(layers []*ThermalSim) string {
	if len(layers) == 0 {
		return ""
	}

	var sb strings.Builder

	// Find global min/max
	minT := math.MaxFloat64
	maxT := -math.MaxFloat64
	for _, layer := range layers {
		if layer.GetMinTemperature() < minT {
			minT = layer.GetMinTemperature()
		}
		if layer.GetMaxTemperature() > maxT {
			maxT = layer.GetMaxTemperature()
		}
	}
	if maxT <= minT {
		maxT = minT + 1
	}

	sb.WriteString("Side View (Layers):\n")
	sb.WriteString("───────────────────\n")

	// Render each layer (top to bottom)
	for i := len(layers) - 1; i >= 0; i-- {
		layer := layers[i]
		grid := layer.GetGridCopy()

		// Take a horizontal slice through middle
		midY := len(grid) / 2
		if midY < len(grid) {
			sb.WriteString(fmt.Sprintf("Layer %d: ", i+1))
			for x := 0; x < len(grid[midY]); x++ {
				temp := grid[midY][x]
				char, color := r.tempToChar(temp, minT, maxT)
				if r.UseColor && color != "" {
					sb.WriteString(color)
					sb.WriteRune(char)
					sb.WriteString("\033[0m")
				} else {
					sb.WriteRune(char)
				}
			}
			sb.WriteString(fmt.Sprintf(" (avg: %.1f°C)\n", layer.GetAverageTemperature()))
		}

		// Heat flow indicator between layers
		if i > 0 {
			sb.WriteString("         ")
			for j := 0; j < len(grid[0])/2; j++ {
				sb.WriteString(" ↕")
			}
			sb.WriteString(" heat flow\n")
		}
	}

	sb.WriteString("         ")
	for j := 0; j < len(layers[0].Grid[0]); j++ {
		sb.WriteString("░")
	}
	sb.WriteString(" Heat Sink\n")

	return sb.String()
}

// RenderStats creates a statistics summary.
func (r *HeatMapRenderer) RenderStats(sim *ThermalSim) string {
	var sb strings.Builder

	sb.WriteString("Thermal Statistics:\n")
	sb.WriteString("───────────────────\n")
	sb.WriteString(fmt.Sprintf("  Grid Size: %d × %d\n", sim.Width, sim.Height))
	sb.WriteString(fmt.Sprintf("  Min Temperature: %.2f°C\n", sim.GetMinTemperature()))
	sb.WriteString(fmt.Sprintf("  Max Temperature: %.2f°C\n", sim.GetMaxTemperature()))
	sb.WriteString(fmt.Sprintf("  Average Temperature: %.2f°C\n", sim.GetAverageTemperature()))
	sb.WriteString(fmt.Sprintf("  Ambient Temperature: %.2f°C\n", sim.AmbientTemp))
	sb.WriteString(fmt.Sprintf("  Max Safe Temperature: %.2f°C\n", sim.MaxTemp))
	sb.WriteString(fmt.Sprintf("  Total Heat Generation: %.2e W/m²\n", sim.TotalHeatGeneration()))

	// Hotspot analysis
	hotspots := sim.FindHotspots(sim.MaxTemp * 0.6)
	sb.WriteString(fmt.Sprintf("  Hotspots (>%.0f°C): %d\n", sim.MaxTemp*0.6, len(hotspots)))

	// Warning check
	warning := sim.CheckThermalWarning()
	if warning != nil {
		sb.WriteString(fmt.Sprintf("\n  [Level %d] %s\n", warning.Level, warning.Message))
	} else {
		sb.WriteString("\n  Status: Normal operating temperature\n")
	}

	return sb.String()
}

// RenderWithOverlay renders the heat map with hotspot markers.
func (r *HeatMapRenderer) RenderWithOverlay(sim *ThermalSim) string {
	grid := sim.GetGridCopy()
	height := len(grid)
	if height == 0 {
		return ""
	}
	width := len(grid[0])

	minT, maxT := r.MinTemp, r.MaxTemp
	if maxT <= minT {
		minT = sim.GetMinTemperature()
		maxT = sim.GetMaxTemperature()
		if maxT <= minT {
			maxT = minT + 1
		}
	}

	// Find hotspots
	hotspots := sim.FindHotspots(sim.MaxTemp * 0.75)
	hotspotMap := make(map[string]bool)
	for _, h := range hotspots {
		key := fmt.Sprintf("%d,%d", h.X, h.Y)
		hotspotMap[key] = true
	}

	var sb strings.Builder

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			temp := grid[y][x]
			key := fmt.Sprintf("%d,%d", x, y)

			if hotspotMap[key] {
				// Mark hotspots with special character
				if r.UseColor {
					sb.WriteString("\033[41;37m") // Red background, white text
					sb.WriteString("!")
					sb.WriteString("\033[0m")
				} else {
					sb.WriteString("!")
				}
			} else {
				char, color := r.tempToChar(temp, minT, maxT)
				if r.UseColor && color != "" {
					sb.WriteString(color)
					sb.WriteRune(char)
					sb.WriteString("\033[0m")
				} else {
					sb.WriteRune(char)
				}
			}
		}
		sb.WriteString("\n")
	}

	if r.ShowScale {
		sb.WriteString("\n")
		sb.WriteString(r.renderScale(minT, maxT))
		sb.WriteString("  !")
		if r.UseColor {
			sb.WriteString("\033[41;37m")
			sb.WriteString("=hotspot")
			sb.WriteString("\033[0m")
		} else {
			sb.WriteString("=hotspot")
		}
	}

	return sb.String()
}

// TemperatureGradient calculates and renders temperature gradient vectors.
func (r *HeatMapRenderer) TemperatureGradient(sim *ThermalSim) string {
	grid := sim.GetGridCopy()
	height := len(grid)
	if height == 0 {
		return ""
	}
	width := len(grid[0])

	var sb strings.Builder
	sb.WriteString("Temperature Gradient (Heat Flow Direction):\n")
	sb.WriteString("────────────────────────────────────────────\n")

	// Sample every other cell for clarity
	for y := 1; y < height-1; y += 2 {
		for x := 1; x < width-1; x += 2 {
			// Calculate gradient
			dTdx := (grid[y][x+1] - grid[y][x-1]) / 2
			dTdy := (grid[y+1][x] - grid[y-1][x]) / 2

			// Heat flows opposite to gradient
			arrow := gradientToArrow(-dTdx, -dTdy)
			sb.WriteString(arrow)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// gradientToArrow converts gradient direction to arrow character.
func gradientToArrow(dx, dy float64) string {
	mag := math.Sqrt(dx*dx + dy*dy)
	if mag < 0.1 {
		return "·"
	}

	angle := math.Atan2(dy, dx) * 180 / math.Pi

	// 8 directions
	if angle < -157.5 || angle >= 157.5 {
		return "←"
	} else if angle >= -157.5 && angle < -112.5 {
		return "↙"
	} else if angle >= -112.5 && angle < -67.5 {
		return "↓"
	} else if angle >= -67.5 && angle < -22.5 {
		return "↘"
	} else if angle >= -22.5 && angle < 22.5 {
		return "→"
	} else if angle >= 22.5 && angle < 67.5 {
		return "↗"
	} else if angle >= 67.5 && angle < 112.5 {
		return "↑"
	} else {
		return "↖"
	}
}
