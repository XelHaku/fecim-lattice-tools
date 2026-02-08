// Package logging provides shared logging utilities for all demos.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Verbosity levels for logging
type VerbosityLevel int

const (
	VerbosityOff   VerbosityLevel = 0 // No debug logging
	VerbosityInfo  VerbosityLevel = 1 // Basic info (startup, shutdown)
	VerbosityDebug VerbosityLevel = 2 // Debug (button clicks, value changes)
	VerbosityTrace VerbosityLevel = 3 // Trace (every UI update, simulation tick)
)

// Global verbosity level - set via SetVerbosity()
var (
	globalVerbosity VerbosityLevel = VerbosityOff
	verbosityMu     sync.RWMutex

	// File logging control - disabled by default, enable with EnableFileLogging()
	fileLoggingEnabled bool
	fileLoggingMu      sync.RWMutex

	// Shared log file for all loggers
	sharedLogFile   *os.File
	sharedLogWriter io.Writer
	sharedLogMu     sync.Mutex
	sharedLogPath   string
)

// lazyWriter writes to the shared log writer if available, otherwise discards.
// This lets loggers created before EnableFileLogging() start emitting once
// the shared writer is initialized.
type lazyWriter struct{}

func (w *lazyWriter) Write(p []byte) (int, error) {
	sharedLogMu.Lock()
	writer := sharedLogWriter
	sharedLogMu.Unlock()
	if writer == nil {
		return len(p), nil
	}
	return writer.Write(p)
}

func ensureSharedLogWriter() {
	sharedLogMu.Lock()
	defer sharedLogMu.Unlock()

	if sharedLogWriter != nil {
		return
	}

	logsDir := getLogsDir()
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		// Fallback to stdout only
		sharedLogWriter = os.Stdout
		return
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	sharedLogPath = filepath.Join(logsDir, timestamp+"-fecim.log")

	var err error
	sharedLogFile, err = os.Create(sharedLogPath)
	if err != nil {
		// Fallback to stdout only
		sharedLogWriter = os.Stdout
		sharedLogPath = ""
		return
	}

	// Write to both file and stdout
	sharedLogWriter = io.MultiWriter(os.Stdout, sharedLogFile)
}

// SetVerbosity sets the global verbosity level
func SetVerbosity(level VerbosityLevel) {
	verbosityMu.Lock()
	globalVerbosity = level
	verbosityMu.Unlock()
}

// GetVerbosity returns the current global verbosity level
func GetVerbosity() VerbosityLevel {
	verbosityMu.RLock()
	defer verbosityMu.RUnlock()
	return globalVerbosity
}

// IsVerbose returns true if verbosity is at least the specified level
func IsVerbose(level VerbosityLevel) bool {
	return GetVerbosity() >= level
}

// EnableFileLogging enables file logging. Must be called before NewLogger
// if you want logs to be written to files. By default, file logging is disabled.
func EnableFileLogging() {
	fileLoggingMu.Lock()
	fileLoggingEnabled = true
	fileLoggingMu.Unlock()
	ensureSharedLogWriter()
}

// IsFileLoggingEnabled returns true if file logging is enabled
func IsFileLoggingEnabled() bool {
	fileLoggingMu.RLock()
	defer fileLoggingMu.RUnlock()
	return fileLoggingEnabled
}

// Logger wraps log.Logger with demo-specific configuration and verbosity support
type Logger struct {
	*log.Logger
	logFile  *os.File
	demoName string
}

// NewNoOpLogger creates a logger that doesn't write to any files
// Used when --logger flag is not provided
func NewNoOpLogger() *Logger {
	return &Logger{
		Logger:   log.New(io.Discard, "", 0),
		logFile:  nil,
		demoName: "noop",
	}
}

