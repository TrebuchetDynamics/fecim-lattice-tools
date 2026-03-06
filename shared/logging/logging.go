// Package logging provides shared logging utilities for all demos.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Verbosity levels for logging
type VerbosityLevel int

const (
	VerbosityOff   VerbosityLevel = 0 // No debug logging
	VerbosityInfo  VerbosityLevel = 1 // Basic info (startup, shutdown)
	VerbosityDebug VerbosityLevel = 2 // Debug (button clicks, value changes)
	VerbosityTrace VerbosityLevel = 3 // Trace (every UI update, simulation tick)
)

// Global verbosity level - set via SetVerbosity() uses manager.go
// File logging control - uses manager.go

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

// CloseShared removed - in manager.go

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
	safeInputs := sanitizeFields(inputs)
	msg := fmt.Sprintf("%s(%s) = %v", funcName, formatParams(safeInputs), result)
	entry := NewEntry(LevelDebug, l.demoName, msg).WithCategory("CALC")
	entry.WithFields(safeInputs).WithField("result", result)
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] CALC: %s(%s) = %v", funcName, formatParams(safeInputs), result)
	}
}

// Input logs function entry with parameters at DEBUG level
// Format: [DEBUG] INPUT: funcName(param1=value1, param2=value2)
func (l *Logger) Input(funcName string, params map[string]interface{}) {
	safeParams := sanitizeFields(params)
	msg := fmt.Sprintf("%s(%s)", funcName, formatParams(safeParams))
	entry := NewEntry(LevelDebug, l.demoName, msg).WithCategory("INPUT")
	entry.WithFields(safeParams)
	AddToBuffer(entry)

	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] INPUT: %s(%s)", funcName, formatParams(safeParams))
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

	safeDetails := sanitizeFields(details)
	var msg string
	if len(safeDetails) > 0 {
		msg = fmt.Sprintf("%s: %s (%s)", operation, errMsg, formatParams(safeDetails))
	} else {
		msg = fmt.Sprintf("%s: %s", operation, errMsg)
	}

	entry := NewEntry(LevelError, l.demoName, msg).WithError(err)
	entry.WithField("operation", operation).WithFields(safeDetails)
	AddToBuffer(entry)

	if len(safeDetails) > 0 {
		l.Printf("[ERROR] %s: %s (%s)", operation, errMsg, formatParams(safeDetails))
	} else {
		l.Printf("[ERROR] %s: %s", operation, errMsg)
	}
}

func sanitizeFields(params map[string]interface{}) map[string]interface{} {
	if len(params) == 0 {
		return params
	}
	out := make(map[string]interface{}, len(params))
	for k, v := range params {
		if isSensitiveKey(k) {
			out[k] = "[REDACTED]"
			continue
		}
		out[k] = v
	}
	return out
}

func isSensitiveKey(k string) bool {
	k = strings.ToLower(strings.TrimSpace(k))
	return strings.Contains(k, "password") || strings.Contains(k, "passwd") || strings.Contains(k, "secret") || strings.Contains(k, "token") || strings.Contains(k, "api_key") || strings.Contains(k, "apikey")
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

// Global convenience functions now in manager.go

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
	if override := strings.TrimSpace(os.Getenv("FECIM_LOGS_DIR")); override != "" {
		return override
	}

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
