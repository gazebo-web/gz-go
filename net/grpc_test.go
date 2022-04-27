package net

import (
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
	"testing"
)

func TestGRPCSuite(t *testing.T) {
	suite.Run(t, &GRPCTestSuite{})
}

type GRPCTestSuite struct {
	suite.Suite
}

func (suite *GRPCTestSuite) TestGRPCCallOK() {
	// nil should be true
	suite.Assert().True(GRPCCallOK(nil))

	// Status with an OK code should be true
	err := status.Error(codes.OK, "")
	suite.Assert().True(GRPCCallOK(err))

	// Status with an Unknown code should be false
	err = status.Error(codes.Unknown, "custom application error")
	suite.Assert().False(GRPCCallOK(err))

	// Status with a non-OK code should be false
	err = status.Error(codes.DeadlineExceeded, "deadline exceeded")
	suite.Assert().False(GRPCCallOK(err))
}

func (suite *GRPCTestSuite) TestGRPCCallSuccessful() {
	// nil should be true
	suite.Assert().True(GRPCCallSuccessful(nil))

	// Status with an OK code should be true
	err := status.Error(codes.OK, "")
	suite.Assert().True(GRPCCallSuccessful(err))

	// Status with an Unknown code should be true
	err = status.Error(codes.Unknown, "custom application error")
	suite.Assert().True(GRPCCallSuccessful(err))

	// Status with a non-OK code should be false
	err = status.Error(codes.DeadlineExceeded, "deadline exceeded")
	suite.Assert().False(GRPCCallSuccessful(err))
}

func (suite *GRPCTestSuite) TestListenAndServerAndConnectToGRPCServer() {
	server := grpc.NewServer()

	listener := ListenAndServeGRPC(server)
	defer server.Stop()

	conn, err := ConnectToGRPCServer(listener)
	suite.Require().NoError(err)

	suite.Assert().Equal(connectivity.Ready, conn.GetState())

}
