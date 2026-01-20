package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"microservice-template/config"
)

func TestRecovery(t *testing.T) {
	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("test panic")
	})

	// Wrap with recovery middleware
	handler := Recovery()(panicHandler)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	// Response should contain error message
	body := w.Body.String()
	if body == "" {
		t.Error("expected error response body, got empty")
	}
}

func TestRecovery_NoError(t *testing.T) {
	// Create a normal handler
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with recovery middleware
	handler := Recovery()(normalHandler)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Should work normally
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "success" {
		t.Errorf("expected body 'success', got '%s'", w.Body.String())
	}
}

func TestLogger(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate processing time
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("logged"))
	})

	// Wrap with logger middleware
	handler := Logger()(testHandler)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "logged" {
		t.Errorf("expected body 'logged', got '%s'", w.Body.String())
	}
}

func TestLogger_CapturesStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		statusFunc func(http.ResponseWriter)
	}{
		{"OK", http.StatusOK, func(w http.ResponseWriter) { w.WriteHeader(http.StatusOK) }},
		{"Created", http.StatusCreated, func(w http.ResponseWriter) { w.WriteHeader(http.StatusCreated) }},
		{"BadRequest", http.StatusBadRequest, func(w http.ResponseWriter) { w.WriteHeader(http.StatusBadRequest) }},
		{"NotFound", http.StatusNotFound, func(w http.ResponseWriter) { w.WriteHeader(http.StatusNotFound) }},
		{"InternalServerError", http.StatusInternalServerError, func(w http.ResponseWriter) { w.WriteHeader(http.StatusInternalServerError) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				time.Sleep(10 * time.Millisecond)
				tt.statusFunc(w)
			})

			handler := Logger()(testHandler)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestCors_Disabled(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.CORSConfig{
		Enabled: false,
	}

	handler := Cors(cfg)(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should not add CORS headers when disabled
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("CORS headers present when disabled")
	}
}

func TestCors_Enabled(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.CORSConfig{
		Enabled:        true,
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         3600,
	}

	handler := Cors(cfg)(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should add CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected CORS origin header, got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected CORS methods header")
	}

	if w.Header().Get("Access-Control-Max-Age") != "3600" {
		t.Errorf("expected max age 3600, got '%s'", w.Header().Get("Access-Control-Max-Age"))
	}
}

func TestCors_WildcardOrigin(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.CORSConfig{
		Enabled:        true,
		AllowedOrigins: []string{"*"},
	}

	handler := Cors(cfg)(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://anysite.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should allow any origin with wildcard
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected CORS origin header for wildcard")
	}
}

func TestCors_PreflightRequest(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.CORSConfig{
		Enabled:        true,
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"},
	}

	handler := Cors(cfg)(testHandler)
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Preflight should return 204 No Content
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d for preflight, got %d", http.StatusNoContent, w.Code)
	}
}

func TestRateLimit_Disabled(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.RateLimitConfig{
		Enabled: false,
	}

	handler := RateLimit(cfg)(testHandler)

	// Should not apply rate limiting
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i, http.StatusOK, w.Code)
		}
	}
}

func TestRateLimit_Enabled(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 2, // 2 requests per second
		Burst:          2, // burst of 2
	}

	handler := RateLimit(cfg)(testHandler)

	// First 2 requests should succeed (burst)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i, http.StatusOK, w.Code)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d for rate limited request, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.RateLimitConfig{
		Enabled:        true,
		RequestsPerSec: 1,
		Burst:          1,
	}

	handler := RateLimit(cfg)(testHandler)

	// Requests from different IPs should not affect each other
	ips := []string{"192.168.1.1:1234", "192.168.1.2:1234", "192.168.1.3:1234"}

	for _, ip := range ips {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request from %s: expected status %d, got %d", ip, http.StatusOK, w.Code)
		}
	}
}
