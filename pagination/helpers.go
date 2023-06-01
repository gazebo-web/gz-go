package pagination

import (
	"encoding"
	"encoding/base64"
	"time"
)

// PageSizeGetter holds a method to return the amount of pages requested by users when listing items in an API call.
type PageSizeGetter interface {
	// GetPageSize returns the desired page size.
	GetPageSize() int32
}

// PageSize returns a valid page size following the AIP-158 for Pagination.
// Reference: https://google.aip.dev/158
//
//	Default value: 50
//	Max value: 1000
//	Min value: 0.
//
//	If no value is passed, it returns 50.
//	If a value greater than 1000 is specified, it caps the result value to 1000.
//	If a negative value is specified, it returns -1.
func PageSize(req PageSizeGetter) int32 {
	if req == nil {
		return defaultPageSize
	}
	if req.GetPageSize() == 0 {
		return defaultPageSize
	}

	if req.GetPageSize() > maxPageSize {
		return maxPageSize
	}
	if req.GetPageSize() < 0 {
		return -1
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
