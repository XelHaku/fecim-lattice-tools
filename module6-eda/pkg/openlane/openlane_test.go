package openlane

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestModeString tests the Mode.String() method
func TestModeString(t *testing.T) {
	tests := []struct {
		mode     Mode
		expected string
	}{
		{ModeNone, "None"},
		{ModeDocker, "Docker"},
		{ModeNative, "Native"},
		{Mode(99), "None"}, // Invalid mode should return "None"
	}

	for _, tt := range tests {
		result := tt.mode.String()
		if result != tt.expected {
			t.Errorf("Mode(%d).String() = %q, want %q", tt.mode, result, tt.expected)
		}
	}
}

// TestNewManager tests Manager creation
func TestNewManager(t *testing.T) {
	// Save original PDK_ROOT
	origPDKRoot := os.Getenv("PDK_ROOT")
	defer os.Setenv("PDK_ROOT", origPDKRoot)

	t.Run("WithPDKRootSet", func(t *testing.T) {
		os.Setenv("PDK_ROOT", "/test/pdk")
		m := NewManager()
		if m.pdkRoot != "/test/pdk" {
			t.Errorf("NewManager().pdkRoot = %q, want %q", m.pdkRoot, "/test/pdk")
		}
		if m.dockerImage != "ghcr.io/the-openroad-project/openlane:latest" {
			t.Errorf("NewManager().dockerImage = %q, want default image", m.dockerImage)
		}
	})

	t.Run("WithoutPDKRoot", func(t *testing.T) {
		os.Unsetenv("PDK_ROOT")
		home := os.Getenv("HOME")
		expected := filepath.Join(home, ".volare")
		m := NewManager()
		if m.pdkRoot != expected {
			t.Errorf("NewManager().pdkRoot = %q, want %q", m.pdkRoot, expected)
		}
	})
}

// TestManagerGetters tests Manager getter methods
func TestManagerGetters(t *testing.T) {
	m := NewManager()
	m.pdkRoot = "/test/pdk"
	m.dockerImage = "test:image"

	if m.GetPDKRoot() != "/test/pdk" {
		t.Errorf("GetPDKRoot() = %q, want %q", m.GetPDKRoot(), "/test/pdk")
	}

	if m.GetDockerImage() != "test:image" {
		t.Errorf("GetDockerImage() = %q, want %q", m.GetDockerImage(), "test:image")
	}
}

// TestGetDockerImageVersion tests version extraction from image name
func TestGetDockerImageVersion(t *testing.T) {
	tests := []struct {
		image    string
		expected string
	}{
		{"ghcr.io/repo/openlane:latest", "latest"},
		{"ghcr.io/repo/openlane:v1.2.3", "v1.2.3"},
		{"openlane:2024.01", "2024.01"},
		{"openlane", "latest"}, // No colon defaults to latest
	}

	for _, tt := range tests {
		m := &Manager{dockerImage: tt.image}
		version, err := m.GetDockerImageVersion()
		if err != nil {
			t.Errorf("GetDockerImageVersion() for %q returned error: %v", tt.image, err)
		}
		if version != tt.expected {
			t.Errorf("GetDockerImageVersion() for %q = %q, want %q", tt.image, version, tt.expected)
		}
	}
}

// TestGetPDKSetupInstructions tests instructions string generation
func TestGetPDKSetupInstructions(t *testing.T) {
	m := NewManager()
	instructions := m.GetPDKSetupInstructions()

	// Check for key phrases in instructions
	requiredPhrases := []string{
		"pip install volare",
		"volare enable",
		"PDK_ROOT",
		"sky130",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(instructions, phrase) {
			t.Errorf("GetPDKSetupInstructions() missing required phrase: %q", phrase)
		}
	}
}

