package gorm

import (
	"database/sql"
	"database/sql/driver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestArrayString_Scanner(t *testing.T) {
	var arrstr ArrayString
	var scanner sql.Scanner

	scanner = &arrstr

	assert.Error(t, scanner.Scan([]byte("test")), "Calling Scan must return an error with an invalid type")
	assert.NoError(t, scanner.Scan("test,test"), "Calling Scan with a string must return no errors")

	assert.Len(t, arrstr, 2, "After calling Scan with \"test,test\", the underlying ArrayString must have n elements")
}

func TestArrayString_Valuer(t *testing.T) {
	var arrstr ArrayString
	var valuer driver.Valuer

	arrstr = []string{"test", "test"}

	valuer = &arrstr

	v, err := valuer.Value()
	assert.NoError(t, err)

	str, ok := v.(string)
	assert.True(t, ok)
	assert.Equal(t, "test,test", str)

}
