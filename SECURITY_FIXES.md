# Security Fixes - December 2025

## Overview

This release fixes critical security vulnerabilities in the Go standard library and updates all dependencies to their latest secure versions.

## Fixed Vulnerabilities

### Critical: crypto/x509 Vulnerabilities (Go stdlib)

#### GO-2025-4175: Improper DNS name constraint verification
- **Severity**: High (CVSS 7.5)
- **Component**: crypto/x509 in Go < 1.25.5
- **Impact**: Could allow TLS certificate validation bypass
- **Fix**: Updated Go from 1.23.0 → 1.24.0/1.25.5
- **Reference**: https://pkg.go.dev/vuln/GO-2025-4175

#### GO-2025-4155: Excessive resource consumption in error handling
- **Severity**: Medium (CVSS 5.3)
- **Component**: crypto/x509 in Go < 1.25.5
- **Impact**: Potential DoS via malicious certificates
- **Fix**: Updated Go from 1.23.0 → 1.24.0/1.25.5
- **Reference**: https://pkg.go.dev/vuln/GO-2025-4155

## Dependency Updates

| Package | Old → New | Reason |
|---------|-----------|---------|
| Go toolchain | 1.23.0 → 1.24.0/1.25.5 | Fix stdlib vulnerabilities |
| google.golang.org/grpc | 1.59.0 → 1.77.0 | Security updates |
| github.com/spf13/cobra | 1.8.0 → 1.10.2 | Security patches |
| github.com/spf13/viper | 1.18.2 → 1.21.0 | Security patches |
| golang.org/x/net | 0.43.0 → 0.47.0 | Security updates |
| golang.org/x/sys | 0.35.0 → 0.38.0 | Security patches |

## Docker Image Updates

- Builder: `golang:1.21-alpine` → `golang:1.25.5-alpine`
- Runtime: `alpine:latest` → `alpine:3.21` (pinned version)

## Verification

Run vulnerability scan:
```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Expected: `No vulnerabilities found.`

## AWS SDK Note

The scan reports 2 vulnerabilities in `aws-sdk-go` S3 Crypto SDK, but these do NOT affect our code as we don't use client-side S3 encryption. Future releases will migrate to AWS SDK v2.
