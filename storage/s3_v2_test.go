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

type s3v2StorageTestSuite struct {
	suite.Suite
	storage    Storage
	server     *httptest.Server
	backend    *s3mem.Backend
	faker      *gofakes3.GoFakeS3
	config     aws.Config
	client     *s3api.Client
	bucketName string
	fsStorage  Storage
}

func TestSuiteS3v2Storage(t *testing.T) {
	suite.Run(t, new(s3v2StorageTestSuite))
}

func (suite *s3v2StorageTestSuite) SetupSuite() {
	suite.backend = s3mem.New()
	suite.faker = gofakes3.New(suite.backend)
	suite.server = httptest.NewServer(suite.faker.Server())
	suite.config = suite.setupS3Config()
	suite.client = s3api.NewFromConfig(suite.config, func(o *s3api.Options) {
		o.UsePathStyle = true
	})
	suite.bucketName = "fuel"
	suite.storage = NewS3v2(suite.client, suite.bucketName)
	suite.fsStorage = newFilesystemStorage(basePath)

	suite.setupTestData()
}

func (suite *s3v2StorageTestSuite) setupTestData() {
	ctx := context.Background()
	_, err := suite.client.CreateBucket(ctx, &s3api.CreateBucketInput{Bucket: aws.String(suite.bucketName)})
	suite.Require().NoError(err)

	suite.Require().NoError(WalkDir(ctx, basePath, UploadFileS3(suite.client, suite.bucketName, nil)))
}

func (suite *s3v2StorageTestSuite) TearDownSuite() {
	ctx := context.Background()

	suite.Require().NoError(WalkDir(ctx, basePath, DeleteFileS3(suite.client, suite.bucketName, nil)))

	_, err := suite.client.DeleteBucket(ctx, &s3api.DeleteBucketInput{Bucket: aws.String(suite.bucketName)})
	suite.Require().NoError(err)
	suite.server.Close()
}

func (suite *s3v2StorageTestSuite) setupS3Config() aws.Config {
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

func (suite *s3v2StorageTestSuite) TestGetFile_InvalidResource() {
	r := invalidResource
	ctx := context.Background()
	content, err := suite.storage.GetFile(ctx, r, "model.sdf")
	suite.Assert().Error(err)
	suite.Assert().Empty(content)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *s3v2StorageTestSuite) TestGetFile_NotFound() {
	r := &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: version,
	}
	ctx := context.Background()
	content, err := suite.storage.GetFile(ctx, r, "model123.sdf")
	suite.Assert().Error(err)
	suite.Assert().Empty(content)
}

func (suite *s3v2StorageTestSuite) TestGetFile_Success() {
	r := &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: version,
	}
	ctx := context.Background()
	content, err := suite.storage.GetFile(ctx, r, "model.sdf")
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(content)

	expected, err := suite.storage.GetFile(ctx, r, "model.sdf")
	suite.Require().NoError(err)
	suite.Assert().Equal(expected, content)
}

func (suite *s3v2StorageTestSuite) TestDownload_InvalidResource() {
	r := invalidResource
	ctx := context.Background()
	content, err := suite.storage.Download(ctx, r)
	suite.Assert().Error(err)
	suite.Assert().Empty(content)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *s3v2StorageTestSuite) TestDownload_NotFound() {
	r := &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: 5,
	}
	ctx := context.Background()
	url, err := suite.storage.Download(ctx, r)
	suite.Assert().Error(err)
	suite.Assert().Empty(url)
}

func (suite *s3v2StorageTestSuite) TestDownload_Success() {
	r := &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: version,
	}
	ctx := context.Background()
	url, err := suite.storage.Download(ctx, r)
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(url)
	suite.Assert().Contains(url, ".zip")
}

func (suite *s3v2StorageTestSuite) TestUploadDir_InvalidResource() {
	r := invalidResource
	ctx := context.Background()
	err := suite.storage.UploadDir(ctx, r, "./testdata/example")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *s3v2StorageTestSuite) TestUploadDir_SourceNotFound() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example1234")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFolderNotFound)
}

func (suite *s3v2StorageTestSuite) TestUploadDir_SourceIsEmpty() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example_empty")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFolderEmpty)
}

func (suite *s3v2StorageTestSuite) TestUploadDir_NotFile() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example/meshes/turtle.dae")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFile)
}

func (suite *s3v2StorageTestSuite) TestUploadDir_Success() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example")
	suite.Assert().NoError(err)

	b, err := suite.storage.GetFile(ctx, r, "meshes/turtle.dae")
	suite.Require().NoError(err)
	suite.Assert().NotEmpty(b)

	suite.Require().NoError(WalkDir(ctx, "./testdata/example", DeleteFileS3(suite.client, suite.bucketName, nonExistentResource)))
}
