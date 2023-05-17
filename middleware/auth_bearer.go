package middleware

import (
	"context"
	"github.com/gazebo-web/auth/pkg/authentication"
	"github.com/golang-jwt/jwt/v5/request"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BearerToken returns a Middleware for authenticating users using Bearer Tokens in JWT format.
func BearerToken(authentication authentication.Authentication) Middleware {
	return newTokenMiddleware(authentication.VerifyJWT, request.BearerExtractor{})
}

// BearerAuthFuncGRPC returns a new grpc_auth.AuthFunc to use with the gazebo-web authentication library.
//
// The passed in context.Context will contain the gRPC metadata.MD object (for header-based authentication) and
// the peer.Peer information that can contain transport-based credentials (e.g. `credentials.AuthInfo`).
//
//	auth := authentication.New[...]()
//
//	srv := grpc.NewServer(
//		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(BearerAuthFuncGRPC(auth))),
//		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(BearerAuthFuncGRPC(auth))),
//	)
func BearerAuthFuncGRPC(auth authentication.Authentication) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}
		if err := auth.VerifyJWT(ctx, token); err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return ctx, nil
	}
}
