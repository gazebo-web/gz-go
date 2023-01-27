package storage

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

var _ Resource = (*testResource)(nil)

type testResource struct {
	uuid    string
	kind    Kind
	owner   string
	version uint64
}

func (t *testResource) GetUUID() string {
	return t.uuid
}

func (t *testResource) GetKind() Kind {
	return t.kind
}

func (t *testResource) GetOwner() string {
	return t.owner
}

func (t *testResource) GetVersion() uint64 {
	return t.version
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

type StorageTestSuite struct {
	suite.Suite
}

func (suite *StorageTestSuite) TestNewS3Storage() {
	storage := NewS3(nil)
	suite.Assert().Implements((*Storage)(nil), storage)
}

func (suite *StorageTestSuite) TestNewGCSStorage() {
	storage := NewGCS()
	suite.Assert().Implements((*Storage)(nil), storage)
}

func (suite *StorageTestSuite) TestNewFilesystemStorage() {
	storage := newFilesystemStorage("./testdata")
	suite.Assert().Implements((*Storage)(nil), storage)
}
