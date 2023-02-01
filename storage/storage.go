package storage

import (
	"context"
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
