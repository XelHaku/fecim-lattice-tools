// pkg/gui/tabs/learn_visuals.go
// Visual components for the Learn tab - OpenLane flow diagram and crossbar visualization

package tabs

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Colors for the diagrams
var (
	colorBgDark      = color.RGBA{0, 30, 60, 255}
	colorBoxStandard = color.RGBA{0, 80, 160, 255}
	colorBoxOurs     = color.RGBA{0, 180, 220, 255} // Cyan for our contribution
	colorArrow       = color.RGBA{100, 150, 200, 255}
	colorText        = color.RGBA{255, 255, 255, 255}
	colorTextMuted   = color.RGBA{180, 200, 220, 255}
	colorWL          = color.RGBA{255, 100, 100, 255} // Red for word lines
	colorBL          = color.RGBA{100, 200, 255, 255} // Blue for bit lines
	colorFeFET       = color.RGBA{255, 200, 50, 255}  // Gold for FeFET devices
	colorHighlight   = color.RGBA{0, 255, 150, 255}   // Green highlight
	colorStorage     = color.RGBA{100, 150, 255, 255}  // Blue for Storage mode
	colorMemory      = color.RGBA{255, 200, 50, 255}   // Gold for Memory mode
	colorCompute     = color.RGBA{0, 255, 150, 255}    // Green for Compute mode
	colorTableHeader = color.RGBA{0, 60, 120, 255}    // Dark blue table headers
	colorTableRow1   = color.RGBA{0, 40, 80, 255}     // Alternating row 1
	colorTableRow2   = color.RGBA{0, 50, 100, 255}    // Alternating row 2
)

// =============================================================================
// OPENLANE FLOW DIAGRAM
// =============================================================================

// OpenLaneFlowDiagram creates a visual pipeline diagram
func OpenLaneFlowDiagram() fyne.CanvasObject {
	// Diagram dimensions - INCREASED 40%
	boxW := float32(140)
	boxH := float32(65)
	spacing := float32(25)
	startX := float32(30)
	startY := float32(50)

	objects := []fyne.CanvasObject{}

	// Title
	title := canvas.NewText("OpenLane RTL-to-GDSII Flow", colorText)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(startX, 8))
	objects = append(objects, title)

	// Subtitle
	subtitle := canvas.NewText("CYAN = Our Array Builder files inject here", colorBoxOurs)
	subtitle.TextSize = 13
	subtitle.TextStyle = fyne.TextStyle{Bold: true}
	subtitle.Move(fyne.NewPos(startX, 30))
	objects = append(objects, subtitle)

	// Pipeline stages - row 1
	stages := []struct {
		name    string
		tool    string
		isOurs  bool
		x, y    float32
	}{
		{"Verilog", "Input", false, startX, startY},
		{"Synthesis", "Yosys", false, startX + boxW + spacing, startY},
		{"Floorplan", "OpenROAD", false, startX + 2*(boxW+spacing), startY},
		{"Placement", "RePlAce", false, startX + 3*(boxW+spacing), startY},
	}

	// Row 2 (reversed flow) - more vertical space
	row2Y := startY + boxH + spacing + 30
	stages = append(stages, []struct {
		name    string
		tool    string
		isOurs  bool
		x, y    float32
	}{
		{"CTS", "TritonCTS", false, startX + 3*(boxW+spacing), row2Y},
		{"Routing", "TritonRoute", false, startX + 2*(boxW+spacing), row2Y},
		{"Signoff", "Magic/LVS", false, startX + boxW + spacing, row2Y},
		{"GDSII", "Output", false, startX, row2Y},
	}...)

	// Mark which stages we contribute to
	stages[0].isOurs = true  // Verilog - we provide
	stages[2].isOurs = true  // Floorplan - our LEF
	stages[3].isOurs = true  // Placement - our DEF (FIXED)

	// Draw stages
	for i, stage := range stages {
		boxColor := colorBoxStandard
		if stage.isOurs {
			boxColor = colorBoxOurs
		}

		// Box
		box := canvas.NewRectangle(boxColor)
		box.Resize(fyne.NewSize(boxW, boxH))
		box.Move(fyne.NewPos(stage.x, stage.y))
		box.CornerRadius = 6
		objects = append(objects, box)

		// Stage name
		nameText := canvas.NewText(stage.name, colorText)
		nameText.TextSize = 13
		nameText.TextStyle = fyne.TextStyle{Bold: true}
		nameText.Move(fyne.NewPos(stage.x+10, stage.y+10))
		objects = append(objects, nameText)

		// Tool name
		toolText := canvas.NewText(stage.tool, colorTextMuted)
		toolText.TextSize = 11
		toolText.Move(fyne.NewPos(stage.x+10, stage.y+30))
		objects = append(objects, toolText)

		// Draw arrows between stages
		if i > 0 && i < 4 {
			// Row 1 arrows (left to right)
			arrow := createArrow(
				stages[i-1].x+boxW, stages[i-1].y+boxH/2,
				stage.x, stage.y+boxH/2,
			)
			objects = append(objects, arrow...)
		} else if i == 4 {
			// Down arrow from Placement to CTS
			arrow := createArrowDown(
				stages[3].x+boxW/2, stages[3].y+boxH,
				stage.x+boxW/2, stage.y,
			)
			objects = append(objects, arrow...)
		} else if i > 4 {
			// Row 2 arrows (right to left)
			arrow := createArrow(
				stages[i-1].x, stages[i-1].y+boxH/2,
				stage.x+boxW, stage.y+boxH/2,
			)
			objects = append(objects, arrow...)
		}
	}

	// Add "Our Files" labels
	// LEF label
	lefLabel := canvas.NewText("Our LEF", colorHighlight)
	lefLabel.TextSize = 10
	lefLabel.Move(fyne.NewPos(stages[2].x+10, stages[2].y-15))
	objects = append(objects, lefLabel)

	// DEF label
	defLabel := canvas.NewText("Our DEF (FIXED)", colorHighlight)
	defLabel.TextSize = 10
	defLabel.Move(fyne.NewPos(stages[3].x+5, stages[3].y-15))
	objects = append(objects, defLabel)

	// Verilog label
	vLabel := canvas.NewText("Our Verilog", colorHighlight)
	vLabel.TextSize = 10
	vLabel.Move(fyne.NewPos(stages[0].x+5, stages[0].y-15))
	objects = append(objects, vLabel)

	// Container with fixed size - INCREASED
	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(720, 300))

	return cont
}

