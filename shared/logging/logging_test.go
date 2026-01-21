package logging

import (
	"os"
	"path/filepath"
	"strings"
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
