package recentfiles

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager(nil)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if len(m.files) != 0 {
		t.Errorf("expected empty files list, got %d", len(m.files))
	}
}

func TestManagerAdd(t *testing.T) {
	m := NewManager(nil)

	m.Add("/path/to/file.json", FileTypeConfig, "hysteresis")

	files := m.ListAll()
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if files[0].Name != "file.json" {
		t.Errorf("expected name 'file.json', got '%s'", files[0].Name)
	}
	if files[0].Type != FileTypeConfig {
		t.Errorf("expected type Config, got '%s'", files[0].Type)
	}
	if files[0].Module != "hysteresis" {
		t.Errorf("expected module 'hysteresis', got '%s'", files[0].Module)
	}
}

func TestManagerAddUpdatesExisting(t *testing.T) {
	m := NewManager(nil)

	m.Add("/path/to/file.json", FileTypeConfig, "hysteresis")
	time.Sleep(10 * time.Millisecond)
	m.Add("/path/to/file.json", FileTypeConfig, "crossbar")

	files := m.ListAll()
	if len(files) != 1 {
		t.Fatalf("expected 1 file (not duplicated), got %d", len(files))
	}

	if files[0].Module != "crossbar" {
		t.Errorf("expected module 'crossbar' (updated), got '%s'", files[0].Module)
	}
}

func TestManagerRemove(t *testing.T) {
	m := NewManager(nil)

	m.Add("/path/to/file1.json", FileTypeConfig, "test")
	m.Add("/path/to/file2.json", FileTypeConfig, "test")

	removed := m.Remove("/path/to/file1.json")
	if !removed {
		t.Error("expected Remove to return true")
	}

	files := m.ListAll()
	if len(files) != 1 {
		t.Fatalf("expected 1 file after removal, got %d", len(files))
	}
	if files[0].Name != "file2.json" {
		t.Errorf("expected 'file2.json' to remain, got '%s'", files[0].Name)
	}
}

func TestManagerClear(t *testing.T) {
	m := NewManager(nil)

	m.Add("/path/to/config.json", FileTypeConfig, "test")
	m.Add("/path/to/export.csv", FileTypeExport, "test")

	m.Clear(FileTypeConfig)

	configs := m.ListConfigs()
	if len(configs) != 0 {
		t.Errorf("expected 0 configs after clear, got %d", len(configs))
	}

	exports := m.ListExports()
	if len(exports) != 1 {
		t.Errorf("expected 1 export to remain, got %d", len(exports))
	}
}

func TestManagerClearAll(t *testing.T) {
	m := NewManager(nil)

	m.Add("/path/to/config.json", FileTypeConfig, "test")
	m.Add("/path/to/export.csv", FileTypeExport, "test")

	m.ClearAll()

	if m.Count(FileTypeAny) != 0 {
		t.Errorf("expected 0 files after ClearAll, got %d", m.Count(FileTypeAny))
	}
}

func TestManagerListByType(t *testing.T) {
	m := NewManager(nil)

	m.Add("/config1.json", FileTypeConfig, "test")
	m.Add("/config2.json", FileTypeConfig, "test")
	m.Add("/export.csv", FileTypeExport, "test")

	configs := m.ListConfigs()
	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}

	exports := m.ListExports()
	if len(exports) != 1 {
		t.Errorf("expected 1 export, got %d", len(exports))
	}

	all := m.ListAll()
	if len(all) != 3 {
		t.Errorf("expected 3 total, got %d", len(all))
	}
}

func TestManagerListByModule(t *testing.T) {
	m := NewManager(nil)

	m.Add("/hyst1.json", FileTypeConfig, "hysteresis")
	m.Add("/hyst2.json", FileTypeConfig, "hysteresis")
	m.Add("/cross1.json", FileTypeConfig, "crossbar")

	hystFiles := m.ListByModule("hysteresis")
	if len(hystFiles) != 2 {
		t.Errorf("expected 2 hysteresis files, got %d", len(hystFiles))
	}

	crossFiles := m.ListByModule("crossbar")
	if len(crossFiles) != 1 {
		t.Errorf("expected 1 crossbar file, got %d", len(crossFiles))
	}
}

func TestManagerMostRecent(t *testing.T) {
	m := NewManager(nil)

	m.Add("/first.json", FileTypeConfig, "test")
	time.Sleep(10 * time.Millisecond)
	m.Add("/second.json", FileTypeConfig, "test")

	recent := m.GetMostRecent(FileTypeConfig)
	if recent == nil {
		t.Fatal("expected a recent file")
	}
	if recent.Name != "second.json" {
		t.Errorf("expected 'second.json' to be most recent, got '%s'", recent.Name)
	}
}

