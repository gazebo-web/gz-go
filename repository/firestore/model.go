package firestore

import (
	"github.com/gazebo-web/gz-go/v7/repository"
	"time"
)

// Modeler is a specific repository.Model interface for Firestore repositories. It includes a set of methods to
// load firestore.DocumentSnapshot metadata into Model.
type Modeler interface {
	repository.Model
	// SetUID sets the given UID as the document identifier.
	SetUID(uid string) Modeler
	// SetCreatedAt sets the given createdAt value as the creation timestamp.
	SetCreatedAt(createdAt time.Time) Modeler
	// SetUpdatedAt sets the given updatedAt value as the last update timestamp.
	SetUpdatedAt(updatedAt time.Time) Modeler
}

// Model implements the repository.Model interface for firestore.
// It provides a set of common generic fields and operations that partially implement the repository.Model interface.
// To use it, embed it in your application-specific repository.Model implementation.
type Model struct {
	// ID contains the Document ID.
	ID string
	// CreatedAt contains the date and time at which this model has been persisted.
	CreatedAt time.Time `firestore:"-"`
	// UpdatedAt contains the last date and time when this model has been updated.
	UpdatedAt time.Time `firestore:"-"`
}

// TableName is included to fulfill the Modeler interface. Application should override this method on each model.
func (m Model) TableName() string {
	return "default"
}

// SetUID sets the given UID as the document identifier.
func (m Model) SetUID(uid string) Modeler {
	m.ID = uid
	return m
}

// SetCreatedAt sets the given createdAt value as the timestamp when this document was created.
func (m Model) SetCreatedAt(createdAt time.Time) Modeler {
	m.CreatedAt = createdAt
	return m
}

// SetUpdatedAt sets the given updatedAt value as the timestamp when this document was last updated.
func (m Model) SetUpdatedAt(updatedAt time.Time) Modeler {
	m.UpdatedAt = updatedAt
	return m
}
