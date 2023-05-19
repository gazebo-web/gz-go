package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/gazebo-web/gz-go/v7/reflect"
	"github.com/gazebo-web/gz-go/v7/repository"
)

// firestoreRepository implements Repository using the firestore client.
type firestoreRepository[T repository.Model] struct {
	client *firestore.Client
}

// FirstOrCreate is not implemented.
func (f *firestoreRepository[T]) FirstOrCreate(entity repository.Model, filters ...repository.Filter) error {
	return repository.ErrMethodNotImplemented
}

// Create is not implemented.
func (f *firestoreRepository[T]) Create(entity repository.Model) (repository.Model, error) {
	return nil, repository.ErrMethodNotImplemented
}

// CreateBulk is not implemented.
func (f *firestoreRepository[T]) CreateBulk(entities []repository.Model) ([]repository.Model, error) {
	return nil, repository.ErrMethodNotImplemented
}

// Find filters entries and stores filtered entries in output.
// output: will contain the result of the query. It must be a pointer to a slice.
// options: configuration options for the search.
func (f *firestoreRepository[T]) Find(output interface{}, options ...repository.Option) error {
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
func (f *firestoreRepository[T]) FindOne(output repository.Model, filters ...repository.Filter) error {
	return repository.ErrMethodNotImplemented
}

// Last is not implemented.
func (f *firestoreRepository[T]) Last(output repository.Model, filters ...repository.Filter) error {
	return repository.ErrMethodNotImplemented
}

// Update is not implemented.
func (f *firestoreRepository[T]) Update(data interface{}, filters ...repository.Filter) error {
	return repository.ErrMethodNotImplemented
}

// Delete is not implemented.
func (f *firestoreRepository[T]) Delete(filters ...repository.Filter) error {
	return repository.ErrMethodNotImplemented
}

// Count is not implemented.
func (f *firestoreRepository[T]) Count(filters ...repository.Filter) (uint64, error) {
	return 0, repository.ErrMethodNotImplemented
}

// Model returns this repository's model.
func (f *firestoreRepository[T]) Model() repository.Model {
	var baseModel T
	return baseModel
}

// NewFirestoreRepository initializes a new Repository implementation for Firestore collections.
func NewFirestoreRepository[T repository.Model](client *firestore.Client) repository.Repository {
	return &firestoreRepository[T]{
		client: client,
	}
}