func TestManagerCount(t *testing.T) {
	m := NewManager(nil)

	m.Add("/config.json", FileTypeConfig, "test")
	m.Add("/export.csv", FileTypeExport, "test")
	m.Add("/project.proj", FileTypeProject, "test")

	if m.Count(FileTypeAny) != 3 {
		t.Errorf("expected count 3 for Any, got %d", m.Count(FileTypeAny))
	}
	if m.Count(FileTypeConfig) != 1 {
		t.Errorf("expected count 1 for Config, got %d", m.Count(FileTypeConfig))
	}
}

func TestManagerEnforcesLimits(t *testing.T) {
	m := NewManager(nil)

	// Add more than maxRecentFiles of one type
	for i := 0; i < 25; i++ {
		m.AddConfig("/config"+string(rune('a'+i))+".json", "test")
	}

	configs := m.ListConfigs()
	if len(configs) > maxRecentFiles {
		t.Errorf("expected at most %d configs, got %d", maxRecentFiles, len(configs))
	}
}

func TestManagerCleanupMissing(t *testing.T) {
	m := NewManager(nil)

	// Create a temp file that exists
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.json")
	os.WriteFile(existingFile, []byte("{}"), 0644)

	m.Add(existingFile, FileTypeConfig, "test")
	m.Add("/nonexistent/file.json", FileTypeConfig, "test")

	removed := m.CleanupMissing()
	if removed != 1 {
		t.Errorf("expected 1 file removed, got %d", removed)
	}

	files := m.ListAll()
	if len(files) != 1 {
		t.Errorf("expected 1 file remaining, got %d", len(files))
	}
	if files[0].Path != existingFile {
		t.Errorf("expected existing file to remain")
	}
}

func TestManagerOnChange(t *testing.T) {
	m := NewManager(nil)

	called := make(chan bool, 1)
	m.OnChange(func(files []*RecentFile) {
		called <- true
	})

	m.Add("/test.json", FileTypeConfig, "test")

	select {
	case <-called:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("OnChange callback not called")
	}
}

func TestFormatAccessTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		time     time.Time
		expected string
	}{
		{now.Add(-30 * time.Second), "Just now"},
		{now.Add(-5 * time.Minute), "5 minutes ago"},
		{now.Add(-1 * time.Hour), "1 hour ago"},
		{now.Add(-3 * time.Hour), "3 hours ago"},
		{now.Add(-25 * time.Hour), "Yesterday"},
		{now.Add(-3 * 24 * time.Hour), "3 days ago"},
		{now.Add(-10 * 24 * time.Hour), "1 week ago"},
		{now.Add(-21 * 24 * time.Hour), "3 weeks ago"},
	}

	for _, tc := range tests {
		result := FormatAccessTime(tc.time)
		if result != tc.expected {
			t.Errorf("FormatAccessTime(%v) = '%s', expected '%s'", tc.time, result, tc.expected)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1 KB"},
		{1536, "1.5 KB"},
		{1048576, "1 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1 GB"},
	}

	for _, tc := range tests {
		result := FormatSize(tc.bytes)
		if result != tc.expected {
			t.Errorf("FormatSize(%d) = '%s', expected '%s'", tc.bytes, result, tc.expected)
		}
	}
}

func TestConvenienceFunctions(t *testing.T) {
	m := NewManager(nil)

	m.AddConfig("/config.json", "hysteresis")
	m.AddExport("/export.csv", "crossbar")
	m.AddProject("/project.proj", "mnist")
	m.AddPreset("/preset.json", "circuits")

	if m.Count(FileTypeConfig) != 1 {
		t.Error("AddConfig failed")
	}
	if m.Count(FileTypeExport) != 1 {
		t.Error("AddExport failed")
	}
	if m.Count(FileTypeProject) != 1 {
		t.Error("AddProject failed")
	}
	if m.Count(FileTypePreset) != 1 {
		t.Error("AddPreset failed")
	}
}

func TestGlobalManager(t *testing.T) {
	// Clear any existing global manager
	SetGlobalManager(nil)

	// Should not panic with nil manager
	GlobalAddConfig("/test.json", "test")

	// Initialize global
	m := InitGlobal(nil)
	if m == nil {
		t.Fatal("InitGlobal returned nil")
	}

	if GetGlobalManager() != m {
		t.Error("GetGlobalManager did not return initialized manager")
	}

	// Global functions should work now
	GlobalAddConfig("/config.json", "test")
	GlobalAddExport("/export.csv", "test")
	GlobalAddProject("/project.proj", "test")
	GlobalAddPreset("/preset.json", "test")

	if m.Count(FileTypeAny) != 4 {
		t.Errorf("expected 4 files, got %d", m.Count(FileTypeAny))
	}

	// Cleanup
	SetGlobalManager(nil)
}