// createArrow creates a horizontal arrow
func createArrow(x1, y1, x2, y2 float32) []fyne.CanvasObject {
	// Line
	line := canvas.NewLine(colorArrow)
	line.StrokeWidth = 2
	line.Position1 = fyne.NewPos(x1, y1)
	line.Position2 = fyne.NewPos(x2-8, y2)

	// Arrowhead
	head := canvas.NewLine(colorArrow)
	head.StrokeWidth = 2
	head.Position1 = fyne.NewPos(x2-12, y2-5)
	head.Position2 = fyne.NewPos(x2-4, y2)

	head2 := canvas.NewLine(colorArrow)
	head2.StrokeWidth = 2
	head2.Position1 = fyne.NewPos(x2-12, y2+5)
	head2.Position2 = fyne.NewPos(x2-4, y2)

	return []fyne.CanvasObject{line, head, head2}
}

// createArrowDown creates a vertical down arrow
func createArrowDown(x1, y1, x2, y2 float32) []fyne.CanvasObject {
	line := canvas.NewLine(colorArrow)
	line.StrokeWidth = 2
	line.Position1 = fyne.NewPos(x1, y1)
	line.Position2 = fyne.NewPos(x2, y2-8)

	head := canvas.NewLine(colorArrow)
	head.StrokeWidth = 2
	head.Position1 = fyne.NewPos(x2-5, y2-12)
	head.Position2 = fyne.NewPos(x2, y2-4)

	head2 := canvas.NewLine(colorArrow)
	head2.StrokeWidth = 2
	head2.Position1 = fyne.NewPos(x2+5, y2-12)
	head2.Position2 = fyne.NewPos(x2, y2-4)

	return []fyne.CanvasObject{line, head, head2}
}

// =============================================================================
// ISOMETRIC CROSSBAR DIAGRAM
// =============================================================================

