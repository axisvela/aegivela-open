package pepsdk

import (
	"errors"
	"fmt"
)

// ErrMalformed marks an authority response that cannot be used as a decision.
var ErrMalformed = errors.New("malformed authorization response")

// APIError records an HTTP authority response, unlike a transport failure.
// It intentionally omits response bodies, which may contain sensitive data.
type APIError struct {
	StatusCode int
	Err        error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("authorization HTTP status %d: %v", e.StatusCode, e.Err)
}

func (e *APIError) Unwrap() error { return e.Err }
