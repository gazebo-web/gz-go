package repository

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/gazebo-web/gz-go/v7/reflect"
)

// firestoreRepository implements Repository using the firestore client.
type firestoreRepository[T Model] struct {
	client *firestore.Client
}

// FirstOrCreate is not implemented.
func (f *firestoreRepository[T]) FirstOrCreate(entity Model, filters ...Filter) error {
	return ErrMethodNotImplemented
}

// Create is not implemented.
func (f *firestoreRepository[T]) Create(entity Model) (Model, error) {
	return nil, ErrMethodNotImplemented
}

// CreateBulk is not implemented.
func (f *firestoreRepository[T]) CreateBulk(entities []Model) ([]Model, error) {
	return nil, ErrMethodNotImplemented
}

// Find filters entries and stores filtered entries in output.
// output: will contain the result of the query. It must be a pointer to a slice.
// options: configuration options for the search.
func (f *firestoreRepository[T]) Find(output interface{}, options ...Option) error {
	iter := f.client.Collection(f.Model().TableName()).Documents(context.Background())
	docs, err := iter.GetAll()
	if err != nil {
		return err
	}

	var element T
	for _, doc := range docs {
		if err := doc.DataTo(&element); err != nil {
			continue
		}

		if err := reflect.AppendToSlice(output, element); err != nil {
			continue
		}
	}

	return nil
}

// FindOne is not implemented.
func (f *firestoreRepository[T]) FindOne(output Model, filters ...Filter) error {
	return ErrMethodNotImplemented
}

// Last is not implemented.
func (f *firestoreRepository[T]) Last(output Model, filters ...Filter) error {
	return ErrMethodNotImplemented
}

// Update is not implemented.
func (f *firestoreRepository[T]) Update(data interface{}, filters ...Filter) error {
	return ErrMethodNotImplemented
}

// Delete is not implemented.
func (f *firestoreRepository[T]) Delete(filters ...Filter) error {
	return ErrMethodNotImplemented
}

// Count is not implemented.
func (f *firestoreRepository[T]) Count(filters ...Filter) (uint64, error) {
	return 0, ErrMethodNotImplemented
}

// Model returns this repository's model.
func (f *firestoreRepository[T]) Model() Model {
	var baseModel T
	return baseModel
}

// NewFirestoreRepository initializes a new Repository implementation for Firestore collections.
func NewFirestoreRepository[T Model](client *firestore.Client) Repository {
	return &firestoreRepository[T]{
		client: client,
	}
}
