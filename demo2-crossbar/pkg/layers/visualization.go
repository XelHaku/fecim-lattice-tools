// Package layers provides visualization utilities for CIM educational demonstrations.
// Implements ASCII-based crossbar visualization, weight heatmaps, and training animations.
package layers

import (
	"fmt"
	"math"
	"strings"
)

// ============================================================================
// CROSSBAR VISUALIZATION
// ============================================================================

// CrossbarVisualizer creates ASCII visualizations of crossbar arrays.
type CrossbarVisualizer struct {
	Rows       int
	Cols       int
	CellWidth  int
	ShowValues bool
}

// NewCrossbarVisualizer creates a new visualizer.
func NewCrossbarVisualizer(rows, cols int) *CrossbarVisualizer {
	return &CrossbarVisualizer{
		Rows:       rows,
		Cols:       cols,
		CellWidth:  6,
		ShowValues: true,
	}
}

// VisualizeCrossbar renders a crossbar array as ASCII art.
func (v *CrossbarVisualizer) VisualizeCrossbar(weights [][]float64, title string) string {
	var sb strings.Builder

	// Title
	if title != "" {
		sb.WriteString(fmt.Sprintf("=== %s ===\n", title))
	}

	rows := min(v.Rows, len(weights))
	cols := 0
	if rows > 0 {
		cols = min(v.Cols, len(weights[0]))
	}

	// Column headers (DAC inputs)
	sb.WriteString("     ")
	for j := 0; j < cols; j++ {
		sb.WriteString(fmt.Sprintf(" V%-4d", j))
	}
	sb.WriteString("\n")

	// Top border
	sb.WriteString("    ┌")
	for j := 0; j < cols; j++ {
		sb.WriteString("─────")
		if j < cols-1 {
			sb.WriteString("┬")
		}
	}
	sb.WriteString("┐\n")

	// Rows with weights
	for i := 0; i < rows; i++ {
		// Row label (wordline)
		sb.WriteString(fmt.Sprintf("WL%d │", i))

		for j := 0; j < cols; j++ {
			w := weights[i][j]
			cell := v.formatCell(w)
			sb.WriteString(cell)
			if j < cols-1 {
				sb.WriteString("│")
			}
		}
		sb.WriteString("│\n")

		// Row separator
		if i < rows-1 {
			sb.WriteString("    ├")
			for j := 0; j < cols; j++ {
				sb.WriteString("─────")
				if j < cols-1 {
					sb.WriteString("┼")
				}
			}
			sb.WriteString("┤\n")
		}
	}

	// Bottom border
	sb.WriteString("    └")
	for j := 0; j < cols; j++ {
		sb.WriteString("─────")
		if j < cols-1 {
			sb.WriteString("┴")
		}
	}
	sb.WriteString("┘\n")

	// Column labels (ADC outputs)
	sb.WriteString("     ")
	for j := 0; j < cols; j++ {
		sb.WriteString(fmt.Sprintf(" I%-4d", j))
	}
	sb.WriteString("\n")

	return sb.String()
}

// formatCell formats a weight value for display.
func (v *CrossbarVisualizer) formatCell(w float64) string {
	if v.ShowValues {
		if math.Abs(w) < 0.01 {
			return "  ·  "
		}
		return fmt.Sprintf("%5.2f", w)
	}
	// Show as intensity block
	intensity := int((w + 1) / 2 * 4) // Map [-1,1] to [0,4]
	blocks := []string{" ", "░", "▒", "▓", "█"}
	if intensity < 0 {
		intensity = 0
	}
	if intensity > 4 {
		intensity = 4
	}
	return fmt.Sprintf("  %s  ", blocks[intensity])
}

