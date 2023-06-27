package storage

import (
	"context"
	"github.com/pkg/errors"
	"io"
	"os"
)

// fileSys is a Storage implementation that uses the host filesystem to store resources.
// It can be used with AWS EFS storage in EC2 instances.
type fileSys struct {
	basePath string
}

// UploadZip uploads the given file as the zip file of the given resource. If the file already exists, it will be
// replaced by the new file.
//
//	Resources can have a compressed representation of the resource itself that acts like a cache, it contains all the
//	files from the said resource. This function uploads that zip file.
func (s *fileSys) UploadZip(ctx context.Context, resource Resource, file *os.File) error {
	if err := validateResource(resource); err != nil {
		return err
	}
	if file == nil {
		return ErrFileNil
	}

	dst := getZipLocation(s.basePath, resource)
	err := gz.RemoveIfFound(dst)
	if err != nil {
		return err
	}

	zipFile, err := os.Create(dst)
	defer gz.Close(zipFile)
	if err != nil {
		return err
	}

	_, err = io.Copy(zipFile, file)
	if err != nil {
		return err
	}

	return nil
}

// UploadDir uploads the assets found in source to the dedicated directory used to store resources.
func (s *fileSys) UploadDir(ctx context.Context, resource Resource, src string) error {
	if err := validateResource(resource); err != nil {
		return err
	}
	var info os.FileInfo
	var err error
	if info, err = os.Stat(src); errors.Is(err, os.ErrNotExist) {
		return errors.Wrap(ErrSourceFolderNotFound, err.Error())
	}
	if !info.IsDir() {
		return ErrSourceFile
	}
	empty, err := gz.IsDirEmpty(src)
	if err != nil {
		return errors.Wrap(ErrSourceFolderEmpty, err.Error())
	}
	if empty {
		return ErrSourceFolderEmpty
	}
	dst := getLocation(s.basePath, resource, "")
	if _, err = os.Stat(dst); errors.Is(err, os.ErrNotExist) {
		err = s.create(ctx, resource.GetOwner(), resource.GetUUID())
		if err != nil {
			return err
		}
	}

	err = gz.CopyDir(dst, src)
	if err != nil {
		return err
	}
	return nil
}

func (s *fileSys) create(ctx context.Context, owner string, uuid string) error {
	if err := validateOwner(owner); err != nil {
		return err
	}
	if err := validateUUID(uuid); err != nil {
		return err
	}

	path := getRootLocation(s.basePath, owner, uuid)
	_, err := os.Stat(path)
	if err == nil {
		return ErrResourceAlreadyExists
	}

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// Download returns a path to the zip file that includes the given resource.
func (s *fileSys) Download(ctx context.Context, resource Resource) (string, error) {
	if err := validateResource(resource); err != nil {
		return "", err
	}

	var info os.FileInfo
	var err error
	path := getLocation(s.basePath, resource, "")
	if info, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "", errors.Wrap(ErrResourceNotFound, err.Error())
	}

	if info.IsDir() {
		empty, err := gz.IsDirEmpty(path)
		if err != nil {
			return "", errors.Wrap(ErrEmptyResource, err.Error())
		}
		if empty {
			return "", ErrEmptyResource
		}
	}

	return s.zip(ctx, resource)
}

// GetFile returns the content of file from a given path.
func (s *fileSys) GetFile(ctx context.Context, resource Resource, path string) ([]byte, error) {
	if err := validateResource(resource); err != nil {
		return nil, err
	}

	path = getLocation(s.basePath, resource, path)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, errors.Wrap(ErrResourceNotFound, err.Error())
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// zip compresses the given resource to a zip file and returns the path to the zip file.
// If the file was already created, it returns a cached file.
func (s *fileSys) zip(ctx context.Context, resource Resource) (string, error) {
	target := getZipLocation(s.basePath, resource)
	source := getLocation(s.basePath, resource, "")
	f, err := gz.Zip(target, source)
	defer gz.Close(f)
	if err != nil {
		return "", err
	}
	return target, nil
}

// newFilesystemStorage initializes a new Storage implementation using the host FileSystem.
// It receives the base path as an argument, where all resources are stored.
func newFilesystemStorage(path string) Storage {
	return &fileSys{
		basePath: path,
	}
}
