package storage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// gcs implements Storage using the Google Cloud Platform - Cloud Storage (CS) service.
//
//	Reference: https://cloud.google.com/storage
//	API: https://cloud.google.com/storage/docs/apis
//	SDK: https://pkg.go.dev/cloud.google.com/go/storage
type gcs struct {
	client *storage.Client
	bucket string
}

func (g *gcs) UploadZip(ctx context.Context, resource Resource, file *os.File) error {
	return UploadZip(ctx, resource, file, uploadFileGCS(g.client, g.bucket, nil))
}

func (g *gcs) UploadDir(ctx context.Context, resource Resource, src string) error {
	return UploadDir(ctx, resource, src, uploadFileGCS(g.client, g.bucket, resource))
}

func (g *gcs) Download(ctx context.Context, resource Resource) (string, error) {
	if err := validateResource(resource); err != nil {
		return "", err
	}

	path := getZipLocation("", resource)
	obj := getObjectGCS(g.client, g.bucket, path)
	if _, err := obj.Attrs(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return "", fmt.Errorf("object %s does not exist in bucket %s", path, g.bucket)
		}
		return "", err
	}

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(1 * time.Minute),
	}

	u, err := g.client.Bucket(g.bucket).SignedURL(path, opts)
	if err != nil {
		return "", err
	}
	return u, nil
}

// GetFile returns the content of file from a given path.
func (g *gcs) GetFile(ctx context.Context, resource Resource, path string) ([]byte, error) {
	return ReadFile(ctx, resource, path, readFileGCS(g.client, g.bucket))
}

// readFileS3v1 generates a function that contains the interaction with S3 to read the contents of a file.
func readFileGCS(client *storage.Client, bucket string) ReadFileFunc {
	return func(ctx context.Context, resource Resource, path string) (io.ReadCloser, error) {
		obj := getObjectGCS(client, bucket, getLocation("", resource, path))
		r, err := obj.NewReader(ctx)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
}

func getObjectGCS(client *storage.Client, bucket string, path string) *storage.ObjectHandle {
	obj := client.Bucket(bucket).Object(path)
	return obj
}

// uploadFileS3v1 generates a function that uploads a single file in a path.
func uploadFileGCS(client *storage.Client, bucket string, resource Resource) WalkDirFunc {
	return func(ctx context.Context, path string, body io.Reader) error {
		// If Resource is nil, it will use the given path as-is, otherwise it will use the given path as a relative path
		// to the given Resource.
		if resource != nil {
			path = getLocation("", resource, path)
		}

		obj := getObjectGCS(client, bucket, path)
		w := obj.NewWriter(ctx)

		if _, err := io.Copy(w, body); err != nil {
			return err
		}

		if err := w.Close(); err != nil {
			return err
		}

		return nil
	}
}

// deleteFileS3v1 generates a function that allows to delete a single file in a path.
func deleteFileGCS(client *storage.Client, bucket string, resource Resource) WalkDirFunc {
	return func(ctx context.Context, path string, _ io.Reader) error {
		// If Resource is nil, it will use the given path as-is, otherwise it will use the given path as a relative path
		// to the given Resource.
		if resource != nil {
			path = getLocation("", resource, path)
		}

		obj := getObjectGCS(client, bucket, path)

		if err := obj.Delete(ctx); err != nil {
			return err
		}
		return nil
	}
}

func NewGCS(client *storage.Client, bucket string) Storage {
	return &gcs{
		client: client,
		bucket: bucket,
	}
}
