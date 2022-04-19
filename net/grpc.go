package net

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCCallOK processes a gRPC response error to verify that a gRPC call returned an OK status.
// The actual contents of the error are ignored. Only the call status error is checked.
func GRPCCallOK(err error) bool {
	return status.Code(err) == codes.OK
}

// GRPCCallSuccessful processes a gRPC response error to verify that a gRPC call returned an OK or Unknown status.
// The actual contents of the error are ignored. Only the call status error is checked.
//
// gRPC method implementations that return anything other than a *Status value as their error value (e.g. an error
// generated with errors.New()) will have the value wrapped inside a *Status value and its error code set to Unknown.
//
// This function considers calls with statuses with unknown errors successful. In addition, if the error value is
// not of type *Status, then the call is considered successful.
func GRPCCallSuccessful(err error) bool {
	code := status.Code(err)

	return code == codes.OK || code == codes.Unknown
}