// IsometricCrossbar creates an isometric view of a crossbar array
func IsometricCrossbar(rows, cols int, showLabels bool) fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// Isometric projection parameters - INCREASED 40%
	cellSize := float32(52)
	isoAngle := float32(30 * math.Pi / 180) // 30 degrees
	cosA := float32(math.Cos(float64(isoAngle)))
	sinA := float32(math.Sin(float64(isoAngle)))

	// Starting position
	startX := float32(180)
	startY := float32(60)

	// Layer separation (Z height) - INCREASED
	layerGap := float32(70)

	// Convert grid coordinates to isometric
	toIso := func(gridX, gridY, z float32) (float32, float32) {
		x := startX + (gridX-gridY)*cellSize*cosA
		y := startY + (gridX+gridY)*cellSize*sinA - z
		return x, y
	}

	// Title
	title := canvas.NewText("PASSIVE Crossbar Structure", colorText)
	title.TextSize = 16
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(20, 10))
	objects = append(objects, title)

	// Draw bottom layer (Word Lines - horizontal in real space)
	for i := 0; i <= rows; i++ {
		x1, y1 := toIso(0, float32(i), 0)
		x2, y2 := toIso(float32(cols), float32(i), 0)

		line := canvas.NewLine(colorWL)
		line.StrokeWidth = 4
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		// WL label
		if showLabels && i < rows {
			label := canvas.NewText("WL"+string(rune('0'+i)), colorWL)
			label.TextSize = 12
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-35, y1-5))
			objects = append(objects, label)
		}
	}

	// Draw top layer (Bit Lines - vertical in real space)
	for j := 0; j <= cols; j++ {
		x1, y1 := toIso(float32(j), 0, layerGap)
		x2, y2 := toIso(float32(j), float32(rows), layerGap)

		line := canvas.NewLine(colorBL)
		line.StrokeWidth = 4
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		// BL label
		if showLabels && j < cols {
			label := canvas.NewText("BL"+string(rune('0'+j)), colorBL)
			label.TextSize = 12
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-5, y1-25))
			objects = append(objects, label)
		}
	}

	// Draw FeFET devices at intersections (vertical pillars)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Bottom connection point (on WL layer)
			x1, y1 := toIso(float32(j)+0.5, float32(i)+0.5, 0)
			// Top connection point (on BL layer)
			x2, y2 := toIso(float32(j)+0.5, float32(i)+0.5, layerGap)

			// Vertical connector (the FeFET)
			pillar := canvas.NewLine(colorFeFET)
			pillar.StrokeWidth = 5
			pillar.Position1 = fyne.NewPos(x1, y1)
			pillar.Position2 = fyne.NewPos(x2, y2)
			objects = append(objects, pillar)

			// Larger circle at center to represent the device
			midX := (x1 + x2) / 2
			midY := (y1 + y2) / 2
			device := canvas.NewCircle(colorFeFET)
			device.Resize(fyne.NewSize(12, 12))
			device.Move(fyne.NewPos(midX-6, midY-6))
			objects = append(objects, device)
		}
	}

	// Legend - moved down
	legendY := float32(240)
	legendX := float32(20)

	// WL legend
	wlBox := canvas.NewRectangle(colorWL)
	wlBox.Resize(fyne.NewSize(20, 12))
	wlBox.Move(fyne.NewPos(legendX, legendY))
	objects = append(objects, wlBox)
	wlText := canvas.NewText("Word Lines (WL) - Row Select", colorTextMuted)
	wlText.TextSize = 12
	wlText.Move(fyne.NewPos(legendX+25, legendY-2))
	objects = append(objects, wlText)

	// BL legend
	blBox := canvas.NewRectangle(colorBL)
	blBox.Resize(fyne.NewSize(20, 12))
	blBox.Move(fyne.NewPos(legendX, legendY+18))
	objects = append(objects, blBox)
	blText := canvas.NewText("Bit Lines (BL) - Data/Output", colorTextMuted)
	blText.TextSize = 12
	blText.Move(fyne.NewPos(legendX+25, legendY+16))
	objects = append(objects, blText)

	// FeFET legend
	feBox := canvas.NewCircle(colorFeFET)
	feBox.Resize(fyne.NewSize(14, 14))
	feBox.Move(fyne.NewPos(legendX+3, legendY+36))
	objects = append(objects, feBox)
	feText := canvas.NewText("FeFET Device (stores weight/data)", colorTextMuted)
	feText.TextSize = 12
	feText.Move(fyne.NewPos(legendX+25, legendY+36))
	objects = append(objects, feText)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(540, 400))

	return cont
}

