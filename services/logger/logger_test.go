package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewLogger()

	tests := []struct {
		name     string
		logFunc  func()
		contains []string
	}{
		{
			name: "Debug log",
			logFunc: func() {
				logger.Debug("debug message", "key", "value")
			},
			contains: []string{"debug message", "key", "value", "DEBUG"},
		},
		{
			name: "Info log",
			logFunc: func() {
				logger.Info("info message", "count", 42)
			},
			contains: []string{"info message", "count", "42", "INFO"},
		},
		{
			name: "Error log",
			logFunc: func() {
				logger.Error("error message", "err", "failed")
			},
			contains: []string{"error message", "err", "failed", "ERROR"},
		},
		{
			name: "With context",
			logFunc: func() {
				contextLogger := logger.With("requestID", "123")
				contextLogger.Info("context message")
			},
			contains: []string{"context message", "requestID", "123", "INFO"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outC := make(chan string)
			go func() {
				var buf bytes.Buffer
				_, _ = buf.ReadFrom(r)
				outC <- buf.String()
			}()

			tt.logFunc()
			w.Close()
			os.Stdout = old
			output := <-outC

			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Fatalf("Failed to parse JSON log output: %v", err)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected log output to contain %q, got %q", expected, output)
				}
			}
		})
	}
}

func TestNewTestLogger(t *testing.T) {
	logger := NewTestLogger()
	if logger == nil {
		t.Error("NewTestLogger returned nil")
	}

	// Verify it doesn't panic when used
	logger.Debug("test debug")
	logger.Info("test info")
	logger.Warn("test warn")
	logger.Error("test error")
	logger.With("key", "value").Info("test with context")
}
