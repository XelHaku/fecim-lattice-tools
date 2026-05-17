//go:build legacy_fyne && !race

package widgets

// isRaceEnabled returns false when the binary was not built with -race.
func isRaceEnabled() bool { return false }
