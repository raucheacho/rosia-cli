package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// DebugLevel for detailed debugging information
	DebugLevel LogLevel = iota
	// InfoLevel for general informational messages
	InfoLevel
	// WarnLevel for warning messages
	WarnLevel
	// ErrorLevel for error messages
	ErrorLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
)

// Logger provides structured logging with color-coded output
type Logger struct {
	mu          sync.Mutex
	level       LogLevel
	output      io.Writer
	colorOutput bool
	verbose     bool
}

// defaultLogger is the global logger instance
var (
	defaultLogger *Logger
	once          sync.Once
)

// init initializes the default logger
func init() {
	defaultLogger = New(InfoLevel, os.Stdout, true)
}

// New creates a new Logger instance
func New(level LogLevel, output io.Writer, colorOutput bool) *Logger {
	return &Logger{
		level:       level,
		output:      output,
		colorOutput: colorOutput,
		verbose:     false,
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetVerbose enables or disables verbose (debug) logging
func (l *Logger) SetVerbose(verbose bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.verbose = verbose
	if verbose {
		l.level = DebugLevel
	}
}

// log writes a log message with the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Skip if level is below threshold
	if level < l.level {
		return
	}

	// Format timestamp
	timestamp := time.Now().Format("15:04:05")

	// Get color for level
	color := l.getColor(level)

	// Format message
	message := fmt.Sprintf(format, args...)

	// Build log line
	var logLine string
	if l.colorOutput {
		logLine = fmt.Sprintf("%s[%s]%s %s%s%s %s\n",
			colorGray, timestamp, colorReset,
			color, level.String(), colorReset,
			message)
	} else {
		logLine = fmt.Sprintf("[%s] %s %s\n", timestamp, level.String(), message)
	}

	// Write to output
	fmt.Fprint(l.output, logLine)
}

// getColor returns the ANSI color code for a log level
func (l *Logger) getColor(level LogLevel) string {
	if !l.colorOutput {
		return ""
	}

	switch level {
	case ErrorLevel:
		return colorRed
	case WarnLevel:
		return colorYellow
	case InfoLevel:
		return colorBlue
	case DebugLevel:
		return colorGray
	default:
		return colorReset
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

// Global logger functions

// SetLevel sets the minimum log level for the default logger
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetVerbose enables or disables verbose logging for the default logger
func SetVerbose(verbose bool) {
	defaultLogger.SetVerbose(verbose)
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// GetDefaultLogger returns the default logger instance
func GetDefaultLogger() *Logger {
	return defaultLogger
}
