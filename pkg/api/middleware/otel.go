// Package middleware provides HTTP middleware for the API.
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// OTelHTTP returns middleware that creates spans for HTTP requests.
// It extracts trace context from incoming headers and records request metadata.
func OTelHTTP(serviceName string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract trace context from incoming request headers
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.target", r.URL.Path),
					attribute.String("http.host", r.Host),
					attribute.String("http.user_agent", r.UserAgent()),
				),
			)
			defer span.End()

			// Wrap response writer to capture status code
			wrw := &otelResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			start := time.Now()
			next.ServeHTTP(wrw, r.WithContext(ctx))
			duration := time.Since(start)

			span.SetAttributes(
				attribute.Int("http.status_code", wrw.statusCode),
				attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
			)

			if wrw.statusCode >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrw.statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}
		})
	}
}

type otelResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the HTTP response status code.
func (w *otelResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
