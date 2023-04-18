package storage

import (
	"context"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type gcsStorageTestSuite struct {
	suite.Suite
	storage    Storage
	server     *fakestorage.Server
	bucketName string
	fsStorage  Storage
}

func TestSuiteGCSStorage(t *testing.T) {
	suite.Run(t, new(gcsStorageTestSuite))
}

func (suite *gcsStorageTestSuite) SetupSuite() {
	suite.server = fakestorage.NewServer(nil)
	suite.bucketName = "fuel"
	suite.storage = NewGCS(suite.server.Client(), suite.bucketName)
	suite.fsStorage = newFilesystemStorage(basePath)

	suite.setupTestData()
}

func (suite *gcsStorageTestSuite) setupTestData() {
	ctx := context.Background()
	suite.Require().NoError(suite.server.Client().Bucket(suite.bucketName).Create(ctx, "", nil))

	suite.Require().NoError(WalkDir(ctx, basePath, uploadFileGCS(suite.server.Client(), suite.bucketName, nil)))
}

func (suite *gcsStorageTestSuite) TearDownSuite() {
	ctx := context.Background()

	suite.Require().NoError(WalkDir(ctx, basePath, deleteFileGCS(suite.server.Client(), suite.bucketName, nil)))

	_ = os.Remove(getZipLocation(basePath, compressibleResource))

	suite.Require().NoError(suite.server.Client().Bucket(suite.bucketName).Delete(ctx))
	suite.server.Stop()
}

func (suite *gcsStorageTestSuite) TestGetFile_InvalidResource() {
	r := invalidResource
	ctx := context.Background()
	content, err := suite.storage.GetFile(ctx, r, "model.sdf")
	suite.Assert().Error(err)
	suite.Assert().Empty(content)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *gcsStorageTestSuite) TestGetFile_NotFound() {
	r := validResource
	ctx := context.Background()
	content, err := suite.storage.GetFile(ctx, r, "model123.sdf")
	suite.Assert().Error(err)
	suite.Assert().Empty(content)
}

func (suite *gcsStorageTestSuite) TestGetFile_Success() {
	r := validResource
	ctx := context.Background()
	content, err := suite.storage.GetFile(ctx, r, "model.sdf")
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(content)

	expected, err := suite.storage.GetFile(ctx, r, "model.sdf")
	suite.Require().NoError(err)
	suite.Assert().Equal(expected, content)
}

func (suite *gcsStorageTestSuite) TestDownload_InvalidResource() {
	r := invalidResource
	ctx := context.Background()
	content, err := suite.storage.Download(ctx, r)
	suite.Assert().Error(err)
	suite.Assert().Empty(content)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *gcsStorageTestSuite) TestDownload_NotFound() {
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

func (suite *gcsStorageTestSuite) TestDownload_Success() {
	r := validResource
	ctx := context.Background()
	url, err := suite.storage.Download(ctx, r)
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(url)
	suite.Assert().Contains(url, ".zip")
}

func (suite *gcsStorageTestSuite) TestUploadDir_InvalidResource() {
	r := invalidResource
	ctx := context.Background()
	err := suite.storage.UploadDir(ctx, r, "./testdata/example")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *gcsStorageTestSuite) TestUploadDir_SourceNotFound() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example1234")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFolderNotFound)
}

func (suite *gcsStorageTestSuite) TestUploadDir_SourceIsEmpty() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example_empty")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFolderEmpty)
}

func (suite *gcsStorageTestSuite) TestUploadDir_SourceIsAFile() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example/meshes/turtle.dae")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFile)
}

func (suite *gcsStorageTestSuite) TestUploadDir_Success() {
	r := nonExistentResource
	ctx := context.Background()

	err := suite.storage.UploadDir(ctx, r, "./testdata/example")
	suite.Assert().NoError(err)

	b, err := suite.storage.GetFile(ctx, r, "meshes/turtle.dae")
	suite.Require().NoError(err)
	suite.Assert().NotEmpty(b)

	suite.Require().NoError(WalkDir(ctx, "./testdata/example", deleteFileGCS(suite.server.Client(), suite.bucketName, nonExistentResource)))
}

func (suite *gcsStorageTestSuite) TestUploadZip_InvalidResource() {
	r := invalidResource
	err := suite.storage.UploadZip(context.Background(), r, nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *gcsStorageTestSuite) TestUploadZip_FileIsNil() {
	r := compressibleResource
	err := suite.storage.UploadZip(context.Background(), r, nil)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrFileNil)
}

func (suite *gcsStorageTestSuite) TestUploadZip_Success() {
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