// Isometric1T1RCrossbar creates an isometric view of a 1T1R crossbar array
func Isometric1T1RCrossbar(rows, cols int) fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// Isometric projection parameters
	cellSize := float32(52)
	isoAngle := float32(30 * math.Pi / 180)
	cosA := float32(math.Cos(float64(isoAngle)))
	sinA := float32(math.Sin(float64(isoAngle)))

	startX := float32(180)
	startY := float32(60)
	layerGap := float32(70)

	toIso := func(gridX, gridY, z float32) (float32, float32) {
		x := startX + (gridX-gridY)*cellSize*cosA
		y := startY + (gridX+gridY)*cellSize*sinA - z
		return x, y
	}

	// Title
	title := canvas.NewText("1T1R Crossbar Structure", colorText)
	title.TextSize = 16
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(20, 10))
	objects = append(objects, title)

	subtitle := canvas.NewText("(Transistor isolates each cell - NO sneak paths)", colorHighlight)
	subtitle.TextSize = 12
	subtitle.Move(fyne.NewPos(20, 30))
	objects = append(objects, subtitle)

	// Draw WL layer (bottom)
	for i := 0; i <= rows; i++ {
		x1, y1 := toIso(0, float32(i), 0)
		x2, y2 := toIso(float32(cols), float32(i), 0)

		line := canvas.NewLine(colorWL)
		line.StrokeWidth = 4
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		if i < rows {
			label := canvas.NewText("WL"+string(rune('0'+i)), colorWL)
			label.TextSize = 12
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-35, y1-5))
			objects = append(objects, label)
		}
	}

	// Draw BL layer (top)
	for j := 0; j <= cols; j++ {
		x1, y1 := toIso(float32(j), 0, layerGap)
		x2, y2 := toIso(float32(j), float32(rows), layerGap)

		line := canvas.NewLine(colorBL)
		line.StrokeWidth = 4
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		if j < cols {
			label := canvas.NewText("BL"+string(rune('0'+j)), colorBL)
			label.TextSize = 12
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-5, y1-25))
			objects = append(objects, label)
		}
	}

	// Source Line color (green)
	colorSL := color.RGBA{100, 255, 100, 255}

	// Draw SL layer (middle - unique to 1T1R)
	slZ := layerGap / 2
	for j := 0; j <= cols; j++ {
		x1, y1 := toIso(float32(j), 0, slZ)
		x2, y2 := toIso(float32(j), float32(rows), slZ)

		line := canvas.NewLine(colorSL)
		line.StrokeWidth = 3
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		if j < cols {
			label := canvas.NewText("SL"+string(rune('0'+j)), colorSL)
			label.TextSize = 10
			label.Move(fyne.NewPos(x2+5, y2-5))
			objects = append(objects, label)
		}
	}

	// Draw 1T1R cells (transistor symbol + FeFET)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			x1, y1 := toIso(float32(j)+0.5, float32(i)+0.5, 0)
			xMid, yMid := toIso(float32(j)+0.5, float32(i)+0.5, slZ)
			x2, y2 := toIso(float32(j)+0.5, float32(i)+0.5, layerGap)

			// Bottom part (transistor) - square symbol
			transistor := canvas.NewRectangle(colorHighlight)
			transistor.Resize(fyne.NewSize(10, 10))
			transistor.Move(fyne.NewPos(x1-5, y1-5))
			objects = append(objects, transistor)

			// Line from transistor to FeFET
			conn := canvas.NewLine(colorFeFET)
			conn.StrokeWidth = 3
			conn.Position1 = fyne.NewPos(x1, y1)
			conn.Position2 = fyne.NewPos(x2, y2)
			objects = append(objects, conn)

			// FeFET device (circle at top)
			device := canvas.NewCircle(colorFeFET)
			device.Resize(fyne.NewSize(12, 12))
			device.Move(fyne.NewPos(xMid-6, yMid-6))
			objects = append(objects, device)
		}
	}

	// Legend
	legendY := float32(240)
	legendX := float32(20)

	// Transistor legend
	tBox := canvas.NewRectangle(colorHighlight)
	tBox.Resize(fyne.NewSize(14, 14))
	tBox.Move(fyne.NewPos(legendX+3, legendY))
	objects = append(objects, tBox)
	tText := canvas.NewText("Select Transistor (gate on WL)", colorTextMuted)
	tText.TextSize = 12
	tText.Move(fyne.NewPos(legendX+25, legendY))
	objects = append(objects, tText)

	// SL legend
	slBox := canvas.NewRectangle(colorSL)
	slBox.Resize(fyne.NewSize(20, 12))
	slBox.Move(fyne.NewPos(legendX, legendY+18))
	objects = append(objects, slBox)
	slText := canvas.NewText("Source Lines (SL) - Current path", colorTextMuted)
	slText.TextSize = 12
	slText.Move(fyne.NewPos(legendX+25, legendY+18))
	objects = append(objects, slText)

	// FeFET legend
	feBox := canvas.NewCircle(colorFeFET)
	feBox.Resize(fyne.NewSize(14, 14))
	feBox.Move(fyne.NewPos(legendX+3, legendY+36))
	objects = append(objects, feBox)
	feText := canvas.NewText("FeFET Device (stores weight)", colorTextMuted)
	feText.TextSize = 12
	feText.Move(fyne.NewPos(legendX+25, legendY+36))
	objects = append(objects, feText)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(540, 400))

	return cont
}

