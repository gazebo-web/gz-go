package storage

import (
	"context"
	s3api "github.com/aws/aws-sdk-go-v2/service/s3"
)

// s3 implements Storage using the Amazon Web Services - Simple Storage Service (S3).
type s3 struct {
	client *s3api.Client
}

func (s *s3) Upload(ctx context.Context, resource Resource, source string) error {
	//TODO implement me
	panic("implement me")
}

func (s *s3) Create(ctx context.Context, owner string, kind Kind, uuid string) error {
	//TODO implement me
	panic("implement me")
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
func NewS3(client *s3api.Client) Storage {
	return &s3{
		client: client,
	}
}
