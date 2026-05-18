package circuits

const (
	OperationRead    = "read"
	OperationWrite   = "write"
	OperationCompute = "compute"
)

const (
	ArchitecturePassive = "0T1R (Passive)"
	Architecture1T1R    = "1T1R (Transistor)"
	Architecture2T1R    = "2T1R (Dual Transistor)"
)

const (
	CouplingIdeal = "Ideal"
	CouplingTierA = "Tier-A"
	CouplingTierB = "Tier-B"
)

const (
	ISPPEngineLevel = "Preisach (Level-based)"
	ISPPEngineLK    = "Landau-Khalatnikov (Physics ODE)"
)

const (
	HalfSelectStateInactive          = "inactive"
	HalfSelectStateColumnWriteActive = "column-write active"
	HalfSelectStateAttenuated        = "attenuated residual"
	HalfSelectStateIsolated          = "isolated"
)

const (
	ActionRunRead                  = "run_read"
	ActionRunWrite                 = "run_write"
	ActionRunCompute               = "run_compute"
	ActionExportOperationLog       = "export_operation_log"
	ActionExportReferenceSpecs     = "export_reference_specs"
	ActionExportReferenceTiming    = "export_reference_timing"
	ActionExportReferenceTimingSVG = "export_reference_timing_svg"
	ActionAnimateReferenceTiming   = "animate_reference_timing"
	ActionToggleISPP               = "toggle_ispp"
	ActionResizeArray              = "resize_array"
	ActionSetOperationMode         = "set_operation_mode"
	ActionSetArchitecture          = "set_architecture"
	ActionSelectCell               = "select_cell"
	ActionSetWriteTarget           = "set_write_target"
	ActionSetDACBits               = "set_dac_bits"
	ActionSetADCBits               = "set_adc_bits"
	ActionSetTIAGain               = "set_tia_gain"
	ActionSetCouplingTier          = "set_coupling_tier"
	ActionSetISPPEngine            = "set_ispp_engine"
	ActionSetTimingOperation       = "set_timing_operation"
	ActionSetLoggerVerbosity       = "set_logger_verbosity"
)

const DefaultQuantLevels = 30
const DefaultDisturbVoltage = 1.4
const PassiveStressPerPulse = 0.0025
const OneTOneRStressAttenuation = 20.0
const OperationLogLimit = 12

var ValidArraySizes = []int{1, 2, 4, 8, 16, 32, 64, 128}

type OperationLogEntry struct {
	Sequence int    `json:"sequence"`
	Kind     string `json:"kind"`
	Message  string `json:"message"`
}

type ComputeRunLog struct {
	Schema        string             `json:"schema"`
	ArraySize     string             `json:"array_size"`
	Material      string             `json:"material"`
	QuantLevels   int                `json:"quant_levels"`
	Architecture  string             `json:"architecture"`
	CouplingTier  string             `json:"coupling_tier"`
	InputVector   []float64          `json:"input_vector_volts"`
	Weights       [][]int            `json:"weight_matrix"`
	Conductances  [][]float64        `json:"conductance_matrix_uS"`
	RowResults    []ComputeRowResult `json:"row_results"`
	ExportedCells int                `json:"exported_cells"`
}

type ComputeRowResult struct {
	Row        int                       `json:"row"`
	Active     bool                      `json:"active"`
	CurrentUA  float64                   `json:"current_uA"`
	TIAVoltage float64                   `json:"tia_voltage_V"`
	ADCLevel   int                       `json:"adc_level"`
	Saturated  bool                      `json:"saturated"`
	CellDetail []ComputeCellContribution `json:"cell_details"`
}

type ComputeCellContribution struct {
	Col           int     `json:"col"`
	Weight        int     `json:"weight"`
	ConductanceUS float64 `json:"conductance_uS"`
	VoltageV      float64 `json:"voltage_V"`
	CurrentUA     float64 `json:"current_uA"`
}

type ReferenceSpecExport struct {
	Schema          string  `json:"schema"`
	Module          string  `json:"module"`
	Rows            int     `json:"rows"`
	Cols            int     `json:"cols"`
	QuantLevels     int     `json:"quant_levels"`
	Cells           int     `json:"cells"`
	BitsPerCell     float64 `json:"bits_per_cell"`
	DACCount        int     `json:"dac_count"`
	TIACount        int     `json:"tia_count"`
	ADCCount        int     `json:"adc_count"`
	DACCodes        int     `json:"dac_codes"`
	ADCCodes        int     `json:"adc_codes"`
	TotalPowerMW    float64 `json:"total_power_mw"`
	LatencyNS       float64 `json:"latency_ns"`
	ThroughputGOPS  float64 `json:"throughput_gops"`
	EfficiencyGOPSW float64 `json:"efficiency_gops_w"`
	Compliance      string  `json:"compliance"`
	BoundaryNotice  string  `json:"boundary_notice"`
}

