package equations

import (
	"encoding/json"
	"testing"
)

func TestEmbeddedEquationAssets_NonEmptyAndValidJSON(t *testing.T) {
	if len(LkEquationSVG) == 0 {
		t.Fatal("LkEquationSVG is empty")
	}
	if len(PreisachEquationSVG) == 0 {
		t.Fatal("PreisachEquationSVG is empty")
	}
	if len(LkHotspotsJSON) == 0 {
		t.Fatal("LkHotspotsJSON is empty")
	}
	var v any
	if err := json.Unmarshal(LkHotspotsJSON, &v); err != nil {
		t.Fatalf("LkHotspotsJSON invalid JSON: %v", err)
	}
}
