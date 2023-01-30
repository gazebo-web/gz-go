package storage

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3api "github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"time"
)

// s3 implements Storage using the Amazon Web Services - Simple Storage Service (S3).
type s3 struct {
	client   *s3api.Client
	presign  *s3api.PresignClient
	bucket   string
	duration time.Duration
}

// Upload the assets found in source to S3.
func (s *s3) Upload(ctx context.Context, resource Resource, source string) error {
	//TODO implement me
	panic("implement me")
}

// Create prepares the bucket to hold a resource identified by UUID that will be uploaded
// by owner and it will of the given kind.
func (s *s3) Create(ctx context.Context, owner string, kind Kind, uuid string) error {
	//TODO implement me
	panic("implement me")
}

// Download downloads a zip version of the given resource from S3.
func (s *s3) Download(ctx context.Context, resource Resource) (string, error) {
	if err := validateResource(resource); err != nil {
		return "", err
	}

	_, err := s.client.HeadObject(ctx, &s3api.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(getZipLocation("", resource)),
	})
	if err != nil {
		return "", err
	}

	out, err := s.presign.PresignGetObject(ctx, &s3api.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(getZipLocation("", resource)),
	}, s3api.WithPresignExpires(s.duration))
	if err != nil {
		return "", err
	}

	return out.URL, nil
}

// GetFile returns the content of file from a given path.
func (s *s3) GetFile(ctx context.Context, resource Resource, path string) ([]byte, error) {
	if err := validateResource(resource); err != nil {
		return nil, err
	}

	out, err := s.client.GetObject(ctx, &s3api.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(getLocation("", resource, path)),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	b, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// NewS3 initializes a new implementation of Storage using the AWS S3 service.
func NewS3(client *s3api.Client, bucket string) Storage {
	return &s3{
		client:   client,
		presign:  s3api.NewPresignClient(client),
		bucket:   bucket,
		duration: 60 * time.Minute,
	}
}
