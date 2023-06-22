package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/gazebo-web/auth/pkg/authentication"
	"github.com/golang-jwt/jwt/v5/request"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"google.golang.org/grpc/codes"
	grpc_metadata "google.golang.org/grpc/metadata"
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
		sub, err := token.GetSubject()
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return InjectGRPCAuthSubject(ctx, sub), nil
	}
}

const (
	metadataSubjectKey = "subject"
)

// ExtractGRPCAuthSubject extracts the authentication subject (sub) claim from the context metadata. This claim is
// usually injected in a middleware such as BearerToken or BearerAuthFuncGRPC, if present.
//
// From the RFC7519, section 4.1.2: https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.2
//
//	The "sub" (subject) claim identifies the principal that is the subject of the JWT. The claims in a JWT are normally
//	statements about the subject. The subject value MUST either be scoped to be locally unique in the context of the
//	issuer or be globally unique. The processing of this claim is generally application specific. The "sub" value is a
//	case-sensitive string containing a StringOrURI value.
//
// This function only works with gRPC requests. It returns an error if the metadata couldn't be parsed or the subject
// is not present.
func ExtractGRPCAuthSubject(ctx context.Context) (string, error) {
	return extractGRPCMetadata(ctx, metadataSubjectKey)
}

// InjectGRPCAuthSubject injects the authentication subject (sub) claim into the given context metadata.
// See ExtractGRPCAuthSubject for information on how to extract this value.
func InjectGRPCAuthSubject(ctx context.Context, sub string) context.Context {
	return injectGRPCMetadata(ctx, metadataSubjectKey, sub)
}

// extractGRPCMetadata extracts the first value of the given key. This only works for gRPC servers, not clients.
func extractGRPCMetadata(ctx context.Context, key string) (string, error) {
	md, ok := grpc_metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("failed to get context metadata")
	}
	var value string
	if len(md.Get(key)) > 0 {
		value = md.Get(key)[0]
	} else {
		return "", fmt.Errorf("context metadata does not have a value for key: %s", key)
	}
	return value, nil
}

// injectGRPCMetadata injects the given key and value into a context using grpc metadata. This only works for
// gRPC servers, not clients.
func injectGRPCMetadata(ctx context.Context, key string, value string) context.Context {
	md, ok := grpc_metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	md.Set(key, value)
	return metadata.MD(md).ToIncoming(ctx)
}
