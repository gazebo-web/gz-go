package encoders

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestSingletons(t *testing.T) {
	require.NotNil(t, JSON)
	require.NotNil(t, ProtoText)
}

type TestValue interface {
	GetData() string
}

func testUnmarshal(t *testing.T, marshaller Marshaller, filepath string, v TestValue) {
	b, err := os.ReadFile(filepath)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	err = marshaller.Unmarshal(b, v)
	assert.NoError(t, err)
	assert.Equal(t, "test", v.GetData())
}

func testMarshal(t *testing.T, marshaller Marshaller, v any) {
	b, err := marshaller.Marshal(v)
	assert.NoError(t, err)
	assert.NotEmpty(t, b)
	assert.Contains(t, string(b), "test")
}
