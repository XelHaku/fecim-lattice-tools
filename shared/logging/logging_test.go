package logging

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestNewLogger(t *testing.T) {
	// Create logger for test
	logger := NewLogger("test-demo")
	defer logger.Close()

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	if logger.Logger == nil {
		t.Error("Logger.Logger is nil")
	}
}

func TestLoggerOutput(t *testing.T) {
	logger := NewLogger("test-output")
	defer logger.Close()

	// Should not panic
	logger.Println("Test message")
	logger.Printf("Test message with format: %d", 42)
}

func TestGetLogsDir(t *testing.T) {
	dir := getLogsDir()

	// Should return a non-empty string
	if dir == "" {
		t.Error("getLogsDir returned empty string")
	}

	// Should be a valid path (or at least not contain illegal characters)
	if strings.ContainsAny(dir, "<>:\"|?*") {
		t.Errorf("getLogsDir returned invalid path: %s", dir)
	}
}

func TestLoggerCreatesLogFile(t *testing.T) {
	// Create a unique logger name to avoid conflicts
	loggerName := "test-file-creation"
	logger := NewLogger(loggerName)

	if logger.logFile != nil {
		defer logger.Close()

		// Get the log file path
		logPath := logger.logFile.Name()

		// Verify the file exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Errorf("Log file was not created: %s", logPath)
		}

		// Verify the filename contains the logger name
		if !strings.Contains(filepath.Base(logPath), loggerName) {
			t.Errorf("Log filename doesn't contain logger name: %s", logPath)
		}

		// Clean up
		logger.Close()
		os.Remove(logPath)
	}
}

func TestLoggerClose(t *testing.T) {
	logger := NewLogger("test-close")

	// Should not panic even if logFile is nil
	logger.Close()

	// Should not panic on double close
	logger.Close()
}

func TestLoggerWithFailedFileCreation(t *testing.T) {
	// Save original function behavior - we can't easily test failed file creation
	// without modifying the filesystem, so we just verify the logger works
	// even when constructed in a fallback mode

	logger := &Logger{
		Logger:  nil,
		logFile: nil,
	}

	// Close should not panic even with nil fields
	logger.Close()
}

// TestCalculation tests the Calculation method for physics/math calculations
func TestCalculation(t *testing.T) {
	// Create a buffer to capture log output
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-calc")

	// Set verbosity to TRACE to enable calculation logging
	oldVerbosity := GetVerbosity()
	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(oldVerbosity)

	// Test calculation logging
	inputs := map[string]interface{}{
		"voltage": 3.3,
		"current": 0.001,
	}
	result := 3.3 * 0.001 // 0.0033 W

	logger.Calculation("calculatePower", inputs, result)

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Calculation log should contain [DEBUG], got: %s", output)
	}
	if !strings.Contains(output, "CALC:") {
		t.Errorf("Calculation log should contain CALC:, got: %s", output)
	}
	if !strings.Contains(output, "calculatePower") {
		t.Errorf("Calculation log should contain function name, got: %s", output)
	}
	if !strings.Contains(output, "voltage") {
		t.Errorf("Calculation log should contain input param name, got: %s", output)
	}
}

// TestCalculationVerbosityFiltering tests that Calculation respects verbosity levels
func TestCalculationVerbosityFiltering(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-calc-filter")

	// Set verbosity below DEBUG - calculation should not log
	SetVerbosity(VerbosityInfo)
	defer SetVerbosity(VerbosityOff)

	inputs := map[string]interface{}{"x": 10}
	logger.Calculation("testFunc", inputs, 100)

	output := buf.String()
	if strings.Contains(output, "CALC:") {
		t.Errorf("Calculation should not log at INFO verbosity, got: %s", output)
	}
}

// TestInput tests the Input method for function entry logging
func TestInput(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-input")

	SetVerbosity(VerbosityDebug)
	defer SetVerbosity(VerbosityOff)

	params := map[string]interface{}{
		"arraySize": 256,
		"threshold": 0.5,
		"enabled":   true,
	}

	logger.Input("processArray", params)

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Input log should contain [DEBUG], got: %s", output)
	}
	if !strings.Contains(output, "INPUT:") {
		t.Errorf("Input log should contain INPUT:, got: %s", output)
	}
	if !strings.Contains(output, "processArray") {
		t.Errorf("Input log should contain function name, got: %s", output)
	}
	if !strings.Contains(output, "arraySize") {
		t.Errorf("Input log should contain param name, got: %s", output)
	}
}

