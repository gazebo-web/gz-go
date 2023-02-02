package storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	s3api "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gazebo-web/gz-go/v7"
	"github.com/pkg/errors"
	"io"
	"io/fs"
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

	err = WalkDir(ctx, src, UploadFileS3v1(s.uploader, s.bucket, resource))
	if err != nil {
		return fmt.Errorf("failed to upload files in directory: %s, error: %w", src, err)
	}
	return nil
}

func (s *s3v1) UploadZip(ctx context.Context, resource Resource, file *os.File) error {
	err := validateResource(resource)
	if err != nil {
		return err
	}
	if file == nil {
		return ErrFileNil
	}
	path := getZipLocation("", resource)
	uploader := UploadFileS3v1(s.uploader, s.bucket, nil)
	err = uploader(ctx, path, file)
	if err != nil {
		return err
	}
	return nil
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
