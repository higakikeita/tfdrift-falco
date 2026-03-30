// Package errors defines sentinel errors for the tfdrift-falco application.
package errors

import "errors"

// Sentinel errors for common failure modes.
var (
	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthorized indicates the request lacks valid authentication.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates the request is not allowed.
	ErrForbidden = errors.New("forbidden")

	// ErrBadRequest indicates invalid input.
	ErrBadRequest = errors.New("bad request")

	// ErrTimeout indicates an operation timed out.
	ErrTimeout = errors.New("operation timed out")

	// ErrConfigInvalid indicates invalid configuration.
	ErrConfigInvalid = errors.New("invalid configuration")

	// ErrProviderUnavailable indicates a cloud provider API is unavailable.
	ErrProviderUnavailable = errors.New("provider unavailable")

	// ErrTerraformState indicates a Terraform state parsing error.
	ErrTerraformState = errors.New("terraform state error")
)
