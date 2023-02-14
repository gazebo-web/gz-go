package storage

import (
	"fmt"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"path/filepath"
	"strconv"
)

var (
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceInvalidFormat = errors.New("invalid resource format")
	ErrEmptyResource         = errors.New("resource has no content")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrSourceFolderNotFound  = errors.New("source folder not found")
	ErrSourceFolderEmpty     = errors.New("source folder is empty")
	ErrSourceFile            = errors.New("source is a file, should be a folder")
	ErrFileNil               = errors.New("no file provided")
)

// Resource represents the resource that a user wants to download from a cloud storage.
type Resource interface {
	// GetUUID returns the UUID v4 that identifies the current Resource.
	GetUUID() string
	// GetOwner returns who is the owner of the current Resource.
	GetOwner() string
	// GetVersion returns the numeric version of the current Resource. Resources increment their version as new
	// updates are introduced to them.
	GetVersion() uint64
}

// NewResource initializes a new Resource with the given values.
func NewResource(uuid string, owner string, version uint64) Resource {
	return &resource{
		uuid:    uuid,
		owner:   owner,
		version: version,
	}
}

// resource is the default implementation of Resource provided  by this package.
type resource struct {
	uuid    string
	owner   string
	version uint64
}

// GetUUID returns this resource's uuid.
func (r *resource) GetUUID() string {
	return r.uuid
}

// GetOwner returns this resource's owner.
func (r *resource) GetOwner() string {
	return r.owner
}

// GetVersion returns this resource's version.
func (r *resource) GetVersion() uint64 {
	return r.version
}

// validateResource validates the given resource, it returns an error if the resource is invalid.
// This validation step only performs sanity checks, but doesn't apply any business rules.
func validateResource(r Resource) error {
	if err := validateOwner(r.GetOwner()); err != nil {
		return err
	}
	if err := validateUUID(r.GetUUID()); err != nil {
		return err
	}
	if err := validateVersion(r.GetVersion()); err != nil {
		return err
	}
	return nil
}

// validateOwner validates the given owner.
func validateOwner(owner string) error {
	if len(owner) == 0 {
		return errors.Wrap(ErrResourceInvalidFormat, "missing owner")
	}
	return nil
}

// validateUUID performs validation to the given UUID.
func validateUUID(id string) error {
	if u, err := uuid.FromString(id); err != nil || u.Version() != uuid.V4 {
		return errors.Wrap(ErrResourceInvalidFormat, "invalid uuid")
	}
	return nil
}

// validateVersion validates the given version.
func validateVersion(v uint64) error {
	if v == 0 {
		return errors.Wrap(ErrResourceInvalidFormat, "invalid version, should be greater than 0")
	}
	return nil
}

// getLocation returns the location of a Resource relative to the base location.
//
//	If path is not empty, it will append the given path to the resulting location of the resource.
func getLocation(base string, r Resource, path string) string {
	location := filepath.Join(base, r.GetOwner(), r.GetUUID(), strconv.FormatUint(r.GetVersion(), 10))
	if len(path) > 0 {
		location = filepath.Join(location, path)
	}
	return location
}

// getZipLocation returns the location of the zip file associated to a Resource relative to the base location.
func getZipLocation(base string, r Resource) string {
	filename := fmt.Sprintf("%d.zip", r.GetVersion())
	return filepath.Join(base, r.GetOwner(), r.GetUUID(), ".zips", filename)
}

// getRootLocation returns the absolute location of where all the versions of the given uuid and the given kind will be
// uploaded for the given owner.
func getRootLocation(base string, owner string, uuid string) string {
	return filepath.Join(base, owner, uuid)
}
