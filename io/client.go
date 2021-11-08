package io

import (
	"context"
	"errors"
	"gitlab.com/ignitionrobotics/web/ign-go/encoders"
	"reflect"
	"time"
)

// ClientOptions is used to define Client settings such as Timeout and the URL.
type ClientOptions struct {
	// Timeout is the max amount of time that an HTTP client can wait until it decides to cancel a request.
	Timeout time.Duration
}

var (
	// ErrNilValuesIO is returned when either the input or the output are nil.
	ErrNilValuesIO = errors.New("nil input or output values")

	// ErrOutputMustBePointer is returned if the output is not a pointer.
	ErrOutputMustBePointer = errors.New("output must be a pointer")
)

// Client is a generic wrapper for creating API clients with different encodings and transport layers
type Client struct {
	// dialer holds a transport layer implementation used to dial to a certain endpoint.
	dialer Dialer

	// serializer holds a encoders.Serializer implementation to encode/decode to/from a specific format.
	serializer encoders.Serializer
}

// Call calls the given endpoint with the given input as payload. If there's a response back from the endpoint, it will be
// stored in the output variable.
func (c *Client) Call(ctx context.Context, endpoint string, in, out interface{}) error {
	if in == nil || out == nil {
		return ErrNilValuesIO
	}
	if reflect.ValueOf(out).Kind() != reflect.Ptr {
		return ErrOutputMustBePointer
	}

	body, err := c.serializer.Marshal(in)
	if err != nil {
		return err
	}

	body, err = c.dialer.Dial(ctx, endpoint, body)
	if err != nil {
		return err
	}

	err = c.serializer.Unmarshal(body, out)
	if err != nil {
		return err
	}

	return nil
}

// NewClient initializes a new Client using the given Dialer and encoders.Serializer.
func NewClient(dialer Dialer, serializer encoders.Serializer) Client {
	return Client{
		dialer:     dialer,
		serializer: serializer,
	}
}
