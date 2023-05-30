package pagination

import (
	"encoding/base64"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type paginationTestValue int32

func (p paginationTestValue) GetPageSize() int32 {
	return int32(p)
}

func TestPageSize(t *testing.T) {
	// The page_size field must not be required.
	ps := PageSize(nil)
	assert.Equal(t, int32(50), ps)

	// If the user does not specify page_size (or specifies 0), the API chooses an appropriate default (50),
	// which the API should document. The API must not return an error.
	ps = PageSize(paginationTestValue(0))
	assert.Equal(t, int32(50), ps)

	// If the user specifies page_size greater than the maximum permitted by the API (1000), the API should coerce down
	// to the maximum permitted page size.
	ps = PageSize(paginationTestValue(1001))
	assert.Equal(t, int32(1000), ps)

	// If the user specifies a negative value for page_size, the API must send an INVALID_ARGUMENT error.
	// The API may return fewer results than the number requested (including zero results), even if not at the end of
	// the collection.
	ps = PageSize(paginationTestValue(-500))
	assert.Equal(t, int32(-1), ps)

	// If the user specifies a page_size = 1, the API chooses an appropriate minimum default (10),
	// which the API should document. The API must not return an error.
	ps = PageSize(paginationTestValue(1))
	assert.Equal(t, int32(10), ps)

	// In all other cases, it should return the input value.
	ps = PageSize(paginationTestValue(100))
	assert.Equal(t, int32(100), ps)
}

type mockTextMarshaller struct {
	mock.Mock
}

func (m *mockTextMarshaller) MarshalText() (text []byte, err error) {
	args := m.Called()
	return []byte(args.String(0)), args.Error(1)
}

func TestGeneratePageToken(t *testing.T) {
	// Returns empty string when a nil argument is passed
	assert.Empty(t, NewPageToken(nil))

	// Returns an empty string when it fails to marshal the data type into a text.
	m := new(mockTextMarshaller)
	m.On("MarshalText").Return("", errors.New("test error"))
	assert.Empty(t, NewPageToken(m))

	// Once time.Time is marshalled, it should encode the value into a valid base64 string, and return a string value that
	// once converted back to time.Time, it should be equal to the original value.
	updatedAt := time.Now()
	token := NewPageToken(updatedAt)
	require.NotEmpty(t, token)

	value, err := base64.StdEncoding.DecodeString(token)
	require.NoError(t, err)

	result, err := time.Parse(time.RFC3339, string(value))
	require.NoError(t, err)
	assert.True(t, result.Equal(updatedAt))
}
