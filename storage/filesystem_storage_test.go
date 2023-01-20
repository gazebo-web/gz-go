package storage

import (
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

const (
	validUUID      = "e6af5323-db4d-4db3-a402-a8992d6c8d99"
	owner          = "OpenRobotics"
	version        = 1
	invalidVersion = 5
)

type FilesystemStorageTestSuite struct {
	suite.Suite
	storage Storage
}

func TestSuiteFilesystemStorage(t *testing.T) {
	suite.Run(t, new(FilesystemStorageTestSuite))
}

func (suite *FilesystemStorageTestSuite) SetupSuite() {
	suite.storage = newFilesystemStorage("./testdata")
}

func (suite *FilesystemStorageTestSuite) TestGetFile_NotFound() {
	var r Resource
	r = &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: invalidVersion,
	}

	var err error
	_, err = suite.storage.GetFile(r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceNotFound)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ResourceInvalidOwner() {
	var r Resource
	r = &testResource{
		uuid:    "",
		kind:    "",
		owner:   "",
		version: 0,
	}

	var err error
	_, err = suite.storage.GetFile(r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ResourceInvalidKind() {
	var r Resource
	r = &testResource{
		uuid:    "",
		kind:    "",
		owner:   owner,
		version: 0,
	}
	var err error
	_, err = suite.storage.GetFile(r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ResourceInvalidUUID() {
	var r Resource
	r = &testResource{
		uuid:    "",
		kind:    KindModels,
		owner:   owner,
		version: 0,
	}

	var err error
	_, err = suite.storage.GetFile(r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ResourceInvalidVersion() {
	var r Resource
	r = &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: 0,
	}

	var err error
	_, err = suite.storage.GetFile(r, "")
	suite.Assert().Error(err)
	suite.Assert().ErrorIs(err, ErrResourceInvalidFormat)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_Exists() {
	var r Resource
	r = &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: version,
	}

	var err error
	_, err = suite.storage.GetFile(r, "/model.sdf")
	suite.Assert().NoError(err)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ContentMatchesInRootFolder() {

	var r Resource
	r = &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: version,
	}

	var err error
	expected, err := os.ReadFile("./testdata/OpenRobotics/models/e6af5323-db4d-4db3-a402-a8992d6c8d99/1/model.sdf")
	suite.Require().NoError(err)
	suite.Require().NotEmpty(expected)

	var b []byte
	path := "/model.sdf"
	b, err = suite.storage.GetFile(r, path)
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(b)
	suite.Assert().Equal(expected, b)
}

func (suite *FilesystemStorageTestSuite) TestGetFile_ContentMatchesInSubFolder() {

	var r Resource
	r = &testResource{
		uuid:    validUUID,
		kind:    KindModels,
		owner:   owner,
		version: version,
	}

	var err error
	expected, err := os.ReadFile("./testdata/OpenRobotics/models/e6af5323-db4d-4db3-a402-a8992d6c8d99/1/meshes/turtle.dae")
	suite.Require().NoError(err)
	suite.Require().NotEmpty(expected)

	var b []byte
	path := "/meshes/turtle.dae"
	b, err = suite.storage.GetFile(r, path)
	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(b)
	suite.Assert().Equal(expected, b)
}
