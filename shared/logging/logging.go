// Package logging provides shared logging utilities for all demos.
package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger wraps log.Logger with demo-specific configuration
type Logger struct {
	*log.Logger
	logFile *os.File
}

// NewLogger creates a new logger for the specified demo
func NewLogger(demoName string) *Logger {
	// Get logs directory relative to executable or working directory
	logsDir := getLogsDir()
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		// Fallback to stdout only
		return &Logger{
			Logger: log.New(os.Stdout, "["+demoName+"] ", log.Ltime|log.Lmicroseconds),
		}
	}

	// Create log file with datetime
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logPath := filepath.Join(logsDir, timestamp+"-"+demoName+".log")

	logFile, err := os.Create(logPath)
	if err != nil {
		// Fallback to stdout only
		return &Logger{
			Logger: log.New(os.Stdout, "["+demoName+"] ", log.Ltime|log.Lmicroseconds),
		}
	}

	// Write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := &Logger{
		Logger:  log.New(multiWriter, "["+demoName+"] ", log.Ltime|log.Lmicroseconds),
		logFile: logFile,
	}
	logger.Printf("Logging to: %s", logPath)
	return logger
}

// Close closes the log file if open
func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}

// getLogsDir returns the logs directory path
func getLogsDir() string {
	// Try to find the logs directory relative to working directory
	paths := []string{
		"logs",
		"../logs",
		"../../logs",
	}

	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err == nil {
			// Check if parent directory exists
			parentDir := filepath.Dir(absPath)
			if _, err := os.Stat(parentDir); err == nil {
				return absPath
			}
		}
	}

	// Default to "logs" in current working directory
	return "logs"
}
