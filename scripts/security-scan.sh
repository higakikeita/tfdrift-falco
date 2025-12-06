#!/bin/bash
# Security scanning script for local development

set -e

echo "=== TFDrift-Falco Security Scan ==="
echo ""

# Check if gosec is installed
if ! command -v gosec &> /dev/null; then
    echo "📦 Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

# Run gosec
echo "🔍 Running GoSec security scanner..."
gosec -fmt=text -exclude=G104 ./...
echo "✅ GoSec scan complete"
echo ""

# Check for nancy
if ! command -v nancy &> /dev/null; then
    echo "📦 Installing nancy..."
    go install github.com/sonatype-nexus-community/nancy@latest
fi

# Run nancy
echo "🔍 Running Nancy dependency scanner..."
go list -json -deps ./... | nancy sleuth
echo "✅ Nancy scan complete"
echo ""

# Check for vulnerabilities in go.mod
echo "🔍 Checking Go module vulnerabilities..."
go list -json -m all | go run golang.org/x/vuln/cmd/govulncheck@latest -mode=binary ./...
echo "✅ Vulnerability check complete"
echo ""

echo "🎉 All security scans passed!"
