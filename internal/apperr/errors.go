// Package apperr holds the sentinel errors shared by every layer.
//
// Adapters translate their own failures into these values; controllers
// translate these values into HTTP status codes (see platform/httpx).
package apperr

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalid        = errors.New("invalid input")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrConflict       = errors.New("conflict")
	ErrRateLimited    = errors.New("rate limited")
	ErrNotImplemented = errors.New("not implemented")
)