// NewLogger creates a new logger for the specified demo
// All loggers share a single log file to avoid creating multiple files
// If file logging is not enabled (via EnableFileLogging()), returns a no-op logger
func NewLogger(demoName string) *Logger {
	if IsFileLoggingEnabled() {
		ensureSharedLogWriter()
	}

	logger := &Logger{
		Logger:   log.New(&lazyWriter{}, "["+demoName+"] ", log.Ldate|log.Ltime|log.Lmicroseconds),
		logFile:  nil, // Don't store file reference - shared file is managed globally
		demoName: demoName,
	}

	// Only log the path once for the first logger
	sharedLogMu.Lock()
	firstPath := sharedLogPath
	if sharedLogPath != "" {
		sharedLogPath = "" // Clear to avoid repeating
	}
	writerReady := sharedLogWriter != nil
	sharedLogMu.Unlock()

	if firstPath != "" {
		logger.Printf("Logging to: %s", firstPath)
	}
	if writerReady {
		// Hook standard log package to write to the shared log file as well
		log.SetOutput(&lazyWriter{})
	}

	return logger
}

// Close is a no-op for individual loggers since they share a file
// Use CloseShared() to close the shared log file
func (l *Logger) Close() {
	// No-op - shared log file is managed globally
}

// CloseShared closes the shared log file
func CloseShared() {
	sharedLogMu.Lock()
	defer sharedLogMu.Unlock()
	if sharedLogFile != nil {
		sharedLogFile.Close()
		sharedLogFile = nil
		sharedLogWriter = nil
	}
}

// Info logs at INFO level (verbosity >= 1)
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// Add to global buffer for LogViewer
	entry := NewEntry(LevelInfo, l.demoName, msg)
	AddToBuffer(entry)

	if IsVerbose(VerbosityInfo) {
		l.Printf("[INFO] "+format, args...)
	}
}

// Debug logs at DEBUG level (verbosity >= 2) - for button clicks, value changes
func (l *Logger) Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// Add to global buffer for LogViewer
	entry := NewEntry(LevelDebug, l.demoName, msg)
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] "+format, args...)
	}
}

// Trace logs at TRACE level (verbosity >= 3) - for frequent updates
func (l *Logger) Trace(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// Add to global buffer for LogViewer
	entry := NewEntry(LevelTrace, l.demoName, msg)
	AddToBuffer(entry)

	if IsVerbose(VerbosityTrace) {
		l.Printf("[TRACE] "+format, args...)
	}
}

// Warn logs at WARN level (always logs, less severe than ERROR)
func (l *Logger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// Add to global buffer for LogViewer
	entry := NewEntry(LevelWarn, l.demoName, msg)
	AddToBuffer(entry)

	l.Printf("[WARN] "+format, args...)
}

// Button logs a button click event at DEBUG level
func (l *Logger) Button(buttonName string) {
	// Add to global buffer for LogViewer
	entry := NewEntry(LevelDebug, l.demoName, buttonName+" clicked").WithCategory("BUTTON")
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] BUTTON: %s clicked", buttonName)
	}
}

// ValueChange logs a value change event at DEBUG level
func (l *Logger) ValueChange(widgetName string, oldValue, newValue interface{}) {
	msg := fmt.Sprintf("%s changed from %v to %v", widgetName, oldValue, newValue)
	entry := NewEntry(LevelDebug, l.demoName, msg).WithCategory("VALUE")
	entry.WithFields(map[string]interface{}{"old": oldValue, "new": newValue, "widget": widgetName})
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] VALUE: %s changed from %v to %v", widgetName, oldValue, newValue)
	}
}

// Selection logs a selection change event at DEBUG level
func (l *Logger) Selection(widgetName string, selected string) {
	msg := fmt.Sprintf("%s = %q", widgetName, selected)
	entry := NewEntry(LevelDebug, l.demoName, msg).WithCategory("SELECT")
	entry.WithField("widget", widgetName).WithField("selected", selected)
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] SELECT: %s = %q", widgetName, selected)
	}
}

// SliderChange logs a slider value change at DEBUG level
func (l *Logger) SliderChange(sliderName string, value float64) {
	msg := fmt.Sprintf("%s = %.4f", sliderName, value)
	entry := NewEntry(LevelDebug, l.demoName, msg).WithCategory("SLIDER")
	entry.WithField("slider", sliderName).WithField("value", value)
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] SLIDER: %s = %.4f", sliderName, value)
	}
}

