// Command release-report generates release notes from git history.
//
// Usage: release-report -from v1.0.0 -to HEAD -output RELEASE_NOTES.md
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type materialPhysicsSnapshot struct {
	MaterialID string `json:"material_id"`
}

type regressionSummary struct {
	Suite            string                  `json:"suite"`
	Model            string                  `json:"model"`
	AllPass          bool                    `json:"all_pass"`
	MaterialSnapshot materialPhysicsSnapshot `json:"material_snapshot"`
}

type perMaterialReport struct {
	MaterialID     string          `json:"material_id"`
	ByModel        map[string]bool `json:"by_model"`
	AllModelsPass  bool            `json:"all_models_pass"`
	ObservedModels []string        `json:"observed_models"`
	MissingModels  []string        `json:"missing_models"`
	ObservedFiles  []string        `json:"observed_files"`
	ObservedSuites []string        `json:"observed_suites"`
	Notes          []string        `json:"notes,omitempty"`
}

type doeCoverage struct {
	ExpectedMaterials int     `json:"expected_materials"`
	ExpectedModels    int     `json:"expected_models"`
	ExpectedCombos    int     `json:"expected_combos"`
	FoundCombos       int     `json:"found_combos"`
	CoverageFrac      float64 `json:"coverage_frac"`
}

type releaseReport struct {
	GeneratedAt time.Time                    `json:"generated_at"`
	InputDir    string                       `json:"input_dir"`
	FileCount   int                          `json:"file_count"`
	DOE         doeCoverage                  `json:"doe"`
	PerMaterial map[string]perMaterialReport `json:"per_material"`
	Errors      []string                     `json:"errors,omitempty"`
}

