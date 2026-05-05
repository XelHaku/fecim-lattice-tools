package circuits

type CircuitsState struct {
	ADCResolution    int     `json:"adc_resolution"`
	DACResolution    int     `json:"dac_resolution"`
	TIAGain          float64 `json:"tia_gain"`
	ChargePumpStages int     `json:"charge_pump_stages"`
	SupplyVoltage    float64 `json:"supply_voltage"`
	ISPPEnabled      bool    `json:"ispp_enabled"`
	ISPPExecuted     bool    `json:"ispp_executed"`
	ISPPTotalAttempts int    `json:"ispp_total_attempts"`
	ISPPConvergedCount int   `json:"ispp_converged_count"`
	ISPPAvgAttempts  float64 `json:"ispp_avg_attempts"`
	ISPPAttempts     []int   `json:"ispp_attempts"`
	ISPPConverged    []bool  `json:"ispp_converged"`
}
