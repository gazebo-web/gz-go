package repository

import (
	"github.com/gazebo-web/gz-go/v6/reflect"
	"github.com/jinzhu/gorm"
)

// NewRepository initializes a new Repository implementation for SQL databases.
func NewRepository(db *gorm.DB, entity Model) Repository {
	return &repositorySQL{
		DB:     db,
		entity: entity,
	}
}

// repositorySQL implements Repository using gorm to support SQL databases.
type repositorySQL struct {
	DB     *gorm.DB
	entity Model
}

// Ensure that repositorySQL implements the Repository interface.
var _ Repository = (*repositorySQL)(nil)

// applyOptions applies operation options to a database query.
func (r *repositorySQL) applyOptions(q *gorm.DB, opts ...Option) {
	for _, opt := range opts {
		opt.(SQLOption)(q)
	}
}

// Create inserts a single entry.
//
//	entity: The entry to insert.
func (r *repositorySQL) Create(entity Model) (Model, error) {
	result, err := r.CreateBulk([]Model{entity})
	if err != nil {
		return nil, err
	}
	return result[0], nil
}

// CreateBulk is a bulk operation to create multiple entries with a single operation.
//
//	entities: should be a slice of the same data structure implementing Model.
func (r *repositorySQL) CreateBulk(entities []Model) ([]Model, error) {
	for _, entity := range entities {
		err := r.DB.Model(r.Model()).Create(entity).Error
		if err != nil {
			return nil, err
		}
	}
	return entities, nil
}

// Find filters entries and stores filtered entries in output.
//
//	output: will contain the result of the query. It must be a pointer to a slice.
//	offset: defines the number of results to skip before loading values to output.
//	limit: defines the maximum number of entries to return. A nil value returns infinite results.
//	filters: filter entries by field value.
func (r *repositorySQL) Find(output interface{}, options ...Option) error {
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
// a data structure implementing Model.
func (r *repositorySQL) FindOne(output Model, filters ...Filter) error {
	if len(filters) == 0 {
		return ErrNoFilter
	}
	q := r.startQuery()
	q = r.setQueryFilters(q, filters)
	q = q.First(output)
	return q.Error
}

// Last gets the last record ordered by primary key desc.
//
//	output: must be a pointer to a Model implementation.
func (r *repositorySQL) Last(output Model, filters ...Filter) error {
	if len(filters) == 0 {
		return ErrNoFilter
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
func (r *repositorySQL) Update(data interface{}, filters ...Filter) error {
	q := r.startQuery()
	q = r.setQueryFilters(q, filters)
	q = q.Update(data)
	return q.Error
}

// Delete removes all the model entries that match filters.
//
//	filters: filter entries that should be deleted.
func (r *repositorySQL) Delete(filters ...Filter) error {
	q := r.startQuery()
	q = r.setQueryFilters(q, filters)
	q = q.Delete(r.Model())
	err := q.Error
	if err != nil {
		return err
	}
	return nil
}

// FirstOrCreate inserts a new entry if the given filters don't find any existing record.
//
//	entity: must be a pointer to a Model implementation. Results will be saved in this argument if the record exists.
func (r *repositorySQL) FirstOrCreate(entity Model, filters ...Filter) error {
	q := r.startQuery()
	q = r.setQueryFilters(q, filters)
	return q.FirstOrCreate(entity).Error
}

// Count counts all the model entries that match filters.
//
//	filters: selection criteria for entries that should be considered when counting entries.
func (r *repositorySQL) Count(filters ...Filter) (uint64, error) {
	var count uint64
	q := r.startQuery()
	q = r.setQueryFilters(q, filters)
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// startQuery inits a gorm query for this repository's model. Multiple filters are ANDd together.
func (r *repositorySQL) startQuery() *gorm.DB {
	return r.DB.Model(r.Model())
}

// setQueryFilters applies the given filters to a gorm query.
func (r *repositorySQL) setQueryFilters(q *gorm.DB, filters []Filter) *gorm.DB {
	for _, f := range filters {
		q = q.Where(f.Template, f.Values...)
	}
	return q
}

// Model returns this repository's model.
func (r *repositorySQL) Model() Model {
	return reflect.NewInstance(r.entity).(Model)
}
