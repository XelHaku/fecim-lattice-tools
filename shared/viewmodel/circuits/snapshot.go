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
		{ID: "logger_verbosity", Label: "Logger Verbosity", Value: loggerVerbosityValue(state)},
		{ID: "logger_detail", Label: "Logger Detail", Value: loggerDetailValue(state)},
		{ID: "half_select_state", Label: "Half-Select", Value: halfSelectStateValue(state)},
		{ID: "half_select_cells", Label: "Disturbed Cells", Value: halfSelectCellsValue(state)},
		{ID: "disturb_voltage", Label: "Disturb Voltage", Value: disturbVoltageValue(state)},
		{ID: "stress_budget", Label: "Stress Budget", Value: stressBudgetValue(state)},
		{ID: "stress_per_pulse", Label: "Stress/Pulse", Value: stressPerPulseValue(state)},
		{ID: "stress_selected_cell", Label: "Stress Target", Value: fmt.Sprintf("[%d,%d]", state.SelectedRow, state.SelectedCol)},
		{ID: "adc", Label: "ADC", Value: fmt.Sprintf("%d-bit SAR", state.ADCResolution)},
		{ID: "dac", Label: "DAC", Value: fmt.Sprintf("%d-bit R-2R", state.DACResolution)},
		{ID: "tia", Label: "TIA", Value: fmt.Sprintf("%.0f kΩ", state.TIAGain/1e3)},
		{ID: "charge_pump", Label: "Charge Pump", Value: fmt.Sprintf("%d-stage Dickson", state.ChargePumpStages)},
		{ID: "ispp", Label: "ISPP", Value: fmt.Sprintf("%v", state.ISPPEnabled)},
		{ID: "supply", Label: "Vdd", Value: fmt.Sprintf("%.1f V", state.SupplyVoltage)},
		{ID: "spec_storage", Label: "Spec Storage", Value: specStorageValue(state)},
		{ID: "spec_components", Label: "Spec Components", Value: specComponentsValue(state)},
		{ID: "spec_power_latency", Label: "Spec Power/Latency", Value: specPowerLatencyValue(state)},
		{ID: "spec_throughput", Label: "Spec Throughput", Value: specThroughputValue(state)},
		{ID: "spec_resolution", Label: "Spec Resolution", Value: specResolutionValue(state, quantLevels)},
		{ID: "spec_compliance", Label: "Spec Compliance", Value: specComplianceValue(state)},
		{ID: "reference_spec_export", Label: "Reference Spec Export", Value: referenceSpecExportStatusValue(state)},
		{ID: "reference_spec_export_path", Label: "Reference Spec Export Target", Value: referenceSpecExportPathValue(state)},
		{ID: "reference_spec_export_bytes", Label: "Reference Spec Export Size", Value: fmt.Sprintf("%d bytes", state.ReferenceSpecExportBytes)},
		{ID: "timing_write", Label: "Write Timing", Value: timingTotalValue(state.TimingWriteTotalNS)},
		{ID: "timing_read", Label: "Read Timing", Value: timingTotalValue(state.TimingReadTotalNS)},
		{ID: "timing_compute", Label: "Compute Timing", Value: timingTotalValue(state.TimingComputeTotalNS)},
		{ID: "timing_active", Label: "Active Timing", Value: timingActiveValue(state)},
		{ID: "timing_active_phases", Label: "Timing Phases", Value: timingActivePhasesValue(state)},
		{ID: "timing_waveform_signals", Label: "Timing Waveform Signals", Value: timingWaveformSignalsValue(state)},
		{ID: "timing_waveform_markers", Label: "Timing Waveform Markers", Value: timingWaveformMarkersValue(state)},
		{ID: "timing_waveform_phases", Label: "Timing Waveform Phases", Value: timingWaveformPhasesValue(state)},
		{ID: "reference_timing_export", Label: "Reference Timing Export", Value: referenceTimingExportStatusValue(state)},
		{ID: "reference_timing_export_path", Label: "Reference Timing Export Target", Value: referenceTimingExportPathValue(state)},
		{ID: "reference_timing_export_bytes", Label: "Reference Timing Export Size", Value: fmt.Sprintf("%d bytes", state.ReferenceTimingExportBytes)},
		{ID: "reference_timing_animation", Label: "Reference Timing Animation", Value: referenceTimingAnimationStatusValue(state)},
		{ID: "reference_timing_animation_step", Label: "Reference Timing Animation Step", Value: referenceTimingAnimationStepValue(state)},
		{ID: "reference_timing_animation_steps", Label: "Reference Timing Animation Steps", Value: referenceTimingAnimationStepsValue(state)},
		{ID: "operation_log_count", Label: "Operation Log", Value: operationLogCountValue(state)},
		{ID: "operation_log_latest", Label: "Latest Log Entry", Value: operationLogLatestValue(state.OperationLog)},
		{ID: "operation_log_recent", Label: "Recent Log Entries", Value: operationLogRecentValue(state.OperationLog)},
		{ID: "operation_log_export", Label: "Operation Log Export", Value: operationLogExportStatusValue(state)},
		{ID: "operation_log_export_path", Label: "Export Target", Value: operationLogExportPathValue(state)},
		{ID: "operation_log_export_bytes", Label: "Export Size", Value: fmt.Sprintf("%d bytes", state.OperationLogExportBytes)},
		{ID: "compute_run", Label: "Compute Run", Value: computeRunValue(state.ComputeRunLog)},
		{ID: "compute_run_peak", Label: "Compute Peak", Value: computeRunPeakValue(state.ComputeRunLog)},
		{ID: "compute_run_input", Label: "Compute Input", Value: computeRunInputValue(state.ComputeRunLog)},
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
			viewmodel.Metric{ID: "pvt_temperature_sweep", Label: "PVT Temperature Sweep", Value: pvtTemperatureSweepValue(state)},
			viewmodel.Metric{ID: "pvt_process_yield", Label: "PVT Process Yield", Value: pvtProcessYieldValue(state)},
			viewmodel.Metric{ID: "pvt_corner_enob", Label: "PVT Corner ENOB", Value: pvtCornerENOBValue(state)},
			viewmodel.Metric{ID: "pvt_noise_ceiling", Label: "Noise-Limited ENOB", Value: pvtNoiseCeilingValue(state)},
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
		ID: "half_select_stress", Title: "Half-Select / Column Stress",
		Body: fmt.Sprintf("%s: %s at %s. Budget: %s. This is a deterministic educational stress budget, not a calibrated endurance claim.",
			halfSelectStateValue(state), halfSelectCellsValue(state), disturbVoltageValue(state), stressBudgetValue(state)),
		Category: "research",
	})
	sections = append(sections, viewmodel.Section{
		ID: "pvt_investigations", Title: "PVT Investigation Summaries",
		Body: fmt.Sprintf("Temperature sweep: %s. Process-yield proxy: %s. Corner ENOB: %s. Thermal-noise ceiling: %s. These are educational summaries of existing Module 4 investigations, not calibrated silicon guarantees.",
			pvtTemperatureSweepValue(state), pvtProcessYieldValue(state), pvtCornerENOBValue(state), pvtNoiseCeilingValue(state)),
		Category: "research",
	})
	sections = append(sections, viewmodel.Section{
		ID: "reference_specs", Title: "Reference Spec / Compliance Summary",
		Body: fmt.Sprintf("%s. %s. %s. %s. %s. Summary-level port of the legacy reference specs; power, area, and latency values are educational estimates.",
			specStorageValue(state), specComponentsValue(state), specPowerLatencyValue(state), specThroughputValue(state), specComplianceValue(state)),
		Category: "design",
	})
	sections = append(sections, viewmodel.Section{
		ID: "reference_timing", Title: "Reference Timing Summary",
		Body: fmt.Sprintf("Write: %s. Read: %s. Compute: %s. Active %s phases: %s. Waveform metadata: %s; %s; %s. Summary-level port of the legacy timing diagrams; no raster/SVG export or timed playback loop is claimed.",
			timingTotalValue(state.TimingWriteTotalNS), timingTotalValue(state.TimingReadTotalNS), timingTotalValue(state.TimingComputeTotalNS), timingActiveValue(state), timingActivePhasesValue(state),
			timingWaveformSignalsValue(state), timingWaveformMarkersValue(state), timingWaveformPhasesValue(state)),
		Category: "design",
	})
	sections = append(sections, viewmodel.Section{
		ID:       "operation_log",
		Title:    "Operation Log",
		Body:     operationLogSectionBody(state),
		Category: "design",
	})
	sections = append(sections, viewmodel.Section{
		ID:       "compute_run_log",
		Title:    "Compute Run Log",
		Body:     computeRunSectionBody(state.ComputeRunLog),
		Category: "design",
	})
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
		{ID: ActionExportOperationLog, Label: "Export Operation Log", Kind: viewmodel.ActionCommand},
		{ID: ActionExportReferenceSpecs, Label: "Export Reference Specs", Kind: viewmodel.ActionCommand},
		{ID: ActionExportReferenceTiming, Label: "Export Reference Timing", Kind: viewmodel.ActionCommand},
		{ID: ActionAnimateReferenceTiming, Label: "Animate Reference Timing", Kind: viewmodel.ActionCommand},
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
		{ID: ActionSetLoggerVerbosity, Label: "Logger Verbosity", Kind: viewmodel.ActionSelect, Payload: map[string]string{"verbosity": loggerVerbosityValue(state)}},
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