// =============================================================================
// CELL COMPARISON TABLE
// =============================================================================

// CellComparisonTable creates a visual comparison table for Passive vs 1T1R cells
func CellComparisonTable() fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// Table dimensions
	colWidths := []float32{100, 100, 100, 100} // Property, Passive, 1T1R, Winner
	rowHeight := float32(24)
	startX := float32(10)
	startY := float32(10)

	// Table data
	type row struct {
		property string
		passive  string
		t1r      string
		winner   string // "Passive" or "1T1R"
	}

	rows := []row{
		{"Property", "Passive", "1T1R", "Winner"}, // Header
		{"Cell Size", "0.46x2.72um", "0.92x2.72um", "Passive"},
		{"Area (um²)", "1.25", "2.50", "Passive"},
		{"Sneak Path", "HIGH", "NONE", "1T1R"},
		{"Max Array", "~32x32", "128x128+", "1T1R"},
		{"Port Count", "WL,BL", "WL,BL,SL", "Passive"},
		{"Fab Complex", "Simple", "Moderate", "Passive"},
	}

	// Draw table
	for i, rowData := range rows {
		y := startY + float32(i)*rowHeight
		x := startX

		// Row background
		var bgColor color.Color
		if i == 0 {
			bgColor = colorTableHeader
		} else if i%2 == 1 {
			bgColor = colorTableRow1
		} else {
			bgColor = colorTableRow2
		}

		rowBg := canvas.NewRectangle(bgColor)
		rowBg.Resize(fyne.NewSize(400, rowHeight))
		rowBg.Move(fyne.NewPos(x, y))
		objects = append(objects, rowBg)

		// Cells
		cells := []string{rowData.property, rowData.passive, rowData.t1r, rowData.winner}
		for j, cellText := range cells {
			cellX := x + float32(j)*colWidths[j]

			// Cell text
			text := canvas.NewText(cellText, colorText)
			if i == 0 {
				text.TextSize = 11
				text.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				text.TextSize = 10
			}

			// Winner column gets special coloring
			if j == 3 && i > 0 {
				if cellText == "1T1R" {
					text.Color = colorHighlight
					text.TextStyle = fyne.TextStyle{Bold: true}
				} else {
					text.Color = colorBoxOurs
					text.TextStyle = fyne.TextStyle{Bold: true}
				}
			}

			text.Move(fyne.NewPos(cellX+5, y+5))
			objects = append(objects, text)

			// Vertical grid line (except after last column)
			if j < len(cells)-1 {
				lineX := cellX + colWidths[j]
				line := canvas.NewLine(colorArrow)
				line.StrokeWidth = 1
				line.Position1 = fyne.NewPos(lineX, y)
				line.Position2 = fyne.NewPos(lineX, y+rowHeight)
				objects = append(objects, line)
			}
		}

		// Horizontal grid line after row
		hLine := canvas.NewLine(colorArrow)
		hLine.StrokeWidth = 1
		hLine.Position1 = fyne.NewPos(x, y+rowHeight)
		hLine.Position2 = fyne.NewPos(x+400, y+rowHeight)
		objects = append(objects, hLine)
	}

	// Outer border
	border := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	border.StrokeColor = colorArrow
	border.StrokeWidth = 2
	border.Resize(fyne.NewSize(400, float32(len(rows))*rowHeight))
	border.Move(fyne.NewPos(startX, startY))
	objects = append(objects, border)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(420, 200))

	return cont
}

