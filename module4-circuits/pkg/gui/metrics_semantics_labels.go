//go:build legacy_fyne

package gui

// readModeMetricLabels returns the canonical READ-mode metric labels shown in UI/docs.
func readModeMetricLabels() []string {
	return []string{
		"I_cell (µA)",
		"V_TIA (V)",
		"ADC Code (0–2^N-1)",
		"Noise RMS (µA)",
		"SNR (dB)",
		"I_LSB (µA/code)",
	}
}