// TabChange logs a tab selection change at DEBUG level
func (l *Logger) TabChange(tabName string) {
	msg := fmt.Sprintf("switched to %q", tabName)
	entry := NewEntry(LevelDebug, l.demoName, msg).WithCategory("TAB")
	entry.WithField("tab", tabName)
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] TAB: switched to %q", tabName)
	}
}

// CheckboxChange logs a checkbox state change at DEBUG level
func (l *Logger) CheckboxChange(checkboxName string, checked bool) {
	msg := fmt.Sprintf("%s = %v", checkboxName, checked)
	entry := NewEntry(LevelDebug, l.demoName, msg).WithCategory("CHECKBOX")
	entry.WithField("checkbox", checkboxName).WithField("checked", checked)
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] CHECKBOX: %s = %v", checkboxName, checked)
	}
}

// EntryChange logs a text entry change at DEBUG level
func (l *Logger) EntryChange(entryName string, text string) {
	msg := fmt.Sprintf("%s = %q", entryName, text)
	entry := NewEntry(LevelDebug, l.demoName, msg).WithCategory("ENTRY")
	entry.WithField("entry", entryName).WithField("text", text)
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] ENTRY: %s = %q", entryName, text)
	}
}

// Calculation logs a physics/math calculation at DEBUG level
// Format: [DEBUG] CALC: funcName(param1=value1, param2=value2) = result
func (l *Logger) Calculation(funcName string, inputs map[string]interface{}, result interface{}) {
	msg := fmt.Sprintf("%s(%s) = %v", funcName, formatParams(inputs), result)
	entry := NewEntry(LevelDebug, l.demoName, msg).WithCategory("CALC")
	entry.WithFields(inputs).WithField("result", result)
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] CALC: %s(%s) = %v", funcName, formatParams(inputs), result)
	}
}

// Input logs function entry with parameters at DEBUG level
// Format: [DEBUG] INPUT: funcName(param1=value1, param2=value2)
func (l *Logger) Input(funcName string, params map[string]interface{}) {
	msg := fmt.Sprintf("%s(%s)", funcName, formatParams(params))
	entry := NewEntry(LevelDebug, l.demoName, msg).WithCategory("INPUT")
	entry.WithFields(params)
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] INPUT: %s(%s)", funcName, formatParams(params))
	}
}

// Output logs function return value at TRACE level
// Format: [TRACE] OUTPUT: funcName -> result
func (l *Logger) Output(funcName string, result interface{}) {
	msg := fmt.Sprintf("%s -> %v", funcName, result)
	entry := NewEntry(LevelTrace, l.demoName, msg).WithCategory("OUTPUT")
	entry.WithField("result", result)
	AddToBuffer(entry)

	if IsVerbose(VerbosityTrace) {
		l.Printf("[TRACE] OUTPUT: %s -> %v", funcName, result)
	}
}

// Error logs an error with context - always logs regardless of verbosity
// Format: [ERROR] context: error message
func (l *Logger) Error(err error, context string) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	} else {
		errMsg = "<nil>"
	}

	msg := fmt.Sprintf("%s: %s", context, errMsg)
	entry := NewEntry(LevelError, l.demoName, msg).WithError(err)
	entry.WithField("context", context)
	AddToBuffer(entry)

	l.Printf("[ERROR] %s: %s", context, errMsg)
}

// ErrorContext logs an error with operation context and additional details
// Format: [ERROR] operation: error message (detail1=val1, detail2=val2)
func (l *Logger) ErrorContext(operation string, err error, details map[string]interface{}) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	} else {
		errMsg = "<nil>"
	}

	var msg string
	if len(details) > 0 {
		msg = fmt.Sprintf("%s: %s (%s)", operation, errMsg, formatParams(details))
	} else {
		msg = fmt.Sprintf("%s: %s", operation, errMsg)
	}

	entry := NewEntry(LevelError, l.demoName, msg).WithError(err)
	entry.WithField("operation", operation).WithFields(details)
	AddToBuffer(entry)

	if len(details) > 0 {
		l.Printf("[ERROR] %s: %s (%s)", operation, errMsg, formatParams(details))
	} else {
		l.Printf("[ERROR] %s: %s", operation, errMsg)
	}
}

