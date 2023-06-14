package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/gazebo-web/gz-go/v7/errors"
	"github.com/gazebo-web/gz-go/v7/reflect"
	"github.com/gazebo-web/gz-go/v7/repository"
	"google.golang.org/api/iterator"
)

// firestoreRepository implements Repository using the firestore client.
type firestoreRepository[T Modeler[T]] struct {
	// client holds a reference to the firestore client in order to read, persist and modify collections.
	client *firestore.Client
}

// FirstOrCreate is not implemented.
func (r *firestoreRepository[T]) FirstOrCreate(entity repository.Model, filters ...repository.Filter) error {
	return errors.ErrMethodNotImplemented
}

// Create is not implemented.
func (r *firestoreRepository[T]) Create(entity repository.Model) error {
	return errors.ErrMethodNotImplemented
}

// CreateBulk is not implemented.
func (r *firestoreRepository[T]) CreateBulk(entities []repository.Model) error {
	return errors.ErrMethodNotImplemented
}

// Find filters entries and stores filtered entries in output.
//
//	output: will contain the result of the query. It must be a pointer to a slice.
//	options: configuration options for the search.
func (r *firestoreRepository[T]) Find(output interface{}, options ...repository.Option) error {
	col := r.client.Collection(r.Model().TableName())
	r.applyOptions(&col.Query, options...)
	iter := col.Documents(context.Background())
	docs, err := iter.GetAll()
	if err != nil {
		return err
	}

	var out []T
	for _, doc := range docs {
		var element T
		element = element.FromDocumentSnapshot(doc)
		out = append(out, element)
	}
	return reflect.SetValue(output, out)
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

// Delete deletes all the entities that match the given options.
//
// This method is not responsible for performing soft deletes.
// Any project using this repository must implement soft deletion at the firestore-level if they're in need of soft
// deletes. Consider using something like https://extensions.dev/extensions/adamnathanlewis/ext-firestore-soft-deletes
// We DO NOT recommend any third-party extension, and they're only presented here as an example of what can be used
// to implement soft deletes.
//
// Delete does not remove all the records at once, it will perform the document removal in small batches. This mechanism
// prevents running into out-of-memory errors.
func (r *firestoreRepository[T]) Delete(options ...repository.Option) error {
	ctx := context.Background()
	col := r.client.Collection(r.Model().TableName())
	r.applyOptions(&col.Query, options...)

	err := r.deleteBatch(ctx, col, 30)
	if err != nil {
		return err
	}

	return nil
}

// deleteBatch is a helper function that allows deleting documents in small batches of the given size.
func (r *firestoreRepository[T]) deleteBatch(ctx context.Context, col *firestore.CollectionRef, size int) error {
	writer := r.client.BulkWriter(ctx)
	for {
		// Get the next batch of documents
		iter := col.Limit(size).Documents(ctx)

		// Track the number of deleted records in this batch
		deleted := 0

		// Iterate over the current batch of documents and delete them
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}

			_, err = writer.Delete(doc.Ref)
			if err != nil {
				return err
			}
			deleted++
		}

		// If no documents were deleted, there are no more documents available and the process is over.
		if deleted == 0 {
			writer.End()
			break
		}

		writer.Flush()
	}
	return nil
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
func NewFirestoreRepository[T Modeler[T]](client *firestore.Client) repository.Repository {
	return &firestoreRepository[T]{
		client: client,
	}
}
