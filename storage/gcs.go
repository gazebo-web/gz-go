package storage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/gazebo-web/gz-go/v7"
	"io"
	"os"
	"time"
)

// gcs implements Storage using the Google Cloud Storage (GCS) service.
//
//	Reference: https://cloud.google.com/storage
//	API: https://cloud.google.com/storage/docs/apis
//	SDK: https://pkg.go.dev/cloud.google.com/go/storage
type gcs struct {
	// client contains a reference to the Google Cloud Storage SDK client.
	client *storage.Client

	// bucket contains the name of the bucket used to upload resources and zip files.
	bucket string

	// privateKey is the Google service account private key. It is obtainable
	// from the Google Developers Console.
	// At https://console.developers.google.com/project/<your-project-id>/apiui/credential,
	// create a service account client ID or reuse one of your existing service account
	// credentials. Click on the "Generate new P12 key" to generate and download
	// a new private key. Once you download the P12 file, use the following command
	// to convert it into a PEM file.
	//
	//    $ openssl pkcs12 -in key.p12 -passin pass:notasecret -out key.pem -nodes
	//
	privateKey []byte

	// accessID represents the authorizer of the signed URL generation. It is typically the Google service account
	// client email address from the Google Developers Console in the form of "xxx@developer.gserviceaccount.com".
	accessID string

	// duration defines the lifespan of a Pre-signed URL.
	duration time.Duration
}

// getObjectGCS is a helper function that gets the object reference in the given bucket identified by the given path.
func getObjectGCS(client *storage.Client, bucket string, path string) *storage.ObjectHandle {
	obj := client.Bucket(bucket).Object(path)
	return obj
}

// UploadZip uploads a zip file of the given resource to GCS. It should be called before any attempts to Download
// the zip file of the given Resource.
//
//	Resources can have a compressed representation of the resource itself that acts like a cache, it contains all the
//	files from the said resource. This function uploads that zip file.
func (g *gcs) UploadZip(ctx context.Context, resource Resource, file *os.File) error {
	return UploadZip(ctx, resource, file, uploadFileGCS(g.client, g.bucket, nil))
}

// UploadDir uploads the entire src directory to GCS.
func (g *gcs) UploadDir(ctx context.Context, resource Resource, src string) error {
	return UploadDir(ctx, resource, src, uploadFileGCS(g.client, g.bucket, resource))
}

// Download returns the URL to a zip file that contains all the contents of the given Resource.
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
		GoogleAccessID: g.accessID,
		PrivateKey:     g.privateKey,
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		Expires:        time.Now().Add(g.duration),
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

// uploadFileGCS generates a function that uploads a single file in a path.
func uploadFileGCS(client *storage.Client, bucket string, resource Resource) WalkDirFunc {
	return func(ctx context.Context, path string, body io.Reader) error {
		// If Resource is nil, it will use the given path as-is, otherwise it will use the given path as a relative path
		// to the given Resource.
		if resource != nil {
			path = getLocation("", resource, path)
		}

		obj := getObjectGCS(client, bucket, path)
		w := obj.NewWriter(ctx)
		defer gz.Close(w)

		if _, err := io.Copy(w, body); err != nil {
			return err
		}

		return nil
	}
}

// deleteFileGCS generates a function that allows to delete a single file in a path.
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

// NewGCS initializes a new implementation of Storage using the Google Cloud Storage service.
func NewGCS(client *storage.Client, bucket string, pk []byte, accessID string) Storage {
	return &gcs{
		client:     client,
		bucket:     bucket,
		accessID:   accessID,
		privateKey: pk,
		duration:   5 * time.Minute,
	}
}
