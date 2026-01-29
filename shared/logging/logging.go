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

	// Shared log file for all loggers
	sharedLogFile   *os.File
	sharedLogWriter io.Writer
	sharedLogMu     sync.Mutex
	sharedLogPath   string
)

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

// Logger wraps log.Logger with demo-specific configuration and verbosity support
type Logger struct {
	*log.Logger
	logFile  *os.File
	demoName string
}

// NewLogger creates a new logger for the specified demo
// All loggers share a single log file to avoid creating multiple files
func NewLogger(demoName string) *Logger {
	sharedLogMu.Lock()
	defer sharedLogMu.Unlock()

	// Initialize shared log file if not already done
	if sharedLogWriter == nil {
		logsDir := getLogsDir()
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			// Fallback to stdout only
			sharedLogWriter = os.Stdout
		} else {
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			sharedLogPath = filepath.Join(logsDir, timestamp+"-fecim.log")

			var err error
			sharedLogFile, err = os.Create(sharedLogPath)
			if err != nil {
				// Fallback to stdout only
				sharedLogWriter = os.Stdout
			} else {
				// Write to both file and stdout
				sharedLogWriter = io.MultiWriter(os.Stdout, sharedLogFile)
			}
		}
	}

	logger := &Logger{
		Logger:   log.New(sharedLogWriter, "["+demoName+"] ", log.Ldate|log.Ltime|log.Lmicroseconds),
		logFile:  nil, // Don't store file reference - shared file is managed globally
		demoName: demoName,
	}

	// Only log the path once for the first logger
	if sharedLogPath != "" {
		logger.Printf("Logging to: %s", sharedLogPath)
		sharedLogPath = "" // Clear to avoid repeating
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
	if IsVerbose(VerbosityInfo) {
		l.Printf("[INFO] "+format, args...)
	}
}

// Debug logs at DEBUG level (verbosity >= 2) - for button clicks, value changes
func (l *Logger) Debug(format string, args ...interface{}) {
	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] "+format, args...)
	}
}

// Trace logs at TRACE level (verbosity >= 3) - for frequent updates
func (l *Logger) Trace(format string, args ...interface{}) {
	if IsVerbose(VerbosityTrace) {
		l.Printf("[TRACE] "+format, args...)
	}
}

// Button logs a button click event at DEBUG level
func (l *Logger) Button(buttonName string) {
	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] BUTTON: %s clicked", buttonName)
	}
}

// ValueChange logs a value change event at DEBUG level
func (l *Logger) ValueChange(widgetName string, oldValue, newValue interface{}) {
	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] VALUE: %s changed from %v to %v", widgetName, oldValue, newValue)
	}
}

// Selection logs a selection change event at DEBUG level
func (l *Logger) Selection(widgetName string, selected string) {
	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] SELECT: %s = %q", widgetName, selected)
	}
}

// SliderChange logs a slider value change at DEBUG level
func (l *Logger) SliderChange(sliderName string, value float64) {
	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] SLIDER: %s = %.4f", sliderName, value)
	}
}

// TabChange logs a tab selection change at DEBUG level
func (l *Logger) TabChange(tabName string) {
	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] TAB: switched to %q", tabName)
	}
}

// CheckboxChange logs a checkbox state change at DEBUG level
func (l *Logger) CheckboxChange(checkboxName string, checked bool) {
	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] CHECKBOX: %s = %v", checkboxName, checked)
	}
}

// EntryChange logs a text entry change at DEBUG level
func (l *Logger) EntryChange(entryName string, text string) {
	if IsVerbose(VerbosityDebug) {
		l.Printf("[DEBUG] ENTRY: %s = %q", entryName, text)
	}
}

// Calculation logs a physics/math calculation at TRACE level
// Format: [TRACE] CALC: funcName(param1=value1, param2=value2) = result
func (l *Logger) Calculation(funcName string, inputs map[string]interface{}, result interface{}) {
	if IsVerbose(VerbosityTrace) {
		l.Printf("[TRACE] CALC: %s(%s) = %v", funcName, formatParams(inputs), result)
	}
}

// Input logs function entry with parameters at TRACE level
// Format: [TRACE] INPUT: funcName(param1=value1, param2=value2)
func (l *Logger) Input(funcName string, params map[string]interface{}) {
	if IsVerbose(VerbosityTrace) {
		l.Printf("[TRACE] INPUT: %s(%s)", funcName, formatParams(params))
	}
}

// Output logs function return value at TRACE level
// Format: [TRACE] OUTPUT: funcName -> result
func (l *Logger) Output(funcName string, result interface{}) {
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
		log.Printf(format, v...)
	}
}

// Println logs to the default logger
func Println(v ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Println(v...)
	} else {
		log.Println(v...)
	}
}

// GlobalInfo logs at INFO level
func GlobalInfo(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	} else {
		log.Printf("[INFO] "+format, args...)
	}
}

// GlobalDebug logs at DEBUG level
func GlobalDebug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	} else if IsVerbose(VerbosityDebug) {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// GlobalError logs at ERROR level
func GlobalError(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Printf("[ERROR] "+format, args...)
	} else {
		log.Printf("[ERROR] "+format, args...)
	}
}

// GlobalCalculation logs a calculation at TRACE level using the default logger
func GlobalCalculation(funcName string, inputs map[string]interface{}, result interface{}) {
	if defaultLogger != nil {
		defaultLogger.Calculation(funcName, inputs, result)
	} else if IsVerbose(VerbosityTrace) {
		log.Printf("[TRACE] CALC: %s(%s) = %v", funcName, formatParams(inputs), result)
	}
}

// GlobalInput logs function entry at TRACE level using the default logger
func GlobalInput(funcName string, params map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.Input(funcName, params)
	} else if IsVerbose(VerbosityTrace) {
		log.Printf("[TRACE] INPUT: %s(%s)", funcName, formatParams(params))
	}
}

// GlobalOutput logs function return at TRACE level using the default logger
func GlobalOutput(funcName string, result interface{}) {
	if defaultLogger != nil {
		defaultLogger.Output(funcName, result)
	} else if IsVerbose(VerbosityTrace) {
		log.Printf("[TRACE] OUTPUT: %s -> %v", funcName, result)
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
