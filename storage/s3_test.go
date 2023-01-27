package storage

import (
	"context"
	"crypto/tls"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	s3api "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type s3StorageTestSuite struct {
	suite.Suite
	storage Storage
	server  *httptest.Server
	backend *s3mem.Backend
	faker   *gofakes3.GoFakeS3
	config  aws.Config
	client  *s3api.Client
}

func TestSuiteS3Storage(t *testing.T) {
	suite.Run(t, new(s3StorageTestSuite))
}

func (suite *s3StorageTestSuite) SetupSuite() {
	suite.backend = s3mem.New()
	suite.faker = gofakes3.New(suite.backend)
	suite.server = httptest.NewServer(suite.faker.Server())
	suite.config = suite.setupS3Config()
	suite.client = s3api.NewFromConfig(suite.config, func(o *s3api.Options) {
		o.UsePathStyle = true
	})
	suite.storage = NewS3(suite.client)
}

func (suite *s3StorageTestSuite) setupS3Config() aws.Config {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("KEY", "SECRET", "SESSION")),
		config.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(_, _ string, _ ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: suite.server.URL}, nil
			}),
		),
	)
	suite.Require().NoError(err)
	return cfg
}

func (suite *s3StorageTestSuite) SetupTest() {

}

func (suite *s3StorageTestSuite) TearDownTest() {
}

func (suite *s3StorageTestSuite) TearDownSuite() {
	suite.server.Close()
}
