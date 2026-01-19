package multilayer

import (
	"fmt"
	"strings"
)

// StackRenderer renders the 3D stack as ASCII art.
type StackRenderer struct {
	ShowVias     bool
	ShowDataFlow bool
	ShowMetrics  bool
	UseColor     bool
	Compact      bool
}

// DefaultRenderer returns a renderer with default settings.
func DefaultRenderer() *StackRenderer {
	return &StackRenderer{
		ShowVias:     true,
		ShowDataFlow: true,
		ShowMetrics:  true,
		UseColor:     true,
		Compact:      false,
	}
}

// Render3DView creates an isometric 3D view of the stack.
func (r *StackRenderer) Render3DView(stack *Stack) string {
	var sb strings.Builder

	numLayers := len(stack.Layers)
	if numLayers == 0 {
		return "Empty stack"
	}

	// Find max dimensions for scaling
	maxCols := 0
	for _, layer := range stack.Layers {
		if layer.Cols > maxCols {
			maxCols = layer.Cols
		}
	}

	// Scale factor for display
	displayWidth := 40
	if maxCols > 0 && maxCols < displayWidth {
		displayWidth = maxCols + 10
	}

	sb.WriteString("3D Stack View (Isometric):\n")
	sb.WriteString("══════════════════════════════════════════\n\n")

	// Render from top layer to bottom
	for i := numLayers - 1; i >= 0; i-- {
		layer := stack.Layers[i]
		indent := strings.Repeat(" ", (numLayers-1-i)*2)

		// Top edge of layer
		width := r.scaleWidth(layer.Cols, maxCols, displayWidth)
		topEdge := "╔" + strings.Repeat("═", width) + "╗"

		sb.WriteString(indent)
		if r.UseColor {
			sb.WriteString("\033[36m") // Cyan
		}
		sb.WriteString(fmt.Sprintf("     %s\n", topEdge))

		// Layer name and dimensions
		sb.WriteString(indent)
		layerInfo := fmt.Sprintf("Layer %d: %s (%d×%d)", i+1, layer.Name, layer.Rows, layer.Cols)
		paddedInfo := r.centerText(layerInfo, width)
		sb.WriteString(fmt.Sprintf("    ╱%s╱│\n", paddedInfo))

		// Layer body
		sb.WriteString(indent)
		sb.WriteString(fmt.Sprintf("   ╔%s╗ │\n", strings.Repeat("═", width)))

		// Cell representation
		sb.WriteString(indent)
		cells := r.renderLayerCells(layer, width)
		sb.WriteString(fmt.Sprintf("   ║%s║╱\n", cells))

		// Bottom edge
		sb.WriteString(indent)
		sb.WriteString(fmt.Sprintf("   ╚%s╝\n", strings.Repeat("═", width)))

		if r.UseColor {
			sb.WriteString("\033[0m")
		}

		// Via connections
		if r.ShowVias && i > 0 {
			sb.WriteString(indent)
			viaCount := layer.Rows
			sb.WriteString(fmt.Sprintf("        ↑ %d vias ↑\n", viaCount))
		}

		if !r.Compact {
			sb.WriteString("\n")
		}
	}

	// Input arrow
	sb.WriteString(strings.Repeat(" ", (numLayers-1)*2))
	sb.WriteString(fmt.Sprintf("           ↑\n"))
	sb.WriteString(strings.Repeat(" ", (numLayers-1)*2))
	sb.WriteString(fmt.Sprintf("       Input (%d)\n", stack.Layers[0].Rows))

	return sb.String()
}

// RenderExplodedView creates an exploded view showing all layers separately.
func (r *StackRenderer) RenderExplodedView(stack *Stack) string {
	var sb strings.Builder

	sb.WriteString("Exploded View (Layer Details):\n")
	sb.WriteString("══════════════════════════════════════════\n\n")

	for i, layer := range stack.Layers {
		sb.WriteString(fmt.Sprintf("┌─── Layer %d: %s ───────────────────────┐\n", i+1, layer.Name))
		sb.WriteString(fmt.Sprintf("│  Dimensions: %d × %d = %d cells          │\n",
			layer.Rows, layer.Cols, layer.Rows*layer.Cols))
		sb.WriteString(fmt.Sprintf("│  Levels: %d (%.1f bits/cell)             │\n",
			layer.Levels, stack.BitsPerCell()))
		sb.WriteString(fmt.Sprintf("│  Activation: %s                        │\n", layer.Activation))
		sb.WriteString("│                                          │\n")

		// Visual representation
		sb.WriteString("│  ")
		sb.WriteString(r.renderMiniLayer(layer))
		sb.WriteString("   │\n")

		sb.WriteString("└──────────────────────────────────────────┘\n")

		// Connection to next layer
		if i < len(stack.Layers)-1 {
			sb.WriteString("                    │\n")
			sb.WriteString(fmt.Sprintf("                    ▼ %d outputs → %d inputs\n",
				layer.Cols, stack.Layers[i+1].Rows))
			sb.WriteString("                    │\n")
		}
	}

	return sb.String()
}

