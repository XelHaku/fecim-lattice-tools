package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type module4KCLScenarioRecord struct {
	Name    string  `json:"name"`
	Size    int     `json:"size"`
	PassKCL bool    `json:"pass_kcl"`
	PassKVL bool    `json:"pass_kvl"`
	MaxKCL  float64 `json:"max_kcl_error_A"`
	MaxKVL  float64 `json:"max_kvl_error_V"`
}

type module4KCLScenarioArtifact struct {
	Results []module4KCLScenarioRecord `json:"results"`
}

func TestModule4KCLKVLScenarioFamily_CompletenessContract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	p := filepath.Join(repoRoot, "validation", "output", "validation", "module4", "kcl_kvl_exhaustive.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var rec module4KCLScenarioArtifact
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("decode %s: %v", p, err)
	}
	if len(rec.Results) == 0 {
		t.Fatalf("%s has empty results", p)
	}

	need := map[string]bool{
		"uniform_4x4":      false,
		"uniform_8x8":      false,
		"uniform_16x16":    false,
		"uniform_32x32":    false,
		"random_4x4":       false,
		"random_8x8":       false,
		"random_16x16":     false,
		"hi_contrast_8x8":  false,
		"hi_contrast_16x16": false,
		"gradient_8x8":     false,
		"gradient_16x16":   false,
		"half_select_8x8":  false,
		"half_select_16x16": false,
		"all_one_write_8x8": false,
	}

	for i, s := range rec.Results {
		if s.Name == "" || s.Size <= 0 {
			t.Fatalf("%s[%d] invalid name/size", p, i)
		}
		if !s.PassKCL || !s.PassKVL {
			t.Fatalf("%s[%d] pass flags false: kcl=%v kvl=%v", p, i, s.PassKCL, s.PassKVL)
		}
		if s.MaxKCL < 0 || s.MaxKVL < 0 {
			t.Fatalf("%s[%d] negative errors: kcl=%g kvl=%g", p, i, s.MaxKCL, s.MaxKVL)
		}
		if _, ok := need[s.Name]; ok {
			need[s.Name] = true
		}
	}

	for name, seen := range need {
		if !seen {
			t.Fatalf("%s missing required scenario %s", p, name)
		}
	}
}