// =============================================================================
// FECIM OPERATION MODES VISUAL
// =============================================================================

// OperationModesVisual creates a visual diagram showing the 3 FeCIM operation modes
func OperationModesVisual() fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// Title
	title := canvas.NewText("FeCIM Operation Modes", colorText)
	title.TextSize = 16
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(110, 10))
	objects = append(objects, title)

	// Box dimensions
	boxW := float32(140)
	boxH := float32(110)
	spacing := float32(15)
	startX := float32(10)
	startY := float32(40)

	// Mode definitions
	modes := []struct {
		name        string
		description string
		color       color.Color
		x           float32
	}{
		{"STORAGE", "Flash-like, ~4.9 bits/cell", colorStorage, startX},
		{"MEMORY", "DRAM-like, Fast R/W", colorMemory, startX + boxW + spacing},
		{"COMPUTE", "MVM/MAC, in-place", colorCompute, startX + 2*(boxW+spacing)},
	}

	// Draw mode boxes
	for _, mode := range modes {
		// Box background
		box := canvas.NewRectangle(mode.color)
		box.Resize(fyne.NewSize(boxW, boxH))
		box.Move(fyne.NewPos(mode.x, startY))
		box.CornerRadius = 8
		objects = append(objects, box)

		// Mode name (centered in box)
		nameText := canvas.NewText(mode.name, colorBgDark)
		nameText.TextSize = 12
		nameText.TextStyle = fyne.TextStyle{Bold: true}
		// Center text horizontally
		nameX := mode.x + (boxW-float32(len(mode.name))*7)/2
		nameText.Move(fyne.NewPos(nameX, startY+25))
		objects = append(objects, nameText)

		// Description (centered below name)
		descText := canvas.NewText(mode.description, colorBgDark)
		descText.TextSize = 10
		// Center description horizontally (approximate)
		descX := mode.x + 8
		descText.Move(fyne.NewPos(descX, startY+50))
		objects = append(objects, descText)
	}

	// Central FeCIM circle position
	circleX := float32(210)  // Center of the diagram
	circleY := float32(165)  // Below the boxes
	circleRadius := float32(20)

	// Draw lines from each box to central circle
	for i, mode := range modes {
		// Start point: bottom center of box
		x1 := mode.x + boxW/2
		y1 := startY + boxH

		// End point: edge of circle (top)
		line := canvas.NewLine(colorArrow)
		line.StrokeWidth = 2
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(circleX, circleY-circleRadius)

		// Adjust line endpoints for visual clarity
		if i == 0 {
			// Left box - line goes to left side of circle
			line.Position2 = fyne.NewPos(circleX-circleRadius*0.7, circleY-circleRadius*0.7)
		} else if i == 2 {
			// Right box - line goes to right side of circle
			line.Position2 = fyne.NewPos(circleX+circleRadius*0.7, circleY-circleRadius*0.7)
		}

		objects = append(objects, line)
	}

	// Central FeCIM circle
	circle := canvas.NewCircle(colorBoxOurs)
	circle.Resize(fyne.NewSize(circleRadius*2, circleRadius*2))
	circle.Move(fyne.NewPos(circleX-circleRadius, circleY-circleRadius))
	objects = append(objects, circle)

	// "FeCIM" label in center
	fecimLabel := canvas.NewText("FeCIM", colorBgDark)
	fecimLabel.TextSize = 12
	fecimLabel.TextStyle = fyne.TextStyle{Bold: true}
	fecimLabel.Move(fyne.NewPos(circleX-22, circleY-8))
	objects = append(objects, fecimLabel)

	// Quote below circle
	quote := canvas.NewText("\"Same device does all\"", colorBoxOurs)
	quote.TextSize = 11
	quote.TextStyle = fyne.TextStyle{Italic: true}
	quote.Move(fyne.NewPos(circleX-75, circleY+30))
	objects = append(objects, quote)

	// Container with fixed size
	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(560, 260))

	return cont
}

