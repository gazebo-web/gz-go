package storage

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	s3api "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
	"time"
)

// s3v1 implements Storage using the Amazon Web Services - Simple Storage Service (S3).
// It uses the first version of the SDK.
//
//	Reference: https://aws.amazon.com/s3/
//	API: https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html
//	SDK: https://github.com/aws/aws-sdk-go
type s3v1 struct {
	client   *s3api.S3
	uploader *s3manager.Uploader
	bucket   string
	duration time.Duration
}

// GetFile reads the content of a file located in path from the given Resource.
func (s *s3v1) GetFile(ctx context.Context, resource Resource, path string) ([]byte, error) {
	return ReadFile(ctx, resource, path, readFileS3v1(s.client, s.bucket))
}

// Download returns the URL of zip file that contains all the contents of the given Resource.
func (s *s3v1) Download(ctx context.Context, resource Resource) (string, error) {
	if err := validateResource(resource); err != nil {
		return "", err
	}

	path := getZipLocation("", resource)
	_, err := s.client.HeadObject(&s3api.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return "", err
	}

	req, _ := s.client.GetObjectRequest(&s3api.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	url, err := req.Presign(s.duration)
	if err != nil {
		return "", err
	}
	return url, nil
}

// UploadDir uploads all the files found in src to S3.
func (s *s3v1) UploadDir(ctx context.Context, resource Resource, src string) error {
	return UploadDir(ctx, resource, src, uploadFileS3v1(s.uploader, s.bucket, resource))
}

// UploadZip uploads a zip file of the given resource to S3. It should be called before any attempts to Download
// the zip file of the given Resource.
func (s *s3v1) UploadZip(ctx context.Context, resource Resource, file *os.File) error {
	return UploadZip(ctx, resource, file, uploadFileS3v1(s.uploader, s.bucket, nil))
}

// NewS3v1 initializes a new implementation of Storage using the AWS S3 v1 service.
func NewS3v1(client *s3api.S3, uploader *s3manager.Uploader, bucket string) Storage {
	return &s3v1{
		client:   client,
		uploader: uploader,
		bucket:   bucket,
		duration: 60 * time.Minute,
	}
}

// readFileS3v1 generates a function that contains the interaction with S3 to read the contents of a file.
func readFileS3v1(client *s3api.S3, bucket string) ReadFileFunc {
	return func(ctx context.Context, resource Resource, path string) (io.ReadCloser, error) {
		out, err := client.GetObject(&s3api.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(getLocation("", resource, path)),
		})
		if err != nil {
			return nil, err
		}
		return out.Body, nil
	}
}

// uploadFileS3v1 generates a function that uploads a single file in a path.
// If Resource is nil, it will use the given path as-is, otherwise it will use the given path as a relative path
// to the given Resource.
func uploadFileS3v1(uploader *s3manager.Uploader, bucket string, resource Resource) WalkDirFunc {
	return func(ctx context.Context, path string, body io.Reader) error {
		if resource != nil {
			path = getLocation("", resource, path)
		}
		_, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(path),
			Body:   body,
		})
		return err
	}
}

// deleteFileS3v1 generates a function that allows to delete a single file in a path.
// If Resource is nil, it will use the given path as-is, otherwise it will use the given path as a relative path
// to the given Resource.
func deleteFileS3v1(client *s3api.S3, bucket string, resource Resource) WalkDirFunc {
	return func(ctx context.Context, path string, _ io.Reader) error {
		if resource != nil {
			path = getLocation("", resource, path)
		}
		_, err := client.DeleteObject(&s3api.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(path),
		})
		return err
	}
}
