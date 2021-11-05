package io

import (
	"context"
	"errors"
	"gitlab.com/ignitionrobotics/web/ign-go/encoders"
	"net/url"
	"time"
)

// ClientOptions is used to define Client settings such as Timeout and the URL.
type ClientOptions struct {
	// Timeout is the max amount of time that an HTTP client can wait until it decides to cancel a request.
	Timeout time.Duration

	// URL is the URL where to client should point to.
	URL *url.URL
}

var (
	// ErrNilValuesIO is returned when either the input or the output are nil.
	ErrNilValuesIO = errors.New("nil input or output values")
	// ErrOutputMustBePointer is returned if the output is not a pointer.
	ErrOutputMustBePointer = errors.New("output must be a pointer")
)

// Client holds a method to call different service methods.
type Client interface {
	// Call performs a call to a certain service method.
	Call(ctx context.Context, method, endpoint string, format encoders.Format, in, out encoders.Serializer) error
}
