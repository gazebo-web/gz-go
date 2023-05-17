package repository

import (
	"cloud.google.com/go/firestore"
	"errors"
)

var (
	ErrMethodNotImplemented = errors.New("method not implemented")
)

type firestoreDB struct {
	entity Model
	client *firestore.Client
}

func (f *firestoreDB) FirstOrCreate(entity Model, filters ...Filter) error {
	return ErrMethodNotImplemented
}

func (f *firestoreDB) Create(entity Model) (Model, error) {
	return nil, ErrMethodNotImplemented
}

func (f *firestoreDB) CreateBulk(entities []Model) ([]Model, error) {
	return nil, ErrMethodNotImplemented
}

func (f *firestoreDB) Find(output interface{}, options ...Option) error {
	return ErrMethodNotImplemented
}

func (f *firestoreDB) FindOne(output Model, filters ...Filter) error {
	return ErrMethodNotImplemented
}

func (f *firestoreDB) Last(output Model, filters ...Filter) error {
	return ErrMethodNotImplemented
}

func (f *firestoreDB) Update(data interface{}, filters ...Filter) error {
	return ErrMethodNotImplemented
}

func (f *firestoreDB) Delete(filters ...Filter) error {
	return ErrMethodNotImplemented
}

func (f *firestoreDB) Count(filters ...Filter) (uint64, error) {
	return 0, ErrMethodNotImplemented
}

func (f *firestoreDB) Model() Model {
	return f.entity
}

// NewFirestoreRepository initializes a new Repository implementation for Firestore collections.
func NewFirestoreRepository(client *firestore.Client, entity Model) Repository {
	return &firestoreDB{
		client: client,
		entity: entity,
	}
}
