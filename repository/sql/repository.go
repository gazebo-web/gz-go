package sql

import (
	"context"
	"github.com/gazebo-web/gz-go/v8/reflect"
	"github.com/gazebo-web/gz-go/v8/repository"
	"gorm.io/gorm"
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
func (r *repositoryGorm) Create(ctx context.Context, entity repository.Model) (repository.Model, error) {
	result, err := r.CreateBulk(ctx, []repository.Model{entity})
	if err != nil {
		return nil, err
	}
	return result[0], nil
}

// CreateBulk creates multiple entries with a single operation.
//
//	entities: should be a slice of the same data structure implementing repository.Model.
func (r *repositoryGorm) CreateBulk(ctx context.Context, entities []repository.Model) ([]repository.Model, error) {
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
//	options: configuration options for the search.
func (r *repositoryGorm) Find(ctx context.Context, output interface{}, options ...repository.Option) error {
	q := r.startQuery(ctx)
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
func (r *repositoryGorm) FindOne(ctx context.Context, output repository.Model, filters ...repository.Filter) error {
	if len(filters) == 0 {
		return repository.ErrNoFilter
	}
	q := r.startQuery(ctx)
	q = r.setQueryFilters(q, filters)
	q = q.First(output)
	return q.Error
}

// Last gets the last record ordered by primary key desc.
//
//	output: must be a pointer to a repository.Model implementation.
func (r *repositoryGorm) Last(ctx context.Context, output repository.Model, filters ...repository.Filter) error {
	if len(filters) == 0 {
		return repository.ErrNoFilter
	}
	q := r.startQuery(ctx)
	q = r.setQueryFilters(q, filters)
	q = q.Last(output)
	return q.Error
}

// Update updates all model entries that match the provided filters with the given data.
//
//		data: must be a map[string]interface{}
//	 filters: filter entries that should be updated.
func (r *repositoryGorm) Update(ctx context.Context, data interface{}, filters ...repository.Filter) error {
	q := r.startQuery(ctx)
	q = r.setQueryFilters(q, filters)
	q = q.Updates(data)
	return q.Error
}

// Delete removes all the model entries that match filters.
//
//	options: configuration options for the removal.
func (r *repositoryGorm) Delete(ctx context.Context, opts ...repository.Option) error {
	q := r.startQuery(ctx)
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
func (r *repositoryGorm) FirstOrCreate(ctx context.Context, entity repository.Model, filters ...repository.Filter) error {
	q := r.startQuery(ctx)
	q = r.setQueryFilters(q, filters)
	return q.FirstOrCreate(entity).Error
}

// Count counts all the model entries that match filters.
//
//	filters: selection criteria for entries that should be considered when counting entries.
func (r *repositoryGorm) Count(ctx context.Context, filters ...repository.Filter) (uint64, error) {
	var count int64
	q := r.startQuery(ctx)
	q = r.setQueryFilters(q, filters)
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return uint64(count), nil
}

// startQuery inits a sql query for this repository's model. Multiple filters are ANDd together.
func (r *repositoryGorm) startQuery(ctx context.Context) *gorm.DB {

	return r.DB.Model(r.Model()).WithContext(ctx)
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