// RenderDataFlow visualizes data flow through the stack.
func (r *StackRenderer) RenderDataFlow(stack *Stack, input []float64) string {
	var sb strings.Builder

	sb.WriteString("Data Flow Visualization:\n")
	sb.WriteString("══════════════════════════════════════════\n\n")

	if len(input) == 0 {
		// Create sample input
		input = make([]float64, stack.Layers[0].Rows)
		for i := range input {
			input[i] = float64(i%10) / 10.0
		}
	}

	activation := input
	for i, layer := range stack.Layers {
		// Show input size
		sb.WriteString(fmt.Sprintf("Layer %d Input: %d values\n", i+1, len(activation)))
		sb.WriteString(r.renderActivationBar(activation, 40))
		sb.WriteString("\n")

		// Show layer operation
		sb.WriteString(fmt.Sprintf("    ┌─%s─┐\n", strings.Repeat("─", 30)))
		sb.WriteString(fmt.Sprintf("    │ %s: %d×%d MVM", layer.Name, layer.Rows, layer.Cols))
		sb.WriteString(strings.Repeat(" ", 18-len(layer.Name)))
		sb.WriteString("│\n")
		sb.WriteString(fmt.Sprintf("    │ %d MAC operations", layer.Rows*layer.Cols))
		sb.WriteString(strings.Repeat(" ", 21-len(fmt.Sprintf("%d", layer.Rows*layer.Cols))))
		sb.WriteString("│\n")
		sb.WriteString(fmt.Sprintf("    └─%s─┘\n", strings.Repeat("─", 30)))
		sb.WriteString("           │\n")
		sb.WriteString("           ▼\n")

		// Simulate forward pass
		output := make([]float64, layer.Cols)
		for j := 0; j < layer.Cols; j++ {
			sum := 0.0
			for k := 0; k < layer.Rows && k < len(activation); k++ {
				weight := float64(layer.Weights[k][j]-15) / 15.0
				sum += activation[k] * weight
			}
			if layer.Activation == "relu" && sum < 0 {
				sum = 0
			}
			output[j] = sum
		}
		activation = output
	}

	sb.WriteString(fmt.Sprintf("Output: %d values\n", len(activation)))
	sb.WriteString(r.renderActivationBar(activation, 40))
	sb.WriteString("\n")

	return sb.String()
}

// RenderMetrics displays stack metrics and comparisons.
func (r *StackRenderer) RenderMetrics(stack *Stack) string {
	var sb strings.Builder

	sb.WriteString("Stack Metrics:\n")
	sb.WriteString("══════════════════════════════════════════\n\n")

	// Basic metrics
	sb.WriteString(fmt.Sprintf("Total Layers:      %d\n", len(stack.Layers)))
	sb.WriteString(fmt.Sprintf("Total Cells:       %d\n", stack.TotalCells()))
	sb.WriteString(fmt.Sprintf("Total Parameters:  %d\n", stack.TotalParameters()))
	sb.WriteString(fmt.Sprintf("Bits per Cell:     %.2f\n", stack.BitsPerCell()))
	sb.WriteString(fmt.Sprintf("Total Bits:        %.0f (%.2f KB)\n",
		stack.TotalBits(), stack.TotalBits()/8/1024))
	sb.WriteString("\n")

	// Physical metrics
	sb.WriteString("Physical Characteristics:\n")
	sb.WriteString(fmt.Sprintf("  Cell Pitch:       %.0f nm\n", stack.CellPitch))
	sb.WriteString(fmt.Sprintf("  Layer Height:     %.0f nm\n", stack.LayerHeight))
	sb.WriteString(fmt.Sprintf("  Stack Height:     %.0f nm (%.3f µm)\n",
		stack.StackHeight(), stack.StackHeight()/1000))
	sb.WriteString(fmt.Sprintf("  Footprint Area:   %.2f µm²\n", stack.FootprintArea()))
	sb.WriteString(fmt.Sprintf("  Areal Density:    %.2f bits/µm²\n", stack.ArealDensity()))
	sb.WriteString(fmt.Sprintf("  Volume Density:   %.2f bits/µm³\n", stack.VolumetricDensity()))
	sb.WriteString("\n")

	// Via network
	viaNet := NewViaNetwork(stack)
	viaStats := viaNet.GetStats(stack.FootprintArea())
	sb.WriteString("Via Network:\n")
	sb.WriteString(fmt.Sprintf("  Total Vias:       %d\n", viaStats.TotalVias))
	sb.WriteString(fmt.Sprintf("  Via Arrays:       %d\n", viaStats.ViaArrays))
	sb.WriteString(fmt.Sprintf("  Total Length:     %.2f µm\n", viaStats.TotalLength))
	sb.WriteString(fmt.Sprintf("  Propagation:      %.3f ps\n", viaStats.PropagationDelay))
	sb.WriteString(fmt.Sprintf("  Via Density:      %.2f vias/µm²\n", viaStats.ViaDensity))
	sb.WriteString("\n")

	// Energy estimates
	sb.WriteString("Energy Comparison (per inference):\n")
	energyEst := stack.EstimateEnergy()
	totalCIM := 0.0
	totalTraditional := 0.0
	for _, e := range energyEst {
		totalCIM += e.TotalEnergy
		totalTraditional += e.TotalEnergy * e.TraditionalComp
	}
	sb.WriteString(fmt.Sprintf("  IronLattice:      %.3f pJ\n", totalCIM))
	sb.WriteString(fmt.Sprintf("  Traditional:      %.1f pJ\n", totalTraditional))
	sb.WriteString(fmt.Sprintf("  Advantage:        %.0fx lower energy!\n", totalTraditional/totalCIM))
	sb.WriteString("\n")

	return sb.String()
}

