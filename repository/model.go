package repository

// Model represents a generic entity. A Model is part of the domain layer and is  persisted by a certain Repository.
type Model interface {
	// GetID returns the ID of the model that has been persisted. It returns -1 if no value has been defined.
	GetID() int
	// TableName returns the table/collection name for a certain model.
	TableName() string
}
