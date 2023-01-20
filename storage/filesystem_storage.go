package storage

import (
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strconv"
)

// fsStorage contains the implementation of a storage manager for resources using the host filesystem.
// It can be used with something like AWS EFS storage in EC2 instances.
type fsStorage struct {
	basePath string
}

// GetFile returns the content of file from a given path.
func (s *fsStorage) GetFile(resource Resource, path string) ([]byte, error) {
	if err := ValidateResource(resource); err != nil {
		return nil, err
	}

	path = filepath.Join(s.basePath, resource.GetOwner(), string(resource.GetKind()), resource.GetUUID(), strconv.FormatUint(resource.GetVersion(), 10), path)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, errors.Wrap(ErrResourceNotFound, err.Error())
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// newFilesystemStorage initializes a new Storage implementation using the host FileSystem.
// It receives the base path as an argument, where all the resources can be found.
func newFilesystemStorage(path string) Storage {
	return &fsStorage{
		basePath: path,
	}
}