// formatParams formats a map of parameters as "key1=value1, key2=value2"
func formatParams(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}

	parts := make([]string, 0, len(params))
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, ", ")
}

// Global singleton logger
var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the global default logger
func Init(demoName string, logPath string) error {
	var err error
	once.Do(func() {
		// Use provided path or default
		if logPath == "" {
			logsDir := getLogsDir()
			if err := os.MkdirAll(logsDir, 0755); err != nil {
				// Fallback to current dir if logs dir creation fails
				logsDir = "."
			}
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			logPath = filepath.Join(logsDir, timestamp+"-"+demoName+".log")
		} else {
			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
				// Log to stderr if we can't create directory
				fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
			}
		}

		// Create log file
		var logFile *os.File
		logFile, err = os.Create(logPath)
		if err != nil {
			return
		}

		// Write to both file and stdout
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		defaultLogger = &Logger{
			Logger:   log.New(multiWriter, "["+demoName+"] ", log.Ldate|log.Ltime|log.Lmicroseconds),
			logFile:  logFile,
			demoName: demoName,
		}
		defaultLogger.Printf("Logging initialized to: %s", logPath)
	})
	return err
}

// Global convenience functions using the default logger

// Printf logs to the default logger
func Printf(format string, v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Printf(format, v...)
	} else {
		// Use shared log writer if available, otherwise stdout
		sharedLogMu.Lock()
		writer := sharedLogWriter
		sharedLogMu.Unlock()
		msg := fmt.Sprintf(format, v...)
		if writer != nil {
			fmt.Fprintf(writer, "%s %s\n", time.Now().Format("2006/01/02 15:04:05"), msg)
		} else {
			log.Printf("%s", msg)
		}
	}
}

// Println logs to the default logger
func Println(v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Println(v...)
	} else {
		// Use shared log writer if available, otherwise stdout
		sharedLogMu.Lock()
		writer := sharedLogWriter
		sharedLogMu.Unlock()
		if writer != nil {
			fmt.Fprint(writer, time.Now().Format("2006/01/02 15:04:05")+" ")
			fmt.Fprintln(writer, v...)
		} else {
			log.Println(v...)
		}
	}
}

// GlobalInfo logs at INFO level
func GlobalInfo(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	} else if IsVerbose(VerbosityInfo) {
		// Use shared log writer if available, otherwise stdout
		sharedLogMu.Lock()
		writer := sharedLogWriter
		sharedLogMu.Unlock()
		msg := fmt.Sprintf(format, args...)
		if writer != nil {
			fmt.Fprintf(writer, "%s [INFO] %s\n", time.Now().Format("2006/01/02 15:04:05"), msg)
		} else {
			log.Printf("[INFO] %s", msg)
		}
	}
}

// GlobalDebug logs at DEBUG level
func GlobalDebug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	} else if IsVerbose(VerbosityDebug) {
		// Use shared log writer if available, otherwise stdout
		sharedLogMu.Lock()
		writer := sharedLogWriter
		sharedLogMu.Unlock()
		msg := fmt.Sprintf(format, args...)
		if writer != nil {
			fmt.Fprintf(writer, "%s [DEBUG] %s\n", time.Now().Format("2006/01/02 15:04:05"), msg)
		} else {
			log.Printf("[DEBUG] %s", msg)
		}
	}
}

// GlobalError logs at ERROR level (always logs regardless of verbosity)
func GlobalError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// Add to global buffer for LogViewer
	entry := NewEntry(LevelError, "global", msg)
	AddToBuffer(entry)

	if defaultLogger != nil {
		defaultLogger.Printf("[ERROR] "+format, args...)
	} else {
		// Use shared log writer if available, otherwise stdout
		sharedLogMu.Lock()
		writer := sharedLogWriter
		sharedLogMu.Unlock()
		if writer != nil {
			fmt.Fprintf(writer, "%s [ERROR] %s\n", time.Now().Format("2006/01/02 15:04:05"), msg)
		} else {
			log.Printf("[ERROR] %s", msg)
		}
	}
}

