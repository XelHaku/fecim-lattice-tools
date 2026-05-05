package circuits

import (
	"fmt"
	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state CircuitsState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "adc", Label: "ADC", Value: fmt.Sprintf("%d-bit SAR", state.ADCResolution)},
		{ID: "dac", Label: "DAC", Value: fmt.Sprintf("%d-bit R-2R", state.DACResolution)},
		{ID: "tia", Label: "TIA", Value: fmt.Sprintf("%.0f kΩ", state.TIAGain/1e3)},
		{ID: "charge_pump", Label: "Charge Pump", Value: fmt.Sprintf("%d-stage Dickson", state.ChargePumpStages)},
		{ID: "ispp", Label: "ISPP", Value: fmt.Sprintf("%v", state.ISPPEnabled)},
		{ID: "supply", Label: "Vdd", Value: fmt.Sprintf("%.1f V", state.SupplyVoltage)},
	}
	if state.ISPPExecuted {
		metrics = append(metrics,
			viewmodel.Metric{ID: "ispp_conv", Label: "ISPP Converged", Value: fmt.Sprintf("%d/%d levels", state.ISPPConvergedCount, 30)},
			viewmodel.Metric{ID: "ispp_avg", Label: "Avg Attempts/Level", Value: fmt.Sprintf("%.1f", state.ISPPAvgAttempts)},
		)
	}
	if state.ENOBtt > 0 {
		metrics = append(metrics,
			viewmodel.Metric{ID: "enob_tt", Label: "ENOB (TT)", Value: fmt.Sprintf("%.2f bits", state.ENOBtt)},
			viewmodel.Metric{ID: "enob_ff", Label: "ENOB (FF)", Value: fmt.Sprintf("%.2f bits", state.ENOBff)},
			viewmodel.Metric{ID: "enob_ss", Label: "ENOB (SS)", Value: fmt.Sprintf("%.2f bits", state.ENOBss)},
			viewmodel.Metric{ID: "snr", Label: "Ideal SNR", Value: fmt.Sprintf("%.1f dB", state.SNRdB)},
		)
	}
	sections := []viewmodel.Section{
		{ID: "read_path", Title: "Read Path", Body: fmt.Sprintf("TIA (%.0f kΩ) → %d-bit SAR ADC. Latency: ~%.1f µs.", state.TIAGain/1e3, state.ADCResolution, float64(state.ADCResolution)*0.5), Category: "research"},
		{ID: "write_path", Title: "Write Path (ISPP)", Body: fmt.Sprintf("%d-stage charge pump → %d-bit DAC → ISPP pulse train.", state.ChargePumpStages, state.DACResolution), Category: "research"},
	}
	sections = append(sections, viewmodel.Section{
		ID: "edu_adc", Title: "How SAR ADC Works",
		Body: fmt.Sprintf("Successive Approximation Register ADC: Binary search over %d levels. Each bit is tested: set bit, compare against input, keep or discard. %d clock cycles to complete. INL/DNL characterize deviation from ideal.", 1<<state.ADCResolution, state.ADCResolution),
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID: "edu_ispp", Title: "ISPP Write-Verify",
		Body: "Incremental Step Pulse Programming: Apply voltage pulse → Wait for settling → Verify conductance → If not at target, increase pulse amplitude → Repeat. Guard-band pulses prevent overshoot. Binary search accelerates convergence.",
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID: "research_pvt", Title: "PVT Variation",
		Body: fmt.Sprintf("Process/Voltage/Temperature corners: TT (typical), FF (fast NMOS/PMOS), SS (slow). ADC INL degrades at SS corner. Charge pump output drops at low Vdd (%.1f V min). All values are educational models.", state.SupplyVoltage*0.9),
		Category: "research",
	})
	sections = append(sections, viewmodel.Section{
		ID: "design_readpath", Title: "Optimizing the Read Path",
		Body: fmt.Sprintf("Latency budget: TIA settling + %d-cycle ADC conversion. Lower resolution = faster but noisier. Design trade: 5-bit ADC for 30-level cells gives 1.7× noise margin. Cross-reference: Module 2 array output feeds this read path.", state.ADCResolution),
		Category: "design",
	})
	actions := []viewmodel.Action{
		{ID: "run_read", Label: "Simulate Read", Kind: viewmodel.ActionCommand},
		{ID: "run_write", Label: "Simulate Write", Kind: viewmodel.ActionCommand},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID: viewmodel.ModuleCircuits, Title: "FeCIM Peripheral Circuits Visualizer",
			Description:    "DAC, ADC, TIA, read path, write path, and ISPP circuit behavior.",
			Status:         viewmodel.StatusFunctional,
			BoundaryNotice: "SIMULATION OUTPUT — Educational circuit models. ADC/DAC/TIA are behavioral abstractions, not calibrated against silicon measurements.",
		},
		Metrics: metrics, Sections: sections, Actions: actions,
		Plots: buildISPPPlots(state),
	}
}

func buildISPPPlots(state CircuitsState) []viewmodel.PlotData {
	if !state.ISPPExecuted || len(state.ISPPAttempts) == 0 {
		return nil
	}
	pts := make([]viewmodel.PlotPoint, len(state.ISPPAttempts))
	for i, a := range state.ISPPAttempts {
		pts[i] = viewmodel.PlotPoint{X: float64(i), Y: float64(a)}
	}
	return []viewmodel.PlotData{{
		ID:     "ispp_convergence",
		Title:  "ISPP Write-Verify Convergence",
		XLabel: "Target Level (0-29)",
		YLabel: "Attempts to Converge",
		Series: []viewmodel.PlotSeries{{Name: "attempts", Points: pts}},
	}}
}