func halfSelectStateValue(state CircuitsState) string {
	if state.HalfSelectState == "" {
		return HalfSelectStateInactive
	}
	return state.HalfSelectState
}

func loggerVerbosityValue(state CircuitsState) string {
	if state.LoggerVerbosity == "" {
		return "off"
	}
	return state.LoggerVerbosity
}

func loggerDetailValue(state CircuitsState) string {
	switch loggerVerbosityValue(state) {
	case "info":
		return "info: startup and shutdown summaries"
	case "debug":
		return "debug: action and input events"
	case "trace":
		return "trace: every UI update and simulation tick"
	default:
		return "off: file/debug logging disabled"
	}
}

func halfSelectCellsValue(state CircuitsState) string {
	switch state.HalfSelectState {
	case HalfSelectStateColumnWriteActive:
		return fmt.Sprintf("%d same-column cells", state.HalfSelectCells)
	case HalfSelectStateAttenuated:
		return fmt.Sprintf("%d gated cells", state.HalfSelectCells)
	default:
		return fmt.Sprintf("%d cells", state.HalfSelectCells)
	}
}

func disturbVoltageValue(state CircuitsState) string {
	if state.DisturbVoltage <= 0 {
		return "0.00 V"
	}
	return fmt.Sprintf("%.2f V", state.DisturbVoltage)
}

