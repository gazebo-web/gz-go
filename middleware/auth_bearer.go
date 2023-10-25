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
func BearerAuthFuncGRPC(auth authentication.Authentication, claimInjector ClaimInjectorJWT) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		raw, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}
		token, err := auth.VerifyJWT(ctx, raw)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return claimInjector(ctx, token)
	}
}

// ClaimInjectorJWT allows authentication layers to inject JWT claims into a context.Context.
//
//	Rules when creating a new claim injector:
//	- Must always return ctx, even in error handlers.
//	- Claim validation might be required depending on the underlying jwt.Claims implementation.
type ClaimInjectorJWT func(ctx context.Context, token jwt.Claims) (context.Context, error)

// ClaimInjectorBehavior is used in combination with ClaimInjectorJWT when grouping
// different claim injectors by using GroupClaimInjectors.
type ClaimInjectorBehavior func(ctx context.Context, err error) (context.Context, error)

// groupClaimInjectors treats all the given injectors as a single one.
//
// By setting a mandatoryInjection, the resulting injector will early return
// if an error is found at any point during the claim injection.
//
// If optionalInjection is used instead, no errors will be returned during the
// claim injection.
//
// If no ClaimInjectorBehavior is provided, it defaults to mandatoryInjection.
//
// Example:
//
//	groupClaimInjectors(mandatoryInjection,
//		groupClaimInjectors(mandatoryInjection, SubjectClaimer),
//		groupClaimInjectors(optionalInjection, EmailClaimer),
//	)
//
// For public-accessible methods, check GroupMandatoryClaimInjectors and
// GroupOptionalClaimInjectors.
func groupClaimInjectors(behavior ClaimInjectorBehavior, injectors ...ClaimInjectorJWT) ClaimInjectorJWT {
	if behavior == nil {
		behavior = mandatoryInjection
	}
	return func(ctx context.Context, token jwt.Claims) (context.Context, error) {
		for _, injector := range injectors {
			var err error
			ctx, err = behavior(injector(ctx, token))
			if err != nil {
				return ctx, err
			}
		}
		return ctx, nil
	}
}

// mandatoryInjection forces a ClaimInjectorJWT to always return an error.
func mandatoryInjection(ctx context.Context, err error) (context.Context, error) {
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

// optionalInjection ignores any errors returned by a ClaimInjectorJWT.
func optionalInjection(ctx context.Context, _ error) (context.Context, error) {
	return ctx, nil
}

// GroupMandatoryClaimInjectors allows injecting mandatory ClaimInjectors as a single one.
// Check groupClaimInjectors to understand how grouping works.
func GroupMandatoryClaimInjectors(injectors ...ClaimInjectorJWT) ClaimInjectorJWT {
	return groupClaimInjectors(mandatoryInjection, injectors...)
}

// GroupOptionalClaimInjectors allows injecting optional ClaimInjectors as a single one.
// Check groupClaimInjectors to understand how grouping works.
func GroupOptionalClaimInjectors(injectors ...ClaimInjectorJWT) ClaimInjectorJWT {
	return groupClaimInjectors(optionalInjection, injectors...)
}

// SubjectClaimer is a ClaimInjectorJWT for the "sub" claim.
func SubjectClaimer(ctx context.Context, token jwt.Claims) (context.Context, error) {
	sub, err := token.GetSubject()
	if err != nil {
		return ctx, status.Error(codes.Unauthenticated, err.Error())
	}
	if len(sub) == 0 {
		return ctx, status.Error(codes.Unauthenticated, "empty subject claim")
	}
	return InjectGRPCAuthSubject(ctx, sub), nil
}

// EmailClaimer is a ClaimInjectorJWT for the "email" custom claim.
func EmailClaimer(ctx context.Context, token jwt.Claims) (context.Context, error) {
	emailClaim, ok := token.(authentication.EmailClaimer)
	if !ok {
		return ctx, status.Error(codes.Unauthenticated, "email not found in JWT")
	}
	email, err := emailClaim.GetEmail()
	if err != nil {
		return ctx, status.Error(codes.Unauthenticated, err.Error())
	}
	if len(email) == 0 {
		return ctx, status.Error(codes.Unauthenticated, "empty email claim")
	}
	return InjectGRPCAuthEmail(ctx, email), nil
}
