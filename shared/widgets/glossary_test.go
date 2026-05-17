//go:build legacy_fyne

package widgets

import (
	"testing"
)

func TestTermsDataNotEmpty(t *testing.T) {
	if len(TermsData) == 0 {
		t.Fatal("TermsData should not be empty")
	}

	expectedMinTerms := 20
	if len(TermsData) < expectedMinTerms {
		t.Errorf("Expected at least %d terms, got %d", expectedMinTerms, len(TermsData))
	}
}

func TestReferencesDataNotEmpty(t *testing.T) {
	if len(ReferencesData) == 0 {
		t.Fatal("ReferencesData should not be empty")
	}

	expectedMinRefs := 5
	if len(ReferencesData) < expectedMinRefs {
		t.Errorf("Expected at least %d references, got %d", expectedMinRefs, len(ReferencesData))
	}
}

func TestQuickTermLookup(t *testing.T) {
	tests := []struct {
		term           string
		expectFound    bool
		containsString string
	}{
		{"FeCIM", true, "Ferroelectric Compute-in-Memory"},
		{"Ec", true, "Coercive Field"},
		{"Pr", true, "Remnant Polarization"},
		{"HZO", true, "Hafnium Zirconium Oxide"},
		{"MAC", true, "Multiply-Accumulate"},
		{"MVM", true, "Matrix-Vector Multiplication"},
		{"NONEXISTENT_TERM", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.term, func(t *testing.T) {
			result := QuickTermLookup(tt.term)

			if tt.expectFound {
				if result == "" {
					t.Errorf("Expected to find definition for %q, got empty string", tt.term)
				}
				if tt.containsString != "" && !contains(result, tt.containsString) {
					t.Errorf("Expected definition to contain %q, got: %s", tt.containsString, result)
				}
			} else {
				if result != "" {
					t.Errorf("Expected empty string for %q, got: %s", tt.term, result)
				}
			}
		})
	}
}

func TestQuickTermLookupCaseInsensitive(t *testing.T) {
	variations := []string{"FeCIM", "fecim", "FECIM", "FeCim"}
	for _, variant := range variations {
		result := QuickTermLookup(variant)
		if result == "" {
			t.Errorf("Case-insensitive lookup failed for %q", variant)
		}
	}
}

func TestGetCategories(t *testing.T) {
	categories := GetCategories()

	if len(categories) == 0 {
		t.Fatal("Expected at least one category")
	}

	expectedCategories := map[string]bool{
		"Physics":      false,
		"Architecture": false,
		"Circuits":     false,
		"Metrics":      false,
	}

	for _, cat := range categories {
		if _, ok := expectedCategories[cat]; ok {
			expectedCategories[cat] = true
		}
	}

	for cat, found := range expectedCategories {
		if !found {
			t.Errorf("Expected category %q not found", cat)
		}
	}
}

func TestGetTermsByCategory(t *testing.T) {
	categories := []string{"Physics", "Architecture", "Circuits", "Metrics"}

	for _, cat := range categories {
		terms := GetTermsByCategory(cat)
		if len(terms) == 0 {
			t.Errorf("Expected at least one term in category %q", cat)
		}

		// Verify all returned terms are in the correct category
		for _, term := range terms {
			if term.Category != cat {
				t.Errorf("Term %q has category %q, expected %q", term.Term, term.Category, cat)
			}
		}
	}
}

func TestGlossaryEntryStructure(t *testing.T) {
	for i, entry := range TermsData {
		if entry.Term == "" {
			t.Errorf("Entry %d has empty Term", i)
		}
		if entry.Definition == "" {
			t.Errorf("Entry %d (%s) has empty Definition", i, entry.Term)
		}
		if entry.Category == "" {
			t.Errorf("Entry %d (%s) has empty Category", i, entry.Term)
		}

		// Check category is valid
		validCategories := []string{"Physics", "Architecture", "Circuits", "Metrics"}
		valid := false
		for _, cat := range validCategories {
			if entry.Category == cat {
				valid = true
				break
			}
		}
		if !valid {
			t.Errorf("Entry %d (%s) has invalid category: %q", i, entry.Term, entry.Category)
		}
	}
}

func TestReferenceEntryStructure(t *testing.T) {
	for i, ref := range ReferencesData {
		if ref.Title == "" {
			t.Errorf("Reference %d has empty Title", i)
		}
		if ref.Citation == "" {
			t.Errorf("Reference %d (%s) has empty Citation", i, ref.Title)
		}
		// URL can be empty, but if present should not be whitespace
		if ref.URL != "" && len(ref.URL) < 3 {
			t.Errorf("Reference %d (%s) has suspiciously short URL: %q", i, ref.Title, ref.URL)
		}
	}
}

func TestKeyTermsPresent(t *testing.T) {
	// Ensure critical terms are defined
	criticalTerms := []string{
		"FeCIM",
		"Ec",
		"Pr",
		"HZO",
		"MAC",
		"MVM",
		"1T1R",
		"DAC",
		"ADC",
		"TIA",
		"TRL",
		"BEOL",
	}

	for _, term := range criticalTerms {
		def := QuickTermLookup(term)
		if def == "" {
			t.Errorf("Critical term %q is missing from glossary", term)
		}
	}
}

func TestGlossaryWidgetCreation(t *testing.T) {
	// Skip widget creation tests - requires Fyne app context
	t.Skip("Widget creation requires Fyne app context")
}

func TestReferencesWidgetCreation(t *testing.T) {
	// Skip widget creation tests - requires Fyne app context
	t.Skip("Widget creation requires Fyne app context")
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
