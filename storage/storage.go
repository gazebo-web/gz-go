package storage

// Storage holds the methods to interact with a Cloud provider storage.
type Storage interface {
	// GetFile returns the content of file from a given path.
	GetFile(resource Resource, path string) ([]byte, error)
}
