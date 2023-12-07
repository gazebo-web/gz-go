package sql

import "time"

// Model implements the repository.Model interface for SQL backends.
// It provides a set of common generic fields and operations that partially implement the repository.Model interface.
// To use it, embed it in your application-specific repository.Model implementation.
type Model struct {
	// ID contains the primary key identifier.
	ID uint `gorm:"primaryKey" firestore:"-"`
	// CreatedAt contains the date and time at which this model has been persisted.
	CreatedAt time.Time `json:"created_at" firestore:"-"`
	// UpdatedAt contains the last date and time when this model has been updated.
	UpdatedAt time.Time `json:"updated_at" firestore:"-"`
	// DeletedAt is used to implement soft record deletion. If set, the record will be considered
	// as deleted.
	DeletedAt *time.Time `json:"deleted_at" gorm:"index" firestore:"-"`
}

// GetID returns the unique identifier for this Model.
func (m Model) GetID() uint {
	return m.ID
}
