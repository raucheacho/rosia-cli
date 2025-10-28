package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		logFunc  func(*Logger, string, ...interface{})
		message  string
		expected bool // whether message should be logged
	}{
		{
			name:     "Debug message at Debug level",
			level:    DebugLevel,
			logFunc:  (*Logger).Debug,
			message:  "debug message",
			expected: true,
		},
		{
			name:     "Debug message at Info level",
			level:    InfoLevel,
			logFunc:  (*Logger).Debug,
			message:  "debug message",
			expected: false,
		},
		{
			name:     "Info message at Info level",
			level:    InfoLevel,
			logFunc:  (*Logger).Info,
			message:  "info message",
			expected: true,
		},
		{
			name:     "Warn message at Error level",
			level:    ErrorLevel,
			logFunc:  (*Logger).Warn,
			message:  "warn message",
			expected: false,
		},
		{
			name:     "Error message at Error level",
			level:    ErrorLevel,
			logFunc:  (*Logger).Error,
			message:  "error message",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := New(tt.level, buf, false) // Disable colors for testing

			tt.logFunc(logger, tt.message)

			output := buf.String()
			if tt.expected {
				if !strings.Contains(output, tt.message) {
					t.Errorf("Expected message %q to be logged, but it wasn't. Output: %s", tt.message, output)
				}
			} else {
				if strings.Contains(output, tt.message) {
					t.Errorf("Expected message %q NOT to be logged, but it was. Output: %s", tt.message, output)
				}
			}
		})
	}
}

func TestLogger_SetVerbose(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(InfoLevel, buf, false)

	// Debug should not be logged initially
	logger.Debug("debug message")
	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message should not be logged at Info level")
	}

	// Enable verbose mode
	buf.Reset()
	logger.SetVerbose(true)
	logger.Debug("debug message after verbose")
	if !strings.Contains(buf.String(), "debug message after verbose") {
		t.Error("Debug message should be logged after enabling verbose mode")
	}
}

func TestLogger_ColorOutput(t *testing.T) {
	tests := []struct {
		name        string
		level       LogLevel
		colorOutput bool
		expectColor bool
	}{
		{
			name:        "Error with color",
			level:       ErrorLevel,
			colorOutput: true,
			expectColor: true,
		},
		{
			name:        "Error without color",
			level:       ErrorLevel,
			colorOutput: false,
			expectColor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := New(tt.level, buf, tt.colorOutput)

			logger.Error("error message")

			output := buf.String()
			hasColor := strings.Contains(output, "\033[")

			if tt.expectColor && !hasColor {
				t.Error("Expected color codes in output, but none found")
			}
			if !tt.expectColor && hasColor {
				t.Error("Expected no color codes in output, but found some")
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGlobalLogger(t *testing.T) {
	// Test that global logger functions work
	buf := &bytes.Buffer{}
	defaultLogger = New(InfoLevel, buf, false)

	Info("test info message")
	if !strings.Contains(buf.String(), "test info message") {
		t.Error("Global Info() function should log message")
	}

	buf.Reset()
	Debug("test debug message")
	if strings.Contains(buf.String(), "test debug message") {
		t.Error("Global Debug() function should not log at Info level")
	}

	SetVerbose(true)
	buf.Reset()
	Debug("test debug after verbose")
	if !strings.Contains(buf.String(), "test debug after verbose") {
		t.Error("Global Debug() function should log after SetVerbose(true)")
	}
}
