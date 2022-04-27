package net

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
)

// ListenAndServeGRPC receives a configured gRPC server and starts listening. This function does not block.
// You should have called the appropriate gRPC server configuration function before calling this.
// After calling this function, you should call `server.Stop()` when done with the server.
func ListenAndServeGRPC(server *grpc.Server) *bufconn.Listener {
	listener := bufconn.Listen(1024 * 1024)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalln("Failed while serving gRPC server:", err)
		}
	}()

	return listener
}

// ConnectToGRPCServer connects to a gRPC server and returns the established connection.
// The returned connection should be used to instance a service client.
func ConnectToGRPCServer(listener *bufconn.Listener) (*grpc.ClientConn, error) {
	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	return grpc.DialContext(
		context.Background(),
		"",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
	)
}

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
