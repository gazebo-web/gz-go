package firestore

import "time"

// Model implements the repository.Model interface for firestore.
// It provides a set of common generic fields and operations that partially implement the repository.Model interface.
// To use it, embed it in your application-specific repository.Model implementation.
type Model struct {
	// CreatedAt contains the date and time at which this model has been persisted.
	CreatedAt time.Time `firestore:"created_at"`
	// UpdatedAt contains the last date and time when this model has been updated.
	UpdatedAt time.Time `firestore:"updated_at"`
	// DeletedAt is used to implement soft record deletion. If set, the record will be considered
	// as deleted.
	DeletedAt *time.Time `firestore:"deleted_at"`
}

// GetID returns the unique identifier for this Model.
func (m Model) GetID() uint {
	return 0
}
