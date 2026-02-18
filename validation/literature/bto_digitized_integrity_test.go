package literature

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBTO2021_DigitizedDatasetIntegrity(t *testing.T) {
	base := filepath.Join("data", "bto2021_cryst11101192_hysteresis_digitized.csv")
	provPath := filepath.Join("data", "bto2021_cryst11101192_hysteresis_digitized.provenance.json")

	f, err := os.Open(base)
	if err != nil {
		t.Fatalf("open digitized csv: %v", err)
	}
	defer f.Close()
	recs, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read digitized csv: %v", err)
	}
	if len(recs) < 100 {
		t.Fatalf("digitized csv too small: got %d rows, want >= 100", len(recs))
	}

	raw, err := os.ReadFile(provPath)
	if err != nil {
		t.Fatalf("read provenance: %v", err)
	}
	var prov struct {
		DatasetID string `json:"dataset_id"`
		Status    string `json:"status"`
		Tier      string `json:"tier"`
		Units     struct {
			FieldDatasetUnit        string `json:"field_dataset_unit"`
			PolarizationDatasetUnit string `json:"polarization_dataset_unit"`
		} `json:"units"`
		Digitization struct {
			PointCount                int  `json:"point_count"`
			IsPlaceholderForRefinement bool `json:"is_placeholder_for_refinement"`
		} `json:"digitization"`
		Uncertainty map[string]any `json:"uncertainty"`
	}
	if err := json.Unmarshal(raw, &prov); err != nil {
		t.Fatalf("parse provenance: %v", err)
	}
	if prov.DatasetID != "bto2021_cryst11101192_hysteresis_digitized" {
		t.Fatalf("dataset_id mismatch: %q", prov.DatasetID)
	}
	if prov.Status != "pixel_digitized_curve" {
		t.Fatalf("status mismatch: %q", prov.Status)
	}
	if prov.Tier != "candidate_tier1" {
		t.Fatalf("tier mismatch: %q", prov.Tier)
	}
	if prov.Units.FieldDatasetUnit != "MV/cm" || prov.Units.PolarizationDatasetUnit != "uC/cm2" {
		t.Fatalf("unit mismatch: field=%q polar=%q", prov.Units.FieldDatasetUnit, prov.Units.PolarizationDatasetUnit)
	}
	if prov.Digitization.PointCount < 100 {
		t.Fatalf("point_count too small: %d", prov.Digitization.PointCount)
	}
	if prov.Digitization.IsPlaceholderForRefinement {
		t.Fatal("digitized dataset must not be placeholder")
	}
	if len(prov.Uncertainty) == 0 {
		t.Fatal("uncertainty block missing")
	}
}