type ReferenceTimingExport struct {
	Schema          string                     `json:"schema"`
	Module          string                     `json:"module"`
	OperationMode   string                     `json:"operation_mode"`
	WriteTotalNS    int                        `json:"write_total_ns"`
	ReadTotalNS     int                        `json:"read_total_ns"`
	ComputeTotalNS  int                        `json:"compute_total_ns"`
	ActiveOperation string                     `json:"active_operation"`
	ActiveTotalNS   int                        `json:"active_total_ns"`
	ActivePhases    string                     `json:"active_phases"`
	Operations      []ReferenceTimingOperation `json:"operations"`
	BoundaryNotice  string                     `json:"boundary_notice"`
}

type ReferenceTimingOperation struct {
	Operation string                 `json:"operation"`
	TotalNS   int                    `json:"total_ns"`
	Phases    []ReferenceTimingPhase `json:"phases"`
}

type ReferenceTimingPhase struct {
	Name       string `json:"name"`
	DurationNS int    `json:"duration_ns"`
}

type ReferenceTimingWaveform struct {
	Operation    string                       `json:"operation"`
	TotalNS      int                          `json:"total_ns"`
	Signals      []ReferenceTimingSignal      `json:"signals"`
	TimeMarkers  []ReferenceTimingTimeMarker  `json:"time_markers"`
	PhaseMarkers []ReferenceTimingPhaseMarker `json:"phase_markers"`
}

type ReferenceTimingSignal struct {
	Name        string                  `json:"name"`
	HighWindows []ReferenceTimingWindow `json:"high_windows"`
}

type ReferenceTimingWindow struct {
	StartPct int `json:"start_pct"`
	EndPct   int `json:"end_pct"`
}

type ReferenceTimingTimeMarker struct {
	Percent int    `json:"percent"`
	Label   string `json:"label"`
	TimeNS  int    `json:"time_ns"`
}

type ReferenceTimingPhaseMarker struct {
	Label      string `json:"label"`
	StartPct   int    `json:"start_pct"`
	EndPct     int    `json:"end_pct"`
	DurationNS int    `json:"duration_ns"`
}

