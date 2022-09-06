package retry

import (
	"context"
	"errors"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

func TestRetrySuite(t *testing.T) {
	suite.Run(t, &RetryTestSuite{})
}

type RetryTestSuite struct {
	suite.Suite
}

func (s *RetryTestSuite) SetupSuite() {
}

func (s *RetryTestSuite) TestRetryFunctionResolves() {
	i := 0
	expected := 3

	f := func() error {
		if i < expected {
			i++
			return errors.New("i is not expected value")
		}
		return nil
	}

	s.Assert().NoError(Retry(context.Background(), time.Millisecond*1, f))
	s.Assert().Equal(expected, i)
}

func (s *RetryTestSuite) TestRetryFunctionTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
	defer cancel()

	f := func() error {
		return errors.New("always fail")
	}

	s.Assert().Error(Retry(ctx, time.Millisecond*1, f))
}

func (s *RetryTestSuite) TestRetryFunctionCancel() {
	ctx, cancel := context.WithCancel(context.Background())

	f := func() error {
		return errors.New("always fail")
	}

	go func() {
		time.Sleep(time.Millisecond * 5)
		cancel()
	}()

	err := Retry(ctx, time.Millisecond*1, f)

	s.Assert().Error(err)
}