// VisualizeMVM shows a matrix-vector multiplication step.
func (v *CrossbarVisualizer) VisualizeMVM(weights [][]float64, input []float64) string {
	var sb strings.Builder

	rows := min(v.Rows, len(weights))
	cols := 0
	if rows > 0 {
		cols = min(v.Cols, len(weights[0]))
	}

	sb.WriteString("=== Matrix-Vector Multiplication ===\n\n")

	// Show input voltages
	sb.WriteString("Input voltages (DAC):\n")
	sb.WriteString("  [")
	for i := 0; i < min(cols, len(input)); i++ {
		sb.WriteString(fmt.Sprintf(" %.2f", input[i]))
	}
	sb.WriteString(" ]\n\n")

	// Show crossbar with current flow
	sb.WriteString("Crossbar operation:\n")
	sb.WriteString("  V_in → [ G ] → I_out\n\n")

	// Compute output currents
	output := make([]float64, rows)
	for i := 0; i < rows; i++ {
		for j := 0; j < min(cols, len(input)); j++ {
			if j < len(weights[i]) {
				output[i] += weights[i][j] * input[j]
			}
		}
	}

	// Show computation
	sb.WriteString("  I[i] = Σ G[i,j] × V[j]\n\n")

	sb.WriteString("Output currents (ADC):\n")
	sb.WriteString("  [")
	for i := 0; i < rows; i++ {
		sb.WriteString(fmt.Sprintf(" %.2f", output[i]))
	}
	sb.WriteString(" ]\n")

	return sb.String()
}

// ============================================================================
// WEIGHT HEATMAP
// ============================================================================

// HeatmapVisualizer creates ASCII heatmaps.
type HeatmapVisualizer struct {
	Width  int
	Height int
	Chars  []rune
}

// NewHeatmapVisualizer creates a new heatmap visualizer.
func NewHeatmapVisualizer(width, height int) *HeatmapVisualizer {
	return &HeatmapVisualizer{
		Width:  width,
		Height: height,
		Chars:  []rune{' ', '·', '░', '▒', '▓', '█'},
	}
}

// Visualize creates a heatmap from a 2D array.
func (h *HeatmapVisualizer) Visualize(data [][]float64, title string, minVal, maxVal float64) string {
	var sb strings.Builder

	if title != "" {
		sb.WriteString(fmt.Sprintf("%s\n", title))
	}

	rows := len(data)
	cols := 0
	if rows > 0 {
		cols = len(data[0])
	}

	// Calculate scaling
	scaleY := float64(rows) / float64(h.Height)
	scaleX := float64(cols) / float64(h.Width)
	valueRange := maxVal - minVal
	if valueRange == 0 {
		valueRange = 1
	}

	// Generate heatmap
	for y := 0; y < h.Height; y++ {
		for x := 0; x < h.Width; x++ {
			// Sample from data
			dataY := int(float64(y) * scaleY)
			dataX := int(float64(x) * scaleX)

			if dataY >= rows {
				dataY = rows - 1
			}
			if dataX >= cols {
				dataX = cols - 1
			}

			value := data[dataY][dataX]
			normalized := (value - minVal) / valueRange
			charIdx := int(normalized * float64(len(h.Chars)-1))
			if charIdx < 0 {
				charIdx = 0
			}
			if charIdx >= len(h.Chars) {
				charIdx = len(h.Chars) - 1
			}

			sb.WriteRune(h.Chars[charIdx])
		}
		sb.WriteString("\n")
	}

	// Legend
	sb.WriteString("\nLegend: ")
	for i, c := range h.Chars {
		val := minVal + float64(i)/float64(len(h.Chars)-1)*valueRange
		sb.WriteString(fmt.Sprintf("%c=%.1f ", c, val))
	}
	sb.WriteString("\n")

	return sb.String()
}

// ============================================================================
// TRAINING VISUALIZATION
// ============================================================================

// TrainingVisualizer shows training progress.
type TrainingVisualizer struct {
	MaxWidth   int
	ShowLoss   bool
	ShowAcc    bool
	History    []TrainingStep
}