func stressBudgetValue(state CircuitsState) string {
	if state.StressCyclesToLevel <= 0 {
		if halfSelectStateValue(state) == HalfSelectStateIsolated {
			return "isolated"
		}
		return "inactive"
	}
	return fmt.Sprintf("%d pulses/level", state.StressCyclesToLevel)
}

func stressPerPulseValue(state CircuitsState) string {
	if state.StressPerPulse <= 0 {
		return "0.000000 level/pulse"
	}
	return fmt.Sprintf("%.6f level/pulse", state.StressPerPulse)
}

func pvtTemperatureSweepValue(state CircuitsState) string {
	if state.PVTTemperatureSweep == "" {
		return "not evaluated"
	}
	return state.PVTTemperatureSweep
}

func pvtProcessYieldValue(state CircuitsState) string {
	if state.PVTSamples <= 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("%.1f%% (%d/%d)", 100*state.PVTProcessYield, state.PVTPassSamples, state.PVTSamples)
}

func pvtCornerENOBValue(state CircuitsState) string {
	if state.ENOBtt <= 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("FF %.2f / TT %.2f / SS %.2f bits", state.ENOBff, state.ENOBtt, state.ENOBss)
}

func pvtNoiseCeilingValue(state CircuitsState) string {
	if state.PVTENOBNoiseCeiling <= 0 || state.PVTENOBCeilingBits <= 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("%.2f bits at %d-bit ADC", state.PVTENOBNoiseCeiling, state.PVTENOBCeilingBits)
}

func specStorageValue(state CircuitsState) string {
	if state.SpecCells <= 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("%d cells / %.2f bits/cell", state.SpecCells, state.SpecBitsPerCell)
}

