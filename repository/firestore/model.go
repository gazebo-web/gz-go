package firestore

import (
	"github.com/gazebo-web/gz-go/v7/repository"
	"time"
)

// Modeler is a specific repository.Model interface for Firestore repositories. It includes a set of methods to
// load firestore.DocumentSnapshot metadata into Model.
type Modeler interface {
	repository.Model
	NewModel() Modeler
	// SetUID sets the given UID as the document identifier.
	SetUID(uid string)
	// SetCreatedAt sets the given createdAt value as the creation timestamp.
	SetCreatedAt(createdAt time.Time)
	// SetUpdatedAt sets the given updatedAt value as the last update timestamp.
	SetUpdatedAt(updatedAt time.Time)
}

// Model implements the repository.Model interface for firestore.
// It provides a set of common generic fields and operations that partially implement the repository.Model interface.
// To use it, embed it in your application-specific repository.Model implementation.
type Model struct {
	// ID contains the Document ID.
	ID string `firestore:"-"`
	// CreatedAt contains the date and time at which this model has been persisted.
	CreatedAt time.Time `firestore:"-"`
	// UpdatedAt contains the last date and time when this model has been updated.
	UpdatedAt time.Time `firestore:"-"`
}

func (m *Model) NewModel() Modeler {
	return new(Model)
}

// TableName is included to fulfill the Modeler interface. Application should override this method on each model.
func (*Model) TableName() string {
	return "default"
}

// SetUID sets the given UID as the document identifier.
func (m *Model) SetUID(uid string) {
	m.ID = uid
}

// SetCreatedAt sets the given createdAt value as the timestamp when this document was created.
func (m *Model) SetCreatedAt(createdAt time.Time) {
	m.CreatedAt = createdAt
}

// SetUpdatedAt sets the given updatedAt value as the timestamp when this document was last updated.
func (m *Model) SetUpdatedAt(updatedAt time.Time) {
	m.UpdatedAt = updatedAt
}
