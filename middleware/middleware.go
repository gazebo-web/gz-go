package middleware

import (
	"context"
	"fmt"
	"github.com/gazebo-web/auth/pkg/authentication"
	"github.com/golang-jwt/jwt/v5/request"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

// Extractor extracts a string value from an HTTP request. It's usually used to extract a header from an HTTP request,
// but can also be used for extracting a user and password from the body.
//
// There are a few implementations already provided by the request package, for example:
// Bearer tokens: request.BearerExtractor
type Extractor = request.Extractor

// Middleware is used to modify or augment the behavior of an HTTP request handler.
type Middleware func(http.Handler) http.Handler

// BearerToken returns a Middleware for authenticating users using Bearer Tokens in JWT format.
func BearerToken(authentication authentication.Authentication) Middleware {
	return newTokenMiddleware(authentication.VerifyJWT, request.BearerExtractor{})
}

// newTokenMiddleware initializes a generic middleware that uses token authentication. It attempts to extract the tokens from
// the HTTP request, and verifies the access token is valid to continue to the next element in the middleware chain.
func newTokenMiddleware(verify authentication.TokenAuthentication, extractors ...Extractor) Middleware {
	e := request.MultiExtractor(extractors)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := e.ExtractToken(r)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to extract token: %s", err), http.StatusBadRequest)
				return
			}
			if len(token) == 0 {
				http.Error(w, "No token provided", http.StatusBadRequest)
				return
			}
			err = verify(r.Context(), token)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to verify token: %s", err), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthFuncGRPC returns a new grpc_auth.AuthFunc to use with the gazebo-web authentication library.
//
// The passed in context.Context will contain the gRPC metadata.MD object (for header-based authentication) and
// the peer.Peer information that can contain transport-based credentials (e.g. `credentials.AuthInfo`).
//
//	auth := authentication.New[...]()
//
//	srv := grpc.NewServer(
//		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(AuthFuncGRPC(auth))),
//		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(AuthFuncGRPC(auth))),
//	)
func AuthFuncGRPC(auth authentication.Authentication) grpc_auth.AuthFunc {
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
