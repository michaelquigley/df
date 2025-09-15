package dl

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "string input",
			input:    "test message",
			expected: "test message",
		},
		{
			name:     "error input",
			input:    errors.New("test error"),
			expected: "test error",
		},
		{
			name:     "int input",
			input:    42,
			expected: "42",
		},
		{
			name:     "nil input",
			input:    nil,
			expected: "<nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertFormattedMessage(t *testing.T) {
	tests := []struct {
		name     string
		format   any
		args     []any
		expected string
	}{
		{
			name:     "string format with args",
			format:   "hello %s",
			args:     []any{"world"},
			expected: "hello world",
		},
		{
			name:     "error format ignores args",
			format:   errors.New("test error"),
			args:     []any{"ignored"},
			expected: "test error",
		},
		{
			name:     "int format ignores args",
			format:   42,
			args:     []any{"ignored"},
			expected: "42",
		},
		{
			name:     "string format no args",
			format:   "no formatting",
			args:     []any{},
			expected: "no formatting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertFormattedMessage(tt.format, tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuilderLoggingMethods(t *testing.T) {
	// create a buffer to capture log output
	var buf bytes.Buffer
	
	// create a text handler that writes to our buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	
	// create a builder with our test logger
	builder := &Builder{
		logger: slog.New(handler),
		attrs:  []slog.Attr{},
	}

	testError := errors.New("test error message")

	tests := []struct {
		name      string
		logFunc   func(any)
		input     any
		level     string
		expectMsg string
	}{
		{
			name:      "Debug with string",
			logFunc:   builder.Debug,
			input:     "debug message",
			level:     "DEBUG",
			expectMsg: "debug message",
		},
		{
			name:      "Debug with error",
			logFunc:   builder.Debug,
			input:     testError,
			level:     "DEBUG",
			expectMsg: "test error message",
		},
		{
			name:      "Info with string",
			logFunc:   builder.Info,
			input:     "info message",
			level:     "INFO",
			expectMsg: "info message",
		},
		{
			name:      "Info with error",
			logFunc:   builder.Info,
			input:     testError,
			level:     "INFO",
			expectMsg: "test error message",
		},
		{
			name:      "Warn with string",
			logFunc:   builder.Warn,
			input:     "warn message",
			level:     "WARN",
			expectMsg: "warn message",
		},
		{
			name:      "Warn with error",
			logFunc:   builder.Warn,
			input:     testError,
			level:     "WARN",
			expectMsg: "test error message",
		},
		{
			name:      "Error with string",
			logFunc:   builder.Error,
			input:     "error message",
			level:     "ERROR",
			expectMsg: "error message",
		},
		{
			name:      "Error with error",
			logFunc:   builder.Error,
			input:     testError,
			level:     "ERROR",
			expectMsg: "test error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.input)
			
			output := buf.String()
			assert.Contains(t, output, tt.level)
			assert.Contains(t, output, tt.expectMsg)
		})
	}
}

func TestBuilderFormattedLoggingMethods(t *testing.T) {
	// create a buffer to capture log output
	var buf bytes.Buffer
	
	// create a text handler that writes to our buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	
	// create a builder with our test logger
	builder := &Builder{
		logger: slog.New(handler),
		attrs:  []slog.Attr{},
	}

	testError := errors.New("formatted error message")

	tests := []struct {
		name      string
		logFunc   func(any, ...any)
		format    any
		args      []any
		level     string
		expectMsg string
	}{
		{
			name:      "Debugf with string format",
			logFunc:   builder.Debugf,
			format:    "debug %s %d",
			args:      []any{"test", 42},
			level:     "DEBUG",
			expectMsg: "debug test 42",
		},
		{
			name:      "Debugf with error format",
			logFunc:   builder.Debugf,
			format:    testError,
			args:      []any{"ignored"},
			level:     "DEBUG",
			expectMsg: "formatted error message",
		},
		{
			name:      "Infof with string format",
			logFunc:   builder.Infof,
			format:    "info %s",
			args:      []any{"message"},
			level:     "INFO",
			expectMsg: "info message",
		},
		{
			name:      "Infof with error format",
			logFunc:   builder.Infof,
			format:    testError,
			args:      []any{"ignored"},
			level:     "INFO",
			expectMsg: "formatted error message",
		},
		{
			name:      "Warnf with string format",
			logFunc:   builder.Warnf,
			format:    "warn %v",
			args:      []any{123},
			level:     "WARN",
			expectMsg: "warn 123",
		},
		{
			name:      "Warnf with error format",
			logFunc:   builder.Warnf,
			format:    testError,
			args:      []any{"ignored"},
			level:     "WARN",
			expectMsg: "formatted error message",
		},
		{
			name:      "Errorf with string format",
			logFunc:   builder.Errorf,
			format:    "error %s occurred",
			args:      []any{"critical"},
			level:     "ERROR",
			expectMsg: "error critical occurred",
		},
		{
			name:      "Errorf with error format",
			logFunc:   builder.Errorf,
			format:    testError,
			args:      []any{"ignored"},
			level:     "ERROR",
			expectMsg: "formatted error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.format, tt.args...)
			
			output := buf.String()
			assert.Contains(t, output, tt.level)
			assert.Contains(t, output, tt.expectMsg)
		})
	}
}

