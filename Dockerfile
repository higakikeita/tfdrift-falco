# TFDrift-Falco Production Dockerfile
# Multi-stage build for optimized production image

# ============================================
# Stage 1: Build stage
# ============================================
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /build

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimization flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o tfdrift \
    ./cmd/tfdrift

# ============================================
# Stage 2: Runtime stage
# ============================================
FROM alpine:3.23.2

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget

# Create non-root user for security
RUN addgroup -g 1000 tfdrift && \
    adduser -D -u 1000 -G tfdrift tfdrift

# Create necessary directories
RUN mkdir -p /app /config /data && \
    chown -R tfdrift:tfdrift /app /config /data

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/tfdrift .

# Copy example config (optional)
COPY --chown=tfdrift:tfdrift examples/config.yaml ./config.example.yaml

# Switch to non-root user
USER tfdrift

# Create volumes for persistent data
VOLUME ["/config", "/data"]

# Expose ports
EXPOSE 8080 9090

# Health check for API server
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default: Run in server mode
ENTRYPOINT ["./tfdrift"]
CMD ["--server", "--api-port", "8080"]
