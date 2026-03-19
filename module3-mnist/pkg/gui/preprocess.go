// Package gui provides Fyne-based GUI components for MNIST visualization.
package gui

import "math"

// InputSource tracks where the current digit input came from.
type InputSource int

const (
	InputProgrammatic InputSource = iota
	InputUser
)

const (
	preprocessThreshold = 0.05
	preprocessTarget    = 20
	preprocessSize      = 28
)

func preprocessIfUserInput(dc *DigitCanvas, pixels []float64) []float64 {
	if dc == nil || dc.LastInputSource() != InputUser {
		return pixels
	}
	return PreprocessDigit(pixels)
}

// PreprocessDigit normalizes a drawn digit to match MNIST-like centering/scaling.
// The input is a flattened 28x28 grayscale array in [0,1].
func PreprocessDigit(pixels []float64) []float64 {
	if len(pixels) < preprocessSize*preprocessSize {
		return pixels
	}

	var img [preprocessSize][preprocessSize]float64
	for y := 0; y < preprocessSize; y++ {
		for x := 0; x < preprocessSize; x++ {
			img[y][x] = pixels[y*preprocessSize+x]
		}
	}

	minX, minY := preprocessSize, preprocessSize
	maxX, maxY := -1, -1
	for y := 0; y < preprocessSize; y++ {
		for x := 0; x < preprocessSize; x++ {
			if img[y][x] > preprocessThreshold {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if maxX < 0 || maxY < 0 {
		return pixels
	}

	cropW := maxX - minX + 1
	cropH := maxY - minY + 1
	crop := make2DFloat(cropH, cropW)
	for y := 0; y < cropH; y++ {
		for x := 0; x < cropW; x++ {
			crop[y][x] = img[minY+y][minX+x]
		}
	}

	scale := float64(preprocessTarget) / float64(max(cropW, cropH))
	newW := int(math.Round(float64(cropW) * scale))
	newH := int(math.Round(float64(cropH) * scale))
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	resized := resizeBilinear(crop, newH, newW)

	// Normalize to max intensity of 1.0.
	maxVal := 0.0
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			if resized[y][x] > maxVal {
				maxVal = resized[y][x]
			}
		}
	}
	if maxVal > 0 {
		for y := 0; y < newH; y++ {
			for x := 0; x < newW; x++ {
				resized[y][x] /= maxVal
			}
		}
	}

	// Center using center-of-mass, clamped to bounds.
	sum, sumX, sumY := 0.0, 0.0, 0.0
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			v := resized[y][x]
			sum += v
			sumX += v * float64(x)
			sumY += v * float64(y)
		}
	}

	x0 := (preprocessSize - newW) / 2
	y0 := (preprocessSize - newH) / 2
	if sum > 0 {
		comX := sumX / sum
		comY := sumY / sum
		x0 = int(math.Round(float64(preprocessSize/2) - comX))
		y0 = int(math.Round(float64(preprocessSize/2) - comY))
	}

	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x0+newW > preprocessSize {
		x0 = preprocessSize - newW
	}
	if y0+newH > preprocessSize {
		y0 = preprocessSize - newH
	}

	out := make([]float64, preprocessSize*preprocessSize)
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			out[(y0+y)*preprocessSize+(x0+x)] = resized[y][x]
		}
	}

	return out
}

func make2DFloat(h, w int) [][]float64 {
	out := make([][]float64, h)
	for i := range out {
		out[i] = make([]float64, w)
	}
	return out
}

func resizeBilinear(src [][]float64, newH, newW int) [][]float64 {
	if newH <= 0 || newW <= 0 {
		return make2DFloat(0, 0)
	}
	srcH := len(src)
	if srcH == 0 {
		return make2DFloat(newH, newW)
	}
	srcW := len(src[0])
	if srcW == 0 {
		return make2DFloat(newH, newW)
	}

	dst := make2DFloat(newH, newW)

	var scaleY, scaleX float64
	if newH == 1 {
		scaleY = 0
	} else {
		scaleY = float64(srcH-1) / float64(newH-1)
	}
	if newW == 1 {
		scaleX = 0
	} else {
		scaleX = float64(srcW-1) / float64(newW-1)
	}

	for y := 0; y < newH; y++ {
		sy := float64(y) * scaleY
		y0 := int(math.Floor(sy))
		y1 := min(y0+1, srcH-1)
		wy := sy - float64(y0)
		for x := 0; x < newW; x++ {
			sx := float64(x) * scaleX
			x0 := int(math.Floor(sx))
			x1 := min(x0+1, srcW-1)
			wx := sx - float64(x0)

			v0 := src[y0][x0]*(1-wx) + src[y0][x1]*wx
			v1 := src[y1][x0]*(1-wx) + src[y1][x1]*wx
			dst[y][x] = v0*(1-wy) + v1*wy
		}
	}

	return dst
}

