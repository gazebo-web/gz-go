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
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type s3v2StorageTestSuite struct {
	suite.Suite
	storage       Storage
	server        *httptest.Server
	backend       *s3mem.Backend
	faker         *gofakes3.GoFakeS3
	config        aws.Config
	client        *s3api.Client
	bucketName    string
	fsStorage     Storage
	presignClient *s3api.PresignClient
}

func TestSuiteS3Storage(t *testing.T) {
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
	suite.presignClient = s3api.NewPresignClient(suite.client)
	suite.bucketName = "fuel"
	suite.storage = NewS3v2(suite.client, suite.bucketName)
	suite.fsStorage = newFilesystemStorage(basePath)

	suite.setupTestData()
}

func (suite *s3v2StorageTestSuite) setupTestData() {
	ctx := context.Background()
	_, err := suite.client.CreateBucket(ctx, &s3api.CreateBucketInput{Bucket: aws.String(suite.bucketName)})
	suite.Require().NoError(err)

	suite.Require().NoError(suite.walkDirWithS3Func(ctx, suite.uploadFile))
}

func (suite *s3v2StorageTestSuite) TearDownSuite() {
	ctx := context.Background()

	suite.Require().NoError(suite.walkDirWithS3Func(ctx, suite.deleteFile))

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

func (suite *s3v2StorageTestSuite) walkDirWithS3Func(ctx context.Context, fn func(ctx context.Context, bucket, path string, body io.Reader) error) error {
	return filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		key, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		err = fn(ctx, suite.bucketName, key, f)
		if err != nil {
			return err
		}
		return nil
	})
}

func (suite *s3v2StorageTestSuite) uploadFile(ctx context.Context, bucket string, path string, body io.Reader) error {
	_, err := suite.client.PutObject(ctx, &s3api.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		Body:   body,
	})
	return err
}

func (suite *s3v2StorageTestSuite) deleteFile(ctx context.Context, bucket string, path string, _ io.Reader) error {
	_, err := suite.client.DeleteObject(ctx, &s3api.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})
	return err
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
