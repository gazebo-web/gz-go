package repository

import (
	"context"
	"errors"
)

var (
	// ErrNoFilter represents an error when no filter are provided.
	ErrNoFilter = errors.New("no filters provided")
	// ErrNoEntriesUpdated represent an error when no entries were updated in the database
	// after an Update operation.
	ErrNoEntriesUpdated = errors.New("no entries were updated")
	// ErrNoEntriesDeleted represent an error when no entries were deleted in the database
	// after a Delete operation.
	ErrNoEntriesDeleted = errors.New("no entries were deleted")
)

// Option is used to define repository operation options for the generic Repository interface.
// It is expected that each Repository implementation will have its own set of concrete options available, and those
// options are not expected to be compatible with other implementations.
type Option interface {
	// IsOption is a dummy method used to identify repository options.
	IsOption()
}

// Repository holds methods to CRUD an entity on a certain persistence layer.
type Repository interface {
	// FirstOrCreate inserts a new entry if the given filters don't find any existing record.
	// entity: must be a pointer to a Model implementation. Results will be saved in this argument if the record exists.
	FirstOrCreate(ctx context.Context, entity Model, filters ...Filter) error
	// Create inserts a single entry.
	// entity: The entry to insert.
	Create(ctx context.Context, entity Model) (Model, error)
	// CreateBulk creates multiple entries with a single operation.
	// entities: should be a slice of a Model implementation.
	CreateBulk(ctx context.Context, entities []Model) ([]Model, error)
	// Find filters entries and stores filtered entries in output.
	// output: will contain the result of the query. It must be a pointer to a slice.
	// options: configuration options for the search. Refer to the implementation's set of options to get a lit of options.
	Find(ctx context.Context, output interface{}, options ...Option) error
	// FindOne filters entries and stores the first filtered entry in output.
	// output: must be a pointer to a Model implementation.
	FindOne(ctx context.Context, output Model, filters ...Filter) error
	// Last gets the last record ordered by primary key desc.
	// output: must be a pointer to a Model implementation.
	Last(ctx context.Context, output Model, filters ...Filter) error
	// Update updates all model entries that match the provided filters with the given data.
	// data: must be a map[string]interface{}
	// filters: selection criteria for entries that should be updated.
	Update(ctx context.Context, data interface{}, filters ...Filter) error
	// Delete removes all the model entries that match filters.
	// filters: selection criteria for entries that should be deleted.
	Delete(ctx context.Context, opts ...Option) error
	// Count counts all the model entries that match filters.
	// filters: selection criteria for entries that should be considered when counting entries.
	Count(ctx context.Context, filters ...Filter) (uint64, error)
	// Model returns this repository's model.
	Model() Model
}