// TestInputVerbosityFiltering tests that Input respects verbosity levels
func TestInputVerbosityFiltering(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-input-filter")

	SetVerbosity(VerbosityInfo)
	defer SetVerbosity(VerbosityOff)

	params := map[string]interface{}{"x": 10}
	logger.Input("testFunc", params)

	output := buf.String()
	if strings.Contains(output, "INPUT:") {
		t.Errorf("Input should not log at INFO verbosity, got: %s", output)
	}
}

// TestOutput tests the Output method for function return logging
func TestOutput(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-output")

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	result := struct {
		Success bool
		Value   float64
	}{true, 42.5}

	logger.Output("computeResult", result)

	output := buf.String()
	if !strings.Contains(output, "[TRACE]") {
		t.Errorf("Output log should contain [TRACE], got: %s", output)
	}
	if !strings.Contains(output, "OUTPUT:") {
		t.Errorf("Output log should contain OUTPUT:, got: %s", output)
	}
	if !strings.Contains(output, "computeResult") {
		t.Errorf("Output log should contain function name, got: %s", output)
	}
	if !strings.Contains(output, "->") {
		t.Errorf("Output log should contain arrow (->), got: %s", output)
	}
}

// TestOutputVerbosityFiltering tests that Output respects verbosity levels
func TestOutputVerbosityFiltering(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-output-filter")

	SetVerbosity(VerbosityOff)
	defer SetVerbosity(VerbosityOff)

	logger.Output("testFunc", "result")

	output := buf.String()
	if strings.Contains(output, "OUTPUT:") {
		t.Errorf("Output should not log at OFF verbosity, got: %s", output)
	}
}

// TestError tests the Error method for error logging
func TestError(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-error")

	// Error should log regardless of verbosity
	SetVerbosity(VerbosityOff)
	defer SetVerbosity(VerbosityOff)

	err := os.ErrNotExist
	logger.Error(err, "failed to load configuration")

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Error log should contain [ERROR], got: %s", output)
	}
	if !strings.Contains(output, "failed to load configuration") {
		t.Errorf("Error log should contain context, got: %s", output)
	}
	if !strings.Contains(output, "file does not exist") {
		t.Errorf("Error log should contain error message, got: %s", output)
	}
}

// TestErrorAlwaysLogs tests that Error logs regardless of verbosity
func TestErrorAlwaysLogs(t *testing.T) {
	verbosityLevels := []VerbosityLevel{VerbosityOff, VerbosityInfo, VerbosityDebug, VerbosityTrace}

	for _, level := range verbosityLevels {
		t.Run(VerbosityString(level), func(t *testing.T) {
			var buf strings.Builder
			logger := createTestLogger(&buf, "test-error-always")

			SetVerbosity(level)
			defer SetVerbosity(VerbosityOff)

			err := os.ErrPermission
			logger.Error(err, "access denied")

			output := buf.String()
			if !strings.Contains(output, "[ERROR]") {
				t.Errorf("Error should log at %s verbosity, got: %s", VerbosityString(level), output)
			}
		})
	}
}

// TestErrorContext tests the ErrorContext method with additional details
func TestErrorContext(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-error-ctx")

	SetVerbosity(VerbosityOff) // Should still log
	defer SetVerbosity(VerbosityOff)

	err := os.ErrNotExist
	details := map[string]interface{}{
		"filename": "config.json",
		"attempts": 3,
	}

	logger.ErrorContext("LoadConfig", err, details)

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("ErrorContext log should contain [ERROR], got: %s", output)
	}
	if !strings.Contains(output, "LoadConfig") {
		t.Errorf("ErrorContext log should contain operation name, got: %s", output)
	}
	if !strings.Contains(output, "filename") {
		t.Errorf("ErrorContext log should contain detail key, got: %s", output)
	}
	if !strings.Contains(output, "config.json") {
		t.Errorf("ErrorContext log should contain detail value, got: %s", output)
	}
}

