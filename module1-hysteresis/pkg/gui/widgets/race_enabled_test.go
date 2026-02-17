//go:build race

package widgets

// isRaceEnabled returns true when the binary was built with -race.
func isRaceEnabled() bool { return true }
