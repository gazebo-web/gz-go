package net

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"time"
)

// Server is a web server to listen to incoming requests. It supports different types of transport mechanisms used through Ignition Robotics projects.
// It currently supports HTTP and gRPC servers.
type Server struct {
	// httpServer holds a group of HTTP servers. All these servers will be listening to different ports.
	httpServers []*http.Server
	// listener holds a reference to a TCP listener for GRPC servers. All gRPC servers will use this listener to listen for incoming requests.
	listener net.Listener
	// grpcServers holds a group of gRPC servers. All these servers will be listening in the same port.
	grpcServers []*grpc.Server
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

// Option defines a configuration option for the server.
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
// A set of gRPC options are passed as a function in order to inject custom middlewares to the gRPC server.
// DefaultServerOptionsGRPC contains a set of basic middlewares to use in any application, but we encourage you to create or extend
// your own opts function.
func GRPC(register func(s grpc.ServiceRegistrar), opts func() []grpc.ServerOption) Option {
	return func(s *Server) *Server {
		grpcServer := newGRPC(opts)
		register(grpcServer)
		s.grpcServers = append(s.grpcServers, grpcServer)
		return s
	}
}

// DefaultServerOptionsGRPC is a predefined set of gRPC server options that can be used by any Server when calling the GRPC function.
// We encourage you to create or extend your own opts function taking this one as a starting point.
func DefaultServerOptionsGRPC() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_opentracing.StreamServerInterceptor(),
				grpc_recovery.StreamServerInterceptor(),
				grpc_validator.StreamServerInterceptor(),
			),
		),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_opentracing.UnaryServerInterceptor(),
				grpc_recovery.UnaryServerInterceptor(),
				grpc_validator.UnaryServerInterceptor(),
			),
		),
	}
}

// newGRPC initializes a new gRPC server and uses the options received from calling opts().
func newGRPC(opts func() []grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(
		opts()...,
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

// ServiceRegister holds a method to register a gRPC service to a gRPC server.
type ServiceRegister interface {
	// Register registers the current service into the given server.
	Register(s grpc.ServiceRegistrar)
}