// =============================================================================
// FILE FORMAT PREVIEW CARDS
// =============================================================================

// FileFormatCard creates a styled card showing file format examples
func FileFormatCard(title, format, content string) fyne.CanvasObject {
	// Header - INCREASED HEIGHT for better readability
	headerBg := canvas.NewRectangle(colorBoxOurs)
	headerBg.Resize(fyne.NewSize(360, 40))
	headerBg.CornerRadius = 6

	titleText := canvas.NewText(title+" (."+format+")", colorBgDark)
	titleText.TextSize = 14
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.Move(fyne.NewPos(12, 10))

	// Content area - INCREASED HEIGHT for better code visibility
	contentBg := canvas.NewRectangle(color.RGBA{0, 25, 50, 255})
	contentBg.Resize(fyne.NewSize(360, 180))
	contentBg.Move(fyne.NewPos(0, 40))
	contentBg.CornerRadius = 4

	// Code content with monospace style and better line spacing
	codeLabel := widget.NewLabel(content)
	codeLabel.Wrapping = fyne.TextWrapOff
	codeLabel.TextStyle = fyne.TextStyle{Monospace: true}

	codeContainer := container.NewPadded(codeLabel)
	codeContainer.Move(fyne.NewPos(8, 44))
	codeContainer.Resize(fyne.NewSize(344, 172))

	card := container.NewWithoutLayout(headerBg, titleText, contentBg, codeContainer)
	card.Resize(fyne.NewSize(360, 224))

	return card
}

// LEFPreviewCard shows a LEF file example
func LEFPreviewCard() fyne.CanvasObject {
	content := `MACRO fecim_bitcell
  SIZE 0.460 BY 2.720 ;
  PIN WL
    DIRECTION INPUT ;
  END WL
END fecim_bitcell`
	return FileFormatCard("LEF", "lef", content)
}

// DEFPreviewCard shows a DEF file example
func DEFPreviewCard() fyne.CanvasObject {
	content := `COMPONENTS 16 ;
  - cell_0_0 fecim_bitcell
    + FIXED ( 10000 10000 ) N ;
  - cell_0_1 fecim_bitcell
    + FIXED ( 10460 10000 ) N ;
END COMPONENTS`
	return FileFormatCard("DEF", "def", content)
}

// VerilogPreviewCard shows a Verilog file example
func VerilogPreviewCard() fyne.CanvasObject {
	content := `module fecim_array_4x4 (
  input  [3:0] WL,
  inout  [3:0] BL
);
  fecim_bitcell c00(.WL(WL[0]),.BL(BL[0]));
  // ... 16 cells
endmodule`
	return FileFormatCard("Verilog", "v", content)
}

// LibertyPreviewCard shows a Liberty timing file example
func LibertyPreviewCard() fyne.CanvasObject {
	content := `library(fecim_cells) {
  cell(fecim_bitcell) {
    area : 1.2512 ;
    pin(WL) { capacitance: 0.002; }
    timing() { cell_rise: 0.1ns; }
  }
}`
	return FileFormatCard("Liberty", "lib", content)
}

// =============================================================================
// REFERENCES CARD
// =============================================================================