// GlobalWarn logs at WARN level (always logs regardless of verbosity)
func GlobalWarn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// Add to global buffer for LogViewer
	entry := NewEntry(LevelWarn, "global", msg)
	AddToBuffer(entry)

	if defaultLogger != nil {
		defaultLogger.Printf("[WARN] "+format, args...)
	} else {
		// Use shared log writer if available, otherwise stdout
		sharedLogMu.Lock()
		writer := sharedLogWriter
		sharedLogMu.Unlock()
		if writer != nil {
			fmt.Fprintf(writer, "%s [WARN] %s\n", time.Now().Format("2006/01/02 15:04:05"), msg)
		} else {
			log.Printf("[WARN] %s", msg)
		}
	}
}

// GlobalCalculation logs a calculation at DEBUG level using the default logger
func GlobalCalculation(funcName string, inputs map[string]interface{}, result interface{}) {
	if defaultLogger != nil {
		defaultLogger.Calculation(funcName, inputs, result)
	} else if IsVerbose(VerbosityDebug) {
		// Use shared log writer if available, otherwise stdout
		sharedLogMu.Lock()
		writer := sharedLogWriter
		sharedLogMu.Unlock()
		if writer != nil {
			fmt.Fprintf(writer, "%s [DEBUG] CALC: %s(%s) = %v\n", time.Now().Format("2006/01/02 15:04:05"), funcName, formatParams(inputs), result)
		} else {
			log.Printf("[DEBUG] CALC: %s(%s) = %v", funcName, formatParams(inputs), result)
		}
	}
}

// GlobalInput logs function entry at DEBUG level using the default logger
func GlobalInput(funcName string, params map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Input(funcName, params)
	} else if IsVerbose(VerbosityDebug) {
		// Use shared log writer if available, otherwise stdout
		sharedLogMu.Lock()
		writer := sharedLogWriter
		sharedLogMu.Unlock()
		if writer != nil {
			fmt.Fprintf(writer, "%s [DEBUG] INPUT: %s(%s)\n", time.Now().Format("2006/01/02 15:04:05"), funcName, formatParams(params))
		} else {
			log.Printf("[DEBUG] INPUT: %s(%s)", funcName, formatParams(params))
		}
	}
}

// GlobalOutput logs function return at TRACE level using the default logger
func GlobalOutput(funcName string, result interface{}) {
	if defaultLogger != nil {
		defaultLogger.Output(funcName, result)
	} else if IsVerbose(VerbosityTrace) {
		// Use shared log writer if available, otherwise stdout
		sharedLogMu.Lock()
		writer := sharedLogWriter
		sharedLogMu.Unlock()
		if writer != nil {
			fmt.Fprintf(writer, "%s [TRACE] OUTPUT: %s -> %v\n", time.Now().Format("2006/01/02 15:04:05"), funcName, result)
		} else {
			log.Printf("[TRACE] OUTPUT: %s -> %v", funcName, result)
		}
	}
}

// CloseGlobal closes the default logger and shared log file
func CloseGlobal() {
	if defaultLogger != nil {
		defaultLogger.Close()
	}
	CloseShared()
}

func ParseVerbosityFlag(s string) VerbosityLevel {
	switch s {
	case "0", "off", "none":
		return VerbosityOff
	case "1", "info":
		return VerbosityInfo
	case "2", "debug":
		return VerbosityDebug
	case "3", "trace", "all":
		return VerbosityTrace
	default:
		return VerbosityOff
	}
}

// VerbosityString returns a human-readable string for the verbosity level
func VerbosityString(level VerbosityLevel) string {
	switch level {
	case VerbosityOff:
		return "off"
	case VerbosityInfo:
		return "info"
	case VerbosityDebug:
		return "debug"
	case VerbosityTrace:
		return "trace"
	default:
		return fmt.Sprintf("unknown(%d)", level)
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

// LogsDir exposes the resolved logs directory path for other packages.
func LogsDir() string {
	return getLogsDir()
}