func TestBuilderWithAttributes(t *testing.T) {
	// create a buffer to capture log output
	var buf bytes.Buffer
	
	// create a text handler that writes to our buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	
	// create a builder with our test logger
	builder := &Builder{
		logger: slog.New(handler),
		attrs:  []slog.Attr{},
	}

	testError := errors.New("test error with context")

	// test chained attributes with error logging
	builder.With("user", "john").With("action", "login").Error(testError)
	
	output := buf.String()
	assert.Contains(t, output, "ERROR")
	assert.Contains(t, output, "test error with context")
	assert.Contains(t, output, "user=john")
	assert.Contains(t, output, "action=login")
}

func TestBuilderNonFatalMethods(t *testing.T) {
	// test that non-fatal methods don't exit the process
	var buf bytes.Buffer
	
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	
	builder := &Builder{
		logger: slog.New(handler),
		attrs:  []slog.Attr{},
	}

	testError := errors.New("non-fatal error")
	
	// these should all complete without exiting
	builder.Debug(testError)
	builder.Info(testError)
	builder.Warn(testError)
	builder.Error(testError)
	
	// verify we got output for all levels
	output := buf.String()
	assert.Contains(t, output, "DEBUG")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "ERROR")
	
	// verify error message appears multiple times
	errorCount := strings.Count(output, "non-fatal error")
	assert.Equal(t, 4, errorCount)
}

// Note: Testing Fatal methods would require special setup to avoid actually exiting,
// so we're skipping those in this test suite. In a real-world scenario, you might
// want to test them with process isolation or by mocking os.Exit.

func TestBareLoggingFunctions(t *testing.T) {
	// create a buffer to capture log output
	var buf bytes.Buffer
	
	// configure logging to use our test buffer
	opts := &Options{
		Output:  &buf,
		UseJSON: false,
		Level:   slog.LevelDebug,
	}
	Init(opts)
	
	testError := errors.New("bare function test error")

	tests := []struct {
		name      string
		logFunc   func(any)
		input     any
		expectMsg string
	}{
		{
			name:      "bare Debug with string",
			logFunc:   Debug,
			input:     "bare debug message",
			expectMsg: "bare debug message",
		},
		{
			name:      "bare Debug with error",
			logFunc:   Debug,
			input:     testError,
			expectMsg: "bare function test error",
		},
		{
			name:      "bare Info with string",
			logFunc:   Info,
			input:     "bare info message",
			expectMsg: "bare info message",
		},
		{
			name:      "bare Info with error",
			logFunc:   Info,
			input:     testError,
			expectMsg: "bare function test error",
		},
		{
			name:      "bare Warn with string",
			logFunc:   Warn,
			input:     "bare warn message",
			expectMsg: "bare warn message",
		},
		{
			name:      "bare Warn with error",
			logFunc:   Warn,
			input:     testError,
			expectMsg: "bare function test error",
		},
		{
			name:      "bare Error with string",
			logFunc:   Error,
			input:     "bare error message",
			expectMsg: "bare error message",
		},
		{
			name:      "bare Error with error",
			logFunc:   Error,
			input:     testError,
			expectMsg: "bare function test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.input)
			
			output := buf.String()
			assert.Contains(t, output, tt.expectMsg)
		})
	}
}

