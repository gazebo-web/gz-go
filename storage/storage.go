package storage

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

// Storage holds the methods to interact with a Cloud provider storage.
type Storage interface {
	// GetFile returns the content of file from a given path.
	GetFile(ctx context.Context, resource Resource, path string) ([]byte, error)
	// Download returns a URL to download a resource from.
	Download(ctx context.Context, resource Resource) (string, error)
	// UploadDir uploads assets located in the given source folder and placed them into the given resource.
	UploadDir(ctx context.Context, resource Resource, source string) error
	// UploadZip uploads a compressed set of assets of the given resource.
	//
	//	Resources can have a compressed representation of the resource itself that acts like a cache, it contains all the
	//	files from the said resource. This function uploads that zip file.
	UploadZip(ctx context.Context, resource Resource, file *os.File) error
}

// ReadFileFunc is used to provide integration with cloud providers while using the same business logic
// when reading the content of a file.
type ReadFileFunc func(ctx context.Context, resource Resource, path string) (io.ReadCloser, error)

// ReadFile reads the content of the file located in path from the given resource.
// The integration the specific storage providers is provided by the ReadFileFunc.
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
// for each file found inside src. They will be uploaded as the assets for the given Resource.
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

// getLocation returns the location of a Resource relative to the base location.
//
//	If path is not empty, it will append the given path to the resulting location of the resource.
func getLocation(base string, r Resource, path string) string {
	location := filepath.Join(base, r.GetOwner(), r.GetUUID(), strconv.FormatUint(r.GetVersion(), 10))
	if len(path) > 0 {
		location = filepath.Join(location, path)
	}
	return location
}

// getZipLocation returns the location of the zip file associated to a Resource relative to the base location.
func getZipLocation(base string, r Resource) string {
	filename := fmt.Sprintf("%d.zip", r.GetVersion())
	return filepath.Join(base, r.GetOwner(), r.GetUUID(), ".zips", filename)
}

// getRootLocation returns the absolute location of where all the versions of the given uuid and the given kind will be
// uploaded for the given owner.
func getRootLocation(base string, owner string, uuid string) string {
	return filepath.Join(base, owner, uuid)
}