func specComponentsValue(state CircuitsState) string {
	if state.SpecDACCount <= 0 || state.SpecTIACount <= 0 || state.SpecADCCount <= 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("DAC %d / TIA %d / ADC %d", state.SpecDACCount, state.SpecTIACount, state.SpecADCCount)
}

func specPowerLatencyValue(state CircuitsState) string {
	if state.SpecTotalPowerMW <= 0 || state.SpecLatencyNS <= 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("%.1f mW / %.0f ns", state.SpecTotalPowerMW, state.SpecLatencyNS)
}

func specThroughputValue(state CircuitsState) string {
	if state.SpecThroughputGOPS <= 0 || state.SpecEfficiencyGOPSW <= 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("%.2f GOPS / %.0f GOPS/W", state.SpecThroughputGOPS, state.SpecEfficiencyGOPSW)
}

func specResolutionValue(state CircuitsState, quantLevels int) string {
	if state.SpecDACCodes <= 0 || state.SpecADCCodes <= 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("DAC %d / ADC %d / %d levels", state.SpecDACCodes, state.SpecADCCodes, quantLevels)
}

func specComplianceValue(state CircuitsState) string {
	if state.SpecCompliance == "" {
		return "not evaluated"
	}
	return state.SpecCompliance
}

func referenceSpecExportStatusValue(state CircuitsState) string {
	if state.ReferenceSpecExportStatus == "" {
		return "not exported"
	}
	return state.ReferenceSpecExportStatus
}

func referenceSpecExportPathValue(state CircuitsState) string {
	if state.ReferenceSpecExportPath == "" {
		return "none"
	}
	return state.ReferenceSpecExportPath
}

func timingTotalValue(totalNS int) string {
	if totalNS <= 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("%d ns total", totalNS)
}

func timingActiveValue(state CircuitsState) string {
	if state.TimingActiveOp == "" || state.TimingActiveTotalNS <= 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("%s %d ns total", state.TimingActiveOp, state.TimingActiveTotalNS)
}

func timingActivePhasesValue(state CircuitsState) string {
	if state.TimingActivePhases == "" {
		return "not evaluated"
	}
	return state.TimingActivePhases
}

func timingWaveformSignalsValue(state CircuitsState) string {
	waveform, ok := activeTimingWaveform(state)
	if !ok || len(waveform.Signals) == 0 {
		return "not evaluated"
	}
	names := make([]string, 0, len(waveform.Signals))
	for _, signal := range waveform.Signals {
		names = append(names, signal.Name)
	}
	return fmt.Sprintf("%s: %s", waveform.Operation, strings.Join(names, ", "))
}

func timingWaveformMarkersValue(state CircuitsState) string {
	waveform, ok := activeTimingWaveform(state)
	if !ok || len(waveform.TimeMarkers) == 0 {
		return "not evaluated"
	}
	labels := make([]string, 0, len(waveform.TimeMarkers))
	for _, marker := range waveform.TimeMarkers {
		labels = append(labels, marker.Label)
	}
	return fmt.Sprintf("%s markers: %s", waveform.Operation, strings.Join(labels, ", "))
}

func timingWaveformPhasesValue(state CircuitsState) string {
	waveform, ok := activeTimingWaveform(state)
	if !ok || len(waveform.PhaseMarkers) == 0 {
		return "not evaluated"
	}
	phases := make([]string, 0, len(waveform.PhaseMarkers))
	for _, marker := range waveform.PhaseMarkers {
		phases = append(phases, fmt.Sprintf("%s %dns", marker.Label, marker.DurationNS))
	}
	return fmt.Sprintf("%s phases: %s", waveform.Operation, strings.Join(phases, ", "))
}

func activeTimingWaveform(state CircuitsState) (ReferenceTimingWaveform, bool) {
	operation := state.TimingActiveOp
	if operation == "" {
		operation = strings.ToUpper(state.OperationMode)
	}
	for _, waveform := range state.TimingWaveforms {
		if waveform.Operation == operation {
			return waveform, true
		}
	}
	return ReferenceTimingWaveform{}, false
}

func referenceTimingExportStatusValue(state CircuitsState) string {
	if state.ReferenceTimingExportStatus == "" {
		return "not exported"
	}
	return state.ReferenceTimingExportStatus
}

