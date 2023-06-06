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
	assert.Equal(t, defaultPageSize, ps)

	// If the user does not specify page_size (or specifies 0), the API chooses an appropriate default (50),
	// which the API should document. The API must not return an error.
	ps = PageSize(paginationTestValue(0))
	assert.Equal(t, defaultPageSize, ps)

	// If the user specifies page_size greater than the maximum permitted by the API (100), the API should coerce down
	// to the maximum permitted page size.
	ps = PageSize(paginationTestValue(101))
	assert.Equal(t, maxPageSize, ps)

	// If the user specifies a negative value for page_size, the API must send an INVALID_ARGUMENT error.
	// The API may return fewer results than the number requested (including zero results), even if not at the end of
	// the collection.
	ps = PageSize(paginationTestValue(-500))
	assert.Equal(t, InvalidValue, ps)

	// In all other cases, it should return the input value.
	ps = PageSize(paginationTestValue(25))
	assert.Equal(t, int32(25), ps)
}

func TestPageSize_Options(t *testing.T) {
	ps := PageSize(paginationTestValue(1001), PageSizeOptions{
		MaxSize: 1001,
	})
	assert.Equal(t, int32(1001), ps)

	ps = PageSize(paginationTestValue(0), PageSizeOptions{
		DefaultSize: 10,
	})
	assert.Equal(t, int32(10), ps)
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

func TestParsePageTokenToTime(t *testing.T) {
	// No token provided
	_, err := ParsePageTokenToTime("")
	assert.Error(t, err)

	// Not a date
	_, err = ParsePageTokenToTime("12345")
	assert.Error(t, err)

	// A valid time.Time should be correctly parsed into a time.Time when using ParsePageTokenToTime.
	updatedAt := time.Now()
	token := NewPageToken(updatedAt)

	result, err := ParsePageTokenToTime(token)
	require.NoError(t, err)
	assert.True(t, result.Equal(updatedAt))
}

func TestGetNextPageToken(t *testing.T) {
	var zero time.Time
	assert.Empty(t, GetNextPageTokenFromTime(zero))

	now := time.Now()
	assert.NotEmpty(t, GetNextPageTokenFromTime(now))
}

func TestGetListAndCursor(t *testing.T) {
	// If page size is 2, but the raw list contains 3 elements, it means that there is a next page.
	raw := generateMockData()
	list, last := GetListAndCursor(raw, &TestPaginationRequest{
		size:  2,
		token: "",
	})
	assert.Len(t, list, 2)
	assert.NotZero(t, last)

	// If the page size is 3, and the raw list  contains 3 elements, it means that there are no more results to show.
	list, last = GetListAndCursor(raw, &TestPaginationRequest{
		size:  3,
		token: "",
	})
	assert.Len(t, list, 3)
	assert.Zero(t, last)
}

type TestPaginationRequest struct {
	size  int32
	token string
}

func (req *TestPaginationRequest) GetPageSize() int32 {
	return req.size
}

func (req *TestPaginationRequest) GetPageToken() string {
	return req.token
}

type testData int

func generateMockData() []testData {
	data := make([]testData, 3)
	for i := 0; i < 3; i++ {
		data[i] = testData(i)
	}
	return data
}
