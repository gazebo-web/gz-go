package repository

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
)

// SQLOption is a SQL-specific Repository Option implementation.
// It is used to configure SQL repository operations.
// TODO This should be extracted into its own package to avoid colliding with other implementations
type SQLOption func(r *gorm.DB)

func (l SQLOption) IsOption() {}

// Fields defines fields to return.
// Passing this Option to a Repository operation overwrites any previous GroupBy options passed.
func Fields(fields ...string) Option {
	return SQLOption(func(q *gorm.DB) {
		*q = *q.Select(strings.Join(fields, ","))
	})
}

// Where filters results based on passed conditions.
// Multiple Filter options can be passed to a single Repository operation. They are logically ANDed together.
func Where(template string, values ...interface{}) Option {
	return SQLOption(func(q *gorm.DB) {
		*q = *q.Where(template, values...)
	})
}

// MaxResults defines the maximum number of results for an operation that can return multiple results.
// Passing this Option to a Repository operation overwrites any previous MaxResults options passed.
func MaxResults(n int) Option {
	return SQLOption(func(q *gorm.DB) {
		*q = *q.Limit(n)
	})
}

// Offset defines a number of results to skip before starting to capture values to return.
// This Option will be ignored if the MaxResults Option is not present.
// Passing this Option to a Repository operation overwrites any previous Offset options passed.
func Offset(offset int) Option {
	return SQLOption(func(q *gorm.DB) {
		*q = *q.Offset(offset)
	})
}

// GroupBy groups results based field values.
// Passing this Option to a Repository operation overwrites any previous GroupBy options passed.
func GroupBy(fields ...string) Option {
	return SQLOption(func(q *gorm.DB) {
		*q = *q.Group(strings.Join(fields, ","))
	})
}

// OrderByField contains order by information for a field
type OrderByField string

// Ascending sorts the passed field in ascending order.
func Ascending(field string) OrderByField {
	return OrderByField(fmt.Sprintf("%s ASC", field))
}

// Descending sorts the passed field in descending order.
func Descending(field string) OrderByField {
	return OrderByField(fmt.Sprintf("%s DESC", field))
}

// OrderBy sorts results based on fields.
// Use the Ascending and Descending functions to pass orders to this Option.
// In situations with multiple orders, they are applied in sequence.
// Multiple OrderBy options can be passed to a single Repository operation. They are appended to any previous orders.
func OrderBy(orders ...OrderByField) Option {
	return SQLOption(func(q *gorm.DB) {
		for _, order := range orders {
			*q = *q.Order(string(order))
		}
	})
}

// Preload allows loading a single related model to have it be included in the returned value.
// The field parameter must match the model field name exactly (case-sensitive).
// An optional filter composed of a template and any number of values can be passed to filter preloaded results.
// Multiple Preload options can be passed to a single Repository operation. They are appended to any previous preloads.
func Preload(field string, filter ...interface{}) Option {
	return SQLOption(func(q *gorm.DB) {
		*q = *q.Preload(field, filter...)
	})
}
