// Package cli provides shared CLI utilities for FeCIM commands.
// It adds common flags (--json, --quiet, --config) and batch processing support.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CommonFlags contains the standard CLI flags used across all commands.
type CommonFlags struct {
	JSON     bool   // Output results as JSON
	Quiet    bool   // Suppress non-essential output
	Config   string // Config file path (YAML or JSON)
	Batch    string // Batch input file (one item per line)
	Output   string // Output file (default: stdout)
	Help     bool   // Show help
	HelpFlag bool   // -h shorthand
}

// OutputWriter manages output formatting based on flags.
type OutputWriter struct {
	flags   *CommonFlags
	writer  io.Writer
	results []interface{} // Collected results for JSON batch output
}

// ConfigLoader handles config file parsing.
type ConfigLoader struct {
	configPath string
}

// NewCommonFlags creates a new CommonFlags with defaults.
func NewCommonFlags() *CommonFlags {
	return &CommonFlags{}
}

// Register adds the common flags to a flag set.
func (cf *CommonFlags) Register(fs *flag.FlagSet) {
	fs.BoolVar(&cf.JSON, "json", false, "Output results as JSON")
	fs.BoolVar(&cf.Quiet, "quiet", false, "Suppress informational output (only show results/errors)")
	fs.BoolVar(&cf.Quiet, "q", false, "Suppress informational output (shorthand)")
	fs.StringVar(&cf.Config, "config", "", "Load configuration from YAML or JSON file")
	fs.StringVar(&cf.Config, "c", "", "Load configuration from file (shorthand)")
	fs.StringVar(&cf.Batch, "batch", "", "Batch input file (one item per line, or JSON array)")
	fs.StringVar(&cf.Output, "output", "", "Output file (default: stdout)")
	fs.StringVar(&cf.Output, "o", "", "Output file (shorthand)")
	fs.BoolVar(&cf.Help, "help", false, "Show help")
	fs.BoolVar(&cf.HelpFlag, "h", false, "Show help (shorthand)")
}

// WantsHelp returns true if help was requested.
func (cf *CommonFlags) WantsHelp() bool {
	return cf.Help || cf.HelpFlag
}

// NewOutputWriter creates an output writer based on flags.
func NewOutputWriter(flags *CommonFlags) (*OutputWriter, error) {
	var w io.Writer = os.Stdout

	if flags.Output != "" {
		f, err := os.Create(flags.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to create output file: %w", err)
		}
		w = f
	}

	return &OutputWriter{
		flags:   flags,
		writer:  w,
		results: make([]interface{}, 0),
	}, nil
}

// Close closes the output writer if it's a file.
func (ow *OutputWriter) Close() error {
	if f, ok := ow.writer.(*os.File); ok && f != os.Stdout {
		return f.Close()
	}
	return nil
}

// Print outputs text (suppressed in quiet mode).
func (ow *OutputWriter) Print(format string, args ...interface{}) {
	if ow.flags.Quiet || ow.flags.JSON {
		return
	}
	fmt.Fprintf(ow.writer, format, args...)
}

// Println outputs a line (suppressed in quiet mode).
func (ow *OutputWriter) Println(args ...interface{}) {
	if ow.flags.Quiet || ow.flags.JSON {
		return
	}
	fmt.Fprintln(ow.writer, args...)
}

// PrintAlways outputs text regardless of quiet mode.
func (ow *OutputWriter) PrintAlways(format string, args ...interface{}) {
	if ow.flags.JSON {
		return // Suppress in JSON mode
	}
	fmt.Fprintf(ow.writer, format, args...)
}

// Result outputs a single result (always shown, formatted based on flags).
func (ow *OutputWriter) Result(result interface{}) error {
	if ow.flags.JSON {
		return ow.outputJSON(result)
	}
	// For non-JSON, print the result directly
	fmt.Fprintln(ow.writer, result)
	return nil
}

// AddResult adds a result to the batch collection for later output.
func (ow *OutputWriter) AddResult(result interface{}) {
	ow.results = append(ow.results, result)
}

// FlushResults outputs all collected batch results.
func (ow *OutputWriter) FlushResults() error {
	if len(ow.results) == 0 {
		return nil
	}

	if ow.flags.JSON {
		return ow.outputJSON(ow.results)
	}

	// For non-JSON, print each result
	for _, r := range ow.results {
		fmt.Fprintln(ow.writer, r)
	}
	return nil
}

