//go:build legacy_fyne

package gui

import "fmt"

func formatMetricVTIAMV(voltageV float64) string {
	return fmt.Sprintf("%+.2f V", voltageV)
}

func formatMetricICellUA(currentUA float64) string {
	return fmt.Sprintf("%+.2f µA", currentUA)
}

func formatMetricADCCode(code int) string {
	return fmt.Sprintf("%d", code)
}

func formatMetricConductanceUS(conductanceUS float64) string {
	return fmt.Sprintf("%.1f µS", conductanceUS)
}

func formatMetricLevel(level int) string {
	return fmt.Sprintf("%d", level)
}

func formatOverlayBottomValue(mode string, value float64) string {
	if mode == "Icell" {
		// value in A -> µA
		return fmt.Sprintf("I: %+.2f µA", value*1e6)
	}
	return fmt.Sprintf("V: %+.2f V", value)
}
