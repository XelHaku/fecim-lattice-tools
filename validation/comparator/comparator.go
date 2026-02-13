package comparator

import (
	"encoding/json"
	"math"

	"fecim-lattice-tools/validation/heracles"
)

// PointComparison stores model polarization values and mismatch to Heracles reference.
type PointComparison struct {
	E                float64 `json:"E_V_m"`
	PreisachP        float64 `json:"preisach_P_C_m2"`
	LKP              float64 `json:"lk_P_C_m2"`
	HeraclesP        float64 `json:"heracles_P_C_m2"`
	PreisachAbsError float64 `json:"preisach_abs_error"`
	LKAbsError       float64 `json:"lk_abs_error"`
}

// ComparatorReport is the JSON-serializable output.
type ComparatorReport struct {
	SweepE         []float64         `json:"sweep_E_V_m"`
	Table          []PointComparison `json:"mismatch_table"`
	PreisachMAE    float64           `json:"preisach_mae"`
	LKMAE          float64           `json:"lk_mae"`
	Recommendation string            `json:"recommendation"`
}

// RunCrossModelComparator runs the same E sweep through Preisach, LK, and Heracles reference.
func RunCrossModelComparator(sweepE []float64) ComparatorReport {
	if len(sweepE) == 0 {
		sweepE = defaultSweep()
	}
	ref := heracles.Reference10nmHfO2_300K()
	table := make([]PointComparison, 0, len(sweepE))
	var sumP, sumL float64
	for _, e := range sweepE {
		h := interpolateHeracles(ref, e)
		pp := preisachModel(e)
		lp := lkModel(e)
		pe := math.Abs(pp - h)
		le := math.Abs(lp - h)
		sumP += pe
		sumL += le
		table = append(table, PointComparison{E: e, PreisachP: pp, LKP: lp, HeraclesP: h, PreisachAbsError: pe, LKAbsError: le})
	}
	maeP := sumP / float64(len(table))
	maeL := sumL / float64(len(table))
	rec := "Preisach"
	if maeL < maeP {
		rec = "LK"
	}
	return ComparatorReport{SweepE: sweepE, Table: table, PreisachMAE: maeP, LKMAE: maeL, Recommendation: rec}
}

// ToJSON serializes report.
func (r ComparatorReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func preisachModel(E float64) float64 {
	ps := 0.30
	ec := 1.1e8
	return ps * math.Tanh(E/ec)
}

func lkModel(E float64) float64 {
	ps := 0.30
	ec := 1.2e8
	return ps * (2 / math.Pi) * math.Atan(2.5*E/ec)
}

func interpolateHeracles(ref heracles.HeraclesReferenceDataset, E float64) float64 {
	eMVcm := E / 1e8
	pts := ref.Ascending
	if len(pts) == 0 {
		return 0
	}
	if eMVcm <= pts[0].E_MVcm {
		return pts[0].P_uCcm * 1e-2
	}
	if eMVcm >= pts[len(pts)-1].E_MVcm {
		return pts[len(pts)-1].P_uCcm * 1e-2
	}
	for i := 1; i < len(pts); i++ {
		a, b := pts[i-1], pts[i]
		if eMVcm >= a.E_MVcm && eMVcm <= b.E_MVcm {
			f := (eMVcm - a.E_MVcm) / (b.E_MVcm - a.E_MVcm)
			p := a.P_uCcm + f*(b.P_uCcm-a.P_uCcm)
			return p * 1e-2
		}
	}
	return pts[len(pts)-1].P_uCcm * 1e-2
}

func defaultSweep() []float64 {
	out := make([]float64, 0, 31)
	for e := -3.0e8; e <= 3.0e8; e += 2.0e7 {
		out = append(out, e)
	}
	return out
}
