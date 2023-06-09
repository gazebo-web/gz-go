package sql

import (
	"github.com/gazebo-web/gz-go/v7/errors"
	"github.com/gazebo-web/gz-go/v7/reflect"
	"github.com/gazebo-web/gz-go/v7/repository"
	"github.com/jinzhu/gorm"
)

// NewRepository initializes a new repository.Repository implementation for SQL databases.
func NewRepository(db *gorm.DB, entity repository.Model) repository.Repository {
	return &repositoryGorm{
		DB:     db,
		entity: entity,
	}
}

// repositoryGorm implements a SQL repository.Repository implementation using Gorm.
type repositoryGorm struct {
	DB     *gorm.DB
	entity repository.Model
}

// Ensure that repositoryGorm implements the repository.Repository interface.
var _ repository.Repository = (*repositoryGorm)(nil)

// applyOptions applies operation options to a database query.
func (r *repositoryGorm) applyOptions(q *gorm.DB, opts ...repository.Option) {
	for _, opt := range opts {
		opt.(Option)(q)
	}
}

// Create inserts a single entry.
//
//	entity: The entry to insert.
func (r *repositoryGorm) Create(entity repository.Model) error {
	err := r.CreateBulk([]repository.Model{entity})
	if err != nil {
		return err
	}
	return nil
}

// CreateBulk creates multiple entries with a single operation.
//
//	entities: should be a slice of the same data structure implementing repository.Model.
func (r *repositoryGorm) CreateBulk(entities []repository.Model) error {
	for _, entity := range entities {
		err := r.DB.Model(r.Model()).Create(entity).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// Find filters entries and stores filtered entries in output.
//
//	output: will contain the result of the query. It must be a pointer to a slice.
//	options: configuration options for the search.
func (r *repositoryGorm) Find(output interface{}, options ...repository.Option) error {
	q := r.startQuery()
	r.applyOptions(q, options...)
	q = q.Find(output)
	err := q.Error
	if err != nil {
		return err
	}
	return nil
}

// FindOne filters entries and stores the first filtered entry in output, it must be a pointer to
// a data structure implementing repository.Model.
func (r *repositoryGorm) FindOne(output repository.Model, filters ...repository.Filter) error {
	if len(filters) == 0 {
		return repository.ErrNoFilter
	}
	q := r.startQuery()
	q = r.setQueryFilters(q, filters)
	q = q.First(output)
	return q.Error
}

// Last gets the last record ordered by primary key desc.
//
//	output: must be a pointer to a repository.Model implementation.
func (r *repositoryGorm) Last(output repository.Model, filters ...repository.Filter) error {
	if len(filters) == 0 {
		return repository.ErrNoFilter
	}
	q := r.startQuery()
	q = r.setQueryFilters(q, filters)
	q = q.Last(output)
	return q.Error
}

// Update updates all model entries that match the provided filters with the given data.
//
//		data: must be a map[string]interface{}
//	 filters: filter entries that should be updated.
func (r *repositoryGorm) Update(data interface{}, filters ...repository.Filter) error {
	q := r.startQuery()
	q = r.setQueryFilters(q, filters)
	q = q.Update(data)
	return q.Error
}

// Delete removes all the model entries that match filters.
//
//	options: configuration options for the removal.
func (r *repositoryGorm) DeleteBulk(opts ...repository.Option) error {
	q := r.startQuery()
	r.applyOptions(q, opts...)
	q = q.Delete(r.Model())
	err := q.Error
	if err != nil {
		return err
	}
	return nil
}

// FirstOrCreate inserts a new entry if the given filters don't find any existing record.
//
//	entity: must be a pointer to a repository.Model implementation. Results will be saved in this argument if the record exists.
func (r *repositoryGorm) FirstOrCreate(entity repository.Model, filters ...repository.Filter) error {
	q := r.startQuery()
	q = r.setQueryFilters(q, filters)
	return q.FirstOrCreate(entity).Error
}

// Count counts all the model entries that match filters.
//
//	filters: selection criteria for entries that should be considered when counting entries.
func (r *repositoryGorm) Count(filters ...repository.Filter) (uint64, error) {
	var count uint64
	q := r.startQuery()
	q = r.setQueryFilters(q, filters)
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Delete removes a single model entry that matches the given id.
func (r *repositoryGorm) Delete(_ interface{}) error {
	return errors.ErrMethodNotImplemented
}

// startQuery inits a sql query for this repository's model. Multiple filters are ANDd together.
func (r *repositoryGorm) startQuery() *gorm.DB {
	return r.DB.Model(r.Model())
}

// setQueryFilters applies the given filters to a sql query.
func (r *repositoryGorm) setQueryFilters(q *gorm.DB, filters []repository.Filter) *gorm.DB {
	for _, f := range filters {
		q = q.Where(f.Template, f.Values...)
	}
	return q
}

// Model returns this repository's model.
func (r *repositoryGorm) Model() repository.Model {
	return reflect.NewInstance(r.entity).(repository.Model)
}