// TestErrorContextAlwaysLogs tests that ErrorContext logs regardless of verbosity
func TestErrorContextAlwaysLogs(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-error-ctx-always")

	SetVerbosity(VerbosityOff)
	defer SetVerbosity(VerbosityOff)

	err := os.ErrClosed
	details := map[string]interface{}{"socket": "tcp:8080"}

	logger.ErrorContext("ConnectSocket", err, details)

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("ErrorContext should log at OFF verbosity, got: %s", output)
	}
}

// TestErrorWithNilError tests Error handles nil error gracefully
func TestErrorWithNilError(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-error-nil")

	// Should not panic with nil error
	logger.Error(nil, "some context")

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Error should still log with nil error, got: %s", output)
	}
	if !strings.Contains(output, "nil") || !strings.Contains(output, "<nil>") {
		// Either "nil" or "<nil>" is acceptable
		if !strings.Contains(output, "nil") {
			t.Errorf("Error should indicate nil error, got: %s", output)
		}
	}
}

// TestErrorContextWithNilDetails tests ErrorContext handles nil details
func TestErrorContextWithNilDetails(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-error-ctx-nil")

	err := os.ErrNotExist
	// Should not panic with nil details
	logger.ErrorContext("Operation", err, nil)

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("ErrorContext should log with nil details, got: %s", output)
	}
}

// TestGlobalCalculation tests the global Calculation convenience function
func TestGlobalCalculation(t *testing.T) {
	var buf strings.Builder
	oldLogger := defaultLogger
	defaultLogger = createTestLogger(&buf, "global-calc")
	defer func() { defaultLogger = oldLogger }()

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	inputs := map[string]interface{}{"a": 1, "b": 2}
	GlobalCalculation("add", inputs, 3)

	output := buf.String()
	if !strings.Contains(output, "CALC:") {
		t.Errorf("GlobalCalculation should log CALC:, got: %s", output)
	}
}

// TestGlobalInput tests the global Input convenience function
func TestGlobalInput(t *testing.T) {
	var buf strings.Builder
	oldLogger := defaultLogger
	defaultLogger = createTestLogger(&buf, "global-input")
	defer func() { defaultLogger = oldLogger }()

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	params := map[string]interface{}{"size": 100}
	GlobalInput("processData", params)

	output := buf.String()
	if !strings.Contains(output, "INPUT:") {
		t.Errorf("GlobalInput should log INPUT:, got: %s", output)
	}
}

// TestGlobalOutput tests the global Output convenience function
func TestGlobalOutput(t *testing.T) {
	var buf strings.Builder
	oldLogger := defaultLogger
	defaultLogger = createTestLogger(&buf, "global-output")
	defer func() { defaultLogger = oldLogger }()

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	GlobalOutput("computeSum", 42)

	output := buf.String()
	if !strings.Contains(output, "OUTPUT:") {
		t.Errorf("GlobalOutput should log OUTPUT:, got: %s", output)
	}
}

// TestCalculationEmptyInputs tests Calculation with empty inputs map
func TestCalculationEmptyInputs(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-calc-empty")

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	logger.Calculation("noArgs", map[string]interface{}{}, "result")

	output := buf.String()
	if !strings.Contains(output, "CALC:") {
		t.Errorf("Calculation should log with empty inputs, got: %s", output)
	}
	if !strings.Contains(output, "noArgs") {
		t.Errorf("Calculation should contain function name, got: %s", output)
	}
}

// TestInputEmptyParams tests Input with empty params map
func TestInputEmptyParams(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-input-empty")

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	logger.Input("noParams", map[string]interface{}{})

	output := buf.String()
	if !strings.Contains(output, "INPUT:") {
		t.Errorf("Input should log with empty params, got: %s", output)
	}
}

// TestOutputNilResult tests Output with nil result
func TestOutputNilResult(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-output-nil")

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	logger.Output("returnsNil", nil)

	output := buf.String()
	if !strings.Contains(output, "OUTPUT:") {
		t.Errorf("Output should log with nil result, got: %s", output)
	}
}

