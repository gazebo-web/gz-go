package middleware

import (
	"fmt"
	"github.com/gazebo-web/auth/pkg/authentication"
	"github.com/golang-jwt/jwt/v5/request"
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
