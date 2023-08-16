package structs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToMap(t *testing.T) {
	type structTest struct {
		TestString string `structs:"s"`
		TestInt    int    `structs:"i"`
		TestStruct struct {
			DeepValue int `structs:"ss_i"`
		} `structs:"ss"`
		TestIgnored string `structs:"-"`
	}

	s := structTest{
		TestString: "test",
		TestInt:    1,
		TestStruct: struct {
			DeepValue int `structs:"ss_i"`
		}{
			DeepValue: 2,
		},
		TestIgnored: "ignored",
	}

	// Not struct
	_, err := ToMap("test")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidInputType)

	// Correct struct tag
	result, err := ToMap(s)
	var expected map[string]any
	assert.IsType(t, expected, result)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Equal(t, s.TestString, result["s"])
	assert.Equal(t, s.TestInt, result["i"])
	assert.NotContains(t, result, "TestIgnored")
	assert.NotZero(t, result["ss"])

	casted, ok := result["ss"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, s.TestStruct.DeepValue, casted["ss_i"])

	// Struct pointer
	result, err = ToMap(&s)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}
