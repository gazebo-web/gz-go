package storage

import "context"

// Storage holds the methods to interact with a Cloud provider storage.
type Storage interface {
	// GetFile returns the content of file from a given path.
	GetFile(ctx context.Context, resource Resource, path string) ([]byte, error)
	// Download returns the URL where to download the given resource from.
	Download(ctx context.Context, resource Resource) (string, error)
	// UploadDir uploads assets located in the given source folder and placed them into the given resource.
	UploadDir(ctx context.Context, resource Resource, source string) error
}
