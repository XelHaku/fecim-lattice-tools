package presets

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPreset(t *testing.T) {
	config := map[string]interface{}{
		"frequency": 0.5,
		"amplitude": 1.5,
		"material":  "HZO",
	}

	preset := NewPreset("Test Preset", "A test preset", ModuleHysteresis, CategoryEducational, config)

	if preset.Metadata.Name != "Test Preset" {
		t.Errorf("Expected name 'Test Preset', got '%s'", preset.Metadata.Name)
	}
	if preset.Metadata.Module != ModuleHysteresis {
		t.Errorf("Expected module 'hysteresis', got '%s'", preset.Metadata.Module)
	}
	if preset.Metadata.Category != CategoryEducational {
		t.Errorf("Expected category 'educational', got '%s'", preset.Metadata.Category)
	}
	if preset.Metadata.BuiltIn {
		t.Error("Expected BuiltIn to be false for new preset")
	}
	if preset.Config["frequency"] != 0.5 {
		t.Errorf("Expected frequency 0.5, got %v", preset.Config["frequency"])
	}
}

func TestPresetGetters(t *testing.T) {
	config := map[string]interface{}{
		"float_val":  1.5,
		"int_val":    10,
		"string_val": "test",
		"bool_val":   true,
	}

	preset := NewPreset("Test", "", ModuleHysteresis, CategoryCustom, config)

	// Test GetFloat
	if v, ok := preset.GetFloat("float_val"); !ok || v != 1.5 {
		t.Errorf("GetFloat failed: got %v, %v", v, ok)
	}

	// Test GetInt
	if v, ok := preset.GetInt("int_val"); !ok || v != 10 {
		t.Errorf("GetInt failed: got %v, %v", v, ok)
	}

	// Test GetString
	if v, ok := preset.GetString("string_val"); !ok || v != "test" {
		t.Errorf("GetString failed: got %v, %v", v, ok)
	}

	// Test GetBool
	if v, ok := preset.GetBool("bool_val"); !ok || !v {
		t.Errorf("GetBool failed: got %v, %v", v, ok)
	}

	// Test missing key
	if _, ok := preset.GetFloat("missing"); ok {
		t.Error("Expected ok=false for missing key")
	}
}

func TestPresetClone(t *testing.T) {
	config := map[string]interface{}{
		"nested": map[string]interface{}{
			"value": 42,
		},
	}

	original := NewPreset("Original", "", ModuleHysteresis, CategoryCustom, config)
	clone := original.Clone()

	// Modify clone
	clone.Config["new_key"] = "new_value"

	// Original should be unchanged
	if _, ok := original.Config["new_key"]; ok {
		t.Error("Clone modified original preset")
	}
}

func TestPresetValidate(t *testing.T) {
	// Valid preset
	valid := NewPreset("Valid", "", ModuleHysteresis, CategoryCustom, map[string]interface{}{})
	if err := valid.Validate(); err != nil {
		t.Errorf("Valid preset failed validation: %v", err)
	}

	// Missing name
	invalid := &Preset{
		Metadata: Metadata{Module: ModuleHysteresis},
		Config:   map[string]interface{}{},
	}
	if err := invalid.Validate(); err == nil {
		t.Error("Expected validation error for missing name")
	}

	// Missing module
	invalid = &Preset{
		Metadata: Metadata{Name: "Test"},
		Config:   map[string]interface{}{},
	}
	if err := invalid.Validate(); err == nil {
		t.Error("Expected validation error for missing module")
	}

	// Missing config
	invalid = &Preset{
		Metadata: Metadata{Name: "Test", Module: ModuleHysteresis},
	}
	if err := invalid.Validate(); err == nil {
		t.Error("Expected validation error for missing config")
	}
}

func TestManager(t *testing.T) {
	// Create temp directory for test
	tmpDir := filepath.Join(os.TempDir(), "fecim-presets-test-"+time.Now().Format("20060102150405"))
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)

	// Test built-in presets are loaded
	builtins := manager.List(FilterByBuiltIn(true))
	if len(builtins) == 0 {
		t.Error("Expected built-in presets to be loaded")
	}

	// Test saving a new preset
	preset := NewPreset("User Preset", "A user-created preset", ModuleHysteresis, CategoryCustom, map[string]interface{}{
		"frequency": 1.0,
	})

	if err := manager.Save(preset); err != nil {
		t.Fatalf("Failed to save preset: %v", err)
	}

	// Test loading
	loaded, err := manager.Load(preset.Metadata.ID)
	if err != nil {
		t.Fatalf("Failed to load preset: %v", err)
	}
	if loaded.Metadata.Name != "User Preset" {
		t.Errorf("Loaded preset has wrong name: %s", loaded.Metadata.Name)
	}

	// Test listing with filters
	hysteresisPresets := manager.List(FilterByModule(ModuleHysteresis))
	found := false
	for _, p := range hysteresisPresets {
		if p.Metadata.ID == preset.Metadata.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Saved preset not found in filtered list")
	}

	// Test delete
	if err := manager.Delete(preset.Metadata.ID); err != nil {
		t.Fatalf("Failed to delete preset: %v", err)
	}

	if _, err := manager.Load(preset.Metadata.ID); err == nil {
		t.Error("Expected error loading deleted preset")
	}

	// Test cannot delete built-in
	if len(builtins) > 0 {
		if err := manager.Delete(builtins[0].Metadata.ID); err == nil {
			t.Error("Expected error deleting built-in preset")
		}
	}
}

