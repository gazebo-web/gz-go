package firestore

import (
	"cloud.google.com/go/firestore"
	"github.com/gazebo-web/gz-go/v7/repository"
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
