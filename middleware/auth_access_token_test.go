package middleware

import (
	"context"
	"testing"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	grpc_test "github.com/grpc-ecosystem/go-grpc-middleware/v2/testing/testpb"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpc_metadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthFuncGRPC_AccessToken(t *testing.T) {
	ss := &TestAuthAccessTokenSuite{
		InterceptorTestSuite: &grpc_test.InterceptorTestSuite{
			TestService: newTestAuthenticationAccessToken(),
			ServerOpts: []grpc.ServerOption{
				grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(BearerAccessTokenAuthFuncGRPC(validateAccessToken))),
				grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(BearerAccessTokenAuthFuncGRPC(validateAccessToken))),
			},
		},
	}
	suite.Run(t, ss)
}

func validateAccessToken(ctx context.Context, token string) error {
	if token != "ey.LfexACqSU5qgYgp9EXSdR4rtnD7BJ0oOCNi8BKIkZ4vt25jRxyu6AXAKVNrtItb1" {
		return status.Error(codes.Unauthenticated, "Invalid access token")
	}
	return nil
}

type TestAuthAccessTokenSuite struct {
	*grpc_test.InterceptorTestSuite
}

func (suite *TestAuthAccessTokenSuite) TestNoBearer() {
	ctx := context.Background()
	client := suite.NewClient()
	res, err := client.Ping(ctx, &grpc_test.PingRequest{})
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, status.Error(codes.Unauthenticated, "Request unauthenticated with bearer"))
	suite.Assert().Nil(res)
}

func (suite *TestAuthAccessTokenSuite) TestNoToken() {
	ctx := context.Background()
	md := grpc_metadata.Pairs("authorization", "bearer")
	ctx = metadata.MD(md).ToOutgoing(ctx)

	client := suite.NewClient()
	res, err := client.Ping(ctx, &grpc_test.PingRequest{})
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, status.Error(codes.Unauthenticated, "Bad authorization string"))
	suite.Assert().Nil(res)
}

func (suite *TestAuthAccessTokenSuite) TestInvalidScheme() {
	ctx := context.Background()
	ctx = ctxWithToken(ctx, "basic_auth", "test:test")

	client := suite.NewClient()
	res, err := client.Ping(ctx, &grpc_test.PingRequest{})
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, status.Error(codes.Unauthenticated, "Request unauthenticated with bearer"))
	suite.Assert().Nil(res)
}

func (suite *TestAuthAccessTokenSuite) TestInvalidToken() {
	ctx := context.Background()
	ctx = ctxWithToken(ctx, "bearer", "ey.FfexACqSU5qgYgp9EXSdR4rtnD7BJ0oOCNi8BKIkZ4vt25jRxyu6AXAKVNrtItb1")

	client := suite.NewClient()
	res, err := client.Ping(ctx, &grpc_test.PingRequest{})
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, status.Error(codes.Unauthenticated, "Invalid access token"))
	suite.Assert().Nil(res)
}

func (suite *TestAuthAccessTokenSuite) TestValidToken() {
	ctx := context.Background()
	ctx = ctxWithToken(ctx, "bearer", "ey.LfexACqSU5qgYgp9EXSdR4rtnD7BJ0oOCNi8BKIkZ4vt25jRxyu6AXAKVNrtItb1")

	client := suite.NewClient()
	res, err := client.Ping(ctx, &grpc_test.PingRequest{})
	suite.Assert().NoError(err)
	suite.Assert().NotNil(res)
}

type testAuthAccessToken struct {
	grpc_test.UnimplementedTestServiceServer
}

func (t *testAuthAccessToken) Ping(ctx context.Context, request *grpc_test.PingRequest) (*grpc_test.PingResponse, error) {
	return &grpc_test.PingResponse{}, nil
}
func newTestAuthenticationAccessToken() grpc_test.TestServiceServer {
	return &testAuthAccessToken{}
}
