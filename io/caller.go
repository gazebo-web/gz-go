package io

import (
	"context"
)

// Caller holds a method to call an endpoint.
type Caller interface {
	// Call establishes a communication with the given endpoint, sending the input bytes as payload.
	// If there is any response, it will be returned as a slice of bytes.
	Call(ctx context.Context, endpoint string, in []byte) ([]byte, error)
}
