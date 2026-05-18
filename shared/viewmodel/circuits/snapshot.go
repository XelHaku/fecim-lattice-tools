package circuits

import (
	"fmt"
	"strings"

	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state CircuitsState) viewmodel.ModuleSnapshot {
	quantLevels := state.QuantLevels
	if quantLevels <= 0 {
		quantLevels = DefaultQuantLevels
	}
	metrics := []viewmodel.Metric{
		{ID: "array", Label: "Array", Value: fmt.Sprintf("%dx%d", state.Rows, state.Cols)},
		{ID: "mode", Label: "Mode", Value: strings.ToUpper(state.OperationMode)},
		{ID: "architecture", Label: "Architecture", Value: state.Architecture},
		{ID: "selected_cell", Label: "Selected Cell", Value: fmt.Sprintf("[%d,%d]", state.SelectedRow, state.SelectedCol)},
		{ID: "write_target", Label: "Write Target", Value: fmt.Sprintf("%d/%d", state.WriteTargetLevel, quantLevels-1)},
		{ID: "coupling", Label: "Coupling", Value: state.CouplingTier},
		{ID: "ispp_engine", Label: "ISPP Engine", Value: state.ISPPEngine},
		{ID: "adc", Label: "ADC", Value: fmt.Sprintf("%d-bit SAR", state.ADCResolution)},
		{ID: "dac", Label: "DAC", Value: fmt.Sprintf("%d-bit R-2R", state.DACResolution)},
		{ID: "tia", Label: "TIA", Value: fmt.Sprintf("%.0f kΩ", state.TIAGain/1e3)},
		{ID: "charge_pump", Label: "Charge Pump", Value: fmt.Sprintf("%d-stage Dickson", state.ChargePumpStages)},
		{ID: "ispp", Label: "ISPP", Value: fmt.Sprintf("%v", state.ISPPEnabled)},
		{ID: "supply", Label: "Vdd", Value: fmt.Sprintf("%.1f V", state.SupplyVoltage)},
	}
	if state.LastOperationStatus != "" {
		metrics = append(metrics, viewmodel.Metric{ID: "last_operation", Label: "Last Operation", Value: state.LastOperationStatus})
	}
	if state.ISPPExecuted {
		metrics = append(metrics,
			viewmodel.Metric{ID: "ispp_conv", Label: "ISPP Converged", Value: fmt.Sprintf("%d/%d levels", state.ISPPConvergedCount, quantLevels)},
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
		{
			ID:    "unified_controls",
			Title: "Unified Circuit Controls",
			Body: fmt.Sprintf(
				"%s mode on a %dx%d %s array. Selected cell [%d,%d], target level %d/%d, %d-bit DAC, %d-bit ADC, %.0f kΩ TIA, %s coupling, %s ISPP.",
				strings.ToUpper(state.OperationMode),
				state.Rows,
				state.Cols,
				state.Architecture,
				state.SelectedRow,
				state.SelectedCol,
				state.WriteTargetLevel,
				quantLevels-1,
				state.DACResolution,
				state.ADCResolution,
				state.TIAGain/1e3,
				state.CouplingTier,
				state.ISPPEngine,
			),
			Category: "design",
		},
		{ID: "read_path", Title: "Read Path", Body: fmt.Sprintf("TIA (%.0f kΩ) → %d-bit SAR ADC. Latency: ~%.1f µs.", state.TIAGain/1e3, state.ADCResolution, float64(state.ADCResolution)*0.5), Category: "research"},
		{ID: "write_path", Title: "Write Path (ISPP)", Body: fmt.Sprintf("%d-stage charge pump → %d-bit DAC → ISPP pulse train.", state.ChargePumpStages, state.DACResolution), Category: "research"},
	}
	sections = append(sections, viewmodel.Section{
		ID: "edu_adc", Title: "How SAR ADC Works",
		Body:     fmt.Sprintf("Successive Approximation Register ADC: Binary search over %d levels. Each bit is tested: set bit, compare against input, keep or discard. %d clock cycles to complete. INL/DNL characterize deviation from ideal.", 1<<state.ADCResolution, state.ADCResolution),
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID: "edu_ispp", Title: "ISPP Write-Verify",
		Body:     "Incremental Step Pulse Programming: Apply voltage pulse → Wait for settling → Verify conductance → If not at target, increase pulse amplitude → Repeat. Guard-band pulses prevent overshoot. Binary search accelerates convergence.",
		Category: "education",
	})
	sections = append(sections, viewmodel.Section{
		ID: "research_pvt", Title: "PVT Variation",
		Body:     fmt.Sprintf("Process/Voltage/Temperature corners: TT (typical), FF (fast NMOS/PMOS), SS (slow). ADC INL degrades at SS corner. Charge pump output drops at low Vdd (%.1f V min). All values are educational models.", state.SupplyVoltage*0.9),
		Category: "research",
	})
	sections = append(sections, viewmodel.Section{
		ID: "design_readpath", Title: "Optimizing the Read Path",
		Body:     fmt.Sprintf("Latency budget: TIA settling + %d-cycle ADC conversion. Lower resolution = faster but noisier. Design trade: 5-bit ADC for 30-level cells gives 1.7× noise margin. Cross-reference: Module 2 array output feeds this read path.", state.ADCResolution),
		Category: "design",
	})
	actions := []viewmodel.Action{
		{ID: ActionRunRead, Label: "Simulate Read", Kind: viewmodel.ActionCommand},
		{ID: ActionRunWrite, Label: "Simulate Write", Kind: viewmodel.ActionCommand},
		{ID: ActionRunCompute, Label: "Simulate Compute", Kind: viewmodel.ActionCommand},
		{ID: ActionToggleISPP, Label: "Toggle ISPP", Kind: viewmodel.ActionToggle, Payload: map[string]string{"enabled": fmt.Sprintf("%v", state.ISPPEnabled)}},
		{ID: ActionResizeArray, Label: "Array Size", Kind: viewmodel.ActionSelect, Payload: map[string]string{"rows": fmt.Sprintf("%d", state.Rows), "cols": fmt.Sprintf("%d", state.Cols)}},
		{ID: ActionSetOperationMode, Label: "Operation Mode", Kind: viewmodel.ActionSelect, Payload: map[string]string{"mode": state.OperationMode}},
		{ID: ActionSetArchitecture, Label: "Architecture", Kind: viewmodel.ActionSelect, Payload: map[string]string{"architecture": state.Architecture}},
		{ID: ActionSelectCell, Label: "Selected Cell", Kind: viewmodel.ActionSelect, Payload: map[string]string{"row": fmt.Sprintf("%d", state.SelectedRow), "col": fmt.Sprintf("%d", state.SelectedCol)}},
		{ID: ActionSetWriteTarget, Label: "Write Target", Kind: viewmodel.ActionSelect, Payload: map[string]string{"level": fmt.Sprintf("%d", state.WriteTargetLevel)}},
		{ID: ActionSetDACBits, Label: "DAC Bits", Kind: viewmodel.ActionSelect, Payload: map[string]string{"bits": fmt.Sprintf("%d", state.DACResolution)}},
		{ID: ActionSetADCBits, Label: "ADC Bits", Kind: viewmodel.ActionSelect, Payload: map[string]string{"bits": fmt.Sprintf("%d", state.ADCResolution)}},
		{ID: ActionSetTIAGain, Label: "TIA Gain", Kind: viewmodel.ActionSelect, Payload: map[string]string{"gain_ohm": fmt.Sprintf("%.0f", state.TIAGain)}},
		{ID: ActionSetCouplingTier, Label: "Coupling Tier", Kind: viewmodel.ActionSelect, Payload: map[string]string{"tier": state.CouplingTier}},
		{ID: ActionSetISPPEngine, Label: "ISPP Engine", Kind: viewmodel.ActionSelect, Payload: map[string]string{"engine": state.ISPPEngine}},
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
