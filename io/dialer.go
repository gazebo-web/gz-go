package io

import (
	"context"
)

// Dialer holds a method to dial an endpoint.
type Dialer interface {
	// Dial establishes a communication with the given endpoint, sending the input bytes as payload.
	// If there is any response, it will be returned as a slice of bytes.
	Dial(ctx context.Context, endpoint string, in []byte) ([]byte, error)
}