// Helper functions

func (r *StackRenderer) scaleWidth(cols, maxCols, displayWidth int) int {
	if maxCols == 0 {
		return displayWidth
	}
	return int(float64(cols) / float64(maxCols) * float64(displayWidth))
}

func (r *StackRenderer) centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
}

func (r *StackRenderer) renderLayerCells(layer *Layer, width int) string {
	// Show weight distribution
	chars := []rune{'░', '▒', '▓', '█'}

	// Sample weights
	sampleSize := width - 2
	if sampleSize > layer.Cols {
		sampleSize = layer.Cols
	}

	var sb strings.Builder
	for i := 0; i < sampleSize; i++ {
		// Sample from middle row
		row := layer.Rows / 2
		col := i * layer.Cols / sampleSize
		if row < len(layer.Weights) && col < len(layer.Weights[row]) {
			level := layer.Weights[row][col]
			charIdx := level * len(chars) / layer.Levels
			if charIdx >= len(chars) {
				charIdx = len(chars) - 1
			}
			sb.WriteRune(chars[charIdx])
		} else {
			sb.WriteRune('·')
		}
	}

	// Pad to width
	for sb.Len() < width {
		sb.WriteRune(' ')
	}

	return sb.String()
}

func (r *StackRenderer) renderMiniLayer(layer *Layer) string {
	// Compact representation
	width := 30
	height := 3

	rows := make([]string, height)
	for h := 0; h < height; h++ {
		var sb strings.Builder
		for w := 0; w < width; w++ {
			// Sample from layer
			row := h * layer.Rows / height
			col := w * layer.Cols / width
			if row < len(layer.Weights) && col < len(layer.Weights[row]) {
				level := layer.Weights[row][col]
				if level > 20 {
					sb.WriteRune('█')
				} else if level > 10 {
					sb.WriteRune('▓')
				} else if level > 5 {
					sb.WriteRune('▒')
				} else {
					sb.WriteRune('░')
				}
			} else {
				sb.WriteRune('·')
			}
		}
		rows[h] = sb.String()
	}

	return strings.Join(rows, "\n│  ")
}

func (r *StackRenderer) renderActivationBar(values []float64, width int) string {
	if len(values) == 0 {
		return "[empty]"
	}

	// Find max for normalization
	maxVal := 0.0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 1.0
	}

	// Sample values
	sampleSize := width
	if len(values) < sampleSize {
		sampleSize = len(values)
	}

	var sb strings.Builder
	sb.WriteString("[")

	chars := []rune{'_', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	for i := 0; i < sampleSize; i++ {
		idx := i * len(values) / sampleSize
		val := values[idx] / maxVal
		charIdx := int(val * float64(len(chars)-1))
		if charIdx >= len(chars) {
			charIdx = len(chars) - 1
		}
		if charIdx < 0 {
			charIdx = 0
		}
		sb.WriteRune(chars[charIdx])
	}

	sb.WriteString("]")
	return sb.String()
}