func TestFilters(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "fecim-presets-filter-test-"+time.Now().Format("20060102150405"))
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)

	// Test module filter
	hysteresisPresets := manager.List(FilterByModule(ModuleHysteresis))
	for _, p := range hysteresisPresets {
		if p.Metadata.Module != ModuleHysteresis && p.Metadata.Module != ModuleGlobal {
			t.Errorf("Module filter failed: got %s", p.Metadata.Module)
		}
	}

	// Test category filter
	eduPresets := manager.List(FilterByCategory(CategoryEducational))
	for _, p := range eduPresets {
		if p.Metadata.Category != CategoryEducational {
			t.Errorf("Category filter failed: got %s", p.Metadata.Category)
		}
	}

	// Test search filter
	searchResults := manager.List(SearchByName("basic"))
	for _, p := range searchResults {
		if !containsIgnoreCase(p.Metadata.Name, "basic") &&
			!containsIgnoreCase(p.Metadata.Description, "basic") {
			t.Errorf("Search filter failed: %s", p.Metadata.Name)
		}
	}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		containsIgnoreCaseImpl(s, substr))
}

func containsIgnoreCaseImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestExportImport(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "fecim-presets-export-test-"+time.Now().Format("20060102150405"))
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)

	// Create and save a preset
	preset := NewPreset("Export Test", "Test export/import", ModuleHysteresis, CategoryCustom, map[string]interface{}{
		"test_key": "test_value",
	})
	if err := manager.Save(preset); err != nil {
		t.Fatalf("Failed to save preset: %v", err)
	}

	// Export
	exportPath := filepath.Join(tmpDir, "exported.json")
	if err := manager.Export(preset.Metadata.ID, exportPath); err != nil {
		t.Fatalf("Failed to export preset: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Error("Export file not created")
	}

	// Import into new manager
	manager2 := NewManager(filepath.Join(tmpDir, "import"))
	imported, err := manager2.Import(exportPath)
	if err != nil {
		t.Fatalf("Failed to import preset: %v", err)
	}

	if imported.Metadata.Name != "Export Test" {
		t.Errorf("Imported preset has wrong name: %s", imported.Metadata.Name)
	}
	if imported.Metadata.BuiltIn {
		t.Error("Imported preset should not be built-in")
	}
	if v, ok := imported.GetString("test_key"); !ok || v != "test_value" {
		t.Error("Imported preset config mismatch")
	}
}

func TestBuiltInPresets(t *testing.T) {
	builtins := getAllBuiltInPresets()

	if len(builtins) == 0 {
		t.Error("No built-in presets defined")
	}

	// Check each built-in is valid
	for _, p := range builtins {
		if err := p.Validate(); err != nil {
			t.Errorf("Built-in preset '%s' is invalid: %v", p.Metadata.Name, err)
		}
		if !p.Metadata.BuiltIn {
			t.Errorf("Built-in preset '%s' has BuiltIn=false", p.Metadata.Name)
		}
	}

	// Check we have presets for each module
	modules := make(map[Module]bool)
	for _, p := range builtins {
		modules[p.Metadata.Module] = true
	}

	expectedModules := []Module{ModuleGlobal, ModuleHysteresis, ModuleCrossbar, ModuleMNIST, ModuleCircuits, ModuleComparison, ModuleEDA}
	for _, m := range expectedModules {
		if !modules[m] {
			t.Errorf("No built-in presets for module: %s", m)
		}
	}
}

func TestCategories(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "fecim-presets-cat-test-"+time.Now().Format("20060102150405"))
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)

	categories := manager.GetCategories()
	if len(categories) == 0 {
		t.Error("Expected categories from built-in presets")
	}
}

func TestTags(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "fecim-presets-tag-test-"+time.Now().Format("20060102150405"))
	defer os.RemoveAll(tmpDir)

	manager := NewManager(tmpDir)

	tags := manager.GetTags()
	if len(tags) == 0 {
		t.Error("Expected tags from built-in presets")
	}

	// Test tag filter
	taggedPresets := manager.List(FilterByTag("learning"))
	if len(taggedPresets) == 0 {
		t.Error("Expected presets with 'learning' tag")
	}
	for _, p := range taggedPresets {
		found := false
		for _, tag := range p.Metadata.Tags {
			if tag == "learning" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Preset '%s' doesn't have 'learning' tag", p.Metadata.Name)
		}
	}
}
