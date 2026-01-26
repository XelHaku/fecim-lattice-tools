// pkg/gui/tabs/learn_visuals_transistor.go
// Transistor and FeFET structure visualizations

package tabs

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

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
	boxW := float32(180)
	boxH := float32(120)
	spacing := float32(20)
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
		descX := mode.x + (boxW - float32(len(mode.description)*6)) / 2
		descText.Move(fyne.NewPos(descX, startY+50))
		objects = append(objects, descText)
	}

	// Central FeCIM circle position
	circleX := float32(300)  // Center of the diagram
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
	cont.Resize(fyne.NewSize(620, 280))

	return cont
}
