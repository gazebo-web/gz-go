package middleware

import (
	"context"
	"github.com/gazebo-web/auth/pkg/authentication"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
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
		raw, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}
		token, err := auth.VerifyJWT(ctx, raw)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return injectClaims(ctx, token)

	}
}

// injectClaims injects claims from token into the given context.
// It returns the context with the claims injected, or an error if the claims could not be obtained.
// Values injected:
//   - Subject (mandatory): jwt.Claims.
//   - Email address (optional): authentication.EmailClaimer.
func injectClaims(ctx context.Context, token jwt.Claims) (context.Context, error) {
	sub, err := token.GetSubject()
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	ctx = InjectGRPCAuthSubject(ctx, sub)

	// Only inject the email claim if the given token fulfills the authentication.EmailClaimer interface.
	if emailClaim, ok := token.(authentication.EmailClaimer); ok {
		email, err := emailClaim.GetEmail()
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		ctx = InjectGRPCAuthEmail(ctx, email)
	}
	return ctx, nil
}
