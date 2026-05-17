//go:build legacy_fyne

// pkg/gui/tabs/learn_visuals_array.go
// Crossbar array layout and isometric visualizations

package tabs

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

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

	// Starting position - shifted right to make room for labels
	startX := float32(240)
	startY := float32(80)

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

		// WL label - moved further left to avoid overlap
		if showLabels && i < rows {
			label := canvas.NewText("WL"+string(rune('0'+i)), colorWL)
			label.TextSize = 14
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-50, y1-5))
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

		// BL label - moved higher to avoid overlap
		if showLabels && j < cols {
			label := canvas.NewText("BL"+string(rune('0'+j)), colorBL)
			label.TextSize = 14
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-5, y1-35))
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

	// Calculate diagram bounds for proper legend placement
	// The rightmost point is at toIso(cols, 0, layerGap) for BL layer
	// The bottommost point is at toIso(cols, rows, 0) for WL layer
	_, bottomY := toIso(float32(cols), float32(rows), 0)

	// Legend - positioned below the lowest diagram element with margin
	legendY := bottomY + float32(40)
	legendX := float32(20)

	// WL legend
	wlBox := canvas.NewRectangle(colorWL)
	wlBox.Resize(fyne.NewSize(20, 12))
	wlBox.Move(fyne.NewPos(legendX, legendY))
	objects = append(objects, wlBox)
	wlText := canvas.NewText("Word Lines (WL) - Row Select", colorTextMuted)
	wlText.TextSize = 14
	wlText.Move(fyne.NewPos(legendX+25, legendY-2))
	objects = append(objects, wlText)

	// BL legend
	blBox := canvas.NewRectangle(colorBL)
	blBox.Resize(fyne.NewSize(20, 12))
	blBox.Move(fyne.NewPos(legendX, legendY+18))
	objects = append(objects, blBox)
	blText := canvas.NewText("Bit Lines (BL) - Data/Output", colorTextMuted)
	blText.TextSize = 14
	blText.Move(fyne.NewPos(legendX+25, legendY+16))
	objects = append(objects, blText)

	// FeFET legend
	feBox := canvas.NewCircle(colorFeFET)
	feBox.Resize(fyne.NewSize(14, 14))
	feBox.Move(fyne.NewPos(legendX+3, legendY+36))
	objects = append(objects, feBox)
	feText := canvas.NewText("FeFET Device (stores weight/data)", colorTextMuted)
	feText.TextSize = 14
	feText.Move(fyne.NewPos(legendX+25, legendY+36))
	objects = append(objects, feText)

	// Calculate proper container size based on diagram bounds
	// Width: need space for labels on left (-50) and diagram extends right
	rightX, _ := toIso(float32(cols), 0, layerGap)
	totalWidth := rightX + float32(50) // Add margin on right

	// Height: legend bottom + text height + margin
	totalHeight := legendY + float32(60)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(totalWidth, totalHeight))

	return cont
}

// Isometric1T1RCrossbar creates an isometric view of a 1T1R crossbar array
func Isometric1T1RCrossbar(rows, cols int) fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// Isometric projection parameters - matched to passive crossbar
	cellSize := float32(52)
	isoAngle := float32(30 * math.Pi / 180)
	cosA := float32(math.Cos(float64(isoAngle)))
	sinA := float32(math.Sin(float64(isoAngle)))

	startX := float32(240)
	startY := float32(80)
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
	subtitle.TextSize = 14
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
			label.TextSize = 14
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-50, y1-5))
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
			label.TextSize = 14
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-5, y1-35))
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
			label.TextSize = 14
			label.Move(fyne.NewPos(x2+8, y2-5))
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

	// Calculate diagram bounds for proper legend placement
	_, bottomY := toIso(float32(cols), float32(rows), 0)

	// Legend - positioned below the lowest diagram element with margin
	legendY := bottomY + float32(40)
	legendX := float32(20)

	// Transistor legend
	tBox := canvas.NewRectangle(colorHighlight)
	tBox.Resize(fyne.NewSize(14, 14))
	tBox.Move(fyne.NewPos(legendX+3, legendY))
	objects = append(objects, tBox)
	tText := canvas.NewText("Select Transistor (gate on WL)", colorTextMuted)
	tText.TextSize = 14
	tText.Move(fyne.NewPos(legendX+25, legendY))
	objects = append(objects, tText)

	// SL legend
	slBox := canvas.NewRectangle(colorSL)
	slBox.Resize(fyne.NewSize(20, 12))
	slBox.Move(fyne.NewPos(legendX, legendY+18))
	objects = append(objects, slBox)
	slText := canvas.NewText("Source Lines (SL) - Current path", colorTextMuted)
	slText.TextSize = 14
	slText.Move(fyne.NewPos(legendX+25, legendY+18))
	objects = append(objects, slText)

	// FeFET legend
	feBox := canvas.NewCircle(colorFeFET)
	feBox.Resize(fyne.NewSize(14, 14))
	feBox.Move(fyne.NewPos(legendX+3, legendY+36))
	objects = append(objects, feBox)
	feText := canvas.NewText("FeFET Device (stores weight)", colorTextMuted)
	feText.TextSize = 14
	feText.Move(fyne.NewPos(legendX+25, legendY+36))
	objects = append(objects, feText)

	// Calculate proper container size based on diagram bounds
	rightX, _ := toIso(float32(cols), 0, layerGap)
	totalWidth := rightX + float32(50) // Add margin on right
	totalHeight := legendY + float32(60)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(totalWidth, totalHeight))

	return cont
}
