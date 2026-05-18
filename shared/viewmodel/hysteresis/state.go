package hysteresis

import (
	"fmt"

	"fecim-lattice-tools/shared/physics"
)

type HysteresisState struct {
	SelectedMaterial string                 `json:"selected_material"`
	Materials        []*physics.HZOMaterial `json:"materials"`
	FieldRange       FieldRange             `json:"field_range"`
	LoopPoints       []LoopPoint            `json:"loop_points"`
	Waveform         string                 `json:"waveform"`
	IsRunning        bool                   `json:"is_running"`
	Pr               float64                `json:"pr"`
	EcPlus           float64                `json:"ec_plus"`
	EcMinus          float64                `json:"ec_minus"`
	Psat             float64                `json:"psat"`
	PsatNeg          float64                `json:"psat_neg"`
	LoopArea         float64                `json:"loop_area"`
	RetentionTimes   []float64              `json:"retention_times"`
	RetentionPr      []float64              `json:"retention_pr"`
	PUND             PUNDSummary            `json:"pund"`
	FORC             FORCSummary            `json:"forc"`
	CSVExportStatus  string                 `json:"csv_export_status"`
	CSVExportPath    string                 `json:"csv_export_path"`
	CSVExportBytes   int                    `json:"csv_export_bytes"`
	CSVExportContent string                 `json:"csv_export_content,omitempty"`
}

type FieldRange struct {
	MinField float64 `json:"min_field"`
	MaxField float64 `json:"max_field"`
}

type LoopPoint struct {
	Field        float64 `json:"field"`
	Polarization float64 `json:"polarization"`
}

type PUNDSummary struct {
	Available         bool    `json:"available"`
	QP_C              float64 `json:"qp_c"`
	QU_C              float64 `json:"qu_c"`
	QN_C              float64 `json:"qn_c"`
	QD_C              float64 `json:"qd_c"`
	SwitchingPositive float64 `json:"switching_positive_c"`
	SwitchingNegative float64 `json:"switching_negative_c"`
	SwitchingRatio    float64 `json:"switching_ratio"`
	SamplesPerPulse   int     `json:"samples_per_pulse"`
	Summary           string  `json:"summary"`
}

type FORCSummary struct {
	Available   bool    `json:"available"`
	Curves      int     `json:"curves"`
	DensityRows int     `json:"density_rows"`
	DensityCols int     `json:"density_cols"`
	PeakDensity float64 `json:"peak_density"`
	PeakEa_Vm   float64 `json:"peak_ea_vm"`
	PeakEb_Vm   float64 `json:"peak_eb_vm"`
	MinDensity  float64 `json:"min_density"`
	MaxDensity  float64 `json:"max_density"`
	Summary     string  `json:"summary"`
}

func materialSummary(mat *physics.HZOMaterial) string {
	if mat == nil {
		return "N/A"
	}
	return fmt.Sprintf("Pr=%.2f µC/cm²  Ec=%.0f kV/cm  Thickness=%.1f nm  β=%.4e  γ=%.4e",
		mat.Pr*1e6, mat.Ec*1e-3, mat.Thickness*1e9, mat.BetaLandau, mat.GammaLandau)
}
