package structs

import (
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"reflect"
)

var (
	// ErrTagEmpty is returned when an empty tag is passed to ToMap.
	ErrTagEmpty = errors.New("tag cannot be empty")
	// ErrInvalidInputType is returned when ToMap receives an input that is not a struct or a pointer to a struct.
	ErrInvalidInputType = errors.New("invalid input type")
)

// ToMap converts the given struct to a map[string]any. It uses the "structs" tag to identify the key for each map field.
// If no tag is present in the Struct field, the field is omitted.
//
// The default key string is the struct field name but can be changed in the struct field's tag value.
// The "structs" key in the struct's field tag value is the key name. Example:
//
//	// Field appears in map as key "myName".
//	Name string `structs:"myName"`
//
// A tag value with the content of "-" ignores that particular field. Example:
//
//	// Field is ignored by this package.
//	Field bool `structs:"-"`
//
// A tag value with the content of "string" uses the stringer to get the value. Example:
//
//	// The value will be output of Animal's String() func.
//	// Map will panic if Animal does not implement String().
//	Field *Animal `structs:"field,string"`
//
// A tag value with the option of "flatten" used in a struct field is to flatten its fields in the output map. Example:
//
//	// The FieldStruct's fields will be flattened into the output map.
//	FieldStruct time.Time `structs:",flatten"`
//
// A tag value with the option of "omitnested" stops iterating further if the type
// is a struct. Example:
//
//	// Field is not processed further by this package.
//	Field time.Time     `structs:"myName,omitnested"`
//	Field *http.Request `structs:",omitnested"`
//
// A tag value with the option of "omitempty" ignores that particular field if
// the field value is empty. Example:
//
//	// Field appears in map as key "myName", but the field is
//	// skipped if empty.
//	Field string `structs:"myName,omitempty"`
//
//	// Field appears in map as key "Field" (the default), but
//	// the field is skipped if empty.
//	Field string `structs:",omitempty"`
func ToMap(s any) (map[string]any, error) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w, got %T", ErrInvalidInputType, v)
	}
	return structs.Map(s), nil
}
