package storage

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	s3api "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/suite"
	"net/http/httptest"
	"os"
	"testing"
)

type s3v1StorageTestSuite struct {
	suite.Suite
	storage    Storage
	server     *httptest.Server
	backend    *s3mem.Backend
	faker      *gofakes3.GoFakeS3
	config     aws.Config
	client     *s3api.S3
	bucketName string
	fsStorage  Storage
	session    *session.Session
	uploader   *s3manager.Uploader
}

func TestSuiteS3v1Storage(t *testing.T) {
	suite.Run(t, new(s3v1StorageTestSuite))
}

func (suite *s3v1StorageTestSuite) SetupSuite() {
	suite.backend = s3mem.New()
	suite.faker = gofakes3.New(suite.backend)
	suite.server = httptest.NewServer(suite.faker.Server())
	suite.config = suite.setupS3Config()
	var err error
	suite.session, err = session.NewSession(&suite.config)
	suite.Require().NoError(err)
	suite.client = s3api.New(suite.session)
	suite.uploader = s3manager.NewUploader(suite.session)
	suite.bucketName = "fuel"
	suite.storage = NewS3v1(suite.client, suite.uploader, suite.bucketName)
	suite.fsStorage = newFilesystemStorage(basePath)

	suite.setupTestData()
}

func (suite *s3v1StorageTestSuite) setupTestData() {
	ctx := context.Background()
	_, err := suite.client.CreateBucket(&s3api.CreateBucketInput{Bucket: aws.String(suite.bucketName)})
	suite.Require().NoError(err)

	suite.Require().NoError(WalkDir(ctx, basePath, UploadFileS3v1(suite.uploader, suite.bucketName, nil)))
}

func (suite *s3v1StorageTestSuite) TearDownSuite() {
	ctx := context.Background()

	suite.Require().NoError(WalkDir(ctx, basePath, DeleteFileS3v1(suite.client, suite.bucketName, nil)))

	_ = os.Remove(getZipLocation(basePath, compressibleResource))

	_, err := suite.client.DeleteBucket(&s3api.DeleteBucketInput{Bucket: aws.String(suite.bucketName)})
	suite.Require().NoError(err)
	suite.server.Close()
}

func (suite *s3v1StorageTestSuite) setupS3Config() aws.Config {
	return aws.Config{
		Credentials:      credentials.NewStaticCredentials("YOUR-ACCESSKEYID", "YOUR-SECRETACCESSKEY", ""),
		Endpoint:         aws.String(suite.server.URL),
		Region:           aws.String("eu-central-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
}

func (suite *s3v1StorageTestSuite) TestGetFile_InvalidResource() {
	r := invalidResource
	ctx := context.Background()
	content, err := suite.storage.GetFile(ctx, r, "model.sdf")
	suite.Assert().Error(err)
	suite.Assert().Empty(content)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *s3v1StorageTestSuite) TestGetFile_NotFound() {
	r := validResource
	ctx := context.Background()
	content, err := suite.storage.GetFile(ctx, r, "model123.sdf")
	suite.Assert().Error(err)
	suite.Assert().Empty(content)
}

func (suite *s3v1StorageTestSuite) TestGetFile_Success() {
	r := validResource
	ctx := context.Background()
	content, err := suite.storage.GetFile(ctx, r, "model.sdf")
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(content)

	expected, err := suite.storage.GetFile(ctx, r, "model.sdf")
	suite.Require().NoError(err)
	suite.Assert().Equal(expected, content)
}

func (suite *s3v1StorageTestSuite) TestDownload_InvalidResource() {
	r := invalidResource
	ctx := context.Background()
	content, err := suite.storage.Download(ctx, r)
	suite.Assert().Error(err)
	suite.Assert().Empty(content)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *s3v1StorageTestSuite) TestDownload_NotFound() {
	r := &resource{
		uuid:    validUUID,
		owner:   owner,
		version: 5,
	}
	ctx := context.Background()
	url, err := suite.storage.Download(ctx, r)
	suite.Assert().Error(err)
	suite.Assert().Empty(url)
}

func (suite *s3v1StorageTestSuite) TestDownload_Success() {
	r := validResource
	ctx := context.Background()
	url, err := suite.storage.Download(ctx, r)
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(url)
	suite.Assert().Contains(url, ".zip")
}

func (suite *s3v1StorageTestSuite) TestUploadDir_InvalidResource() {
	r := invalidResource
	ctx := context.Background()
	err := suite.storage.UploadDir(ctx, r, "./testdata/example")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *s3v1StorageTestSuite) TestUploadDir_SourceNotFound() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example1234")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFolderNotFound)
}

func (suite *s3v1StorageTestSuite) TestUploadDir_SourceIsEmpty() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example_empty")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFolderEmpty)
}

func (suite *s3v1StorageTestSuite) TestUploadDir_SourceIsAFile() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example/meshes/turtle.dae")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFile)
}

func (suite *s3v1StorageTestSuite) TestUploadDir_Success() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example")
	suite.Assert().NoError(err)

	b, err := suite.storage.GetFile(ctx, r, "meshes/turtle.dae")
	suite.Require().NoError(err)
	suite.Assert().NotEmpty(b)

	suite.Require().NoError(WalkDir(ctx, "./testdata/example", DeleteFileS3v1(suite.client, suite.bucketName, nonExistentResource)))
}

func (suite *s3v1StorageTestSuite) TestUploadZip_InvalidResource() {
	r := invalidResource
	err := suite.storage.UploadZip(context.Background(), r, nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *s3v1StorageTestSuite) TestUploadZip_FileIsNil() {
	r := compressibleResource
	err := suite.storage.UploadZip(context.Background(), r, nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrFileNil)
}

func (suite *s3v1StorageTestSuite) TestUploadZip_Success() {
	ctx := context.Background()
	r := compressibleResource
	// Since the zip file wasn't created yet, it should fail to download the resource.
	_, err := suite.storage.Download(ctx, r)
	suite.Require().Error(err)

	// Generate the zip file locally by downloading through the local storage implementation:
	// The filesystem implementation produces zip files automatically, this is not the case in cloud storages
	// as files need to be uploaded individually, and can't perform zip operations in the cloud automatically.
	path, err := suite.fsStorage.Download(ctx, compressibleResource)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(path)

	// Open the file in order to run UploadZip with the file handler
	f, err := os.Open(path)
	suite.Require().NoError(err)

	// Test
	err = suite.storage.UploadZip(ctx, r, f)
	suite.Assert().NoError(err)

	// The zip file should not be available in the cloud
	link, err := suite.storage.Download(ctx, r)
	suite.Require().NoError(err)
	suite.Require().Contains(link, "2.zip")
}
