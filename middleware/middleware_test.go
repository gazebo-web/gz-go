package middleware

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/gazebo-web/auth/pkg/authentication"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestBearerToken_NoAuthorizationHeader(t *testing.T) {
	handler := setupBearerTokenTest(t)

	wr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "https://gazebosim.org", nil)

	handler.ServeHTTP(wr, r)
	assert.Equal(t, http.StatusBadRequest, wr.Code)
}

func TestBearerToken_EmptyAuthorizationHeader(t *testing.T) {
	handler := setupBearerTokenTest(t)

	wr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "https://gazebosim.org", nil)
	r.Header.Set("Authorization", "")

	handler.ServeHTTP(wr, r)
	assert.Equal(t, http.StatusBadRequest, wr.Code)
}

func TestBearerToken_InvalidAuthorizationHeader(t *testing.T) {
	handler := setupBearerTokenTest(t)

	wr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "https://gazebosim.org", nil)
	r.Header.Set("Authorization", "JWT test1234")

	handler.ServeHTTP(wr, r)
	assert.Equal(t, http.StatusBadRequest, wr.Code)
}

func TestBearerToken_EmptyBearerToken(t *testing.T) {
	handler := setupBearerTokenTest(t)

	wr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "https://gazebosim.org", nil)
	r.Header.Set("Authorization", "Bearer ")

	handler.ServeHTTP(wr, r)
	assert.Equal(t, http.StatusBadRequest, wr.Code)
}

func TestBearerToken_InvalidBearerTokenRandomString(t *testing.T) {
	handler := setupBearerTokenTest(t)

	wr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "https://gazebosim.org", nil)
	r.Header.Set("Authorization", "Bearer 1234")

	handler.ServeHTTP(wr, r)
	assert.Equal(t, http.StatusUnauthorized, wr.Code)
}

func TestBearerToken_InvalidBearerTokenSignedByAnother(t *testing.T) {
	handler := setupBearerTokenTest(t)

	wr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "https://gazebosim.org", nil)
	r.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")

	handler.ServeHTTP(wr, r)
	assert.Equal(t, http.StatusUnauthorized, wr.Code)
}

func TestBearerToken(t *testing.T) {
	handler := setupBearerTokenTest(t)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(map[string]interface{}{
		"exp": time.Hour,
		"iat": time.Now(),
	}))

	// Read private key in order to sign the JWT
	pk, err := os.ReadFile("./testdata/key.private.pem")
	require.NoError(t, err)

	// Use the PEM decoder and parse the private key
	block, _ := pem.Decode(pk)

	// Parse private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	// Sign the JWT
	signedToken, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// Create request
	r := httptest.NewRequest("GET", "https://gazebosim.org", nil)

	// Set authorization header with the recently signed token
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", signedToken))

	// Create a new response recorder
	wr := httptest.NewRecorder()

	handler.ServeHTTP(wr, r)
	assert.Equal(t, http.StatusOK, wr.Code)
}

func setupBearerTokenTest(t *testing.T) http.Handler {
	// Set up public key for Authentication service
	publicKey, err := os.ReadFile("./testdata/key.pem")
	require.NoError(t, err)

	// Set up a Bearer token middleware
	mw := BearerToken(authentication.NewAuth0(publicKey))

	// Define handler that will be wrapped by the middleware
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("It's all good, man!"))
		if err != nil {
			http.Error(w, "Test fail", http.StatusInternalServerError)
		}
	}))
	return handler
}
