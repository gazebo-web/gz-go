package pagination

import (
	"encoding"
	"encoding/base64"
	"time"
)

// Pagination performs result pagination. 
// It is intended to be implemented by client request types.
type Pagination interface {
	PageSizeGetter
	PageTokenGetter
}

// PageSizeGetter provides the amount of results to return when paginating results.
type PageSizeGetter interface {
	// GetPageSize returns the desired page size. It's using int32 in order to match the method signature from the
	// generated Go stubs that also return an int32 value.
	GetPageSize() int32
}

// PageTokenGetter provides a page token used to paginate results.
type PageTokenGetter interface {
	// GetPageToken returns the requested page token. 
	// If the user requests a specific page, this value must not be empty.
	// For cursor-based pagination, this value must be in base64.
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
func PageSize(req PageSizeGetter, opts ...PageSizeOptions) int32 {
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
// based on a `time.Time` type field.
func GetNextPageTokenFromTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return NewPageToken(t)
}

// GetListAndCursor takes a slice of elements and returns a page of results in the form of a slice and a page cursor.
// 
// The page of results is the result of the user query and is to be returned to the user as is.
// 
// The page cursor is used to determine whether there are additional result pages available. 
// The cursor is an element that is part of the collection, but not part of the current page. 
// 
// If the cursor:
// 
// * Is a non-zero value, then there are additional pages available. 
// * Is a zero value, then the returned page is the last page.
// 
// The returned cursor is typically processed by `GetNextPageToken[...]()` type functions to generate a page token.
//
// See firestore.setMaxResults to understand why this function is being used.
func GetListAndCursor[T any](raw []T, sg PageSizeGetter) ([]T, T) {
	if len(raw) > 0 && len(raw) > int(PageSize(sg)) {
		last := raw[len(raw)-1]
		return raw[:len(raw)-1], last
	}
	var zero T
	return raw, zero
}
