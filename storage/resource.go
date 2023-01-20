package storage

import (
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

var (
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceInvalidFormat = errors.New("invalid resource format")
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
