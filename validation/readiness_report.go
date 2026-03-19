package validation

import (
	"encoding/json"
	"math"

	"fecim-lattice-tools/shared/mathutil"
)

// ReadinessInputs captures normalized [0..1] inputs.
type ReadinessInputs struct {
	TestPassRate          float64
	CalibrationStatus     float64
	ExportQuality         float64
	DocumentationCoverage float64
}

// ReadinessReport is the executive summary output.
type ReadinessReport struct {
	EducationScore int      `json:"education_score"`
	ResearchScore  int      `json:"research_score"`
	DesignScore    int      `json:"design_score"`
	NextTasks      []string `json:"next_tasks"`
}

// BuildExecutiveReadinessReport generates scorecard + next 5 tasks.
func BuildExecutiveReadinessReport(in ReadinessInputs) ReadinessReport {
	test := mathutil.Clamp01(in.TestPassRate)
	cal := mathutil.Clamp01(in.CalibrationStatus)
	exp := mathutil.Clamp01(in.ExportQuality)
	doc := mathutil.Clamp01(in.DocumentationCoverage)

	education := int(math.Round((0.30*test + 0.20*cal + 0.15*exp + 0.35*doc) * 100))
	research := int(math.Round((0.35*test + 0.35*cal + 0.20*exp + 0.10*doc) * 100))
	design := int(math.Round((0.30*test + 0.20*cal + 0.40*exp + 0.10*doc) * 100))

	next := []string{
		"Raise short-test pass rate above 95% with flaky-test cleanup",
		"Complete calibration pack for all default materials",
		"Increase export validator coverage for JSON/CSV/SPICE/Verilog",
		"Close documentation gaps in module-level PHYSICS and FEATURES pages",
		"Run end-to-end readiness review and publish release checklist",
	}

	return ReadinessReport{
		EducationScore: education,
		ResearchScore:  research,
		DesignScore:    design,
		NextTasks:      next,
	}
}

// ReadinessReportJSON marshals the report to JSON for downstream pipelines.
func ReadinessReportJSON(in ReadinessInputs) ([]byte, error) {
	report := BuildExecutiveReadinessReport(in)
	return json.MarshalIndent(report, "", "  ")
}