// TestCalculationWithComplexTypes tests Calculation with slices and nested structures
func TestCalculationWithComplexTypes(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-calc-complex")

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	inputs := map[string]interface{}{
		"weights": []float64{0.1, 0.2, 0.3},
		"config": map[string]int{
			"layers": 3,
		},
	}
	result := []float64{0.5, 0.6, 0.7}

	logger.Calculation("forwardPass", inputs, result)

	output := buf.String()
	if !strings.Contains(output, "CALC:") {
		t.Errorf("Calculation should handle complex types, got: %s", output)
	}
}

// createTestLogger creates a logger that writes to the provided buffer
func createTestLogger(buf *strings.Builder, name string) *Logger {
	return &Logger{
		Logger:   log.New(buf, "["+name+"] ", log.Ldate|log.Ltime|log.Lmicroseconds),
		logFile:  nil,
		demoName: name,
	}
}

// TestGlobalCalculationFallback tests GlobalCalculation when defaultLogger is nil
func TestGlobalCalculationFallback(t *testing.T) {
	oldLogger := defaultLogger
	defaultLogger = nil
	defer func() { defaultLogger = oldLogger }()

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	// Should not panic when defaultLogger is nil
	GlobalCalculation("testFunc", map[string]interface{}{"x": 1}, 2)
}

// TestGlobalInputFallback tests GlobalInput when defaultLogger is nil
func TestGlobalInputFallback(t *testing.T) {
	oldLogger := defaultLogger
	defaultLogger = nil
	defer func() { defaultLogger = oldLogger }()

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	// Should not panic when defaultLogger is nil
	GlobalInput("testFunc", map[string]interface{}{"param": "value"})
}

// TestGlobalOutputFallback tests GlobalOutput when defaultLogger is nil
func TestGlobalOutputFallback(t *testing.T) {
	oldLogger := defaultLogger
	defaultLogger = nil
	defer func() { defaultLogger = oldLogger }()

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	// Should not panic when defaultLogger is nil
	GlobalOutput("testFunc", "result")
}

// TestErrorContextWithEmptyDetails tests ErrorContext with empty (not nil) details
func TestErrorContextWithEmptyDetails(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-error-ctx-empty")

	err := os.ErrNotExist
	logger.ErrorContext("Operation", err, map[string]interface{}{})

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("ErrorContext should log with empty details, got: %s", output)
	}
	// Should not contain parentheses for empty details
	if strings.Contains(output, "()") {
		t.Errorf("ErrorContext should not have empty parens, got: %s", output)
	}
}

// TestErrorContextWithNilError tests ErrorContext handles nil error
func TestErrorContextWithNilError(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-error-ctx-nil-err")

	details := map[string]interface{}{"key": "value"}
	logger.ErrorContext("Operation", nil, details)

	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("ErrorContext should log with nil error, got: %s", output)
	}
	if !strings.Contains(output, "<nil>") {
		t.Errorf("ErrorContext should show <nil> for nil error, got: %s", output)
	}
}

// TestSetVerbosity tests the SetVerbosity function
func TestSetVerbosity(t *testing.T) {
	// Save original
	original := GetVerbosity()
	defer SetVerbosity(original)

	levels := []VerbosityLevel{VerbosityOff, VerbosityInfo, VerbosityDebug, VerbosityTrace}
	for _, level := range levels {
		SetVerbosity(level)
		if GetVerbosity() != level {
			t.Errorf("SetVerbosity(%v) failed, GetVerbosity() = %v", level, GetVerbosity())
		}
	}
}

// TestIsVerbose tests the IsVerbose function
func TestIsVerbose(t *testing.T) {
	original := GetVerbosity()
	defer SetVerbosity(original)

	SetVerbosity(VerbosityDebug)

	if !IsVerbose(VerbosityOff) {
		t.Error("IsVerbose(VerbosityOff) should return true when verbosity is Debug")
	}
	if !IsVerbose(VerbosityInfo) {
		t.Error("IsVerbose(VerbosityInfo) should return true when verbosity is Debug")
	}
	if !IsVerbose(VerbosityDebug) {
		t.Error("IsVerbose(VerbosityDebug) should return true when verbosity is Debug")
	}
	if IsVerbose(VerbosityTrace) {
		t.Error("IsVerbose(VerbosityTrace) should return false when verbosity is Debug")
	}
}

