package middleware

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/gazebo-web/auth/pkg/authentication"
	"github.com/golang-jwt/jwt/v5"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	grpc_test "github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	grpc_metadata "google.golang.org/grpc/metadata"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
		"sub": "gazebo-web",
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

func TestAuthFuncGRPC(t *testing.T) {
	auth := newTestAuthentication()
	ss := &TestAuthSuite{
		auth: auth,
		InterceptorTestSuite: &grpc_test.InterceptorTestSuite{
			TestService: auth,
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(BearerAuthFuncGRPC(auth))),
				grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(BearerAuthFuncGRPC(auth))),
			},
		},
	}
	suite.Run(t, ss)
}

type TestAuthSuite struct {
	*grpc_test.InterceptorTestSuite
	auth *testAuthService
}

func (suite *TestAuthSuite) TestVerifyJWT_FailsNoBearer() {
	ctx := context.Background()

	client := suite.NewClient()
	res, err := client.Ping(ctx, &grpc_test.PingRequest{})
	suite.Assert().Error(err)
	suite.Assert().ErrorContains(err, "unauthenticated with bearer")
	suite.Assert().Nil(res)
}

func (suite *TestAuthSuite) TestVerifyJWT_FailsVerifyJWTError() {
	ctx := ctxWithToken(context.Background(), "bearer", "1234")
	expectedError := errors.New("failed to verify token")

	expectedCtx := mock.AnythingOfType("*context.valueCtx")
	suite.auth.On("VerifyJWT", expectedCtx, "1234").Return(jwt.Claims(&jwt.MapClaims{}), expectedError).Once()
	client := suite.NewClient()

	res, err := client.Ping(ctx, &grpc_test.PingRequest{})
	suite.Assert().Error(err)
	suite.Assert().ErrorContains(err, expectedError.Error())
	suite.Assert().Nil(res)
}

func (suite *TestAuthSuite) TestVerifyJWT_Success() {
	ctx := ctxWithToken(context.Background(), "bearer", "1234")

	expectedCtx := mock.AnythingOfType("*context.valueCtx")
	testToken := authentication.NewFirebaseTestToken()
	suite.auth.On("VerifyJWT", expectedCtx, "1234").Return(authentication.NewFirebaseClaims(testToken), error(nil)).Once()
	client := suite.NewClient()

	res, err := client.Ping(ctx, &grpc_test.PingRequest{})
	suite.Assert().NoError(err)
	suite.Assert().NotNil(res)
	suite.Assert().NotEmpty(res.GetValue())
	suite.Assert().Equal("gazebo-web;test@gazebosim.org", res.GetValue())
}

func ctxWithToken(ctx context.Context, scheme string, token string) context.Context {
	md := grpc_metadata.Pairs("authorization", fmt.Sprintf("%s %v", scheme, token))
	return metadata.MD(md).ToOutgoing(ctx)
}

type testAuthService struct {
	grpc_test.UnimplementedTestServiceServer
	mock.Mock
}

func (s *testAuthService) Ping(ctx context.Context, _ *grpc_test.PingRequest) (*grpc_test.PingResponse, error) {
	sub, err := ExtractGRPCAuthSubject(ctx)
	if err != nil {
		return nil, err
	}
	email, err := ExtractGRPCAuthEmail(ctx)
	if err != nil {
		return nil, err
	}
	return &grpc_test.PingResponse{
		Value: strings.Join([]string{sub, email}, ";"),
	}, nil
}

func (s *testAuthService) VerifyJWT(ctx context.Context, token string) (jwt.Claims, error) {
	args := s.Called(ctx, token)
	return args.Get(0).(jwt.Claims), args.Error(1)
}

func newTestAuthentication() *testAuthService {
	return &testAuthService{}
}
