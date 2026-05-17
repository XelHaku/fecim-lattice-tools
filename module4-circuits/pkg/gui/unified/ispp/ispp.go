//go:build legacy_fyne

package ispp

// ClampTargetLevel keeps write target within valid quantized range.
func ClampTargetLevel(target, levels int) int {
	if levels <= 0 {
		return 0
	}
	if target < 0 {
		return 0
	}
	if target >= levels {
		return levels - 1
	}
	return target
}
