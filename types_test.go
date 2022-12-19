package gz

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInt(t *testing.T) {
	r := Int(1)
	assert.Equal(t, 1, *r)
}

func TestInt64(t *testing.T) {
	r := Int64(1)
	assert.Equal(t, int64(1), *r)
}

func TestFloat64(t *testing.T) {
	r := Float64(1.0)
	assert.Equal(t, 1.0, *r)
}

func TestString(t *testing.T) {
	r := String("test")
	assert.Equal(t, "test", *r)
}

func TestStringSlice(t *testing.T) {
	input := []string{"a", "b", "c"}
	slice := StringSlice(input)
	for i, s := range slice {
		assert.Equal(t, input[i], *s)
	}
}

func TestBool(t *testing.T) {
	r := Bool(true)
	assert.Equal(t, true, *r)
}