func referenceTimingExportPathValue(state CircuitsState) string {
	if state.ReferenceTimingExportPath == "" {
		return "none"
	}
	return state.ReferenceTimingExportPath
}

func referenceTimingAnimationStatusValue(state CircuitsState) string {
	if state.TimingAnimationStatus == "" {
		return "not animated"
	}
	return state.TimingAnimationStatus
}

func referenceTimingAnimationStepValue(state CircuitsState) string {
	if state.TimingAnimationCurrentStep == "" {
		return "none"
	}
	return state.TimingAnimationCurrentStep
}

func referenceTimingAnimationStepsValue(state CircuitsState) string {
	if state.TimingAnimationStepTotal <= 0 {
		return "0 steps"
	}
	return fmt.Sprintf("%d steps", state.TimingAnimationStepTotal)
}

func operationLogCountValue(state CircuitsState) string {
	return fmt.Sprintf("%d total / %d shown", state.OperationLogTotal, len(state.OperationLog))
}

func operationLogLatestValue(log []OperationLogEntry) string {
	if len(log) == 0 {
		return "none"
	}
	return operationLogEntryValue(log[len(log)-1])
}

func operationLogRecentValue(log []OperationLogEntry) string {
	if len(log) == 0 {
		return "none"
	}
	start := len(log) - 3
	if start < 0 {
		start = 0
	}
	entries := make([]string, 0, len(log)-start)
	for _, entry := range log[start:] {
		entries = append(entries, operationLogEntryValue(entry))
	}
	return strings.Join(entries, " | ")
}

func operationLogSectionBody(state CircuitsState) string {
	if len(state.OperationLog) == 0 {
		return "No operation events recorded. Successful controls and operations will appear here."
	}
	lines := []string{operationLogCountValue(state)}
	for _, entry := range state.OperationLog {
		lines = append(lines, operationLogEntryValue(entry))
	}
	return strings.Join(lines, "\n")
}

func operationLogEntryValue(entry OperationLogEntry) string {
	kind := entry.Kind
	if kind == "" {
		kind = "event"
	}
	return fmt.Sprintf("%s #%d: %s", kind, entry.Sequence, entry.Message)
}

func operationLogExportStatusValue(state CircuitsState) string {
	if state.OperationLogExportStatus == "" {
		return "not exported"
	}
	return state.OperationLogExportStatus
}

func operationLogExportPathValue(state CircuitsState) string {
	if state.OperationLogExportPath == "" {
		return "none"
	}
	return state.OperationLogExportPath
}

func computeRunValue(log *ComputeRunLog) string {
	if log == nil {
		return "not evaluated"
	}
	return fmt.Sprintf("%s / %d rows / %d cells", log.ArraySize, len(log.RowResults), log.ExportedCells)
}

func computeRunPeakValue(log *ComputeRunLog) string {
	if log == nil || len(log.RowResults) == 0 {
		return "not evaluated"
	}
	row, current := computeRunPeak(log)
	return fmt.Sprintf("%.3f uA row %d", current, row)
}

func computeRunInputValue(log *ComputeRunLog) string {
	if log == nil || len(log.InputVector) == 0 {
		return "not evaluated"
	}
	return fmt.Sprintf("%.3f..%.3f V", log.InputVector[0], log.InputVector[len(log.InputVector)-1])
}

func computeRunSectionBody(log *ComputeRunLog) string {
	if log == nil {
		return "No compute run recorded. Run compute to populate deterministic MVM row and cell summaries."
	}
	row, current := computeRunPeak(log)
	return fmt.Sprintf("Rows: %d. Cells: %d. Peak current: %.3f uA at row %d. Input vector: %s. Summary-level deterministic MVM log; not a calibrated hardware trace.",
		len(log.RowResults), log.ExportedCells, current, row, computeRunInputValue(log))
}

func computeRunPeak(log *ComputeRunLog) (int, float64) {
	peakRow := 0
	peakCurrent := log.RowResults[0].CurrentUA
	for _, row := range log.RowResults[1:] {
		if row.CurrentUA > peakCurrent {
			peakRow = row.Row
			peakCurrent = row.CurrentUA
		}
	}
	return peakRow, peakCurrent
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
