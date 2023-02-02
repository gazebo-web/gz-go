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

type s3v1 struct {
	client   *s3api.S3
	uploader *s3manager.Uploader
	bucket   string
	duration time.Duration
}

func (s *s3v1) GetFile(ctx context.Context, resource Resource, path string) ([]byte, error) {
	if err := validateResource(resource); err != nil {
		return nil, err
	}

	out, err := s.client.GetObject(&s3api.GetObjectInput{
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

func (s *s3v1) UploadDir(ctx context.Context, resource Resource, src string) error {
	return UploadDir(ctx, resource, src, UploadFileS3v1(s.uploader, s.bucket, resource))
}

func (s *s3v1) UploadZip(ctx context.Context, resource Resource, file *os.File) error {
	return UploadZip(ctx, resource, file, UploadFileS3v1(s.uploader, s.bucket, nil))
}

func NewS3v1(client *s3api.S3, uploader *s3manager.Uploader, bucket string) Storage {
	return &s3v1{
		client:   client,
		uploader: uploader,
		bucket:   bucket,
		duration: 60 * time.Minute,
	}
}

func UploadFileS3v1(uploader *s3manager.Uploader, bucket string, resource Resource) WalkDirFunc {
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

func DeleteFileS3v1(client *s3api.S3, bucket string, resource Resource) WalkDirFunc {
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
