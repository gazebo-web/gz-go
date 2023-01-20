package storage

// gcs implements Storage using the Google Cloud Platform - Cloud Storage (CS) service.
//
//	Reference: https://cloud.google.com/storage
//	API: https://cloud.google.com/storage/docs/apis
//	SDK: https://pkg.go.dev/cloud.google.com/go/storage
type gcs struct{}

// GetFile returns the content of file from a given path.
func (g *gcs) GetFile(resource Resource, path string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func NewGCS() Storage {
	return &gcs{}
}
