//go:build legacy_fyne

package gui

import (
	"fmt"
	"image"
	"image/color"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

type unifiedArrayLayout struct {
	rows, cols              int
	cellSize                int
	offsetX, offsetY        int
	gridW, gridH            int
	leftMargin, rightMargin int
	topMargin, bottomMargin int
	is1T1R, is2T1R          bool
}

func buildUnifiedArrayLayout(w, h, rows, cols int, arch string, zoom float64) unifiedArrayLayout {
	if zoom == 0 || zoom < 0.5 {
		zoom = 1.0
	}
	l := unifiedArrayLayout{rows: rows, cols: cols, topMargin: 65, rightMargin: 130, bottomMargin: 30, leftMargin: 30}
	l.is1T1R = arch == sharedwidgets.Architecture1T1R
	l.is2T1R = arch == sharedwidgets.Architecture2T1R
	if l.is1T1R || l.is2T1R {
		l.leftMargin = 55
	}
	if l.is2T1R {
		l.bottomMargin = 55
	}
	availableW := w - l.leftMargin - l.rightMargin
	availableH := h - l.topMargin - l.bottomMargin
	maxCellSize := int(float64(70) * zoom)
	minCellSize := int(float64(18) * zoom)
	if cols > 32 || rows > 32 {
		maxCellSize = int(float64(30) * zoom)
		minCellSize = int(float64(8) * zoom)
	} else if cols > 16 || rows > 16 {
		maxCellSize = int(float64(40) * zoom)
		minCellSize = int(float64(12) * zoom)
	}
	cellW := availableW / cols
	cellH := availableH / rows
	l.cellSize = min(cellW, cellH)
	if l.cellSize > maxCellSize {
		l.cellSize = maxCellSize
	}
	if l.cellSize < minCellSize {
		l.cellSize = minCellSize
	}
	l.gridW = cols * l.cellSize
	l.gridH = rows * l.cellSize
	l.offsetX = l.leftMargin + (availableW-l.gridW)/2
	l.offsetY = l.topMargin + (availableH-l.gridH)/2
	return l
}

func (ca *CircuitsApp) staticCacheKey(w, h int, l unifiedArrayLayout, levels int) string {
	return fmt.Sprintf("%dx%d:%dx%d:%d:%s:%d", w, h, l.rows, l.cols, l.cellSize, ca.architecture, levels)
}

func (ca *CircuitsApp) getUnifiedStaticLayer(w, h int, l unifiedArrayLayout, levels int) *image.RGBA {
	key := ca.staticCacheKey(w, h, l, levels)
	ca.mu.Lock()
	defer ca.mu.Unlock()
	if ca.unifiedStaticCache != nil && ca.unifiedStaticCacheKey == key {
		return cloneRGBA(ca.unifiedStaticCache)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bgTop := color.RGBA{12, 20, 35, 255}
	bgBottom := color.RGBA{8, 14, 28, 255}
	drawGradientRect(img, 0, 0, w, h, bgTop, bgBottom)
	panelColor := color.RGBA{18, 28, 45, 255}
	drawRoundedRect(img, l.offsetX-6, l.offsetY-6, l.gridW+12, l.gridH+12, 8, panelColor)

	gridColor := color.RGBA{55, 70, 95, 180}
	for c := 0; c <= l.cols; c++ {
		x := l.offsetX + c*l.cellSize
		for y := l.offsetY; y <= l.offsetY+l.gridH; y++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, gridColor)
			}
		}
	}
	for r := 0; r <= l.rows; r++ {
		y := l.offsetY + r*l.cellSize
		for x := l.offsetX; x <= l.offsetX+l.gridW; x++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, gridColor)
			}
		}
	}

	drawSimpleText(img, "BL", l.offsetX+l.gridW/2-6, l.offsetY-35, color.RGBA{100, 180, 255, 200})
	drawSimpleText(img, "WL", l.offsetX-25, l.offsetY+l.gridH/2-3, color.RGBA{255, 180, 100, 200})
	if l.is2T1R {
		drawSimpleText(img, "SL", l.offsetX+l.gridW/2-6, l.offsetY+l.gridH+45, color.RGBA{100, 220, 255, 200})
	}

	ca.unifiedStaticCache = cloneRGBA(img)
	ca.unifiedStaticCacheKey = key
	ca.unifiedStaticRedrawCount++
	return img
}

func cloneRGBA(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Rect)
	copy(dst.Pix, src.Pix)
	return dst
}
