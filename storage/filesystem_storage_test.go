package storage

import (
	"context"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

const (
	validUUID      = "e6af5323-db4d-4db3-a402-a8992d6c8d99"
	owner          = "OpenRobotics"
	version        = 1
	invalidVersion = 5
	basePath       = "./testdata"
)

var (
	compressibleResource = &testResource{
		uuid:    "e6af5323-db4d-4db3-a402-a8992d6c8d99",
		kind:    KindModels,
		owner:   owner,
		version: 2,
	}

	nonExistentResource = &testResource{
		uuid:    uuid.NewV4().String(),
		kind:    KindModels,
		owner:   "TestOrg",
		version: 1,
	}

	invalidResource = &testResource{
		uuid:    "",
		kind:    "",
		owner:   "",
		version: 0,
	}

	validResource = &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: version,
	}
)

type FilesystemStorageTestSuite struct {
	suite.Suite
	storage Storage
}

func TestSuiteFilesystemStorage(t *testing.T) {
	suite.Run(t, new(FilesystemStorageTestSuite))
}

func (suite *FilesystemStorageTestSuite) SetupSuite() {
	suite.storage = newFilesystemStorage(basePath)
}

func (suite *FilesystemStorageTestSuite) SetupTest() {

}

func (suite *FilesystemStorageTestSuite) TearDownTest() {
	_ = os.Remove(getZipLocation(basePath, compressibleResource))
	_ = os.RemoveAll(getRootLocation(basePath, "TestOrg", "", ""))
}

func (suite *FilesystemStorageTestSuite) TestGetFile_NotFound() {
	r := &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: invalidVersion,
	}

	var err error
	_, err = suite.storage.GetFile(context.Background(), r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceNotFound)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ResourceInvalidOwner() {
	r := &testResource{
		uuid:    "",
		kind:    "",
		owner:   "",
		version: 0,
	}

	var err error
	_, err = suite.storage.GetFile(context.Background(), r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ResourceInvalidKind() {
	r := &testResource{
		uuid:    "",
		kind:    "",
		owner:   owner,
		version: 0,
	}
	var err error
	_, err = suite.storage.GetFile(context.Background(), r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ResourceInvalidUUID() {
	r := &testResource{
		uuid:    "",
		kind:    KindModels,
		owner:   owner,
		version: 0,
	}

	var err error
	_, err = suite.storage.GetFile(context.Background(), r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ResourceInvalidVersion() {
	r := &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: 0,
	}

	var err error
	_, err = suite.storage.GetFile(context.Background(), r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_Exists() {
	r := validResource

	var err error
	_, err = suite.storage.GetFile(context.Background(), r, "/model.sdf")
	suite.Assert().NoError(err)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ContentMatches() {

	r := validResource

	var err error
	expected, err := os.ReadFile("./testdata/OpenRobotics/models/e6af5323-db4d-4db3-a402-a8992d6c8d99/1/model.sdf")
	suite.Require().NoError(err)
	suite.Require().NotEmpty(expected)

	var b []byte
	path := "/model.sdf"
	b, err = suite.storage.GetFile(context.Background(), r, path)
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(b)
	suite.Assert().Equal(expected, b)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ContentMatchesSubFolder() {
	r := validResource

	var err error
	expected, err := os.ReadFile("./testdata/OpenRobotics/models/e6af5323-db4d-4db3-a402-a8992d6c8d99/1/meshes/turtle.dae")
	suite.Require().NoError(err)
	suite.Require().NotEmpty(expected)

	var b []byte
	path := "/meshes/turtle.dae"
	b, err = suite.storage.GetFile(context.Background(), r, path)
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(b)
	suite.Assert().Equal(expected, b)
}

func (suite *FilesystemStorageTestSuite) TestDownload_InvalidResource() {
	r := &testResource{
		uuid:    "31f64dd2-e867-45a7-9a8c-10d9733de2b3",
		kind:    KindModels,
		owner:   owner,
		version: 0, // Invalid version
	}

	var err error
	_, err = suite.storage.Download(context.Background(), r)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestDownload_NotFound() {
	r := &testResource{
		uuid:    "31f64dd2-e867-45a7-9a8c-10d9733de2b3",
		kind:    KindModels,
		owner:   owner,
		version: 3, // Valid version but doesn't exist
	}

	var err error
	_, err = suite.storage.Download(context.Background(), r)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceNotFound)
}

func (suite *FilesystemStorageTestSuite) TestDownload_EmptyFolder() {
	r := &testResource{
		uuid:    "31f64dd2-e867-45a7-9a8c-10d9733de2b3",
		kind:    KindModels,
		owner:   owner,
		version: 1,
	}

	suite.Require().NoError(os.MkdirAll(getLocation(basePath, r, ""), os.ModePerm))

	var err error
	_, err = suite.storage.Download(context.Background(), r)
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrEmptyResource)
}

func (suite *FilesystemStorageTestSuite) TestDownload_PathToZip() {
	r := compressibleResource

	zip, err := suite.storage.Download(context.Background(), r)
	suite.Assert().NoError(err)
	suite.Assert().Contains(zip, ".zip")
}

func (suite *FilesystemStorageTestSuite) TestDownload_ValidPath() {
	zip, err := suite.storage.Download(context.Background(), compressibleResource)
	suite.Assert().NoError(err)

	info, err := os.Stat(zip)
	suite.Require().NoError(err)
	suite.Assert().False(info.IsDir())
	suite.Assert().NotZero(info.Size())
}

func (suite *FilesystemStorageTestSuite) TestUploadDir_InvalidOwner() {
	r := &testResource{
		uuid:    "31f64dd2-e867-45a7-9a8c-10d9733de2b3",
		kind:    KindModels,
		owner:   "",
		version: 1,
	}
	err := suite.storage.UploadDir(context.Background(), r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestUploadDir_InvalidKind() {
	r := &testResource{
		uuid:    "31f64dd2-e867-45a7-9a8c-10d9733de2b3",
		kind:    "",
		owner:   "OpenRobotics",
		version: 1,
	}
	err := suite.storage.UploadDir(context.Background(), r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestUploadDir_InvalidUUID() {
	r := &testResource{
		uuid:    "",
		kind:    KindModels,
		owner:   "OpenRobotics",
		version: 1,
	}
	err := suite.storage.UploadDir(context.Background(), r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestUploadDir_SourceIsNotAFolder() {
	// ./testdata/example/model.config is a file
	err := suite.storage.UploadDir(context.Background(), nonExistentResource, "./testdata/example/model.config")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFile)
}

func (suite *FilesystemStorageTestSuite) TestUploadDir_SourceIsEmpty() {
	// Folder: ./testdata/example_empty is empty
	suite.Require().NoError(os.MkdirAll("./testdata/example_empty", os.ModePerm))
	err := suite.storage.UploadDir(context.Background(), nonExistentResource, "./testdata/example_empty")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrSourceFolderEmpty)
}

func (suite *FilesystemStorageTestSuite) TestUploadDir_Success() {
	// Let's upload the assets from ./testdata/example
	err := suite.storage.UploadDir(context.Background(), nonExistentResource, "./testdata/example")
	suite.Assert().NoError(err)
}
