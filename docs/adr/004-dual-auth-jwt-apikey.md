# ADR-004: JWT + API Key Dual Authentication Strategy

## Status

Accepted

## Date

2026-03-22

## Context

TFDrift-Falco's REST API was initially open without authentication, suitable for development but not for enterprise deployment. We needed to add authentication that supports both:

- **Interactive users** accessing the Dashboard UI (session-based, time-limited)
- **Programmatic clients** (CI/CD pipelines, monitoring systems, scripts) that need persistent credentials

Options considered:

1. **JWT only** — Token-based auth for all clients
2. **API Key only** — Static keys for all clients
3. **OAuth 2.0** — Full OAuth flow with external identity provider
4. **JWT + API Key dual authentication** — JWT for interactive, API Key for programmatic

## Decision

We implement dual authentication supporting both JWT Bearer tokens and API Keys (`pkg/api/middleware/auth.go`):

- **JWT tokens** are issued via `POST /api/v1/auth/token` with configurable expiry and HMAC-SHA256 signing. They carry a subject (user identity) and are validated on each request.
- **API Keys** use a `tfd_` prefix with 32 random hex bytes, validated via the `X-API-Key` header with constant-time comparison. Each key has a name, optional scopes, and creation timestamp.
- When both are provided, JWT takes precedence.
- Authentication is disabled by default (`auth.enabled: false`) for development ease.

## Consequences

### Positive

- Two auth methods cover interactive and programmatic use cases
- JWT provides time-limited tokens suitable for browser sessions
- API Keys provide long-lived credentials for automation
- Scoped API Keys enable fine-grained access control
- Constant-time comparison prevents timing attacks on API Keys
- Auth can be disabled for development/testing

### Negative

- Two auth methods increase middleware complexity
- JWT secret must be securely managed (Kubernetes Secret, env var)
- API Keys stored in config file — no built-in rotation mechanism yet
- No integration with external identity providers (LDAP, OIDC) yet

### Neutral

- Rate limiting (`pkg/api/middleware/ratelimit.go`) uses the authenticated identity for per-client tracking
- Public endpoints (`/health`, `/version`) bypass authentication
