//go:build legacy_fyne

package gui

import (
	"fmt"
	"image"
)

type overlayCellLabels struct {
	TopLabel    string
	BottomLabel string
}

// overlayDrawableDims returns safe drawable dimensions constrained by backing weights.
func (ca *CircuitsApp) overlayDrawableDims(rows, cols int, weights [][]int) (int, int) {
	drawRows := rows
	if len(weights) < drawRows {
		drawRows = len(weights)
	}
	drawCols := cols
	for r := 0; r < drawRows; r++ {
		if r >= len(weights) {
			break
		}
		if len(weights[r]) < drawCols {
			drawCols = len(weights[r])
		}
	}
	if drawRows < 0 {
		drawRows = 0
	}
	if drawCols < 0 {
		drawCols = 0
	}
	return drawRows, drawCols
}

// overlayCellInfo composes dual-line overlay labels for selected cell diagnostics.
func (ca *CircuitsApp) overlayCellInfo(level, quantLevels int, value float64, mode string) overlayCellLabels {
	_ = quantLevels // kept for compatibility with call sites/tests
	top := fmt.Sprintf("L: %d", level)
	bottom := formatOverlayBottomValue(mode, value)
	return overlayCellLabels{TopLabel: top, BottomLabel: bottom}
}

func cellRectFor(row, col, offsetX, offsetY, cellSize int) image.Rectangle {
	x0 := offsetX + col*cellSize
	y0 := offsetY + row*cellSize
	return image.Rect(x0, y0, x0+cellSize, y0+cellSize)
}

func selectedHighlightRectFor(row, col, offsetX, offsetY, cellSize int) image.Rectangle {
	cell := cellRectFor(row, col, offsetX, offsetY, cellSize)
	return cell.Inset(1)
}
