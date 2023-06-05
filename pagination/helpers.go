package pagination

import (
	"encoding"
	"encoding/base64"
	"errors"
	"github.com/gazebo-web/gz-go/v7/repository"
	"github.com/gazebo-web/gz-go/v7/repository/firestore"
	"time"
)

// Pagination determines what are the methods needed in a request object to perform pagination.
type Pagination interface {
	pageSizeGetter
	pageTokenGetter
}

// pageSizeGetter holds a method to return the amount of pages requested by users when listing items in an API call.
type pageSizeGetter interface {
	// GetPageSize returns the desired page size. It's using int32 in order to match the method signature from the
	// generated Go stubs that also return an int32 value.
	GetPageSize() int32
}

// pageTokenGetter holds a method to the page requested by a user.
type pageTokenGetter interface {
	// GetPageToken returns the requested page token. If the user requests a specific page, this value is not empty.
	GetPageToken() string
}

// PageSizeOptions allows developers to pass new MaxSize and DefaultSize values to the PageSize function.
// The options described in this struct override the default PageSize values provided in this library.
type PageSizeOptions struct {
	// MaxSize is the maximum number of elements that can be returned by the API.
	MaxSize int32
	// DefaultSize is the number of elements that the page list will default to.
	DefaultSize int32
}

// PageSize returns a valid page size following the AIP-158 proposal for Pagination.
// Reference: https://google.aip.dev/158
//
//	Default value: 50
//	Default max value: 100
//
//	If no value is passed, it returns the default value.
//	If a value greater than the max page size is specified, it caps the result value to the max page size.
//	If a negative value is specified, it returns -1.
func PageSize(req pageSizeGetter, opts ...PageSizeOptions) int32 {
	if req == nil {
		return defaultPageSize
	}
	defaultSize := defaultPageSize
	maxSize := maxPageSize

	if len(opts) > 0 {
		if v := opts[len(opts)-1].DefaultSize; v > 0 {
			defaultSize = v
		}
		if v := opts[len(opts)-1].MaxSize; v > 0 {
			maxSize = v
		}
	}

	if req.GetPageSize() == 0 {
		return defaultSize
	}
	if req.GetPageSize() > maxSize {
		return maxSize
	}
	if req.GetPageSize() < 0 {
		return InvalidValue
	}
	return req.GetPageSize()
}

// NewPageToken generates a page token in base64.
// It returns an empty string if the input is nil or the conversion to string fails.
func NewPageToken(input encoding.TextMarshaler) string {
	if input == nil {
		return ""
	}
	src, err := input.MarshalText()
	if err != nil {
		return ""
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(dst, src)
	return string(dst)
}

// ParsePageTokenToTime converts the given token and returns a valid time.Time.
// The token provided is usually the value used in cursor-based pagination.
func ParsePageTokenToTime(token string) (time.Time, error) {
	value, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, err
	}
	result, err := time.Parse(time.RFC3339, string(value))
	if err != nil {
		return time.Time{}, err
	}
	return result, err
}

// GetNextPageTokenFromTime generates a page token. This function should be used when generating a next page token
// in a List operation on a time.Time field.
func GetNextPageTokenFromTime(updatedAt time.Time) string {
	if updatedAt.IsZero() {
		return ""
	}
	return NewPageToken(updatedAt)
}

// SetCurrentPage generates a set of repository.Option to retrieve results for a specific page.
func SetCurrentPage(opts []repository.Option, p Pagination) ([]repository.Option, error) {
	opts = append(opts, firestore.OrderBy(firestore.Ascending("updated_at")))
	opts, err := setMaxResults(opts, p)
	if err != nil {
		return nil, err
	}
	if p == nil || len(p.GetPageToken()) == 0 {
		return opts, nil
	}
	updatedAt, err := ParsePageTokenToTime(p.GetPageToken())
	if err != nil {
		return nil, err
	}
	opts = append(opts, firestore.StartAt(updatedAt))
	return opts, nil
}

// setMaxResults establishes the max number of items that should be returned from firestore.
// This function includes an extra element in the MaxResults option, this last element is used for pagination.
// The last element should be discarded before the list is returned to the user.
// See GetListAndCursor for more information.
func setMaxResults(opts []repository.Option, sg pageSizeGetter) ([]repository.Option, error) {
	p := PageSize(sg)
	if p == InvalidValue {
		return nil, errors.New("invalid page size")
	}
	return append(opts, firestore.MaxResults(int(p)+1)), nil
}

// GetListAndCursor is used for pagination. This function generates a list of elements that should be returned to the
// user, and if there's a next page available, it returns the last element as a separate return value.
//
// GetListAndCursor extracts the actual list that was requested by the user, and the last element from the Find output.
// This way the List operation checks if there's a next page available.
// If there's not, a zero value is returned, making the getNextPageToken return an empty string.
//
// See setMaxResults to understand why this function is being used.
// This function is usually used alongside GetNextPageToken.
func GetListAndCursor[T any](raw []T, sg pageSizeGetter) ([]T, T) {
	if len(raw) > 0 && len(raw) > int(PageSize(sg)) {
		last := raw[len(raw)-1]
		return raw[:len(raw)-1], last
	}
	var zero T
	return raw, zero
}
