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

// Kind describes the subfolder where different types of resources will be uploaded for a single
// owner. Developers can create different resources, but a set of basic kinds are provided in this library.
type Kind string

const (
	KindModels      Kind = "models"
	KindWorlds      Kind = "worlds"
	KindCollections Kind = "collections"
)

// Resource represents the resource that a user wants to download from a cloud storage.
type Resource interface {
	// GetUUID returns the UUID v4 that identifies the current Resource.
	GetUUID() string
	// GetKind identifies of what type the current Resource is.
	GetKind() Kind
	// GetOwner returns who's the owner of the current Resource.
	GetOwner() string
	// GetVersion returns the numeric version of the current Resource. Resources increment their version as new
	// updates are introduced to them.
	GetVersion() uint64
}

// validateResource validates the given resource, it returns an error if the resource is invalid.
// This validation step only performs sanity checks, but doesn't apply any business rules.
func validateResource(r Resource) error {
	if err := validateOwner(r.GetOwner()); err != nil {
		return err
	}
	if err := validateKind(r.GetKind()); err != nil {
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

// validateOwner performs validation to the given owner.
func validateOwner(owner string) error {
	if len(owner) == 0 {
		return errors.Wrap(ErrResourceInvalidFormat, "missing owner")
	}
	return nil
}

// validateKind performs validation to the given kind.
func validateKind(kind Kind) error {
	if len(kind) == 0 {
		return errors.Wrap(ErrResourceInvalidFormat, "missing kind")
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

// validateVersion performs validation to the given version.
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
	location := filepath.Join(base, r.GetOwner(), string(r.GetKind()), r.GetUUID(), strconv.FormatUint(r.GetVersion(), 10))
	if len(path) > 0 {
		location = filepath.Join(location, path)
	}
	return location
}

// getZipLocation returns the location of the zip file associated to a Resource relative to the base location.
func getZipLocation(base string, r Resource) string {
	filename := fmt.Sprintf("%d.zip", r.GetVersion())
	return filepath.Join(base, r.GetOwner(), string(r.GetKind()), r.GetUUID(), ".zips", filename)
}

// getRootLocation returns the absolute location of where all the versions of the given uuid and the given kind will be
// uploaded for the given owner.
func getRootLocation(base string, owner string, kind Kind, uuid string) string {
	return filepath.Join(base, owner, string(kind), uuid)
}
