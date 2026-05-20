package internal

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestStructuredLogger(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer

	// Create a JSON handler that writes to our buffer
	jsonHandler := slog.NewJSONHandler(&logBuffer, nil)
	logger := slog.New(jsonHandler)

	// Create a test HTTP handler that returns a specific status code
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("test response"))
	})

	// Wrap our test handler with the StructuredLogger middleware
	middlewareHandler := StructuredLogger(logger)(testHandler)

	// Create a test request
	req := httptest.NewRequest("POST", "/test-path", nil)
	req.Header.Set("User-Agent", "test-agent")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	middlewareHandler.ServeHTTP(rr, req)

	// Check the response
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	if body := rr.Body.String(); body != "test response" {
		t.Errorf("handler returned unexpected body: got %v want %v", body, "test response")
	}

	// Parse the log output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(logBuffer.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify log fields
	if msg, ok := logEntry["msg"].(string); !ok || msg != "WEB_REQUEST" {
		t.Errorf("Expected log message 'WEB_REQUEST', got %v", msg)
	}

	if method, ok := logEntry["method"].(string); !ok || method != "POST" {
		t.Errorf("Expected method 'POST', got %v", method)
	}

	if path, ok := logEntry["path"].(string); !ok || path != "/test-path" {
		t.Errorf("Expected path '/test-path', got %v", path)
	}

	if status, ok := logEntry["status"].(float64); !ok || int(status) != http.StatusCreated {
		t.Errorf("Expected status %d, got %v", http.StatusCreated, status)
	}

	if userAgent, ok := logEntry["user_agent"].(string); !ok || userAgent != "test-agent" {
		t.Errorf("Expected user_agent 'test-agent', got %v", userAgent)
	}

	if _, ok := logEntry["duration_ms"]; !ok {
		t.Error("Expected duration field in log output")
	}
}

func TestResponseWriterWriteHeader(t *testing.T) {
	// Create a mock ResponseWriter
	mockRW := httptest.NewRecorder()

	// Create our responseWriter wrapper
	rw := &responseWriter{
		ResponseWriter: mockRW,
		status:         http.StatusOK, // Default status
	}

	// Call WriteHeader with a different status
	rw.WriteHeader(http.StatusNotFound)

	// Check that our wrapper captured the status
	if rw.status != http.StatusNotFound {
		t.Errorf("responseWriter did not capture status code: got %v want %v",
			rw.status, http.StatusNotFound)
	}

	// Check that the underlying ResponseWriter got the status
	if mockRW.Code != http.StatusNotFound {
		t.Errorf("underlying ResponseWriter did not get status code: got %v want %v",
			mockRW.Code, http.StatusNotFound)
	}
}

func TestResponseWriterWrite(t *testing.T) {
	// Create a mock ResponseWriter
	mockRW := httptest.NewRecorder()

	// Create our responseWriter wrapper
	rw := &responseWriter{
		ResponseWriter: mockRW,
		status:         http.StatusOK, // Default status
	}

	// Test data to write
	testData := []byte("test data")

	// Call Write method
	n, err := rw.Write(testData)

	// Check for errors
	if err != nil {
		t.Errorf("responseWriter.Write returned error: %v", err)
	}

	// Check that the correct number of bytes was reported as written
	if n != len(testData) {
		t.Errorf("responseWriter.Write returned wrong byte count: got %v want %v",
			n, len(testData))
	}

	// Check that the data was written to the underlying ResponseWriter
	if mockRW.Body.String() != string(testData) {
		t.Errorf("responseWriter did not write data correctly: got %v want %v",
			mockRW.Body.String(), string(testData))
	}

	// Check that status remains unchanged
	if rw.status != http.StatusOK {
		t.Errorf("responseWriter status changed unexpectedly: got %v want %v",
			rw.status, http.StatusOK)
	}
}

func TestStructuredLoggerWithDefaultStatus(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer

	// Create a JSON handler that writes to our buffer
	jsonHandler := slog.NewJSONHandler(&logBuffer, nil)
	logger := slog.New(jsonHandler)

	// Create a test HTTP handler that doesn't explicitly set a status code
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("default status test"))
	})

	// Wrap our test handler with the StructuredLogger middleware
	middlewareHandler := StructuredLogger(logger)(testHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/default-status", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	middlewareHandler.ServeHTTP(rr, req)

	// Check the response has default status 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Parse the log output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(logBuffer.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Verify the status in the log is 200 OK
	if status, ok := logEntry["status"].(float64); !ok || int(status) != http.StatusOK {
		t.Errorf("Expected status %d, got %v", http.StatusOK, status)
	}

	// Verify remote address is logged
	if remote, ok := logEntry["remote"].(string); !ok || remote != "192.168.1.1:12345" {
		t.Errorf("Expected remote '192.168.1.1:12345', got %v", remote)
	}
}

func TestStructuredLogger_HealthCheck(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		logHealth string
		shouldLog bool
	}{
		{
			name:      "Health check should not log by default",
			path:      "/health",
			logHealth: "",
			shouldLog: false,
		},
		{
			name:      "Health check should not log when LOG_HEALTH is false",
			path:      "/health",
			logHealth: "false",
			shouldLog: false,
		},
		{
			name:      "Health check should log when LOG_HEALTH is true",
			path:      "/health",
			logHealth: "true",
			shouldLog: true,
		},
		{
			name:      "Other paths should always log",
			path:      "/other",
			logHealth: "false",
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save current env and restore after test
			oldLogHealth, exists := os.LookupEnv("LOG_HEALTH")
			defer func() {
				if exists {
					os.Setenv("LOG_HEALTH", oldLogHealth)
				} else {
					os.Unsetenv("LOG_HEALTH")
				}
			}()

			if tt.logHealth != "" {
				os.Setenv("LOG_HEALTH", tt.logHealth)
			} else {
				os.Unsetenv("LOG_HEALTH")
			}

			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := StructuredLogger(logger)
			finalHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()

			finalHandler.ServeHTTP(rr, req)

			hasLog := buf.Len() > 0
			if hasLog != tt.shouldLog {
				t.Errorf("expected log: %v, got: %v for path %s and LOG_HEALTH=%s", tt.shouldLog, hasLog, tt.path, tt.logHealth)
			}
		})
	}
}
