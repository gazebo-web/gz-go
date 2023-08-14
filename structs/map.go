package structs

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrTagEmpty is returned when an empty tag is passed to Map.
	ErrTagEmpty = errors.New("tag cannot be empty")
	// ErrInvalidInputType is returned when Map receives an input that is not a struct or a pointer to a struct.
	ErrInvalidInputType = errors.New("invalid input type")
)

// Map converts the given struct to a map[string]any. It uses the given tag to identify the key for each map field.
// If no tag is present in the Struct field, the field is omitted.
//
// Considering the following struct:
//
//	type Test struct {
//		Field1 string `test:"field_1"`
//		Field2 int `test:"field_2"`
//		Field3 any `test:"field_3"`
//	}
//
// And the following variable:
//
//	input := Test{
//		Field1: "test",
//		Field2: 1,
//		Field3: nil,
//	}
//
// Map(input, "test") will return a map as follows:
//
//	map[string]any{
//		"field_1": "test",
//		"field_2": 1,
//		"field_3": nil,
//	}
//
// If Map is called with a non-existent tag, an empty map will be returned.
//
//	m := Map(input, "not_found")
//	len(m) == 0 (true)
func Map(s any, tag string) (map[string]any, error) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w, got %T", ErrInvalidInputType, v)
	}
	if len(tag) == 0 {
		return nil, ErrTagEmpty
	}
	t := v.Type()
	out := make(map[string]any)
	for i := 0; i < v.NumField(); i++ {
		fi := t.Field(i)
		if tagValue := fi.Tag.Get(tag); tagValue != "" {
			out[tagValue] = v.Field(i).Interface()
		}
	}
	return out, nil
}
