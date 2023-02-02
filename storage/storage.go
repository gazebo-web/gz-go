package storage

import (
	"context"
	"fmt"
	"github.com/gazebo-web/gz-go/v7"
	"github.com/pkg/errors"
	"io"
	"io/fs"
	"os"
)

// Storage holds the methods to interact with a Cloud provider storage.
type Storage interface {
	// GetFile returns the content of file from a given path.
	GetFile(ctx context.Context, resource Resource, path string) ([]byte, error)
	// Download returns the URL where to download the given resource from.
	Download(ctx context.Context, resource Resource) (string, error)
	// UploadDir uploads assets located in the given source folder and placed them into the given resource.
	UploadDir(ctx context.Context, resource Resource, source string) error
	// UploadZip uploads a compressed set of assets of the given resource.
	UploadZip(ctx context.Context, resource Resource, file *os.File) error
}

type ReadFileFunc func(ctx context.Context, resource Resource, path string) (io.ReadCloser, error)

func ReadFile(ctx context.Context, resource Resource, path string, fn ReadFileFunc) ([]byte, error) {
	if err := validateResource(resource); err != nil {
		return nil, err
	}
	body, err := fn(ctx, resource, path)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// UploadDir uploads the directory and all the sub elements found in src using the provided WalkDirFunc
// for each file found inside src. It will be uploaded as the assets for the given Resource.
func UploadDir(ctx context.Context, resource Resource, src string, fn WalkDirFunc) error {
	err := validateResource(resource)
	if err != nil {
		return err
	}

	// Check src exists locally
	var info fs.FileInfo
	if info, err = os.Stat(src); errors.Is(err, os.ErrNotExist) {
		return errors.Wrap(ErrSourceFolderNotFound, err.Error())
	}

	// Check it's a directory
	if !info.IsDir() {
		return ErrSourceFile
	}

	// Check it's not empty
	empty, err := gz.IsDirEmpty(src)
	if err != nil {
		return errors.Wrap(ErrSourceFolderEmpty, err.Error())
	}
	if empty {
		return ErrSourceFolderEmpty
	}

	err = WalkDir(ctx, src, fn)
	if err != nil {
		return fmt.Errorf("failed to upload files in directory: %s, error: %w", src, err)
	}
	return nil
}

// UploadZip uploads the given file to where the given resource is stored.
func UploadZip(ctx context.Context, resource Resource, file *os.File, fn WalkDirFunc) error {
	err := validateResource(resource)
	if err != nil {
		return err
	}
	if file == nil {
		return ErrFileNil
	}
	path := getZipLocation("", resource)
	err = fn(ctx, path, file)
	if err != nil {
		return err
	}
	return nil
}
