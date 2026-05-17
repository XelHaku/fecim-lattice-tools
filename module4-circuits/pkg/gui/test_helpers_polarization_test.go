//go:build legacy_fyne

package gui

func sharedPolarization(g, gmin, gmax, ps float64) float64 {
	if gmax <= gmin || ps == 0 {
		return 0
	}
	norm := 2*(g-gmin)/(gmax-gmin) - 1
	if norm > 1 {
		norm = 1
	}
	if norm < -1 {
		norm = -1
	}
	return norm * ps
}
