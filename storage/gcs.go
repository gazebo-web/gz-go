package storage

import (
	"context"
	"os"
)

// gcs implements Storage using the Google Cloud Platform - Cloud Storage (CS) service.
//
//	Reference: https://cloud.google.com/storage
//	API: https://cloud.google.com/storage/docs/apis
//	SDK: https://pkg.go.dev/cloud.google.com/go/storage
type gcs struct{}

func (g *gcs) UploadZip(ctx context.Context, resource Resource, file *os.File) error {
	//TODO implement me
	panic("implement me")
}

func (g *gcs) UploadDir(ctx context.Context, resource Resource, source string) error {
	//TODO implement me
	panic("implement me")
}

func (g *gcs) Download(ctx context.Context, resource Resource) (string, error) {
	//TODO implement me
	panic("implement me")
}

// GetFile returns the content of file from a given path.
func (g *gcs) GetFile(ctx context.Context, resource Resource, path string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func NewGCS() Storage {
	return &gcs{}
}
