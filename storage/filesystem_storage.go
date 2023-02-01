package storage

import (
	"archive/zip"
	"context"
	"github.com/gazebo-web/gz-go/v7"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
)

// fsStorage contains the implementation of a storage manager for resources using the host filesystem.
// It can be used with AWS EFS storage in EC2 instances.
type fsStorage struct {
	basePath string
}

func (s *fsStorage) UploadZip(ctx context.Context, resource Resource, file *os.File) error {
	//TODO implement me
	panic("implement me")
}

func (s *fsStorage) UploadDir(ctx context.Context, resource Resource, src string) error {
	if err := validateResource(resource); err != nil {
		return err
	}
	var info os.FileInfo
	var err error
	if info, err = os.Stat(src); errors.Is(err, os.ErrNotExist) {
		err = s.create(ctx, resource.GetOwner(), resource.GetKind(), resource.GetUUID())
		if err != nil {
			return err
		}
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

	err = gz.CopyDir(dst, src)
	if err != nil {
		return err
	}
	return nil
}

func (s *fsStorage) create(ctx context.Context, owner string, kind Kind, uuid string) error {
	if err := validateOwner(owner); err != nil {
		return err
	}
	if err := validateKind(kind); err != nil {
		return err
	}
	if err := validateUUID(uuid); err != nil {
		return err
	}

	path := getRootLocation(s.basePath, owner, kind, uuid)
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

// Download returns the path to the zip file that includes the given resource.
func (s *fsStorage) Download(ctx context.Context, resource Resource) (string, error) {
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
func (s *fsStorage) GetFile(ctx context.Context, resource Resource, path string) ([]byte, error) {
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

// zip compress the given resource to a  zip file. It returns the location to the zip file.
// If the file was already created, it returns a cached file.
func (s *fsStorage) zip(ctx context.Context, resource Resource) (string, error) {
	target := getZipLocation(s.basePath, resource)
	if _, err := os.Stat(target); errors.Is(err, os.ErrExist) {
		return target, nil
	}

	f, err := os.Create(target)
	if err != nil {
		return "", err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	source := getLocation(s.basePath, resource, "")
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Set compression
		header.Method = zip.Deflate

		// Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// Create writer for the file header and save content of the file
		headerWriter, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
	if err != nil {
		return "", err
	}

	return target, nil
}

// newFilesystemStorage initializes a new Storage implementation using the host FileSystem.
// It receives the base path as an argument, where all the resources can be found.
func newFilesystemStorage(path string) Storage {
	return &fsStorage{
		basePath: path,
	}
}
