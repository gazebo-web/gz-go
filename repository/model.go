package repository

import (
	"time"
)

// Model represents a generic entity. A Model is part of the domain layer and is  persisted by a certain Repository.
type Model interface {
	// TableName returns the table/collection name for a certain model.
	TableName() string
	// GetID returns the unique identifier for this Model.
	// It returns the ID of the model that has been persisted. It returns 0 if no value has been defined.
	// For SQL environments, this would be the primary key.
	GetID() uint
}

// ModelSQL implements Model for SQL backends.
// It provides a set of common generic fields and operations that partially implement the repository.Model interface.
// To use it, embed it in your application-specific repository.Model implementation.
type ModelSQL struct {
	// ID contains the primary key identifier.
	ID uint `gorm:"primary_key"`
	// CreatedAt contains the date and time at which this model has been persisted.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt contains the last date and time when this model has been updated.
	UpdatedAt time.Time `json:"updated_at"`
	// DeletedAt is used to implement soft record deletion. If set, the record will be considered
	// as deleted.
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}

// GetID returns the unique identifier for this Model.
func (m ModelSQL) GetID() uint {
	return m.ID
}
