package logging

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		verbose   bool
		accessLog bool
	}{
		{"default", false, true},
		{"verbose", true, true},
		{"no access log", false, false},
		{"all options", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.verbose, tt.accessLog)
			if logger == nil {
				t.Fatal("NewLogger() returned nil")
			}
			if logger.AccessLog() != tt.accessLog {
				t.Errorf("AccessLog() = %v, want %v", logger.AccessLog(), tt.accessLog)
			}
		})
	}
}

func TestLoggerWithRequestID(t *testing.T) {
	logger := NewLogger(false, true)
	withID := logger.WithRequestID("test-123")

	if withID == nil {
		t.Fatal("WithRequestID() returned nil")
	}
	if withID.AccessLog() != logger.AccessLog() {
		t.Error("AccessLog() changed after WithRequestID()")
	}
}

func TestLogRequest(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewLogger(false, true)
	ctx := context.WithValue(context.Background(), RequestIDKey, "test-request-123")

	logger.LogRequest(ctx, LogRequestParams{
		Method:       "GET",
		Host:         "8000.localhost",
		TargetPort:   8000,
		StatusCode:   200,
		Latency:      100 * time.Millisecond,
		BytesWritten: 1234,
		Error:        nil,
	})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	output := buf.String()
	
	expectedParts := []string{
		"method=GET",
		"host=8000.localhost",
		"status=200",
		"request_id=test-request-123",
		"msg=\"request completed\"",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("log output missing expected part: %s, output: %s", part, output)
		}
	}
}

func TestLogRequestDisabled(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewLogger(false, false)
	ctx := context.Background()

	logger.LogRequest(ctx, LogRequestParams{
		Method:       "GET",
		Host:         "8000.localhost",
		TargetPort:   8000,
		StatusCode:   200,
		Latency:      100 * time.Millisecond,
		BytesWritten: 1234,
		Error:        nil,
	})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if buf.Len() > 0 {
		t.Error("LogRequest() wrote output when access log disabled")
	}
}

func TestContextKey(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, "my-request-id")

	value := ctx.Value(RequestIDKey)
	if value != "my-request-id" {
		t.Errorf("Context value = %v, want my-request-id", value)
	}
}

func TestLoggerMethods(t *testing.T) {
	logger := NewLogger(true, true)

	logger.ServerStarting(":443", "cert.pem", "key.pem", false)
	logger.ServerStarted(":443")
	logger.CertGenerated("cert.pem", "key.pem")
	logger.PortDenied("req-1", 22, "denied by policy")
	logger.InvalidHost("req-2", "bad.host", "invalid format")
	logger.ProxyError("req-3", 8000, errors.New("connection refused"))
	logger.WebSocketUpgrade("req-4", 8080)
}
