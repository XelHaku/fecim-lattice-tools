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

var ValidArraySizes = []int{1, 2, 4, 8, 16, 32, 64, 128}

type CircuitsState struct {
	Rows                int     `json:"rows"`
	Cols                int     `json:"cols"`
	OperationMode       string  `json:"operation_mode"`
	Architecture        string  `json:"architecture"`
	SelectedRow         int     `json:"selected_row"`
	SelectedCol         int     `json:"selected_col"`
	WriteTargetLevel    int     `json:"write_target_level"`
	QuantLevels         int     `json:"quant_levels"`
	CouplingTier        string  `json:"coupling_tier"`
	ISPPEngine          string  `json:"ispp_engine"`
	LastOperationStatus string  `json:"last_operation_status"`
	ADCResolution       int     `json:"adc_resolution"`
	DACResolution       int     `json:"dac_resolution"`
	TIAGain             float64 `json:"tia_gain"`
	ChargePumpStages    int     `json:"charge_pump_stages"`
	SupplyVoltage       float64 `json:"supply_voltage"`
	ISPPEnabled         bool    `json:"ispp_enabled"`
	ISPPExecuted        bool    `json:"ispp_executed"`
	ISPPTotalAttempts   int     `json:"ispp_total_attempts"`
	ISPPConvergedCount  int     `json:"ispp_converged_count"`
	ISPPAvgAttempts     float64 `json:"ispp_avg_attempts"`
	ISPPAttempts        []int   `json:"ispp_attempts"`
	ISPPConverged       []bool  `json:"ispp_converged"`
	ENOBtt              float64 `json:"enob_tt"`
	ENOBff              float64 `json:"enob_ff"`
	ENOBss              float64 `json:"enob_ss"`
	SNRdB               float64 `json:"snr_db"`
	ADCNoiseLSB         float64 `json:"adc_noise_lsb"`
}
