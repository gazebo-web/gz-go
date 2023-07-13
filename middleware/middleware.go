package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/gazebo-web/auth/pkg/authentication"
	"github.com/golang-jwt/jwt/v5/request"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	grpc_metadata "google.golang.org/grpc/metadata"
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
			_, err = verify(r.Context(), token)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to verify token: %s", err), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

const (
	metadataSubjectKey = "subject"
	metadataEmailKey   = "email"
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

// ExtractGRPCAuthEmail extracts the custom email (email) claim from the context metadata. This claim is
// usually injected in a middleware such as BearerToken or BearerAuthFuncGRPC, if present.
//
// This claim is expected in those provider that inject an email address in their JWT. Not all providers
// do such thing.
//
// This function only works with gRPC requests. It returns an error if the metadata couldn't be parsed or the email
// is not present.
func ExtractGRPCAuthEmail(ctx context.Context) (string, error) {
	return extractGRPCMetadata(ctx, metadataEmailKey)
}

// InjectGRPCAuthEmail injects the custom email (email) claim into the given context metadata.
// See ExtractGRPCAuthSubject for information on how to extract this value.
func InjectGRPCAuthEmail(ctx context.Context, email string) context.Context {
	return injectGRPCMetadata(ctx, metadataEmailKey, email)
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
