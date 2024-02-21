package net

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"google.golang.org/grpc"
)

// Server is a web server to listen to incoming requests. It supports different types of transport mechanisms used through Gazebo projects.
// It currently supports HTTP and gRPC servers.
type Server struct {
	// httpServer holds a group of HTTP servers. All these servers will be listening to different ports.
	httpServers []*http.Server
	// grpcServers holds a group of gRPC servers. All these servers will be listening in the same port.
	grpcServers []*grpc.Server
	// listener holds a reference to a TCP listener for gRPC servers.
	// All gRPC servers will use this listener to listen for incoming requests.
	listener net.Listener
}

// ListenAndServe starts listening for incoming requests for the different HTTP and gRPC servers.
// Returns a channel that will receive an error from any of the current underlying transport mechanisms.
// Each underlying server will be launched in a different go routine.
func (s *Server) ListenAndServe() <-chan error {
	errs := make(chan error, 1)

	// Start HTTP servers
	for _, srv := range s.httpServers {
		go func(srv *http.Server, errs chan<- error) {
			if err := srv.ListenAndServe(); err != nil {
				errs <- err
				close(errs)
			}
		}(srv, errs)
	}

	// Start gRPC servers
	if s.grpcServers != nil && len(s.grpcServers) > 0 && s.listener == nil {
		errs <- errors.New("a listener must be provided to the Server through an Option to support gRPC")
		return errs
	}
	for _, srv := range s.grpcServers {
		go func(srv *grpc.Server, listener net.Listener, errs chan<- error) {
			if err := srv.Serve(listener); err != nil {
				errs <- err
				close(errs)
			}
		}(srv, s.listener, errs)
	}

	return errs
}

// Close gracefully closes all the underlying servers (HTTP & gRPC).
func (s *Server) Close() {
	for _, srv := range s.httpServers {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		if err := srv.Shutdown(ctx); err != nil {
			log.Println("Failed to shutdown HTTP server:", err)
		}
		cancel()
	}
	for _, srv := range s.grpcServers {
		srv.GracefulStop()
	}
}

// Option contains logic that can be passed to a server in its initializer to modify it.
type Option func(*Server) *Server

// HTTP initializes a new HTTP server to serve endpoints defined in handler on a certain port.
func HTTP(handler http.Handler, port uint) Option {
	return func(s *Server) *Server {
		s.httpServers = append(s.httpServers, &http.Server{
			Handler: handler,
			Addr:    fmt.Sprintf(":%d", port),
		})

		return s
	}
}

// GRPC initializes and adds a new gRPC server to the Server.
// Multiple gRPC servers use the same ListenerTCP. ListenerTCP is required if this Option is passed.
// A set of gRPC interceptors for streams and unaries are passed as arguments in order to inject custom middlewares
// to the gRPC server. When calling this function, a set of default interceptors are already added to the gRPC server.
// Please check the NewServerOptionsGRPC implementation.
func GRPC(register func(s grpc.ServiceRegistrar), streams []grpc.StreamServerInterceptor, unaries []grpc.UnaryServerInterceptor) Option {
	return func(s *Server) *Server {
		opts := NewServerOptionsGRPC(streams, unaries)
		grpcServer := newGRPC(opts)
		register(grpcServer)
		s.grpcServers = append(s.grpcServers, grpcServer)
		return s
	}
}

// DefaultStreamInterceptorsGRPC defines the base streams interceptors we usually use for our gRPC servers.
func DefaultStreamInterceptorsGRPC() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		grpc_recovery.StreamServerInterceptor(),
		grpc_validator.StreamServerInterceptor(),
	}
}

// GenerateStreamServerInterceptorsChainWithBase appends the given interceptors to the base interceptors defined in DefaultStreamInterceptorsGRPC.
func GenerateStreamServerInterceptorsChainWithBase(interceptors ...grpc.StreamServerInterceptor) grpc.ServerOption {
	interceptors = append(DefaultStreamInterceptorsGRPC(), interceptors...)
	return grpc.ChainStreamInterceptor(interceptors...)
}

// DefaultUnaryInterceptorsGRPC defines the base streams interceptors we usually use for our gRPC servers.
func DefaultUnaryInterceptorsGRPC() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(),
		grpc_validator.UnaryServerInterceptor(),
	}
}

// GenerateUnaryServerInterceptorsChainWithBase appends the given interceptors to the base interceptors defined in DefaultUnaryInterceptorsGRPC.
func GenerateUnaryServerInterceptorsChainWithBase(interceptors ...grpc.UnaryServerInterceptor) grpc.ServerOption {
	interceptors = append(DefaultUnaryInterceptorsGRPC(), interceptors...)
	return grpc.ChainUnaryInterceptor(interceptors...)
}

// DefaultServerOptionsGRPC is a predefined set of gRPC server options that can be used by any Server when calling the GRPC function.
// We encourage you to create or extend your own opts function taking this one as a starting point.
func DefaultServerOptionsGRPC() []grpc.ServerOption {
	return NewServerOptionsGRPC(nil, nil)
}

// NewServerOptionsGRPC initializes a new set of ServerOption with the given streams and unaries interceptors.
// Calling this function already uses
func NewServerOptionsGRPC(streams []grpc.StreamServerInterceptor, unaries []grpc.UnaryServerInterceptor) []grpc.ServerOption {
	return []grpc.ServerOption{
		GenerateStreamServerInterceptorsChainWithBase(streams...),
		GenerateUnaryServerInterceptorsChainWithBase(unaries...),
	}
}

// newGRPC initializes a new gRPC server and uses the options received from calling opts().
func newGRPC(opts []grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(
		opts...,
	)
}

// ListenerTCP initializes a new Listener for server.
// This Option is required for GRPC servers when registered in Server.
func ListenerTCP(port uint) Option {
	return func(s *Server) *Server {
		var err error
		s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			panic(fmt.Sprintf("failed to initialize listener: %s", err))
		}
		return s
	}
}

// NewServer initializes a new server.
func NewServer(opts ...Option) *Server {
	var s Server
	for _, o := range opts {
		o(&s)
	}

	return &s
}

// ServiceRegistrator registers a gRPC service in a gRPC server.
type ServiceRegistrator interface {
	// Register registers the current service into the given server.
	Register(s grpc.ServiceRegistrar)
}