// TrainingStep records one training step.
type TrainingStep struct {
	Epoch    int
	Batch    int
	Loss     float64
	Accuracy float64
}

// NewTrainingVisualizer creates a new training visualizer.
func NewTrainingVisualizer() *TrainingVisualizer {
	return &TrainingVisualizer{
		MaxWidth: 50,
		ShowLoss: true,
		ShowAcc:  true,
		History:  make([]TrainingStep, 0),
	}
}

// RecordStep adds a training step.
func (t *TrainingVisualizer) RecordStep(epoch, batch int, loss, accuracy float64) {
	t.History = append(t.History, TrainingStep{
		Epoch:    epoch,
		Batch:    batch,
		Loss:     loss,
		Accuracy: accuracy,
	})
}

// VisualizeLossCurve shows loss over training.
func (t *TrainingVisualizer) VisualizeLossCurve() string {
	var sb strings.Builder
	sb.WriteString("=== Training Loss ===\n\n")

	if len(t.History) == 0 {
		sb.WriteString("No training data recorded.\n")
		return sb.String()
	}

	// Find min/max loss
	minLoss := t.History[0].Loss
	maxLoss := t.History[0].Loss
	for _, step := range t.History {
		if step.Loss < minLoss {
			minLoss = step.Loss
		}
		if step.Loss > maxLoss {
			maxLoss = step.Loss
		}
	}

	lossRange := maxLoss - minLoss
	if lossRange == 0 {
		lossRange = 1
	}

	// Draw loss curve
	height := 10
	for row := height - 1; row >= 0; row-- {
		threshold := minLoss + float64(row)/float64(height-1)*lossRange
		sb.WriteString(fmt.Sprintf("%6.3f │", threshold))

		// Sample history
		step := len(t.History) / t.MaxWidth
		if step == 0 {
			step = 1
		}

		for i := 0; i < t.MaxWidth && i*step < len(t.History); i++ {
			loss := t.History[i*step].Loss
			normalized := (loss - minLoss) / lossRange * float64(height-1)
			if int(normalized) >= row {
				sb.WriteString("█")
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	// X-axis
	sb.WriteString("       └")
	sb.WriteString(strings.Repeat("─", t.MaxWidth))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("        Epoch 0%sEpoch %d\n",
		strings.Repeat(" ", t.MaxWidth-15),
		t.History[len(t.History)-1].Epoch))

	return sb.String()
}

// VisualizeAccuracyCurve shows accuracy over training.
func (t *TrainingVisualizer) VisualizeAccuracyCurve() string {
	var sb strings.Builder
	sb.WriteString("=== Training Accuracy ===\n\n")

	if len(t.History) == 0 {
		sb.WriteString("No training data recorded.\n")
		return sb.String()
	}

	height := 10
	for row := height - 1; row >= 0; row-- {
		threshold := float64(row) / float64(height-1)
		sb.WriteString(fmt.Sprintf("%5.1f%% │", threshold*100))

		step := len(t.History) / t.MaxWidth
		if step == 0 {
			step = 1
		}

		for i := 0; i < t.MaxWidth && i*step < len(t.History); i++ {
			acc := t.History[i*step].Accuracy
			normalized := acc * float64(height-1)
			if int(normalized) >= row {
				sb.WriteString("▓")
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	// X-axis
	sb.WriteString("       └")
	sb.WriteString(strings.Repeat("─", t.MaxWidth))
	sb.WriteString("\n")

	return sb.String()
}

// ============================================================================
// POLARIZATION VISUALIZATION
// ============================================================================

// PolarizationVisualizer shows ferroelectric P-E hysteresis.
type PolarizationVisualizer struct {
	Width  int
	Height int
}

// NewPolarizationVisualizer creates a new polarization visualizer.
func NewPolarizationVisualizer() *PolarizationVisualizer {
	return &PolarizationVisualizer{
		Width:  60,
		Height: 20,
	}
}

// VisualizeHysteresis draws a P-E hysteresis curve.
func (p *PolarizationVisualizer) VisualizeHysteresis(Ec, Pr, Ps float64) string {
	var sb strings.Builder

	sb.WriteString("=== Ferroelectric P-E Hysteresis ===\n\n")
	sb.WriteString(fmt.Sprintf("Parameters: Ec=%.2f V/nm, Pr=%.2f µC/cm², Ps=%.2f µC/cm²\n\n", Ec, Pr, Ps))

	// Create canvas
	canvas := make([][]rune, p.Height)
	for i := range canvas {
		canvas[i] = make([]rune, p.Width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	// Draw axes
	midY := p.Height / 2
	midX := p.Width / 2

	for x := 0; x < p.Width; x++ {
		canvas[midY][x] = '─'
	}
	for y := 0; y < p.Height; y++ {
		canvas[y][midX] = '│'
	}
	canvas[midY][midX] = '┼'

	// Draw hysteresis loop using tanh approximation
	for x := 0; x < p.Width; x++ {
		E := (float64(x) - float64(midX)) / float64(midX) * 2 * Ec

		// Upper branch (increasing E)
		P_upper := Ps * math.Tanh((E+Ec)/(Ec*0.5))
		y_upper := midY - int(P_upper/Ps*float64(midY-1))
		if y_upper >= 0 && y_upper < p.Height {
			canvas[y_upper][x] = '●'
		}

		// Lower branch (decreasing E)
		P_lower := Ps * math.Tanh((E-Ec)/(Ec*0.5))
		y_lower := midY - int(P_lower/Ps*float64(midY-1))
		if y_lower >= 0 && y_lower < p.Height {
			canvas[y_lower][x] = '○'
		}
	}

	// Render canvas
	sb.WriteString("  P (µC/cm²)\n")
	sb.WriteString(fmt.Sprintf("  +%.1f ┐\n", Ps))
	for _, row := range canvas {
		sb.WriteString("      ")
		sb.WriteString(string(row))
		sb.WriteString("\n")
	}
	sb.WriteString(fmt.Sprintf("  -%.1f ┘\n", Ps))
	sb.WriteString(fmt.Sprintf("        -%.1f          E (V/nm)          +%.1f\n", Ec*2, Ec*2))

	sb.WriteString("\nLegend: ● = increasing E, ○ = decreasing E\n")

	return sb.String()
}

// ============================================================================
// NETWORK ARCHITECTURE VISUALIZATION
// ============================================================================

// NetworkVisualizer shows neural network architecture.
type NetworkVisualizer struct{}

// NewNetworkVisualizer creates a new network visualizer.
func NewNetworkVisualizer() *NetworkVisualizer {
	return &NetworkVisualizer{}
}

// VisualizeArchitecture draws network layers.
func (n *NetworkVisualizer) VisualizeArchitecture(layerSizes []int, layerNames []string) string {
	var sb strings.Builder

	sb.WriteString("=== Neural Network Architecture ===\n\n")

	maxSize := 0
	for _, size := range layerSizes {
		if size > maxSize {
			maxSize = size
		}
	}

	// Draw each layer
	maxHeight := 15
	for i, size := range layerSizes {
		// Layer header
		name := fmt.Sprintf("Layer %d", i)
		if i < len(layerNames) {
			name = layerNames[i]
		}
		sb.WriteString(fmt.Sprintf("%-12s (%d neurons)\n", name, size))

		// Calculate display height
		height := int(float64(size) / float64(maxSize) * float64(maxHeight))
		if height < 1 {
			height = 1
		}

		// Draw neurons
		for h := 0; h < height; h++ {
			sb.WriteString("  ")
			if h == 0 {
				sb.WriteString("┌")
			} else if h == height-1 {
				sb.WriteString("└")
			} else {
				sb.WriteString("│")
			}

			width := min(size, 30)
			for w := 0; w < width; w++ {
				sb.WriteString("○")
			}
			if size > 30 {
				sb.WriteString("...")
			}

			if h == 0 {
				sb.WriteString("┐")
			} else if h == height-1 {
				sb.WriteString("┘")
			} else {
				sb.WriteString("│")
			}
			sb.WriteString("\n")
		}

		// Connection to next layer
		if i < len(layerSizes)-1 {
			sb.WriteString("      │\n")
			sb.WriteString("      ▼\n")
		}
	}

	return sb.String()
}

// ============================================================================
// CROSSBAR MAPPING VISUALIZATION
// ============================================================================

// MappingVisualizer shows how layers map to crossbar arrays.
type MappingVisualizer struct{}

// NewMappingVisualizer creates a new mapping visualizer.
func NewMappingVisualizer() *MappingVisualizer {
	return &MappingVisualizer{}
}

// VisualizeTiling shows weight tiling across arrays.
func (m *MappingVisualizer) VisualizeTiling(layerRows, layerCols, arrayRows, arrayCols int) string {
	var sb strings.Builder

	sb.WriteString("=== Weight Tiling Visualization ===\n\n")
	sb.WriteString(fmt.Sprintf("Layer size: %d × %d\n", layerRows, layerCols))
	sb.WriteString(fmt.Sprintf("Array size: %d × %d\n\n", arrayRows, arrayCols))

	tilesY := int(math.Ceil(float64(layerRows) / float64(arrayRows)))
	tilesX := int(math.Ceil(float64(layerCols) / float64(arrayCols)))

	sb.WriteString(fmt.Sprintf("Required tiles: %d × %d = %d arrays\n\n", tilesY, tilesX, tilesY*tilesX))

	// Draw tile layout
	for ty := 0; ty < tilesY; ty++ {
		// Top border
		for tx := 0; tx < tilesX; tx++ {
			if tx == 0 {
				sb.WriteString("┌")
			} else {
				sb.WriteString("┬")
			}
			sb.WriteString("───────")
		}
		sb.WriteString("┐\n")

		// Tile content
		startRow := ty * arrayRows
		endRow := min((ty+1)*arrayRows, layerRows)
		for tx := 0; tx < tilesX; tx++ {
			startCol := tx * arrayCols
			endCol := min((tx+1)*arrayCols, layerCols)
			sb.WriteString(fmt.Sprintf("│[%d:%d,", startRow, endRow))
			sb.WriteString(fmt.Sprintf("%d:%d]", startCol, endCol))
		}
		sb.WriteString("│\n")

		// Array ID
		for tx := 0; tx < tilesX; tx++ {
			arrayID := ty*tilesX + tx
			sb.WriteString(fmt.Sprintf("│Array%-2d", arrayID))
		}
		sb.WriteString("│\n")
	}

	// Bottom border
	for tx := 0; tx < tilesX; tx++ {
		if tx == 0 {
			sb.WriteString("└")
		} else {
			sb.WriteString("┴")
		}
		sb.WriteString("───────")
	}
	sb.WriteString("┘\n")

	return sb.String()
}

// ============================================================================
// SPARSITY VISUALIZATION
// ============================================================================

// SparsityVisualizer shows weight sparsity patterns.
type SparsityVisualizer struct{}

// NewSparsityVisualizer creates a new sparsity visualizer.
func NewSparsityVisualizer() *SparsityVisualizer {
	return &SparsityVisualizer{}
}

// VisualizeSparsity shows sparsity pattern in weights.
func (s *SparsityVisualizer) VisualizeSparsity(weights [][]float64, threshold float64) string {
	var sb strings.Builder

	sb.WriteString("=== Weight Sparsity Pattern ===\n\n")

	rows := len(weights)
	cols := 0
	if rows > 0 {
		cols = len(weights[0])
	}

	// Count sparsity
	total := 0
	zeros := 0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			total++
			if math.Abs(weights[i][j]) < threshold {
				zeros++
			}
		}
	}

	sparsity := float64(zeros) / float64(total) * 100
	sb.WriteString(fmt.Sprintf("Sparsity: %.1f%% (threshold=%.3f)\n", sparsity, threshold))
	sb.WriteString(fmt.Sprintf("Non-zero: %d / %d\n\n", total-zeros, total))

	// Visualize pattern (downsampled)
	displayRows := min(20, rows)
	displayCols := min(60, cols)

	scaleY := float64(rows) / float64(displayRows)
	scaleX := float64(cols) / float64(displayCols)

	for y := 0; y < displayRows; y++ {
		for x := 0; x < displayCols; x++ {
			dataY := int(float64(y) * scaleY)
			dataX := int(float64(x) * scaleX)

			if dataY >= rows {
				dataY = rows - 1
			}
			if dataX >= cols {
				dataX = cols - 1
			}

			if math.Abs(weights[dataY][dataX]) < threshold {
				sb.WriteString("·")
			} else {
				sb.WriteString("█")
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nLegend: █ = non-zero, · = zero\n")

	return sb.String()
}

// ============================================================================
// DEVICE CHARACTERISTICS
// ============================================================================

// DeviceVisualizer shows device-level characteristics.
type DeviceVisualizer struct{}

// NewDeviceVisualizer creates a new device visualizer.
func NewDeviceVisualizer() *DeviceVisualizer {
	return &DeviceVisualizer{}
}

// VisualizeConductanceLevels shows multi-level cell states.
func (d *DeviceVisualizer) VisualizeConductanceLevels(numLevels int, Gmin, Gmax float64) string {
	var sb strings.Builder

	sb.WriteString("=== FeFET Conductance Levels ===\n\n")
	sb.WriteString(fmt.Sprintf("Levels: %d (%.0f-bit)\n", numLevels, math.Log2(float64(numLevels))))
	sb.WriteString(fmt.Sprintf("Range: %.2e - %.2e S\n\n", Gmin, Gmax))

	// Draw levels
	barWidth := 40
	for level := 0; level < numLevels; level++ {
		G := Gmin + float64(level)/float64(numLevels-1)*(Gmax-Gmin)
		filled := int(float64(level) / float64(numLevels-1) * float64(barWidth))

		sb.WriteString(fmt.Sprintf("L%02d │", level))
		sb.WriteString(strings.Repeat("█", filled))
		sb.WriteString(strings.Repeat("░", barWidth-filled))
		sb.WriteString(fmt.Sprintf("│ %.2e S\n", G))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("On/Off ratio: %.0f\n", Gmax/Gmin))

	return sb.String()
}

// VisualizeEndurance shows endurance characteristics.
func (d *DeviceVisualizer) VisualizeEndurance(cycles []float64, retention []float64) string {
	var sb strings.Builder

	sb.WriteString("=== FeFET Endurance & Retention ===\n\n")

	// Endurance
	sb.WriteString("Endurance (cycles vs conductance window):\n")
	height := 8
	for row := height - 1; row >= 0; row-- {
		sb.WriteString("  │")
		for i := 0; i < len(cycles); i++ {
			normalized := cycles[i] / cycles[0] * float64(height-1)
			if int(normalized) >= row {
				sb.WriteString("█")
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("  └" + strings.Repeat("─", len(cycles)) + "\n")
	sb.WriteString("   10⁰        10⁶       10¹² cycles\n\n")

	// Retention
	sb.WriteString("Retention (time vs state separation):\n")
	for row := height - 1; row >= 0; row-- {
		sb.WriteString("  │")
		for i := 0; i < len(retention); i++ {
			normalized := retention[i] / retention[0] * float64(height-1)
			if int(normalized) >= row {
				sb.WriteString("▓")
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}
	sb.WriteString("  └" + strings.Repeat("─", len(retention)) + "\n")
	sb.WriteString("   1s      1day    10yr retention\n")

	return sb.String()
}
