package net

import (
	"context"
)

// Caller calls external service endpoints.
type Caller interface {
	// Call establishes a communication with the given endpoint, sending the input bytes as payload.
	// If there is any response, it will be returned as a slice of bytes.
	// Implementations should expect to receive the target service endpoint name through the `endpoint` parameter.
	Call(ctx context.Context, endpoint string, in []byte) ([]byte, error)
}