// ReferencesCard creates an IEEE-formatted references section with category headers
func ReferencesCard() fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// Card dimensions
	cardWidth := float32(420)
	cardHeight := float32(310)

	// Category header dimensions
	headerHeight := float32(26)
	categorySpacing := float32(8)

	// Starting positions
	startX := float32(0)
	currentY := float32(0)

	// Category 1: EDA / OpenLane
	header1 := canvas.NewRectangle(colorBoxOurs)
	header1.Resize(fyne.NewSize(cardWidth, headerHeight))
	header1.Move(fyne.NewPos(startX, currentY))
	header1.CornerRadius = 4
	objects = append(objects, header1)

	headerText1 := canvas.NewText("EDA / OpenLane", colorText)
	headerText1.TextSize = 12
	headerText1.TextStyle = fyne.TextStyle{Bold: true}
	headerText1.Move(fyne.NewPos(startX+10, currentY+6))
	objects = append(objects, headerText1)

	currentY += headerHeight

	// References for EDA / OpenLane
	refs1 := `[1] M. Shalan et al., "OpenLANE: Digital ASIC Flow," WOSET, 2020.
[2] LEF/DEF 5.8 Specification, Si2 Coalition.`

	refBg1 := canvas.NewRectangle(colorBgDark)
	refBg1.Resize(fyne.NewSize(cardWidth, 50))
	refBg1.Move(fyne.NewPos(startX, currentY))
	objects = append(objects, refBg1)

	refLabel1 := widget.NewLabel(refs1)
	refLabel1.Wrapping = fyne.TextWrapWord
	refLabel1.TextStyle = fyne.TextStyle{}
	refContainer1 := container.NewPadded(refLabel1)
	refContainer1.Move(fyne.NewPos(startX+5, currentY+2))
	refContainer1.Resize(fyne.NewSize(cardWidth-10, 46))
	objects = append(objects, refContainer1)

	currentY += 50 + categorySpacing

	// Category 2: FeCIM Device Physics
	header2 := canvas.NewRectangle(colorBoxOurs)
	header2.Resize(fyne.NewSize(cardWidth, headerHeight))
	header2.Move(fyne.NewPos(startX, currentY))
	header2.CornerRadius = 4
	objects = append(objects, header2)

	headerText2 := canvas.NewText("FeCIM Device Physics", colorText)
	headerText2.TextSize = 12
	headerText2.TextStyle = fyne.TextStyle{Bold: true}
	headerText2.Move(fyne.NewPos(startX+10, currentY+6))
	objects = append(objects, headerText2)

	currentY += headerHeight

	// References for FeCIM Device Physics
	refs2 := `[3] S. Shin et al., "Flash In2Se3," Adv. Electron. Mater., 2025.
[4] U. Schroeder et al., "Roadmap Ferroelectric HfZrO," APL Mater., 2023.`

	refBg2 := canvas.NewRectangle(colorBgDark)
	refBg2.Resize(fyne.NewSize(cardWidth, 50))
	refBg2.Move(fyne.NewPos(startX, currentY))
	objects = append(objects, refBg2)

	refLabel2 := widget.NewLabel(refs2)
	refLabel2.Wrapping = fyne.TextWrapWord
	refLabel2.TextStyle = fyne.TextStyle{}
	refContainer2 := container.NewPadded(refLabel2)
	refContainer2.Move(fyne.NewPos(startX+5, currentY+2))
	refContainer2.Resize(fyne.NewSize(cardWidth-10, 46))
	objects = append(objects, refContainer2)

	currentY += 50 + categorySpacing

	// Category 3: CIM Architecture
	header3 := canvas.NewRectangle(colorBoxOurs)
	header3.Resize(fyne.NewSize(cardWidth, headerHeight))
	header3.Move(fyne.NewPos(startX, currentY))
	header3.CornerRadius = 4
	objects = append(objects, header3)

	headerText3 := canvas.NewText("CIM Architecture", colorText)
	headerText3.TextSize = 12
	headerText3.TextStyle = fyne.TextStyle{Bold: true}
	headerText3.Move(fyne.NewPos(startX+10, currentY+6))
	objects = append(objects, headerText3)

	currentY += headerHeight

	// References for CIM Architecture
	refs3 := `[5] Y. Chen, "Sneak Path Solutions," RSC Nanoscale Adv., 2020.
[6] P. Chen, NeuroSim, Georgia Tech, 2021.`

	refBg3 := canvas.NewRectangle(colorBgDark)
	refBg3.Resize(fyne.NewSize(cardWidth, 50))
	refBg3.Move(fyne.NewPos(startX, currentY))
	objects = append(objects, refBg3)

	refLabel3 := widget.NewLabel(refs3)
	refLabel3.Wrapping = fyne.TextWrapWord
	refLabel3.TextStyle = fyne.TextStyle{}
	refContainer3 := container.NewPadded(refLabel3)
	refContainer3.Move(fyne.NewPos(startX+5, currentY+2))
	refContainer3.Resize(fyne.NewSize(cardWidth-10, 46))
	objects = append(objects, refContainer3)

	// Container with fixed size
	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(cardWidth, cardHeight))

	return cont
}
