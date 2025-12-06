# Build stage
FROM golang:1.25.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o tfdrift ./cmd/tfdrift

# Runtime stage
FROM alpine:3.21

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/tfdrift .

# Copy example config
COPY examples/config.yaml ./config.example.yaml

# Create volume for config
VOLUME ["/config"]

# Expose metrics port (future)
EXPOSE 9090

ENTRYPOINT ["./tfdrift"]
CMD ["--config", "/config/config.yaml"]