func main() {
	var inDir string
	var outJSON string
	var outMD string
	var materialsCSV string
	var modelsCSV string

	flag.StringVar(&inDir, "in", "", "input directory containing regression JSON artifacts")
	flag.StringVar(&outJSON, "out-json", "", "output JSON path")
	flag.StringVar(&outMD, "out-md", "", "output markdown path")
	flag.StringVar(&materialsCSV, "materials", "fecim_hzo,literature_superlattice,default_hzo", "expected material ids (comma separated)")
	flag.StringVar(&modelsCSV, "models", "preisach,landau-khalatnikov", "expected models (comma separated)")
	flag.Parse()

	if inDir == "" {
		// Prefer repo artifacts if present; otherwise, use $TMPDIR/fecim-regression.
		candidate := filepath.Join("output", "regression")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			inDir = candidate
		} else {
			inDir = filepath.Join(os.TempDir(), "fecim-regression")
		}
	}
	if outJSON == "" {
		outJSON = filepath.Join("docs", "release", "rg-val-05-release-report.json")
	}
	if outMD == "" {
		outMD = filepath.Join("docs", "release", "rg-val-05-release-report.md")
	}

	expectedMaterials := splitCSV(materialsCSV)
	expectedModels := splitCSV(modelsCSV)

	report := releaseReport{
		GeneratedAt: time.Now().UTC(),
		InputDir:    inDir,
		PerMaterial: map[string]perMaterialReport{},
	}

	observedCombos := map[string]bool{}

	var files []string
	_ = filepath.WalkDir(inDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("walk %s: %v", path, err))
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		files = append(files, path)
		return nil
	})

	sort.Strings(files)
	report.FileCount = len(files)

	for _, p := range files {
		b, err := os.ReadFile(p)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("read %s: %v", p, err))
			continue
		}
		var s regressionSummary
		if err := json.Unmarshal(b, &s); err != nil {
			// Not a regression summary artifact; ignore silently.
			continue
		}
		matID := strings.TrimSpace(s.MaterialSnapshot.MaterialID)
		model := strings.TrimSpace(s.Model)
		if matID == "" || model == "" {
			continue
		}

		key := matID + "|" + model
		observedCombos[key] = true

		r := report.PerMaterial[matID]
		if r.MaterialID == "" {
			r.MaterialID = matID
			r.ByModel = map[string]bool{}
		}
		r.ByModel[model] = s.AllPass
		r.ObservedFiles = append(r.ObservedFiles, p)
		if s.Suite != "" {
			r.ObservedSuites = append(r.ObservedSuites, s.Suite)
		}
		report.PerMaterial[matID] = r
	}

	for matID, r := range report.PerMaterial {
		// observed models list
		var observed []string
		for m := range r.ByModel {
			observed = append(observed, m)
		}
		sort.Strings(observed)
		r.ObservedModels = observed

		// missing models list
		missing := difference(expectedModels, observed)
		r.MissingModels = missing

		// AllModelsPass means all expected models are present AND pass.
		allPass := true
		for _, m := range expectedModels {
			v, ok := r.ByModel[m]
			if !ok || !v {
				allPass = false
				break
			}
		}
		r.AllModelsPass = allPass
		report.PerMaterial[matID] = r
	}

	report.DOE = doeCoverage{
		ExpectedMaterials: len(expectedMaterials),
		ExpectedModels:    len(expectedModels),
		ExpectedCombos:    len(expectedMaterials) * len(expectedModels),
		FoundCombos:       len(observedCombos),
	}
	if report.DOE.ExpectedCombos > 0 {
		report.DOE.CoverageFrac = float64(report.DOE.FoundCombos) / float64(report.DOE.ExpectedCombos)
	}

	if err := os.MkdirAll(filepath.Dir(outJSON), 0o755); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(outMD), 0o755); err != nil {
		fatal(err)
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(outJSON, b, 0o644); err != nil {
		fatal(err)
	}

	md := renderMarkdown(report, expectedMaterials, expectedModels)
	if err := os.WriteFile(outMD, []byte(md), 0o644); err != nil {
		fatal(err)
	}
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func difference(expected, observed []string) []string {
	set := map[string]bool{}
	for _, o := range observed {
		set[o] = true
	}
	var out []string
	for _, e := range expected {
		if !set[e] {
			out = append(out, e)
		}
	}
	return out
}

func renderMarkdown(r releaseReport, expectedMaterials, expectedModels []string) string {
	var sb strings.Builder
	sb.WriteString("# RG-VAL-05 Release Report\n\n")
	sb.WriteString(fmt.Sprintf("GeneratedAt (UTC): %s\n\n", r.GeneratedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("InputDir: `%s`\n\n", r.InputDir))
	sb.WriteString("## DOE Coverage\n\n")
	sb.WriteString(fmt.Sprintf("- Expected materials: %d (%s)\n", r.DOE.ExpectedMaterials, strings.Join(expectedMaterials, ", ")))
	sb.WriteString(fmt.Sprintf("- Expected models: %d (%s)\n", r.DOE.ExpectedModels, strings.Join(expectedModels, ", ")))
	sb.WriteString(fmt.Sprintf("- Expected combos: %d\n", r.DOE.ExpectedCombos))
	sb.WriteString(fmt.Sprintf("- Found combos: %d\n", r.DOE.FoundCombos))
	sb.WriteString(fmt.Sprintf("- Coverage: %.1f%%\n\n", 100.0*r.DOE.CoverageFrac))

	sb.WriteString("## Per-material pass map\n\n")

	matIDs := make([]string, 0, len(r.PerMaterial))
	for k := range r.PerMaterial {
		matIDs = append(matIDs, k)
	}
	sort.Strings(matIDs)

	for _, matID := range matIDs {
		pm := r.PerMaterial[matID]
		sb.WriteString(fmt.Sprintf("### %s\n\n", matID))
		sb.WriteString(fmt.Sprintf("- All expected models pass: %v\n", pm.AllModelsPass))
		sb.WriteString(fmt.Sprintf("- Observed models: %s\n", strings.Join(pm.ObservedModels, ", ")))
		if len(pm.MissingModels) > 0 {
			sb.WriteString(fmt.Sprintf("- Missing models: %s\n", strings.Join(pm.MissingModels, ", ")))
		}
		// Deterministic model order.
		for _, m := range expectedModels {
			v, ok := pm.ByModel[m]
			if !ok {
				continue
			}
			sb.WriteString(fmt.Sprintf("- %s: %v\n", m, v))
		}
		sb.WriteString("\n")
	}

	if len(r.Errors) > 0 {
		sb.WriteString("## Errors\n\n")
		for _, e := range r.Errors {
			sb.WriteString("- " + e + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
