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
)

type Kind string

const (
	KindModels Kind = "models"
	KindWorlds Kind = "worlds"
)

// Resource represents the resource that a user wants to download from a cloud storage.
type Resource interface {
	GetUUID() string
	GetKind() Kind
	GetOwner() string
	GetVersion() uint64
}

// validateResource validates the given resource, it returns an error if the resource is invalid.
// This validation step only performs sanity checks, but doesn't apply any business rules.
func validateResource(r Resource) error {
	if len(r.GetOwner()) == 0 {
		return errors.Wrap(ErrResourceInvalidFormat, "missing owner")
	}
	if len(r.GetKind()) == 0 {
		return errors.Wrap(ErrResourceInvalidFormat, "missing kind")
	}
	if u, err := uuid.FromString(r.GetUUID()); err != nil || u.Version() != uuid.V4 {
		return errors.Wrap(ErrResourceInvalidFormat, "invalid uuid")
	}
	if r.GetVersion() == 0 {
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
