package storage

import (
	"context"
	s3api "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path/filepath"
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
	storage := NewS3v2(&s3api.Client{}, "")
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

func TestUploadDir(t *testing.T) {
	ctx := context.Background()
	r := validResource
	src := "./testdata/example"
	var paths []string
	assert.NoError(t, UploadDir(ctx, r, src, func(ctx context.Context, path string, body io.Reader) error {
		paths = append(paths, path)
		return nil
	}))
	assert.Len(t, paths, 4)
	assert.Contains(t, paths, "meshes/turtle.dae")
	assert.Contains(t, paths, "thumbnails/1.png")
	assert.Contains(t, paths, "model.config")
	assert.Contains(t, paths, "model.sdf")
}

func TestReadFile(t *testing.T) {
	ctx := context.Background()
	r := validResource
	path := filepath.Join(t.TempDir(), "test.txt")
	f, err := os.Create(path)
	require.NoError(t, err)
	n, err := io.WriteString(f, "test")
	require.Equal(t, 4, n)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	b, err := ReadFile(ctx, r, path, func(ctx context.Context, resource Resource, path string) (io.ReadCloser, error) {
		open, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		return open, nil
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, b)
	assert.Equal(t, "test", string(b))
}

func TestUploadZip(t *testing.T) {
	ctx := context.Background()
	r := validResource
	location := getZipLocation("./testdata", r)
	f, err := os.Open(location)
	require.NoError(t, err)

	var content []byte
	assert.NoError(t, UploadZip(ctx, r, f, func(ctx context.Context, p string, body io.Reader) error {
		var err error
		content, err = io.ReadAll(body)
		return err
	}))
	assert.NotEmpty(t, content)
}