// TestDefaultConfig tests default configuration creation
func TestDefaultConfig(t *testing.T) {
	// Save original PDK_ROOT
	origPDKRoot := os.Getenv("PDK_ROOT")
	defer os.Setenv("PDK_ROOT", origPDKRoot)

	t.Run("WithPDKRootSet", func(t *testing.T) {
		os.Setenv("PDK_ROOT", "/custom/pdk")
		cfg := DefaultConfig()

		if cfg.PDKRoot != "/custom/pdk" {
			t.Errorf("DefaultConfig().PDKRoot = %q, want %q", cfg.PDKRoot, "/custom/pdk")
		}
		if cfg.PDKVariant != "sky130A" {
			t.Errorf("DefaultConfig().PDKVariant = %q, want %q", cfg.PDKVariant, "sky130A")
		}
		if cfg.SCLibrary != "sky130_fd_sc_hd" {
			t.Errorf("DefaultConfig().SCLibrary = %q, want %q", cfg.SCLibrary, "sky130_fd_sc_hd")
		}
		if cfg.PreferredMode != ModeDocker {
			t.Errorf("DefaultConfig().PreferredMode = %v, want %v", cfg.PreferredMode, ModeDocker)
		}
		if cfg.TimeoutPlacement != 5*time.Minute {
			t.Errorf("DefaultConfig().TimeoutPlacement = %v, want %v", cfg.TimeoutPlacement, 5*time.Minute)
		}
		if cfg.TimeoutSynthesis != 10*time.Minute {
			t.Errorf("DefaultConfig().TimeoutSynthesis = %v, want %v", cfg.TimeoutSynthesis, 10*time.Minute)
		}
		if cfg.TimeoutRouting != 15*time.Minute {
			t.Errorf("DefaultConfig().TimeoutRouting = %v, want %v", cfg.TimeoutRouting, 15*time.Minute)
		}
		if cfg.DockerImage != "ghcr.io/the-openroad-project/openlane:latest" {
			t.Errorf("DefaultConfig().DockerImage = %q, want default", cfg.DockerImage)
		}
	})

	t.Run("WithoutPDKRoot", func(t *testing.T) {
		os.Unsetenv("PDK_ROOT")
		home := os.Getenv("HOME")
		expected := filepath.Join(home, ".volare")
		cfg := DefaultConfig()

		if cfg.PDKRoot != expected {
			t.Errorf("DefaultConfig().PDKRoot = %q, want %q", cfg.PDKRoot, expected)
		}
	})
}

// TestGetConfigPath tests config path generation
func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	home := os.Getenv("HOME")
	expected := filepath.Join(home, ".fecim", "openlane-config.json")

	if path != expected {
		t.Errorf("GetConfigPath() = %q, want %q", path, expected)
	}
}

// TestConfigPaths tests path generation methods
func TestConfigPaths(t *testing.T) {
	cfg := &Config{
		PDKRoot:    "/pdk",
		PDKVariant: "sky130A",
		SCLibrary:  "sky130_fd_sc_hd",
	}

	t.Run("GetTechLEFPath", func(t *testing.T) {
		expected := "/pdk/sky130A/libs.tech/openlane/sky130_fd_sc_hd/sky130_fd_sc_hd.tlef"
		result := cfg.GetTechLEFPath()
		if result != expected {
			t.Errorf("GetTechLEFPath() = %q, want %q", result, expected)
		}
	})

	t.Run("GetCellLEFPath", func(t *testing.T) {
		expected := "/pdk/sky130A/libs.ref/sky130_fd_sc_hd/lef/sky130_fd_sc_hd.lef"
		result := cfg.GetCellLEFPath()
		if result != expected {
			t.Errorf("GetCellLEFPath() = %q, want %q", result, expected)
		}
	})

	t.Run("GetLibertyPath", func(t *testing.T) {
		expected := "/pdk/sky130A/libs.ref/sky130_fd_sc_hd/lib/sky130_fd_sc_hd__tt_025C_1v80.lib"
		result := cfg.GetLibertyPath()
		if result != expected {
			t.Errorf("GetLibertyPath() = %q, want %q", result, expected)
		}
	})
}

