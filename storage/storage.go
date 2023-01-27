package storage

import "context"

// Storage holds the methods to interact with a Cloud provider storage.
type Storage interface {
	// GetFile returns the content of file from a given path.
	GetFile(ctx context.Context, resource Resource, path string) ([]byte, error)
	Download(ctx context.Context, resource Resource) (string, error)
	Create(ctx context.Context, owner string, kind Kind, uuid string) error
	Upload(ctx context.Context, resource Resource, source string) error
}
