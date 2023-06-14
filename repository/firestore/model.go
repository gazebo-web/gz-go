package firestore

import (
	"cloud.google.com/go/firestore"
	"github.com/gazebo-web/gz-go/v7/repository"
	"time"
)

// Modeler is a specific repository.Model interface for Firestore repositories. It includes a set of methods to
// load firestore.DocumentSnapshot metadata into Model.
type Modeler[T any] interface {
	repository.Model
	FromDocumentSnapshot(doc *firestore.DocumentSnapshot) T
}

var _ Modeler[Model] = (*Model)(nil)

// Model implements the repository.Model interface for firestore.
// It provides a set of common generic fields and operations that partially implement the repository.Model interface.
// To use it, embed it in your application-specific repository.Model implementation.
// The following methods should be added to your application-specific model:
//
//   - TableName, if not, it defaults to the "default" table name.
//   - FromDocumentSnapshot: This method should call doc.DataTo(&entity) inside the application-specific method, as well as
//     the Model.FromDocumentSnapshot method. See the firestore_test.go file, there's a Test struct that serves as an example.
type Model struct {
	// ID contains the Document ID.
	ID string `firestore:"-"`
	// CreatedAt contains the date and time at which this model has been persisted.
	CreatedAt time.Time `firestore:"-"`
	// UpdatedAt contains the last date and time when this model has been updated.
	UpdatedAt time.Time `firestore:"-"`
}

// FromDocumentSnapshot parses the given DocumentSnapshot into the current Model.
func (m Model) FromDocumentSnapshot(doc *firestore.DocumentSnapshot) Model {
	if doc == nil {
		return Model{}
	}
	m.ID = doc.Ref.ID
	m.CreatedAt = doc.CreateTime
	m.UpdatedAt = doc.UpdateTime
	return m
}

// TableName is included to fulfill the Modeler interface. Application should override this method on each model.
func (*Model) TableName() string {
	return "default"
}