// Error outputs an error message.
func (ow *OutputWriter) Error(format string, args ...interface{}) {
	if ow.flags.JSON {
		errResult := map[string]interface{}{
			"error": fmt.Sprintf(format, args...),
		}
		ow.outputJSON(errResult)
		return
	}
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

// outputJSON marshals and outputs data as JSON.
func (ow *OutputWriter) outputJSON(data interface{}) error {
	encoder := json.NewEncoder(ow.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// IsJSON returns true if JSON output is enabled.
func (ow *OutputWriter) IsJSON() bool {
	return ow.flags.JSON
}

// IsQuiet returns true if quiet mode is enabled.
func (ow *OutputWriter) IsQuiet() bool {
	return ow.flags.Quiet
}

// NewConfigLoader creates a config loader for the given path.
func NewConfigLoader(path string) *ConfigLoader {
	return &ConfigLoader{configPath: path}
}

// Load loads and parses a config file into the target struct.
// Supports YAML (.yaml, .yml) and JSON (.json) formats.
func (cl *ConfigLoader) Load(target interface{}) error {
	if cl.configPath == "" {
		return nil // No config file specified
	}

	data, err := os.ReadFile(cl.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(cl.configPath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, target); err != nil {
			return fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, target); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		// Try YAML first (it's a superset of JSON)
		if err := yaml.Unmarshal(data, target); err != nil {
			return fmt.Errorf("failed to parse config (unknown format %s): %w", ext, err)
		}
	}

	return nil
}

// BatchProcessor handles batch input processing.
type BatchProcessor struct {
	items []string
}

// NewBatchProcessor creates a batch processor from a file.
// The file can contain one item per line, or be a JSON array of strings.
func NewBatchProcessor(batchFile string) (*BatchProcessor, error) {
	if batchFile == "" {
		return nil, nil // No batch processing
	}

	data, err := os.ReadFile(batchFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read batch file: %w", err)
	}

	bp := &BatchProcessor{}

	// Try JSON array first
	var jsonItems []string
	if err := json.Unmarshal(data, &jsonItems); err == nil {
		bp.items = jsonItems
		return bp, nil
	}

	// Fall back to line-based parsing
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			bp.items = append(bp.items, line)
		}
	}

	return bp, nil
}

// Items returns the batch items to process.
func (bp *BatchProcessor) Items() []string {
	if bp == nil {
		return nil
	}
	return bp.items
}

// Count returns the number of batch items.
func (bp *BatchProcessor) Count() int {
	if bp == nil {
		return 0
	}
	return len(bp.items)
}

// HasItems returns true if there are batch items to process.
func (bp *BatchProcessor) HasItems() bool {
	return bp != nil && len(bp.items) > 0
}

// Result represents a generic CLI result for JSON output.
type Result struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewResult creates a successful result.
func NewResult(data interface{}) Result {
	return Result{Success: true, Data: data}
}

// NewErrorResult creates an error result.
func NewErrorResult(err error) Result {
	return Result{Success: false, Error: err.Error()}
}

// BatchResult represents results from batch processing.
type BatchResult struct {
	Total     int      `json:"total"`
	Succeeded int      `json:"succeeded"`
	Failed    int      `json:"failed"`
	Results   []Result `json:"results"`
}

// NewBatchResult creates an empty batch result.
func NewBatchResult() *BatchResult {
	return &BatchResult{Results: make([]Result, 0)}
}

// Add adds a result to the batch.
func (br *BatchResult) Add(r Result) {
	br.Total++
	if r.Success {
		br.Succeeded++
	} else {
		br.Failed++
	}
	br.Results = append(br.Results, r)
}

// UsageHeader returns a formatted usage header for CLI commands.
func UsageHeader(name, description string) string {
	return fmt.Sprintf("%s\n\n%s\n", name, description)
}

// CommonUsage returns common flags usage text.
func CommonUsage() string {
	return `
Common Options:
  --json            Output results as JSON
  -q, --quiet       Suppress informational output
  -c, --config FILE Load configuration from YAML/JSON file
  --batch FILE      Process multiple items from file
  -o, --output FILE Write output to file (default: stdout)
  -h, --help        Show this help message
`
}
