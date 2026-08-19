# driftwire Production Dockerfile
# Multi-stage build for optimized production image

# ============================================
# Stage 1: Build stage
# ============================================
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /build

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimization flags.
# VERSION comes from the workflow (tag name); git describe is only a
# fallback and yields 'dev' in CI because .dockerignore excludes .git.
ARG VERSION=
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-s -w -X main.version=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')}" \
    -o driftwire \
    ./cmd/driftwire

# ============================================
# Stage 2: Runtime stage
# ============================================
FROM alpine:3.24.1

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget

# Create non-root user for security
RUN addgroup -g 1000 driftwire && \
    adduser -D -u 1000 -G driftwire driftwire

# Create necessary directories
RUN mkdir -p /app /config /data && \
    chown -R driftwire:driftwire /app /config /data

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/driftwire .

# Copy example config (optional)
COPY --chown=driftwire:driftwire config.yaml.example ./config.example.yaml

# Switch to non-root user
USER driftwire

# Create volumes for persistent data
VOLUME ["/config", "/data"]

# Expose ports
EXPOSE 8080 9090

# Health check for API server
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default: Run in server mode
ENTRYPOINT ["./driftwire"]
CMD ["--server", "--api-port", "8080"]
