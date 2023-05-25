package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/gazebo-web/gz-go/v7/errors"
	"github.com/gazebo-web/gz-go/v7/reflect"
	"github.com/gazebo-web/gz-go/v7/repository"
)

// firestoreRepository implements Repository using the firestore client.
type firestoreRepository[T repository.Model] struct {
	client *firestore.Client
}

// FirstOrCreate is not implemented.
func (r *firestoreRepository[T]) FirstOrCreate(entity repository.Model, filters ...repository.Filter) error {
	return errors.ErrMethodNotImplemented
}

// Create is not implemented.
func (r *firestoreRepository[T]) Create(entity repository.Model) (repository.Model, error) {
	return nil, errors.ErrMethodNotImplemented
}

// CreateBulk is not implemented.
func (r *firestoreRepository[T]) CreateBulk(entities []repository.Model) ([]repository.Model, error) {
	return nil, errors.ErrMethodNotImplemented
}

// Find filters entries and stores filtered entries in output.
// output: will contain the result of the query. It must be a pointer to a slice.
// options: configuration options for the search.
func (r *firestoreRepository[T]) Find(output interface{}, options ...repository.Option) error {
	col := r.client.Collection(r.Model().TableName())
	r.applyOptions(&col.Query, options...)
	iter := col.Documents(context.Background())
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
func (r *firestoreRepository[T]) FindOne(output repository.Model, filters ...repository.Filter) error {
	return errors.ErrMethodNotImplemented
}

// Last is not implemented.
func (r *firestoreRepository[T]) Last(output repository.Model, filters ...repository.Filter) error {
	return errors.ErrMethodNotImplemented
}

// Update is not implemented.
func (r *firestoreRepository[T]) Update(data interface{}, filters ...repository.Filter) error {
	return errors.ErrMethodNotImplemented
}

// Delete is not implemented.
func (r *firestoreRepository[T]) Delete(filters ...repository.Filter) error {
	return errors.ErrMethodNotImplemented
}

// Count is not implemented.
func (r *firestoreRepository[T]) Count(filters ...repository.Filter) (uint64, error) {
	return 0, errors.ErrMethodNotImplemented
}

// Model returns this repository's model.
func (r *firestoreRepository[T]) Model() repository.Model {
	var baseModel T
	return baseModel
}

func (r *firestoreRepository[T]) applyOptions(q *firestore.Query, opts ...repository.Option) {
	for _, opt := range opts {
		opt.(Option)(q)
	}
}

// NewFirestoreRepository initializes a new Repository implementation for Firestore collections.
func NewFirestoreRepository[T repository.Model](client *firestore.Client) repository.Repository {
	return &firestoreRepository[T]{
		client: client,
	}
}
