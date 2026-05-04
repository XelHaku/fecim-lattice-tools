package hysteresis

import (
	"fmt"

	"fecim-lattice-tools/shared/physics"
)

type HysteresisState struct {
	SelectedMaterial string                `json:"selected_material"`
	Materials        []*physics.HZOMaterial `json:"materials"`
	FieldRange       FieldRange            `json:"field_range"`
	LoopPoints       []LoopPoint           `json:"loop_points"`
	Waveform         string                `json:"waveform"`
	IsRunning        bool                  `json:"is_running"`
}

type FieldRange struct {
	MinField float64 `json:"min_field"`
	MaxField float64 `json:"max_field"`
}

type LoopPoint struct {
	Field        float64 `json:"field"`
	Polarization float64 `json:"polarization"`
}

func materialSummary(mat *physics.HZOMaterial) string {
	if mat == nil {
		return "N/A"
	}
	return fmt.Sprintf("Pr=%.2f µC/cm²  Ec=%.0f kV/cm  Thickness=%.1f nm  β=%.4e  γ=%.4e",
		mat.Pr*1e6, mat.Ec*1e-3, mat.Thickness*1e9, mat.BetaLandau, mat.GammaLandau)
}
