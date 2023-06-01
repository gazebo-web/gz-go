package pagination

import (
	"encoding"
	"encoding/base64"
)

// PageSizeGetter holds a method to return the amount of pages requested by users when listing items in an API call.
type PageSizeGetter interface {
	// GetPageSize returns the desired page size. It's using int32 in order to match the method signature from the
	// generated Go stubs that also return an int32 value.
	GetPageSize() int32
}

type PageSizeOptions struct {
	MaxSize     int32
	DefaultSize int32
}

// PageSize returns a valid page size following the AIP-158 proposal for Pagination.
// Reference: https://google.aip.dev/158
//
//	Default value: 50
//	Max value: 1000
//
//	If no value is passed, it returns the default value.
//	If a value greater than the max page size is specified, it caps the result value to the max page size.
//	If a negative value is specified, it returns -1.
func PageSize(req PageSizeGetter, opts ...PageSizeOptions) int32 {
	if req == nil {
		return defaultPageSize
	}
	defaultSize := int32(defaultPageSize)
	maxSize := int32(maxPageSize)

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
