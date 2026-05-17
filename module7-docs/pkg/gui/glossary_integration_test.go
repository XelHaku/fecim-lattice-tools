//go:build legacy_fyne

package gui

import "testing"

func TestDetectGlossaryTerms_CommonAliasesMapToCanonicalTerms(t *testing.T) {
	content := "Ferroelectric devices rely on hysteresis and stable polarization states."
	terms := DetectGlossaryTerms(content)

	required := map[string]bool{
		"FeCIM":                false,
		"Hysteresis Loop":      false,
		"Remnant Polarization": false,
	}

	for _, term := range terms {
		if _, ok := required[term]; ok {
			required[term] = true
		}
	}

	for term, found := range required {
		if !found {
			t.Fatalf("expected canonical term %q from alias detection, got %v", term, terms)
		}
	}
}
