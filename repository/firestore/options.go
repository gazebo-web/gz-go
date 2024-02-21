package firestore

import (
	"errors"

	"cloud.google.com/go/firestore"
	"github.com/gazebo-web/gz-go/v10/pagination"
	"github.com/gazebo-web/gz-go/v10/repository"
)

// Option is a Firestore-specific repository.Option implementation.
// It is used to configure Firestore repository operations.
type Option func(q *firestore.Query)

func (o Option) IsOption() {}

// MaxResults defines the maximum number of results for an operation that can return multiple results.
// Passing this Option to a Repository operation overwrites any previous MaxResults options passed.
func MaxResults(n int) repository.Option {
	return Option(func(q *firestore.Query) {
		*q = q.Limit(n)
	})
}

// Offset defines a number of results to skip before starting to capture values to return.
// This Option will be ignored if the MaxResults Option is not present.
// Passing this Option to a Repository operation overwrites any previous Offset options passed.
func Offset(offset int) repository.Option {
	return Option(func(q *firestore.Query) {
		*q = q.Offset(offset)
	})
}

// OrderByField contains order by information for a field
type OrderByField struct {
	Field     string
	Direction firestore.Direction
}

// Ascending sorts the passed field in ascending order.
func Ascending(field string) OrderByField {
	return OrderByField{
		Field:     field,
		Direction: firestore.Asc,
	}
}

// Descending sorts the passed field in descending order.
func Descending(field string) OrderByField {
	return OrderByField{
		Field:     field,
		Direction: firestore.Desc,
	}
}

// OrderBy sorts results based on fields.
// Use the Ascending and Descending functions to pass orders to this Option.
// In situations with multiple orders, they are applied in sequence.
// Multiple OrderBy options can be passed to a single Repository operation. They are appended to any previous orders.
func OrderBy(orders ...OrderByField) repository.Option {
	return Option(func(q *firestore.Query) {
		for _, order := range orders {
			*q = q.OrderBy(order.Field, order.Direction)
		}
	})
}

// Where filters results based on passed conditions.
// Multiple Where options can be passed to a single Repository operation. They are logically ANDed together.
// The op argument must be one of "==", "!=", "<", "<=", ">", ">=",
// "array-contains", "array-contains-any", "in" or "not-in".
func Where(field string, op string, value interface{}) repository.Option {
	return Option(func(q *firestore.Query) {
		*q = q.Where(field, op, value)
	})
}

// StartAfter initializes a new option that specifies that results should start right after
// the document with the given field values.
//
// StartAfter should be called with one field value for each OrderBy clause,
// in the order that they appear. For example, in
//
//	Repository.Find(&list, OrderBy(Descending("Value"), Ascending("Name")), StartAfter(2, "Test"))
//
// list will begin at the first document where Value = <2> + 1.
//
// Calling StartAfter overrides a previous call to StartAfter.
func StartAfter(fieldValues ...any) repository.Option {
	return Option(func(q *firestore.Query) {
		*q = q.StartAfter(fieldValues...)
	})
}

// StartAt initializes a new option that specifies that results should start at
// document with the given field values.
//
// StartAt should be called with one field value for each OrderBy clause,
// in the order that they appear. For example, in
//
//	Repository.Find(&list, OrderBy(Descending("Value"), Ascending("Name")), StartAt(1, "Test"))
//
// list will begin at the first document where Value = <1> + 1.
//
// Calling StartAt overrides a previous call to StartAt.
func StartAt(fieldValues ...any) repository.Option {
	return Option(func(q *firestore.Query) {
		*q = q.StartAt(fieldValues...)
	})
}

// NoOp is an option that performs no action.
// It is defined so that options functions can validate inputs and choose to do nothing because of some internal logic.
func NoOp() repository.Option {
	return Option(func(q *firestore.Query) {})
}

// In generates a new option that allows selecting all the elements where the given field contains any of the given
// values.
//
// Multiple In options can be passed to a single Repository operation. They are logically ANDed together.
//
//	Repository.Find(&list, In[string]("Name", []string{"Andrew", "John"]))
func In[T any](field string, values []T) repository.Option {
	if len(values) == 0 {
		return NoOp()
	}
	return Where(field, "in", values)
}

// setMaxResults establishes the max number of items that should be returned from a firestore query
// based on pagination configuration.
//
// In order to determine whether there are additional pages of results available, this function requests one
// extra element than the maximum page size. If this element exists, then there is an additional page available, if not,
// then this is the last page.
//
// The last element should be discarded before the list is returned to the user.
// See pagination.GetListAndCursor for more information.
func setMaxResults(opts []repository.Option, sg pagination.PageSizeGetter) ([]repository.Option, error) {
	p := pagination.PageSize(sg)
	if p == pagination.InvalidValue {
		return nil, errors.New("invalid page size")
	}
	return append(opts, MaxResults(int(p)+1)), nil
}

// SetCurrentPage generates a set of repository.Option to retrieve results for a specific page.
func SetCurrentPage(p pagination.Pagination) ([]repository.Option, error) {
	var opts []repository.Option
	opts = append(opts, OrderBy(Ascending("updated_at")))
	opts, err := setMaxResults(opts, p)
	if err != nil {
		return nil, err
	}
	if p == nil || len(p.GetPageToken()) == 0 {
		return opts, nil
	}
	updatedAt, err := pagination.ParsePageTokenToTime(p.GetPageToken())
	if err != nil {
		return nil, err
	}
	opts = append(opts, StartAt(updatedAt))
	return opts, nil
}