func TestBareFormattedLoggingFunctions(t *testing.T) {
	// create a buffer to capture log output
	var buf bytes.Buffer
	
	// configure logging to use our test buffer
	opts := &Options{
		Output:  &buf,
		UseJSON: false,
		Level:   slog.LevelDebug,
	}
	Init(opts)
	
	testError := errors.New("bare formatted test error")

	tests := []struct {
		name      string
		logFunc   func(any, ...any)
		format    any
		args      []any
		expectMsg string
	}{
		{
			name:      "bare Debugf with string format",
			logFunc:   Debugf,
			format:    "bare debug %s %d",
			args:      []any{"test", 42},
			expectMsg: "bare debug test 42",
		},
		{
			name:      "bare Debugf with error format",
			logFunc:   Debugf,
			format:    testError,
			args:      []any{"ignored"},
			expectMsg: "bare formatted test error",
		},
		{
			name:      "bare Infof with string format",
			logFunc:   Infof,
			format:    "bare info %s",
			args:      []any{"message"},
			expectMsg: "bare info message",
		},
		{
			name:      "bare Infof with error format",
			logFunc:   Infof,
			format:    testError,
			args:      []any{"ignored"},
			expectMsg: "bare formatted test error",
		},
		{
			name:      "bare Warnf with string format",
			logFunc:   Warnf,
			format:    "bare warn %v",
			args:      []any{123},
			expectMsg: "bare warn 123",
		},
		{
			name:      "bare Warnf with error format",
			logFunc:   Warnf,
			format:    testError,
			args:      []any{"ignored"},
			expectMsg: "bare formatted test error",
		},
		{
			name:      "bare Errorf with string format",
			logFunc:   Errorf,
			format:    "bare error %s occurred",
			args:      []any{"critical"},
			expectMsg: "bare error critical occurred",
		},
		{
			name:      "bare Errorf with error format",
			logFunc:   Errorf,
			format:    testError,
			args:      []any{"ignored"},
			expectMsg: "bare formatted test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.format, tt.args...)
			
			output := buf.String()
			assert.Contains(t, output, tt.expectMsg)
		})
	}
}

func TestBareLoggingCallStack(t *testing.T) {
	// create a buffer to capture log output  
	var buf bytes.Buffer
	
	// configure logging to use our test buffer with pretty formatting
	opts := &Options{
		Output:          &buf,
		UseJSON:         false,
		Level:           slog.LevelDebug,
		TrimPrefix:      "",
		AbsoluteTime:    false,
		TimestampFormat: "15:04:05.000",
	}
	Init(opts)
	
	// create a test function that calls the bare logging functions
	testFunction := func() {
		Info("call stack test message")
	}
	
	buf.Reset()
	testFunction()
	
	output := buf.String()
	
	// verify the output contains the test function name, not the bare Info function
	assert.Contains(t, output, "call stack test message")
	// the output should show "testFunction" in the call stack, not "Info"
	assert.Contains(t, output, "TestBareLoggingCallStack")
	// it should NOT contain the bare function name "dl.Info"
	assert.NotContains(t, output, "dl.Info")
}