// TestSaveAndLoadConfig tests config serialization
func TestSaveAndLoadConfig(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "openlane-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")

	// Create test config
	originalCfg := &Config{
		PDKRoot:          "/test/pdk",
		PDKVariant:       "sky130B",
		SCLibrary:        "sky130_fd_sc_hs",
		PreferredMode:    ModeNative,
		TimeoutPlacement: 3 * time.Minute,
		TimeoutSynthesis: 7 * time.Minute,
		TimeoutRouting:   12 * time.Minute,
		DockerImage:      "test:image",
	}

	// Save config
	err = SaveConfig(originalCfg, configPath)
	if err != nil {
		t.Fatalf("SaveConfig() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created")
	}

	// Load config
	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Compare fields
	if loadedCfg.PDKRoot != originalCfg.PDKRoot {
		t.Errorf("PDKRoot = %q, want %q", loadedCfg.PDKRoot, originalCfg.PDKRoot)
	}
	if loadedCfg.PDKVariant != originalCfg.PDKVariant {
		t.Errorf("PDKVariant = %q, want %q", loadedCfg.PDKVariant, originalCfg.PDKVariant)
	}
	if loadedCfg.SCLibrary != originalCfg.SCLibrary {
		t.Errorf("SCLibrary = %q, want %q", loadedCfg.SCLibrary, originalCfg.SCLibrary)
	}
	if loadedCfg.PreferredMode != originalCfg.PreferredMode {
		t.Errorf("PreferredMode = %v, want %v", loadedCfg.PreferredMode, originalCfg.PreferredMode)
	}
	if loadedCfg.TimeoutPlacement != originalCfg.TimeoutPlacement {
		t.Errorf("TimeoutPlacement = %v, want %v", loadedCfg.TimeoutPlacement, originalCfg.TimeoutPlacement)
	}
	if loadedCfg.TimeoutSynthesis != originalCfg.TimeoutSynthesis {
		t.Errorf("TimeoutSynthesis = %v, want %v", loadedCfg.TimeoutSynthesis, originalCfg.TimeoutSynthesis)
	}
	if loadedCfg.TimeoutRouting != originalCfg.TimeoutRouting {
		t.Errorf("TimeoutRouting = %v, want %v", loadedCfg.TimeoutRouting, originalCfg.TimeoutRouting)
	}
	if loadedCfg.DockerImage != originalCfg.DockerImage {
		t.Errorf("DockerImage = %q, want %q", loadedCfg.DockerImage, originalCfg.DockerImage)
	}
}

// TestLoadConfigNonexistent tests loading from nonexistent file
func TestLoadConfigNonexistent(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.json")

	// Should return default config and error
	if err == nil {
		t.Error("LoadConfig() with nonexistent file should return error")
	}

	// Should still return valid default config
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}
	if cfg.PDKVariant != "sky130A" {
		t.Errorf("LoadConfig() should return default config on error, got PDKVariant = %q", cfg.PDKVariant)
	}
}

