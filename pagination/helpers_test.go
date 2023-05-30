package pagination

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
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

	// In all other cases, it should return the input value.
	ps = PageSize(paginationTestValue(100))
	assert.Equal(t, int32(100), ps)
}

func TestGeneratePageToken(t *testing.T) {
	assert.Empty(t, NewPageToken(nil))

	updatedAt := time.Now()
	token := NewPageToken(updatedAt)
	require.NotEmpty(t, token)

	value, err := base64.StdEncoding.DecodeString(token)
	require.NoError(t, err)

	result, err := time.Parse(time.RFC3339, string(value))
	require.NoError(t, err)
	assert.True(t, result.Equal(updatedAt))
}
