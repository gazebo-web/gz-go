package net

import (
	"context"
	"errors"
	"gitlab.com/ignitionrobotics/web/ign-go/encoders"
	"reflect"
)

var (
	// ErrNilValuesIO is returned when either the input or the output are nil.
	ErrNilValuesIO = errors.New("nil input or output values")

	// ErrOutputMustBePointer is returned if the output is not a pointer.
	ErrOutputMustBePointer = errors.New("output must be a pointer")
)

// Client is a generic wrapper for creating API clients with different encodings and transport layers
type Client struct {
	// caller holds a transport layer implementation used to call to a certain endpoint.
	caller Caller

	// marshaller holds a encoders.Marshaller implementation to convert to and from an intermediate representation
	// to transfer through the wire.
	marshaller encoders.Marshaller
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

	body, err := c.marshaller.Marshal(in)
	if err != nil {
		return err
	}

	body, err = c.caller.Call(ctx, endpoint, body)
	if err != nil {
		return err
	}

	err = c.marshaller.Unmarshal(body, out)
	if err != nil {
		return err
	}

	return nil
}

// NewClient initializes a new Client using the given Caller and encoders.Marshaller.
func NewClient(caller Caller, serializer encoders.Marshaller) Client {
	return Client{
		caller:     caller,
		marshaller: serializer,
	}
}
