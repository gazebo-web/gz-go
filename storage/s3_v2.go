package storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3api "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gazebo-web/gz-go/v7"
	"github.com/pkg/errors"
	"io"
	"io/fs"
	"os"
	"time"
)

// s3v2 implements Storage using the Amazon Web Services - Simple Storage Service (S3).
// It uses the second version of the SDK.
//
//	Reference: https://aws.amazon.com/s3/
//	API: https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html
//	SDK: https://github.com/aws/aws-sdk-go-v2
type s3v2 struct {
	client   *s3api.Client
	presign  *s3api.PresignClient
	bucket   string
	duration time.Duration
}

// UploadDir the assets found in source to S3.
func (s *s3v2) UploadDir(ctx context.Context, resource Resource, src string) error {
	err := validateResource(resource)
	if err != nil {
		return err
	}

	// Check src exists locally
	var info fs.FileInfo
	if info, err = os.Stat(src); errors.Is(err, os.ErrNotExist) {
		return errors.Wrap(ErrSourceFolderNotFound, err.Error())
	}

	// Check it's a directory
	if !info.IsDir() {
		return ErrSourceFile
	}

	// Check it's not empty
	empty, err := gz.IsDirEmpty(src)
	if err != nil {
		return errors.Wrap(ErrSourceFolderEmpty, err.Error())
	}
	if empty {
		return ErrSourceFolderEmpty
	}

	err = WalkDir(ctx, src, UploadFileS3(s.client, s.bucket, resource))
	if err != nil {
		return fmt.Errorf("failed to upload files in directory: %s, error: %w", src, err)
	}
	return nil
}

// Download downloads a zip version of the given resource from S3.
func (s *s3v2) Download(ctx context.Context, resource Resource) (string, error) {
	if err := validateResource(resource); err != nil {
		return "", err
	}

	path := getZipLocation("", resource)
	_, err := s.client.HeadObject(ctx, &s3api.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return "", err
	}

	out, err := s.presign.PresignGetObject(ctx, &s3api.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	}, s3api.WithPresignExpires(s.duration))
	if err != nil {
		return "", err
	}

	return out.URL, nil
}

// GetFile returns the content of file from a given path.
func (s *s3v2) GetFile(ctx context.Context, resource Resource, path string) ([]byte, error) {
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

// NewS3v2 initializes a new implementation of Storage using the AWS S3 service.
func NewS3v2(client *s3api.Client, bucket string) Storage {
	return &s3v2{
		client:   client,
		presign:  s3api.NewPresignClient(client),
		bucket:   bucket,
		duration: 60 * time.Minute,
	}
}

func UploadFileS3(client *s3api.Client, bucket string, resource Resource) WalkDirFunc {
	return func(ctx context.Context, path string, body io.Reader) error {
		if resource != nil {
			path = getLocation("", resource, path)
		}
		_, err := client.PutObject(ctx, &s3api.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(path),
			Body:   body,
		})
		return err
	}
}

func DeleteFileS3(client *s3api.Client, bucket string, resource Resource) WalkDirFunc {
	return func(ctx context.Context, path string, body io.Reader) error {
		if resource != nil {
			path = getLocation("", resource, path)
		}
		_, err := client.DeleteObject(ctx, &s3api.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(path),
		})
		return err
	}
}
