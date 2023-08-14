package structs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type structTest struct {
	Test string `field:"test"`
}

func TestMap(t *testing.T) {
	s := structTest{
		Test: "value",
	}

	// Not struct
	_, err := Map("test", "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInputType)

	// Empty tag
	_, err = Map(s, "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTagEmpty)

	// Wrong struct tag returns an empty map
	wrong, err := Map(s, "invalid")
	assert.NoError(t, err)
	assert.Empty(t, wrong)

	// Correct struct tag
	result, err := Map(s, "field")
	var expected map[string]any
	assert.IsType(t, expected, result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, s.Test, result["test"])
}
