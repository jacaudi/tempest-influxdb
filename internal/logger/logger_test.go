package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/jacaudi/tempest_influx/internal/config"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
		want   slog.Level
	}{
		{
			name:   "default info level",
			config: &config.Config{Debug: false},
			want:   slog.LevelInfo,
		},
		{
			name:   "debug level when debug enabled",
			config: &config.Config{Debug: true},
			want:   slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.config)

			if logger == nil {
				t.Fatal("New() returned nil logger")
			}

			if logger.Logger == nil {
				t.Fatal("Logger.Logger is nil")
			}

			// Test that the logger is working by capturing output
			var buf bytes.Buffer

			// Create a new logger with a custom handler for testing
			handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: tt.want,
			})
			testLogger := &AppLogger{Logger: slog.New(handler)}

			// Test logging at different levels
			testLogger.Info("test info message")
			testLogger.Debug("test debug message")

			output := buf.String()

			// Info should always be present
			if !strings.Contains(output, "test info message") {
				t.Error("Info message not found in output")
			}

			// Debug should only be present when debug level is set
			hasDebugOutput := strings.Contains(output, "test debug message")
			expectDebugOutput := tt.want == slog.LevelDebug

			if hasDebugOutput != expectDebugOutput {
				if expectDebugOutput {
					t.Error("Expected debug message in output when debug enabled")
				} else {
					t.Error("Debug message found in output when debug disabled")
				}
			}
		})
	}
}

func TestNewLoggerJSONHandler(t *testing.T) {

	// Capture the output from the logger
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := &AppLogger{Logger: slog.New(handler)}

	logger.Info("test json message", "key", "value", "number", 42)

	output := buf.String()

	// Verify it's valid JSON
	var jsonData map[string]any
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Check for expected fields
	if jsonData["msg"] != "test json message" {
		t.Errorf("Expected msg='test json message', got %v", jsonData["msg"])
	}

	if jsonData["key"] != "value" {
		t.Errorf("Expected key='value', got %v", jsonData["key"])
	}

	if jsonData["number"].(float64) != 42 {
		t.Errorf("Expected number=42, got %v", jsonData["number"])
	}
}

func TestNewLoggerTextHandler(t *testing.T) {
	// Capture the output from the logger
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &AppLogger{Logger: slog.New(handler)}

	logger.Debug("test text message", "key", "value")

	output := buf.String()

	// Verify it contains expected text format elements
	if !strings.Contains(output, "test text message") {
		t.Error("Message not found in text output")
	}

	if !strings.Contains(output, "key=value") {
		t.Error("Key-value pair not found in text output")
	}
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := &AppLogger{Logger: slog.New(handler)}

	// Test all log levels
	logger.Debug("debug message", "level", "debug")
	logger.Info("info message", "level", "info")
	logger.Warn("warn message", "level", "warn")
	logger.Error("error message", "level", "error")

	output := buf.String()

	expectedMessages := []string{
		"debug message",
		"info message",
		"warn message",
		"error message",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected message '%s' not found in output", msg)
		}
	}
}

func TestLoggerWithStructuredFields(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := &AppLogger{Logger: slog.New(handler)}

	logger.Info("structured message",
		slog.String("string_field", "test"),
		slog.Int("int_field", 123),
		slog.Bool("bool_field", true),
		slog.Float64("float_field", 3.14),
	)

	output := buf.String()

	var jsonData map[string]any
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify structured fields
	if jsonData["string_field"] != "test" {
		t.Errorf("Expected string_field='test', got %v", jsonData["string_field"])
	}

	if jsonData["int_field"].(float64) != 123 {
		t.Errorf("Expected int_field=123, got %v", jsonData["int_field"])
	}

	if jsonData["bool_field"] != true {
		t.Errorf("Expected bool_field=true, got %v", jsonData["bool_field"])
	}

	if jsonData["float_field"].(float64) != 3.14 {
		t.Errorf("Expected float_field=3.14, got %v", jsonData["float_field"])
	}
}

// Benchmark tests
func BenchmarkLoggerInfo(b *testing.B) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := &AppLogger{Logger: slog.New(handler)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i, "data", "test")
	}
}

func BenchmarkLoggerJSON(b *testing.B) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := &AppLogger{Logger: slog.New(handler)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i, "data", "test")
	}
}
