// pkg/gui/tabs/learn_visuals_cell.go
// Cell structure visualization and comparison tables

package tabs

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// =============================================================================
// CELL COMPARISON TABLE
// =============================================================================

// CellComparisonTable creates a visual comparison table for Passive vs 1T1R cells
func CellComparisonTable() fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// Table dimensions - increased row height for better readability
	colWidths := []float32{110, 120, 120, 90} // Property, Passive, 1T1R, Winner
	rowHeight := float32(28)
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
		rowBg.Resize(fyne.NewSize(440, rowHeight))
		rowBg.Move(fyne.NewPos(x, y))
		objects = append(objects, rowBg)

		// Cells
		cells := []string{rowData.property, rowData.passive, rowData.t1r, rowData.winner}
		for j, cellText := range cells {
			cellX := x + float32(j)*colWidths[j]

			// Cell text - improved sizing
			text := canvas.NewText(cellText, colorText)
			if i == 0 {
				text.TextSize = 12
				text.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				text.TextSize = 11
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

			text.Move(fyne.NewPos(cellX+5, y+7))
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
		hLine.Position2 = fyne.NewPos(x+440, y+rowHeight)
		objects = append(objects, hLine)
	}

	// Outer border
	border := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	border.StrokeColor = colorArrow
	border.StrokeWidth = 2
	border.Resize(fyne.NewSize(440, float32(len(rows))*rowHeight))
	border.Move(fyne.NewPos(startX, startY))
	objects = append(objects, border)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(460, 210))

	return cont
}
