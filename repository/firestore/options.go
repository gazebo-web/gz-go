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
