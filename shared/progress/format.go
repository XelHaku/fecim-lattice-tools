package progress

import "fmt"

// formatRate formats the processing rate for display.
func formatRate(rate float64, total int64) string {
	if rate < 1 {
		return fmt.Sprintf("%.2f /sec", rate)
	}
	if rate < 1000 {
		return fmt.Sprintf("%.1f /sec", rate)
	}
	return fmt.Sprintf("%.1fk /sec", rate/1000)
}