// TestParseVerbosityFlag tests parsing verbosity flags
func TestParseVerbosityFlag(t *testing.T) {
	tests := []struct {
		input  string
		expect VerbosityLevel
	}{
		{"0", VerbosityOff},
		{"off", VerbosityOff},
		{"none", VerbosityOff},
		{"1", VerbosityInfo},
		{"info", VerbosityInfo},
		{"2", VerbosityDebug},
		{"debug", VerbosityDebug},
		{"3", VerbosityTrace},
		{"trace", VerbosityTrace},
		{"all", VerbosityTrace},
		{"invalid", VerbosityOff},
		{"", VerbosityOff},
	}

	for _, tt := range tests {
		result := ParseVerbosityFlag(tt.input)
		if result != tt.expect {
			t.Errorf("ParseVerbosityFlag(%q) = %v, expected %v", tt.input, result, tt.expect)
		}
	}
}

// TestVerbosityString tests verbosity level to string conversion
func TestVerbosityString(t *testing.T) {
	tests := []struct {
		level  VerbosityLevel
		expect string
	}{
		{VerbosityOff, "off"},
		{VerbosityInfo, "info"},
		{VerbosityDebug, "debug"},
		{VerbosityTrace, "trace"},
		{VerbosityLevel(99), "unknown(99)"},
	}

	for _, tt := range tests {
		result := VerbosityString(tt.level)
		if result != tt.expect {
			t.Errorf("VerbosityString(%v) = %q, expected %q", tt.level, result, tt.expect)
		}
	}
}

// TestNewNoOpLogger tests the no-op logger
func TestNewNoOpLogger(t *testing.T) {
	logger := NewNoOpLogger()

	if logger == nil {
		t.Fatal("NewNoOpLogger should return non-nil")
	}
	if logger.demoName != "noop" {
		t.Errorf("expected demoName='noop', got %q", logger.demoName)
	}

	// Should not panic
	logger.Println("test")
	logger.Printf("test %d", 42)
	logger.Close()
}

// TestLoggerButton tests Button logging
func TestLoggerButton(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-button")

	SetVerbosity(VerbosityDebug)
	defer SetVerbosity(VerbosityOff)

	logger.Button("StartButton")

	output := buf.String()
	if !strings.Contains(output, "BUTTON:") {
		t.Errorf("Button should log BUTTON:, got: %s", output)
	}
	if !strings.Contains(output, "StartButton") {
		t.Errorf("Button should log button name, got: %s", output)
	}
}

// TestLoggerValueChange tests ValueChange logging
func TestLoggerValueChange(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-value")

	SetVerbosity(VerbosityDebug)
	defer SetVerbosity(VerbosityOff)

	logger.ValueChange("Slider", 0.5, 0.8)

	output := buf.String()
	if !strings.Contains(output, "VALUE:") {
		t.Errorf("ValueChange should log VALUE:, got: %s", output)
	}
}

// TestLoggerSelection tests Selection logging
func TestLoggerSelection(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-select")

	SetVerbosity(VerbosityDebug)
	defer SetVerbosity(VerbosityOff)

	logger.Selection("Dropdown", "Option1")

	output := buf.String()
	if !strings.Contains(output, "SELECT:") {
		t.Errorf("Selection should log SELECT:, got: %s", output)
	}
}

// TestLoggerSliderChange tests SliderChange logging
func TestLoggerSliderChange(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-slider")

	SetVerbosity(VerbosityDebug)
	defer SetVerbosity(VerbosityOff)

	logger.SliderChange("VolumeSlider", 0.75)

	output := buf.String()
	if !strings.Contains(output, "SLIDER:") {
		t.Errorf("SliderChange should log SLIDER:, got: %s", output)
	}
}

// TestLoggerTabChange tests TabChange logging
func TestLoggerTabChange(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-tab")

	SetVerbosity(VerbosityDebug)
	defer SetVerbosity(VerbosityOff)

	logger.TabChange("Settings")

	output := buf.String()
	if !strings.Contains(output, "TAB:") {
		t.Errorf("TabChange should log TAB:, got: %s", output)
	}
}