type CircuitsState struct {
	Rows                           int                       `json:"rows"`
	Cols                           int                       `json:"cols"`
	OperationMode                  string                    `json:"operation_mode"`
	Architecture                   string                    `json:"architecture"`
	SelectedRow                    int                       `json:"selected_row"`
	SelectedCol                    int                       `json:"selected_col"`
	WriteTargetLevel               int                       `json:"write_target_level"`
	QuantLevels                    int                       `json:"quant_levels"`
	CouplingTier                   string                    `json:"coupling_tier"`
	ISPPEngine                     string                    `json:"ispp_engine"`
	LoggerVerbosity                string                    `json:"logger_verbosity"`
	LoggerVerbosityLevel           int                       `json:"logger_verbosity_level"`
	LastOperationStatus            string                    `json:"last_operation_status"`
	OperationLogTotal              int                       `json:"operation_log_total"`
	OperationLog                   []OperationLogEntry       `json:"operation_log"`
	OperationLogExportStatus       string                    `json:"operation_log_export_status"`
	OperationLogExportPath         string                    `json:"operation_log_export_path"`
	OperationLogExportBytes        int                       `json:"operation_log_export_bytes"`
	OperationLogExportJSON         string                    `json:"operation_log_export_json"`
	ComputeRunLog                  *ComputeRunLog            `json:"compute_run_log,omitempty"`
	ReferenceSpecExportStatus      string                    `json:"reference_spec_export_status"`
	ReferenceSpecExportPath        string                    `json:"reference_spec_export_path"`
	ReferenceSpecExportBytes       int                       `json:"reference_spec_export_bytes"`
	ReferenceSpecExportJSON        string                    `json:"reference_spec_export_json"`
	ReferenceTimingExportStatus    string                    `json:"reference_timing_export_status"`
	ReferenceTimingExportPath      string                    `json:"reference_timing_export_path"`
	ReferenceTimingExportBytes     int                       `json:"reference_timing_export_bytes"`
	ReferenceTimingExportJSON      string                    `json:"reference_timing_export_json"`
	ReferenceTimingSVGExportStatus string                    `json:"reference_timing_svg_export_status"`
	ReferenceTimingSVGExportPath   string                    `json:"reference_timing_svg_export_path"`
	ReferenceTimingSVGExportBytes  int                       `json:"reference_timing_svg_export_bytes"`
	ReferenceTimingSVGExport       string                    `json:"reference_timing_svg_export"`
	TimingAnimationOperation       string                    `json:"timing_animation_operation"`
	TimingAnimationStatus          string                    `json:"timing_animation_status"`
	TimingAnimationStepIndex       int                       `json:"timing_animation_step_index"`
	TimingAnimationStepTotal       int                       `json:"timing_animation_step_total"`
	TimingAnimationCurrentStep     string                    `json:"timing_animation_current_step"`
	TimingAnimationSteps           []string                  `json:"timing_animation_steps"`
	HalfSelectState                string                    `json:"half_select_state"`
	HalfSelectCells                int                       `json:"half_select_cells"`
	DisturbVoltage                 float64                   `json:"disturb_voltage"`
	StressPerPulse                 float64                   `json:"stress_per_pulse"`
	StressCyclesToLevel            int                       `json:"stress_cycles_to_level"`
	ADCResolution                  int                       `json:"adc_resolution"`
	DACResolution                  int                       `json:"dac_resolution"`
	TIAGain                        float64                   `json:"tia_gain"`
	ChargePumpStages               int                       `json:"charge_pump_stages"`
	SupplyVoltage                  float64                   `json:"supply_voltage"`
	ISPPEnabled                    bool                      `json:"ispp_enabled"`
	ISPPExecuted                   bool                      `json:"ispp_executed"`
	ISPPTotalAttempts              int                       `json:"ispp_total_attempts"`
	ISPPConvergedCount             int                       `json:"ispp_converged_count"`
	ISPPAvgAttempts                float64                   `json:"ispp_avg_attempts"`
	ISPPAttempts                   []int                     `json:"ispp_attempts"`
	ISPPConverged                  []bool                    `json:"ispp_converged"`
	ENOBtt                         float64                   `json:"enob_tt"`
	ENOBff                         float64                   `json:"enob_ff"`
	ENOBss                         float64                   `json:"enob_ss"`
	SNRdB                          float64                   `json:"snr_db"`
	ADCNoiseLSB                    float64                   `json:"adc_noise_lsb"`
	PVTTemperatureSweep            string                    `json:"pvt_temperature_sweep"`
	PVTProcessYield                float64                   `json:"pvt_process_yield"`
	PVTPassSamples                 int                       `json:"pvt_pass_samples"`
	PVTSamples                     int                       `json:"pvt_samples"`
	PVTENOBNoiseCeiling            float64                   `json:"pvt_enob_noise_ceiling"`
	PVTENOBCeilingBits             int                       `json:"pvt_enob_ceiling_bits"`
	SpecCells                      int                       `json:"spec_cells"`
	SpecBitsPerCell                float64                   `json:"spec_bits_per_cell"`
	SpecDACCount                   int                       `json:"spec_dac_count"`
	SpecTIACount                   int                       `json:"spec_tia_count"`
	SpecADCCount                   int                       `json:"spec_adc_count"`
	SpecDACCodes                   int                       `json:"spec_dac_codes"`
	SpecADCCodes                   int                       `json:"spec_adc_codes"`
	SpecTotalPowerMW               float64                   `json:"spec_total_power_mw"`
	SpecLatencyNS                  float64                   `json:"spec_latency_ns"`
	SpecThroughputGOPS             float64                   `json:"spec_throughput_gops"`
	SpecEfficiencyGOPSW            float64                   `json:"spec_efficiency_gops_w"`
	SpecCompliance                 string                    `json:"spec_compliance"`
	TimingWriteTotalNS             int                       `json:"timing_write_total_ns"`
	TimingReadTotalNS              int                       `json:"timing_read_total_ns"`
	TimingComputeTotalNS           int                       `json:"timing_compute_total_ns"`
	TimingOperation                string                    `json:"timing_operation"`
	TimingActiveOp                 string                    `json:"timing_active_op"`
	TimingActiveTotalNS            int                       `json:"timing_active_total_ns"`
	TimingActivePhases             string                    `json:"timing_active_phases"`
	TimingWaveforms                []ReferenceTimingWaveform `json:"timing_waveforms"`
}
