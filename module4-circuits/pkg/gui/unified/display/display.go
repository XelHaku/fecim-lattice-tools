//go:build legacy_fyne

package display

// ShouldShowSelectedOnly returns true when overlay should show on the selected cell only.
// For cellSize >= 45, overlay shows on ALL cells (handled by ShouldShowOverlay).
func ShouldShowSelectedOnly(overlayEnabled bool, isSelected bool, cellSize int) bool {
	return overlayEnabled && isSelected && cellSize >= 36 && cellSize < 45
}

// ShouldShowOverlay returns true when overlay should display on this cell.
// For large cells (>=45), show on ALL cells. For medium cells (36-44), only selected.
func ShouldShowOverlay(overlayEnabled bool, isSelected bool, cellSize int) bool {
	if !overlayEnabled {
		return false
	}
	if cellSize >= 45 {
		return true // Large cells: show overlay on ALL cells
	}
	return isSelected && cellSize >= 36 // Medium cells: selected only
}