// TestLoggerCheckboxChange tests CheckboxChange logging
func TestLoggerCheckboxChange(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-checkbox")

	SetVerbosity(VerbosityDebug)
	defer SetVerbosity(VerbosityOff)

	logger.CheckboxChange("EnableFeature", true)

	output := buf.String()
	if !strings.Contains(output, "CHECKBOX:") {
		t.Errorf("CheckboxChange should log CHECKBOX:, got: %s", output)
	}
}

// TestLoggerEntryChange tests EntryChange logging
func TestLoggerEntryChange(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-entry")

	SetVerbosity(VerbosityDebug)
	defer SetVerbosity(VerbosityOff)

	logger.EntryChange("UsernameField", "testuser")

	output := buf.String()
	if !strings.Contains(output, "ENTRY:") {
		t.Errorf("EntryChange should log ENTRY:, got: %s", output)
	}
}

// TestLoggerInfo tests Info logging
func TestLoggerInfo(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-info")

	SetVerbosity(VerbosityInfo)
	defer SetVerbosity(VerbosityOff)

	logger.Info("Application started")

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Info should log [INFO], got: %s", output)
	}
}

// TestLoggerDebug tests Debug logging
func TestLoggerDebug(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-debug")

	SetVerbosity(VerbosityDebug)
	defer SetVerbosity(VerbosityOff)

	logger.Debug("Debug message")

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Debug should log [DEBUG], got: %s", output)
	}
}

// TestLoggerTrace tests Trace logging
func TestLoggerTrace(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "test-trace")

	SetVerbosity(VerbosityTrace)
	defer SetVerbosity(VerbosityOff)

	logger.Trace("Trace message")

	output := buf.String()
	if !strings.Contains(output, "[TRACE]") {
		t.Errorf("Trace should log [TRACE], got: %s", output)
	}
}

// TestGlobalPrintf tests global Printf
func TestGlobalPrintf(t *testing.T) {
	// Should not panic with nil defaultLogger
	oldLogger := defaultLogger
	defaultLogger = nil
	Printf("test %d", 42)
	defaultLogger = oldLogger
}

// TestGlobalPrintln tests global Println
func TestGlobalPrintln(t *testing.T) {
	// Should not panic with nil defaultLogger
	oldLogger := defaultLogger
	defaultLogger = nil
	Println("test message")
	defaultLogger = oldLogger
}

// TestGlobalInfo tests global GlobalInfo
func TestGlobalInfo(t *testing.T) {
	SetVerbosity(VerbosityInfo)
	defer SetVerbosity(VerbosityOff)

	// Should not panic
	GlobalInfo("info message")
}

// TestGlobalDebug tests global GlobalDebug
func TestGlobalDebug(t *testing.T) {
	SetVerbosity(VerbosityDebug)
	defer SetVerbosity(VerbosityOff)

	// Should not panic
	GlobalDebug("debug message")
}

// TestGlobalError tests global GlobalError
func TestGlobalError(t *testing.T) {
	// Should not panic
	GlobalError("error: %s", "something went wrong")
}

// TestEnableFileLogging tests file logging toggle
func TestEnableFileLogging(t *testing.T) {
	// Check initial state
	wasEnabled := IsFileLoggingEnabled()

	// Enable
	EnableFileLogging()

	if !IsFileLoggingEnabled() {
		t.Error("IsFileLoggingEnabled should return true after EnableFileLogging")
	}

	// Note: We can't easily disable it for cleanup, but this tests the mechanism
	_ = wasEnabled
}

// TestFormatParams tests the formatParams helper
func TestFormatParams(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{"empty params", map[string]interface{}{}},
		{"single param", map[string]interface{}{"key": "value"}},
		{"multiple params", map[string]interface{}{"a": 1, "b": "two", "c": 3.14}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatParams(tt.params)
			// Just verify it doesn't panic and returns a string
			_ = result
		})
	}
}

