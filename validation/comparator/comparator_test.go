package comparator

import (
	"encoding/json"
	"testing"
)

func TestComparatorRunsAndOutputsJSON(t *testing.T) {
	report := RunCrossModelComparator(nil)
	if len(report.Table) == 0 {
		t.Fatal("empty mismatch table")
	}
	blob, err := report.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(blob, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded["recommendation"] == "" {
		t.Fatal("missing recommendation")
	}
}
