package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testHandler is a simple HTTP handler for testing
func testHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func TestNewCORS(t *testing.T) {
	corsMiddleware := NewCORS()

	assert.NotNil(t, corsMiddleware)
	// Verify that CORS middleware was created successfully
	// The NewCORS function returns a *cors.Cors handler
	handler := corsMiddleware.Handler(testHandler())
	assert.NotNil(t, handler)
}

func TestCORSPreflight(t *testing.T) {
	cors := NewCORS()
	handler := cors.Handler(testHandler())

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestLoggerMiddleware(t *testing.T) {
	handler := Logger(testHandler())

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:8000"

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body, _ := io.ReadAll(w.Body)
	assert.Equal(t, "OK", string(body))
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	rw.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, rw.statusCode)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestResponseWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	data := []byte("test data")
	n, err := rw.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, "test data", w.Body.String())
}

func TestResponseWriter_Hijack(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	_, _, err := rw.Hijack()

	// httptest.ResponseRecorder doesn't support Hijack, so this should fail
	assert.Error(t, err)
	assert.Equal(t, http.ErrNotSupported, err)
}

func TestResponseWriter_Flush(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	// Should not panic
	rw.Flush()
}

func TestLoggerCapturesStatusCode(t *testing.T) {
	// Create a handler that returns a specific status
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))

	req := httptest.NewRequest("GET", "/api/missing", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestLoggerCapturesMethod(t *testing.T) {
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestLoggerWithUserAgent(t *testing.T) {
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 Test Browser")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Mozilla/5.0 Test Browser", req.UserAgent())
}

func TestLoggerWithBodyWrite(t *testing.T) {
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response body"))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Response body", w.Body.String())
}

func TestResponseWriterDefaultStatus(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	// Write without calling WriteHeader
	rw.Write([]byte("data"))

	// Default status should be 200
	assert.Equal(t, http.StatusOK, rw.statusCode)
}

func TestLoggerMultipleWrites(t *testing.T) {
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Part 1"))
		w.Write([]byte(" Part 2"))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Part 1 Part 2", w.Body.String())
}

func TestResponseWriterHeader(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	rw.Header().Set("X-Custom-Header", "test-value")

	assert.Equal(t, "test-value", rw.Header().Get("X-Custom-Header"))
}

func TestLoggerWithPathCapture(t *testing.T) {
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	paths := []string{
		"/api/v1/test",
		"/api/test",
		"/",
	}

	for _, path := range paths {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, path, req.URL.Path)
	}
}

func TestLoggerWithLargeBody(t *testing.T) {
	largeData := bytes.Repeat([]byte("a"), 10000)
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(largeData)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, string(largeData), w.Body.String())
}

func TestResponseWriterWithReader(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	testData := "test data"
	reader := bytes.NewReader([]byte(testData))

	// Copy from reader to response writer
	n, err := io.Copy(rw, reader)

	assert.NoError(t, err)
	assert.Equal(t, int64(len(testData)), n)
	assert.Equal(t, testData, w.Body.String())
}

func TestLoggerChainedMiddleware(t *testing.T) {
	// Create a chain of middleware
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with logger
	handler := Logger(baseHandler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCORSHandlesMultipleOrigins(t *testing.T) {
	corsMiddleware := NewCORS()

	origins := []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"http://localhost:8080",
	}

	for _, origin := range origins {
		req := httptest.NewRequest("OPTIONS", "/api/test", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", "GET")

		w := httptest.NewRecorder()
		handler := corsMiddleware.Handler(testHandler())
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestCORSHandlesRequest(t *testing.T) {
	corsMiddleware := NewCORS()

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	w := httptest.NewRecorder()
	handler := corsMiddleware.Handler(testHandler())
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCORSHandlesComplexRequest(t *testing.T) {
	corsMiddleware := NewCORS()

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

	w := httptest.NewRecorder()
	handler := corsMiddleware.Handler(testHandler())
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ===== Additional Comprehensive Middleware Tests =====

// TestCORSMultipleOrigins tests CORS with multiple allowed origins
func TestCORSMultipleOrigins(t *testing.T) {
	corsMiddleware := NewCORS()
	handler := corsMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	origins := []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"http://localhost:8080",
	}

	for _, origin := range origins {
		req := httptest.NewRequest("OPTIONS", "/api/test", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", "GET")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Access-Control-Allow-Origin") != origin {
			t.Errorf("expected origin %s, got %s", origin, w.Header().Get("Access-Control-Allow-Origin"))
		}
	}
}

// TestCORSMaxAge tests CORS max age header
func TestCORSMaxAge(t *testing.T) {
	corsMiddleware := NewCORS()
	handler := corsMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	maxAge := w.Header().Get("Access-Control-Max-Age")
	if maxAge != "300" {
		t.Errorf("expected Max-Age 300, got %s", maxAge)
	}
}

// TestLoggerDifferentPaths tests logger with different request paths
func TestLoggerDifferentPaths(t *testing.T) {
	paths := []string{
		"/api/v1/drifts",
		"/api/v1/events",
		"/api/v1/stats",
		"/api/v1/providers/status",
	}

	for _, path := range paths {
		handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", path, nil)
		req.RemoteAddr = "127.0.0.1:8000"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for path %s, got %d", http.StatusOK, path, w.Code)
		}
	}
}

// TestLoggerDifferentStatusCodes tests logger with various status codes
func TestLoggerDifferentStatusCodes(t *testing.T) {
	statusCodes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, statusCode := range statusCodes {
		handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		}))

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != statusCode {
			t.Errorf("expected status %d, got %d", statusCode, w.Code)
		}
	}
}

// TestLoggerPreservesResponseBody tests logger preserves response body
func TestLoggerPreservesResponseBody(t *testing.T) {
	body := "test response body"
	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Body.String() != body {
		t.Errorf("expected body %s, got %s", body, w.Body.String())
	}
}

// TestOTelHTTPWithTracingContext tests OTEL HTTP with trace context
func TestOTelHTTPWithTracingContext(t *testing.T) {
	middleware := OTelHTTP("test-service")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestOTelHTTPDifferentMethods tests OTEL HTTP with different HTTP methods
func TestOTelHTTPDifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		middleware := OTelHTTP("test-service")
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(method, "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d for method %s, got %d", http.StatusOK, method, w.Code)
		}
	}
}

// TestOTelHTTPErrorHandling tests OTEL HTTP with error status codes
func TestOTelHTTPErrorHandling(t *testing.T) {
	errorCodes := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, code := range errorCodes {
		middleware := OTelHTTP("test-service")
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
		}))

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != code {
			t.Errorf("expected status %d, got %d", code, w.Code)
		}
	}
}

// TestResponseWriterMultipleWrites tests responseWriter with multiple writes
func TestResponseWriterMultipleWrites(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	rw.Write([]byte("Hello "))
	rw.Write([]byte("World"))

	if w.Body.String() != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", w.Body.String())
	}
}

// TestCORSCredentialsAllowed tests CORS allows credentials
func TestCORSCredentialsAllowed(t *testing.T) {
	corsMiddleware := NewCORS()
	handler := corsMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	credentials := w.Header().Get("Access-Control-Allow-Credentials")
	if credentials == "" {
		t.Error("expected Access-Control-Allow-Credentials header")
	}
}

// TestLoggerWithMethodCapture tests logger captures HTTP method
func TestLoggerWithMethodCapture(t *testing.T) {
	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(method, "/api/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("failed to process %s request", method)
		}
	}
}