func TestStructuredLoggingOutputsValidJSON(t *testing.T) {
	entry := NewEntry(LevelInfo, "logging-test", "structured log").
		WithCategory("TEST").
		WithField("request_id", "req-123").
		WithField("count", 7)

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("expected entry to marshal to JSON, got error: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("expected valid JSON payload, got error: %v", err)
	}

	if decoded["Source"] != "logging-test" {
		t.Fatalf("unexpected source in JSON: %v", decoded["Source"])
	}
	if decoded["Message"] != "structured log" {
		t.Fatalf("unexpected message in JSON: %v", decoded["Message"])
	}
}

func TestLogLevelsFilterCorrectly(t *testing.T) {
	var buf strings.Builder
	logger := createTestLogger(&buf, "level-filter")

	original := GetVerbosity()
	defer SetVerbosity(original)

	SetVerbosity(VerbosityInfo)
	logger.Info("visible info")
	logger.Debug("hidden debug")
	logger.Trace("hidden trace")

	output := buf.String()
	if !strings.Contains(output, "[INFO] visible info") {
		t.Fatalf("expected INFO log to be visible at INFO level, output=%q", output)
	}
	if strings.Contains(output, "hidden debug") || strings.Contains(output, "hidden trace") {
		t.Fatalf("expected DEBUG/TRACE logs to be filtered at INFO level, output=%q", output)
	}

	buf.Reset()
	SetVerbosity(VerbosityTrace)
	logger.Info("info at trace")
	logger.Debug("debug at trace")
	logger.Trace("trace at trace")

	output = buf.String()
	for _, want := range []string{"[INFO] info at trace", "[DEBUG] debug at trace", "[TRACE] trace at trace"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected %q in output at TRACE level, output=%q", want, output)
		}
	}
}

func TestFileLoggingCreatesValidLogFiles(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "app.log")

	if err := Init("file-test", logPath); err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	SetVerbosity(VerbosityInfo)
	GlobalInfo("file logging smoke test")
	CloseGlobal()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("expected log file to exist, read error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected log file to be non-empty")
	}
	content := string(data)
	if !strings.Contains(content, "Logging initialized to:") {
		t.Fatalf("expected initialization line in log file, content=%q", content)
	}
	if !strings.Contains(content, "file logging smoke test") {
		t.Fatalf("expected application log line in log file, content=%q", content)
	}
}

func TestLogRotationWorksWhenConfigured(t *testing.T) {
	bufferMu.Lock()
	origBuffer := buffer
	origBufferSize := bufferSize
	origWriteAt := bufferWriteAt
	origCount := bufferCount

	bufferSize = 3
	buffer = make([]*Entry, bufferSize)
	bufferWriteAt = 0
	bufferCount = 0
	bufferMu.Unlock()

	defer func() {
		bufferMu.Lock()
		buffer = origBuffer
		bufferSize = origBufferSize
		bufferWriteAt = origWriteAt
		bufferCount = origCount
		bufferMu.Unlock()
	}()

	for i := 0; i < 5; i++ {
		AddToBuffer(NewEntry(LevelInfo, "rotation", fmt.Sprintf("entry-%d", i)))
	}

	entries := ReadBuffer(10)
	if len(entries) != 3 {
		t.Fatalf("expected rotated buffer to keep 3 entries, got %d", len(entries))
	}
	if entries[0].Message != "entry-2" || entries[1].Message != "entry-3" || entries[2].Message != "entry-4" {
		t.Fatalf("unexpected rotated entries: got [%s, %s, %s]", entries[0].Message, entries[1].Message, entries[2].Message)
	}
}

func TestConcurrentLoggingDoesNotProduceInterleavedOutput(t *testing.T) {
	var out bytes.Buffer
	logger := &Logger{Logger: log.New(&out, "", 0), demoName: "concurrency"}

	const workers = 8
	const perWorker = 50

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				logger.Printf("worker=%d seq=%d", worker, i)
			}
		}(w)
	}
	wg.Wait()

	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		if !strings.Contains(line, "worker=") || !strings.Contains(line, "seq=") {
			t.Fatalf("detected malformed/interleaved line: %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("failed scanning log output: %v", err)
	}

	expected := workers * perWorker
	if lineCount != expected {
		t.Fatalf("expected %d complete lines, got %d", expected, lineCount)
	}
}
