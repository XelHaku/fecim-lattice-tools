//go:build legacy_fyne

package sense

import "math"

// ADCCodePercent returns full-scale percentage for a code.
func ADCCodePercent(code, levels int) float64 {
	if levels <= 1 {
		return 0
	}
	return (float64(code) / float64(levels-1)) * 100.0
}

// ClampADCRange enforces valid ADC reference window.
func ClampADCRange(vmin, vmax, minRef, maxRef, minSpan float64) (float64, float64) {
	if vmin < minRef {
		vmin = minRef
	}
	if vmax > maxRef {
		vmax = maxRef
	}
	if vmin > maxRef-minSpan {
		vmin = maxRef - minSpan
	}
	if vmax <= vmin+minSpan {
		vmax = vmin + minSpan
	}
	return vmin, vmax
}

func ClampNearZero(v, eps float64) float64 {
	if math.Abs(v) < eps {
		return 0
	}
	return v
}
