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
	ActionRunRead          = "run_read"
	ActionRunWrite         = "run_write"
	ActionRunCompute       = "run_compute"
	ActionToggleISPP       = "toggle_ispp"
	ActionResizeArray      = "resize_array"
	ActionSetOperationMode = "set_operation_mode"
	ActionSetArchitecture  = "set_architecture"
	ActionSelectCell       = "select_cell"
	ActionSetWriteTarget   = "set_write_target"
	ActionSetDACBits       = "set_dac_bits"
	ActionSetADCBits       = "set_adc_bits"
	ActionSetTIAGain       = "set_tia_gain"
	ActionSetCouplingTier  = "set_coupling_tier"
	ActionSetISPPEngine    = "set_ispp_engine"
)

const DefaultQuantLevels = 30
const DefaultDisturbVoltage = 1.4
const PassiveStressPerPulse = 0.0025
const OneTOneRStressAttenuation = 20.0

var ValidArraySizes = []int{1, 2, 4, 8, 16, 32, 64, 128}

type CircuitsState struct {
	Rows                 int     `json:"rows"`
	Cols                 int     `json:"cols"`
	OperationMode        string  `json:"operation_mode"`
	Architecture         string  `json:"architecture"`
	SelectedRow          int     `json:"selected_row"`
	SelectedCol          int     `json:"selected_col"`
	WriteTargetLevel     int     `json:"write_target_level"`
	QuantLevels          int     `json:"quant_levels"`
	CouplingTier         string  `json:"coupling_tier"`
	ISPPEngine           string  `json:"ispp_engine"`
	LastOperationStatus  string  `json:"last_operation_status"`
	HalfSelectState      string  `json:"half_select_state"`
	HalfSelectCells      int     `json:"half_select_cells"`
	DisturbVoltage       float64 `json:"disturb_voltage"`
	StressPerPulse       float64 `json:"stress_per_pulse"`
	StressCyclesToLevel  int     `json:"stress_cycles_to_level"`
	ADCResolution        int     `json:"adc_resolution"`
	DACResolution        int     `json:"dac_resolution"`
	TIAGain              float64 `json:"tia_gain"`
	ChargePumpStages     int     `json:"charge_pump_stages"`
	SupplyVoltage        float64 `json:"supply_voltage"`
	ISPPEnabled          bool    `json:"ispp_enabled"`
	ISPPExecuted         bool    `json:"ispp_executed"`
	ISPPTotalAttempts    int     `json:"ispp_total_attempts"`
	ISPPConvergedCount   int     `json:"ispp_converged_count"`
	ISPPAvgAttempts      float64 `json:"ispp_avg_attempts"`
	ISPPAttempts         []int   `json:"ispp_attempts"`
	ISPPConverged        []bool  `json:"ispp_converged"`
	ENOBtt               float64 `json:"enob_tt"`
	ENOBff               float64 `json:"enob_ff"`
	ENOBss               float64 `json:"enob_ss"`
	SNRdB                float64 `json:"snr_db"`
	ADCNoiseLSB          float64 `json:"adc_noise_lsb"`
	PVTTemperatureSweep  string  `json:"pvt_temperature_sweep"`
	PVTProcessYield      float64 `json:"pvt_process_yield"`
	PVTPassSamples       int     `json:"pvt_pass_samples"`
	PVTSamples           int     `json:"pvt_samples"`
	PVTENOBNoiseCeiling  float64 `json:"pvt_enob_noise_ceiling"`
	PVTENOBCeilingBits   int     `json:"pvt_enob_ceiling_bits"`
	SpecCells            int     `json:"spec_cells"`
	SpecBitsPerCell      float64 `json:"spec_bits_per_cell"`
	SpecDACCount         int     `json:"spec_dac_count"`
	SpecTIACount         int     `json:"spec_tia_count"`
	SpecADCCount         int     `json:"spec_adc_count"`
	SpecDACCodes         int     `json:"spec_dac_codes"`
	SpecADCCodes         int     `json:"spec_adc_codes"`
	SpecTotalPowerMW     float64 `json:"spec_total_power_mw"`
	SpecLatencyNS        float64 `json:"spec_latency_ns"`
	SpecThroughputGOPS   float64 `json:"spec_throughput_gops"`
	SpecEfficiencyGOPSW  float64 `json:"spec_efficiency_gops_w"`
	SpecCompliance       string  `json:"spec_compliance"`
	TimingWriteTotalNS   int     `json:"timing_write_total_ns"`
	TimingReadTotalNS    int     `json:"timing_read_total_ns"`
	TimingComputeTotalNS int     `json:"timing_compute_total_ns"`
	TimingActiveOp       string  `json:"timing_active_op"`
	TimingActiveTotalNS  int     `json:"timing_active_total_ns"`
	TimingActivePhases   string  `json:"timing_active_phases"`
}
