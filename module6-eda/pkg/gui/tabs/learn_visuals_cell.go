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

// CellComparisonTable creates a visual comparison table for Passive vs 1T1R vs 2T1R cells
func CellComparisonTable() fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// 5-column table: Property, Passive, 1T1R, 2T1R, Best
	colWidths := []float32{100, 100, 100, 110, 70}
	rowHeight := float32(28)
	startX := float32(10)
	startY := float32(10)

	var tableWidth float32
	for _, w := range colWidths {
		tableWidth += w
	}

	// Table data — 5 fields: property, passive, 1T1R, 2T1R, best
	type row struct {
		property string
		passive  string
		t1r      string
		t2r      string
		best     string // "Passive", "1T1R", "2T1R"
	}

	rows := []row{
		{"Property", "Passive", "1T1R", "2T1R", "Best"},
		{"Cell Width", "0.46 µm", "0.92 µm", "~1.38 µm", "Passive"},
		{"Area (µm²)", "1.25", "~3.74", "~5.62", "Passive"},
		{"Sneak Path", "HIGH", "Row-only", "NONE", "2T1R"},
		{"Max Array", "~32×32", "128×128+", "512×512+", "2T1R"},
		{"Ports", "WL,BL", "WL,BL,SL", "WL,CSL,BL,SL", "1T1R"},
		{"Fab Complexity", "Simple", "Moderate", "High", "Passive"},
	}

	// Draw table
	for i, rowData := range rows {
		y := startY + float32(i)*rowHeight
		x := startX

		var bgColor color.Color
		if i == 0 {
			bgColor = colorTableHeader
		} else if i%2 == 1 {
			bgColor = colorTableRow1
		} else {
			bgColor = colorTableRow2
		}

		rowBg := canvas.NewRectangle(bgColor)
		rowBg.Resize(fyne.NewSize(tableWidth, rowHeight))
		rowBg.Move(fyne.NewPos(x, y))
		objects = append(objects, rowBg)

		cells := []string{rowData.property, rowData.passive, rowData.t1r, rowData.t2r, rowData.best}
		cellX := x
		for j, cellText := range cells {
			text := canvas.NewText(cellText, colorText)
			if i == 0 {
				text.TextSize = 13
				text.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				text.TextSize = 13
			}

			// Best column gets colour-coded highlighting
			if j == 4 && i > 0 {
				switch cellText {
				case "2T1R":
					text.Color = color.RGBA{180, 80, 200, 255} // purple for 2T1R
					text.TextStyle = fyne.TextStyle{Bold: true}
				case "1T1R":
					text.Color = colorHighlight
					text.TextStyle = fyne.TextStyle{Bold: true}
				default: // Passive
					text.Color = colorBoxOurs
					text.TextStyle = fyne.TextStyle{Bold: true}
				}
			}

			text.Move(fyne.NewPos(cellX+4, y+7))
			objects = append(objects, text)

			if j < len(cells)-1 {
				lineX := cellX + colWidths[j]
				line := canvas.NewLine(colorArrow)
				line.StrokeWidth = 1
				line.Position1 = fyne.NewPos(lineX, y)
				line.Position2 = fyne.NewPos(lineX, y+rowHeight)
				objects = append(objects, line)
			}

			cellX += colWidths[j]
		}

		hLine := canvas.NewLine(colorArrow)
		hLine.StrokeWidth = 1
		hLine.Position1 = fyne.NewPos(x, y+rowHeight)
		hLine.Position2 = fyne.NewPos(x+tableWidth, y+rowHeight)
		objects = append(objects, hLine)
	}

	tableHeight := float32(len(rows)) * rowHeight

	border := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	border.StrokeColor = colorArrow
	border.StrokeWidth = 2
	border.Resize(fyne.NewSize(tableWidth, tableHeight))
	border.Move(fyne.NewPos(startX, startY))
	objects = append(objects, border)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(startX+tableWidth+float32(10), startY+tableHeight+float32(10)))

	return cont
}
