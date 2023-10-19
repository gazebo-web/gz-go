package validate

import "reflect"

// IsZero checks returns true if the provided value is a zero value.
func IsZero(value any) bool {
	switch v := value.(type) {
	case string:
		return v == ""
	case int:
		return v == 0
	case int8:
		return v == int8(0)
	case int16:
		return v == int16(0)
	case int32:
		return v == int32(0)
	case int64:
		return v == int64(0)
	case uint:
		return v == uint(0)
	case uint8:
		return v == uint8(0)
	case uint16:
		return v == uint16(0)
	case uint32:
		return v == uint32(0)
	case uint64:
		return v == uint64(0)
	case float32:
		return v == float32(0)
	case float64:
		return v == float64(0)
	case complex64:
		return v == complex64(0)
	case complex128:
		return v == complex128(0)
	case bool:
		return !v
	default:
		return reflect.ValueOf(value).IsZero()
	}
}

// HasZero checks returns true if any of the provided values is a zero value.
func HasZero(values ...any) bool {
	for _, value := range values {
		if IsZero(value) {
			return true
		}
	}
	return false
}
