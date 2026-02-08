package cli

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestCommonFlags_Register(t *testing.T) {
	cf := NewCommonFlags()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cf.Register(fs)

	// Parse test arguments
	args := []string{"--json", "--quiet", "--config", "test.yaml", "--batch", "batch.txt", "--output", "out.txt"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("failed to parse args: %v", err)
	}

	if !cf.JSON {
		t.Error("expected JSON flag to be true")
	}
	if !cf.Quiet {
		t.Error("expected Quiet flag to be true")
	}
	if cf.Config != "test.yaml" {
		t.Errorf("expected Config to be 'test.yaml', got %q", cf.Config)
	}
	if cf.Batch != "batch.txt" {
		t.Errorf("expected Batch to be 'batch.txt', got %q", cf.Batch)
	}
	if cf.Output != "out.txt" {
		t.Errorf("expected Output to be 'out.txt', got %q", cf.Output)
	}
}

func TestCommonFlags_Shorthands(t *testing.T) {
	cf := NewCommonFlags()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cf.Register(fs)

	args := []string{"-q", "-c", "config.json", "-o", "output.txt", "-h"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("failed to parse args: %v", err)
	}

	if !cf.Quiet {
		t.Error("expected Quiet flag (via -q) to be true")
	}
	if cf.Config != "config.json" {
		t.Errorf("expected Config (via -c) to be 'config.json', got %q", cf.Config)
	}
	if cf.Output != "output.txt" {
		t.Errorf("expected Output (via -o) to be 'output.txt', got %q", cf.Output)
	}
	if !cf.WantsHelp() {
		t.Error("expected WantsHelp() to be true")
	}
}

func TestOutputWriter_Print(t *testing.T) {
	tests := []struct {
		name     string
		quiet    bool
		json     bool
		expected string
	}{
		{"normal", false, false, "Hello World\n"},
		{"quiet", true, false, ""},
		{"json", false, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			flags := &CommonFlags{Quiet: tt.quiet, JSON: tt.json}
			ow := &OutputWriter{flags: flags, writer: &buf}

			ow.Println("Hello World")

			if buf.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, buf.String())
			}
		})
	}
}

func TestOutputWriter_Result_JSON(t *testing.T) {
	var buf bytes.Buffer
	flags := &CommonFlags{JSON: true}
	ow := &OutputWriter{flags: flags, writer: &buf}

	data := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}

	if err := ow.Result(data); err != nil {
		t.Fatalf("Result() failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if parsed["name"] != "test" {
		t.Errorf("expected name to be 'test', got %v", parsed["name"])
	}
}

func TestOutputWriter_BatchResults(t *testing.T) {
	var buf bytes.Buffer
	flags := &CommonFlags{JSON: true}
	ow := &OutputWriter{flags: flags, writer: &buf, results: make([]interface{}, 0)}

	ow.AddResult(map[string]int{"value": 1})
	ow.AddResult(map[string]int{"value": 2})
	ow.AddResult(map[string]int{"value": 3})

	if err := ow.FlushResults(); err != nil {
		t.Fatalf("FlushResults() failed: %v", err)
	}

	var parsed []map[string]int
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(parsed) != 3 {
		t.Errorf("expected 3 results, got %d", len(parsed))
	}
}

func TestConfigLoader_YAML(t *testing.T) {
	// Create temp YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
name: test
size: 64
enabled: true
levels:
  - 8
  - 16
  - 32
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	type TestConfig struct {
		Name    string `yaml:"name"`
		Size    int    `yaml:"size"`
		Enabled bool   `yaml:"enabled"`
		Levels  []int  `yaml:"levels"`
	}

	var cfg TestConfig
	loader := NewConfigLoader(configPath)
	if err := loader.Load(&cfg); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Name != "test" {
		t.Errorf("expected Name to be 'test', got %q", cfg.Name)
	}
	if cfg.Size != 64 {
		t.Errorf("expected Size to be 64, got %d", cfg.Size)
	}
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if len(cfg.Levels) != 3 {
		t.Errorf("expected 3 levels, got %d", len(cfg.Levels))
	}
}

func TestConfigLoader_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	jsonContent := `{
  "name": "test",
  "size": 128,
  "options": ["a", "b", "c"]
}`
	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	type TestConfig struct {
		Name    string   `json:"name"`
		Size    int      `json:"size"`
		Options []string `json:"options"`
	}

	var cfg TestConfig
	loader := NewConfigLoader(configPath)
	if err := loader.Load(&cfg); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Name != "test" {
		t.Errorf("expected Name to be 'test', got %q", cfg.Name)
	}
	if cfg.Size != 128 {
		t.Errorf("expected Size to be 128, got %d", cfg.Size)
	}
}

func TestBatchProcessor_Lines(t *testing.T) {
	tmpDir := t.TempDir()
	batchPath := filepath.Join(tmpDir, "batch.txt")

	batchContent := `item1
item2
# comment line
item3

item4
`
	if err := os.WriteFile(batchPath, []byte(batchContent), 0644); err != nil {
		t.Fatalf("failed to write batch file: %v", err)
	}

	bp, err := NewBatchProcessor(batchPath)
	if err != nil {
		t.Fatalf("NewBatchProcessor() failed: %v", err)
	}

	items := bp.Items()
	if len(items) != 4 {
		t.Errorf("expected 4 items, got %d: %v", len(items), items)
	}

	expected := []string{"item1", "item2", "item3", "item4"}
	for i, exp := range expected {
		if items[i] != exp {
			t.Errorf("item[%d]: expected %q, got %q", i, exp, items[i])
		}
	}
}

func TestBatchProcessor_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	batchPath := filepath.Join(tmpDir, "batch.json")

	batchContent := `["item1", "item2", "item3"]`
	if err := os.WriteFile(batchPath, []byte(batchContent), 0644); err != nil {
		t.Fatalf("failed to write batch file: %v", err)
	}

	bp, err := NewBatchProcessor(batchPath)
	if err != nil {
		t.Fatalf("NewBatchProcessor() failed: %v", err)
	}

	if bp.Count() != 3 {
		t.Errorf("expected 3 items, got %d", bp.Count())
	}
}

func TestBatchResult(t *testing.T) {
	br := NewBatchResult()

	br.Add(NewResult("success1"))
	br.Add(NewResult("success2"))
	br.Add(NewErrorResult(os.ErrNotExist))

	if br.Total != 3 {
		t.Errorf("expected Total to be 3, got %d", br.Total)
	}
	if br.Succeeded != 2 {
		t.Errorf("expected Succeeded to be 2, got %d", br.Succeeded)
	}
	if br.Failed != 1 {
		t.Errorf("expected Failed to be 1, got %d", br.Failed)
	}
}

func TestNewResult(t *testing.T) {
	r := NewResult(map[string]int{"value": 42})
	if !r.Success {
		t.Error("expected Success to be true")
	}
	if r.Error != "" {
		t.Error("expected Error to be empty")
	}
}

func TestNewErrorResult(t *testing.T) {
	r := NewErrorResult(os.ErrNotExist)
	if r.Success {
		t.Error("expected Success to be false")
	}
	if r.Error == "" {
		t.Error("expected Error to be non-empty")
	}
}