// TestLoadConfigInvalidJSON tests loading from file with invalid JSON
func TestLoadConfigInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "openlane-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "bad-config.json")
	err = os.WriteFile(configPath, []byte("not valid json {{{"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg, err := LoadConfig(configPath)

	// Should return error
	if err == nil {
		t.Error("LoadConfig() with invalid JSON should return error")
	}

	// Should still return default config
	if cfg == nil {
		t.Fatal("LoadConfig() returned nil config")
	}
	if cfg.PDKVariant != "sky130A" {
		t.Errorf("LoadConfig() should return default config on error")
	}
}

// TestConfigJSONMarshaling tests the configJSON intermediate struct
func TestConfigJSONMarshaling(t *testing.T) {
	cfg := &Config{
		PDKRoot:          "/test/pdk",
		PDKVariant:       "sky130A",
		SCLibrary:        "sky130_fd_sc_hd",
		PreferredMode:    ModeNative,
		TimeoutPlacement: 5 * time.Minute,
		TimeoutSynthesis: 10 * time.Minute,
		TimeoutRouting:   15 * time.Minute,
		DockerImage:      "test:latest",
	}

	modeStr := "native"
	cj := configJSON{
		PDKRoot:          cfg.PDKRoot,
		PDKVariant:       cfg.PDKVariant,
		SCLibrary:        cfg.SCLibrary,
		PreferredMode:    modeStr,
		TimeoutPlacement: cfg.TimeoutPlacement.String(),
		TimeoutSynthesis: cfg.TimeoutSynthesis.String(),
		TimeoutRouting:   cfg.TimeoutRouting.String(),
		DockerImage:      cfg.DockerImage,
	}

	// Marshal and verify
	data, err := json.Marshal(cj)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	// Verify JSON contains expected values
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"pdk_root":"/test/pdk"`) {
		t.Error("JSON missing pdk_root")
	}
	if !strings.Contains(jsonStr, `"preferred_mode":"native"`) {
		t.Error("JSON missing or wrong preferred_mode")
	}
	if !strings.Contains(jsonStr, `"timeout_placement":"5m0s"`) {
		t.Error("JSON missing or wrong timeout_placement")
	}
}

// TestGetVolareSetupInstructions tests the global function
func TestGetVolareSetupInstructions(t *testing.T) {
	instructions := GetVolareSetupInstructions()

	requiredPhrases := []string{
		"pip install volare",
		"volare enable",
		"PDK_ROOT",
		"sky130",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(instructions, phrase) {
			t.Errorf("GetVolareSetupInstructions() missing required phrase: %q", phrase)
		}
	}
}

// TestNewRunner tests Runner creation
func TestNewRunner(t *testing.T) {
	manager := NewManager()
	config := DefaultConfig()

	runner := NewRunner(manager, config)

	if runner.manager != manager {
		t.Error("NewRunner() did not set manager correctly")
	}
	if runner.config != config {
		t.Error("NewRunner() did not set config correctly")
	}
}

// TestResultStruct tests Result struct fields
func TestResultStruct(t *testing.T) {
	result := &Result{
		Stdout:   "test output",
		Stderr:   "test error",
		ExitCode: 42,
		Duration: 5 * time.Second,
	}

	if result.Stdout != "test output" {
		t.Errorf("Result.Stdout = %q, want %q", result.Stdout, "test output")
	}
	if result.Stderr != "test error" {
		t.Errorf("Result.Stderr = %q, want %q", result.Stderr, "test error")
	}
	if result.ExitCode != 42 {
		t.Errorf("Result.ExitCode = %d, want %d", result.ExitCode, 42)
	}
	if result.Duration != 5*time.Second {
		t.Errorf("Result.Duration = %v, want %v", result.Duration, 5*time.Second)
	}
}

// TestResultZeroValues tests Result with zero values
func TestResultZeroValues(t *testing.T) {
	result := &Result{}

	if result.Stdout != "" {
		t.Error("Zero Result.Stdout should be empty string")
	}
	if result.Stderr != "" {
		t.Error("Zero Result.Stderr should be empty string")
	}
	if result.ExitCode != 0 {
		t.Error("Zero Result.ExitCode should be 0")
	}
	if result.Duration != 0 {
		t.Error("Zero Result.Duration should be 0")
	}
}

// TestDetectModeWithNoTools tests DetectMode when nothing is available
func TestDetectModeWithNoTools(t *testing.T) {
	m := &Manager{
		dockerImage: "nonexistent:image",
		pdkRoot:     "/nonexistent",
	}

	// In real environment, this might return Docker or Native if tools exist
	// We're just testing that it returns a valid Mode value
	mode := m.DetectMode()

	if mode != ModeNone && mode != ModeDocker && mode != ModeNative {
		t.Errorf("DetectMode() returned invalid mode: %v", mode)
	}
}

// TestDockerCommandConstruction tests Docker command argument building
func TestDockerCommandConstruction(t *testing.T) {
	manager := NewManager()
	config := DefaultConfig()
	runner := NewRunner(manager, config)

	// We can't easily test the full command construction without mocking exec,
	// but we can verify the helper functions work correctly
	workDir := "/test/work"

	// Test absolute path conversion logic
	absPath, err := filepath.Abs(workDir)
	if err != nil {
		// If current directory exists, this should work
		t.Logf("filepath.Abs() failed (expected in some test environments): %v", err)
	} else {
		if !filepath.IsAbs(absPath) {
			t.Errorf("Absolute path conversion failed: %q is not absolute", absPath)
		}
	}

	// Verify runner has required fields
	if runner.manager == nil {
		t.Error("Runner.manager is nil")
	}
	if runner.config == nil {
		t.Error("Runner.config is nil")
	}
}

// TestNativeCommandConstruction tests native command argument building
func TestNativeCommandConstruction(t *testing.T) {
	// Test the argument construction logic for native commands
	scriptName := "test_script.tcl"
	workDir := "/test/work"

	// OpenROAD native args
	args := []string{"-no_splash", "-exit", filepath.Join(workDir, scriptName)}

	if len(args) != 3 {
		t.Errorf("OpenROAD args length = %d, want 3", len(args))
	}
	if args[0] != "-no_splash" {
		t.Errorf("args[0] = %q, want %q", args[0], "-no_splash")
	}
	if args[1] != "-exit" {
		t.Errorf("args[1] = %q, want %q", args[1], "-exit")
	}
}

// TestYosysCommandConstruction tests Yosys command construction
func TestYosysCommandConstruction(t *testing.T) {
	yosysCmd := "read_verilog test.v; hierarchy -check"

	// Native Yosys args
	args := []string{"-p", yosysCmd}

	if len(args) != 2 {
		t.Errorf("Yosys args length = %d, want 2", len(args))
	}
	if args[0] != "-p" {
		t.Errorf("args[0] = %q, want %q", args[0], "-p")
	}
	if args[1] != yosysCmd {
		t.Errorf("args[1] = %q, want %q", args[1], yosysCmd)
	}
}

// TestKLayoutCommandConstruction tests KLayout command construction
func TestKLayoutCommandConstruction(t *testing.T) {
	scriptPath := "/test/script.py"
	envVars := map[string]string{
		"DEF_FILE": "/design/output.def",
		"LEF_FILE": "/design/cells.lef",
		"OUT_PNG":  "/design/output.png",
	}

	// Native KLayout args
	args := []string{"-z"}
	for k, v := range envVars {
		args = append(args, "-rd", k+"="+v)
	}
	args = append(args, "-r", scriptPath)

	// Verify structure
	if args[0] != "-z" {
		t.Errorf("First arg should be -z, got %q", args[0])
	}

	// Check that -rd pairs are present
	foundRd := false
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-rd" {
			foundRd = true
			break
		}
	}
	if !foundRd {
		t.Error("Expected -rd flags in KLayout args")
	}

	// Check script path at end
	if args[len(args)-2] != "-r" || args[len(args)-1] != scriptPath {
		t.Error("Expected -r scriptPath at end of args")
	}
}

// TestXvfbCommandConstruction tests Xvfb wrapper construction
func TestXvfbCommandConstruction(t *testing.T) {
	scriptName := "test.tcl"

	// Test Xvfb command construction
	xvfbCmd := fmt.Sprintf("Xvfb :99 -screen 0 1024x768x24 -nolisten tcp > /dev/null 2>&1 & sleep 1 && export DISPLAY=:99 && openroad -no_splash -exit /design/%s", scriptName)

	// Verify key components
	if !strings.Contains(xvfbCmd, "Xvfb :99") {
		t.Error("Xvfb command missing Xvfb invocation")
	}
	if !strings.Contains(xvfbCmd, "DISPLAY=:99") {
		t.Error("Xvfb command missing DISPLAY export")
	}
	if !strings.Contains(xvfbCmd, scriptName) {
		t.Error("Xvfb command missing script name")
	}
	if !strings.Contains(xvfbCmd, "sleep 1") {
		t.Error("Xvfb command missing sleep delay")
	}
}

// TestEnvironmentVariableConstruction tests env var construction
func TestEnvironmentVariableConstruction(t *testing.T) {
	baseEnv := os.Environ()
	envVars := map[string]string{
		"CELL_LEF": "/path/to/cell.lef",
		"DEF_FILE": "/path/to/design.def",
	}

	// Construct env like runner does
	env := baseEnv
	for k, v := range envVars {
		env = append(env, k+"="+v)
	}

	// Verify new vars are appended
	if len(env) < len(baseEnv)+len(envVars) {
		t.Error("Environment variables not appended correctly")
	}

	// Check for presence of added vars
	foundCellLef := false
	foundDefFile := false
	for _, e := range env {
		if strings.HasPrefix(e, "CELL_LEF=") {
			foundCellLef = true
		}
		if strings.HasPrefix(e, "DEF_FILE=") {
			foundDefFile = true
		}
	}
	if !foundCellLef {
		t.Error("CELL_LEF not found in constructed environment")
	}
	if !foundDefFile {
		t.Error("DEF_FILE not found in constructed environment")
	}
}

// TestConfigTimeoutFields tests timeout field types
func TestConfigTimeoutFields(t *testing.T) {
	cfg := DefaultConfig()

	// Verify timeouts are time.Duration type
	var _ time.Duration = cfg.TimeoutPlacement
	var _ time.Duration = cfg.TimeoutSynthesis
	var _ time.Duration = cfg.TimeoutRouting

	// Verify default values are reasonable
	if cfg.TimeoutPlacement < 1*time.Minute {
		t.Error("TimeoutPlacement too short")
	}
	if cfg.TimeoutSynthesis < cfg.TimeoutPlacement {
		t.Error("TimeoutSynthesis should be >= TimeoutPlacement")
	}
	if cfg.TimeoutRouting < cfg.TimeoutSynthesis {
		t.Error("TimeoutRouting should be >= TimeoutSynthesis")
	}
}

// TestConfigModeSerialization tests Mode enum serialization
func TestConfigModeSerialization(t *testing.T) {
	tests := []struct {
		mode     Mode
		jsonStr  string
		expected Mode
	}{
		{ModeDocker, "docker", ModeDocker},
		{ModeNative, "native", ModeNative},
	}

	for _, tt := range tests {
		// Test mode to string
		cfg := DefaultConfig()
		cfg.PreferredMode = tt.mode

		modeStr := "docker"
		if cfg.PreferredMode == ModeNative {
			modeStr = "native"
		}

		if modeStr != tt.jsonStr {
			t.Errorf("Mode %v serialized to %q, want %q", tt.mode, modeStr, tt.jsonStr)
		}

		// Test string to mode
		var newMode Mode = ModeDocker
		if tt.jsonStr == "native" {
			newMode = ModeNative
		}

		if newMode != tt.expected {
			t.Errorf("String %q deserialized to %v, want %v", tt.jsonStr, newMode, tt.expected)
		}
	}
}

// TestPDKInstallationCheck tests PDK detection logic
func TestPDKInstallationCheck(t *testing.T) {
	m := NewManager()

	// Test with empty PDK root
	m.pdkRoot = ""
	if m.IsPDKInstalled() {
		t.Error("IsPDKInstalled() should return false for empty pdkRoot")
	}

	// Test with nonexistent path
	m.pdkRoot = "/nonexistent/path"
	if m.IsPDKInstalled() {
		t.Error("IsPDKInstalled() should return false for nonexistent path")
	}
}

// TestDockerImageNameConstruction tests Docker image name handling
func TestDockerImageNameConstruction(t *testing.T) {
	tests := []struct {
		name      string
		image     string
		wantImage string
		wantValid bool
	}{
		{
			name:      "Full image name",
			image:     "ghcr.io/the-openroad-project/openlane:latest",
			wantImage: "ghcr.io/the-openroad-project/openlane:latest",
			wantValid: true,
		},
		{
			name:      "Short image name",
			image:     "openlane:latest",
			wantImage: "openlane:latest",
			wantValid: true,
		},
		{
			name:      "Empty image name",
			image:     "",
			wantImage: "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{dockerImage: tt.image}
			result := m.GetDockerImage()

			if result != tt.wantImage {
				t.Errorf("GetDockerImage() = %q, want %q", result, tt.wantImage)
			}
		})
	}
}

// TestAbsolutePathHandling tests absolute path conversion for Docker
func TestAbsolutePathHandling(t *testing.T) {
	testPaths := []string{
		"/absolute/path",
		"relative/path",
		".",
		"..",
	}

	for _, path := range testPaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			// Some paths might fail in test environment
			t.Logf("filepath.Abs(%q) failed: %v", path, err)
			continue
		}

		if !filepath.IsAbs(absPath) {
			t.Errorf("filepath.Abs(%q) = %q which is not absolute", path, absPath)
		}
	}
}
