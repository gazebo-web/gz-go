package validate

import "reflect"

// IsZero checks returns true if the provided value is a zero value.
func IsZero(value any) bool {
	switch value.(type) {
	case string:
		return value == ""
	case int:
		return value == 0
	case int8:
		return value == int8(0)
	case int16:
		return value == int16(0)
	case int32:
		return value == int32(0)
	case int64:
		return value == int64(0)
	case uint:
		return value == uint(0)
	case uint8:
		return value == uint8(0)
	case uint16:
		return value == uint16(0)
	case uint32:
		return value == uint32(0)
	case uint64:
		return value == uint64(0)
	case float32:
		return value == float32(0)
	case float64:
		return value == float64(0)
	case complex64:
		return value == complex64(0)
	case complex128:
		return value == complex128(0)
	case bool:
		return !value.(bool)
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
