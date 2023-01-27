package storage

import "context"

// s3 implements Storage using the Amazon Web Services - Simple Storage Service (S3).
type s3 struct {
}

func (s *s3) Download(ctx context.Context, resource Resource) (string, error) {
	//TODO implement me
	panic("implement me")
}

// GetFile returns the content of file from a given path.
func (s *s3) GetFile(ctx context.Context, resource Resource, path string) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

// NewS3 initializes a new implementation of Storage using the AWS S3 service.
func NewS3() Storage {
	return &s3{}
}
