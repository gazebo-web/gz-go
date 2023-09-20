package gorm

import (
	"database/sql/driver"
	"errors"
	"strings"
)

// ArrayString is a custom data type for SQL databases. It adds support for an array of strings to MySQL.
type ArrayString []string

// Value takes the current ArrayString and converts it to a valid SQL value.
//
// See driver.Valuer for more details.
func (arr *ArrayString) Value() (driver.Value, error) {
	return strings.Join(*arr, ","), nil
}

// Scan fills out the current ArrayString with the string provided as source. It returns an error if source if not a
// string.
//
// See sql.Scanner for more details.
func (arr *ArrayString) Scan(src any) error {
	v, ok := src.(string)
	if !ok {
		return errors.New("invalid data type, must be a string")
	}
	*arr = strings.Split(v, ",")
	return nil
}
