package pagination

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestGenerateLinkHeader(t *testing.T) {
	u, err := url.Parse("https://gazebosim.org")
	require.NoError(t, err)

	assert.Equal(t, "<https://gazebosim.org>; rel=\"test\"", generateLinkHeader(u, "test"))
}

func TestReadRequest_NoQueryParams(t *testing.T) {
	u, err := url.Parse("https://gazebosim.org")
	require.NoError(t, err)

	req := ReadRequest(&http.Request{URL: u})
	assert.Equal(t, uint64(1), req.Page, "Must default to 1")
	assert.Equal(t, uint64(30), req.PageSize, "Must default to 30")
}

func TestReadRequest_WithPage(t *testing.T) {
	u, err := url.Parse("https://gazebosim.org/?page=3")
	require.NoError(t, err)

	req := ReadRequest(&http.Request{URL: u})
	assert.Equal(t, uint64(3), req.Page)
	assert.Equal(t, uint64(30), req.PageSize, "Must default to 30")
}

func TestReadRequest_WithPageSize(t *testing.T) {
	u, err := url.Parse("https://gazebosim.org/?page_size=50")
	require.NoError(t, err)

	req := ReadRequest(&http.Request{URL: u})
	assert.Equal(t, uint64(1), req.Page)
	assert.Equal(t, uint64(50), req.PageSize)
}

func TestReadRequest_WithPageAndPageSize(t *testing.T) {
	u, err := url.Parse("https://gazebosim.org/?page=3&page_size=50")
	require.NoError(t, err)

	req := ReadRequest(&http.Request{URL: u})
	assert.Equal(t, uint64(3), req.Page)
	assert.Equal(t, uint64(50), req.PageSize)
}

func TestPagingResponse_ToPagingLinks(t *testing.T) {
	u, err := url.Parse("https://gazebosim.org/")
	require.NoError(t, err)

	res := PagingResponse{
		BaseURL:    u,
		Page:       2,
		PageSize:   10,
		Count:      10,
		TotalCount: 50,
	}

	links := res.ToPagingLinks()

	assert.NotNil(t, links.First)
	assert.Equal(t, "https://gazebosim.org/?page=1&page_size=10", links.First.String())

	assert.NotNil(t, links.Previous)
	assert.Equal(t, "https://gazebosim.org/?page=1&page_size=10", links.Previous.String())

	assert.NotNil(t, links.Next)
	assert.Equal(t, "https://gazebosim.org/?page=3&page_size=10", links.Next.String())

	assert.NotNil(t, links.Last)
	assert.Equal(t, "https://gazebosim.org/?page=5&page_size=10", links.Last.String())
}

func TestPagingResponse_ToPagingLinksMissingPrevious(t *testing.T) {
	u, err := url.Parse("https://gazebosim.org/")
	require.NoError(t, err)

	res := PagingResponse{
		BaseURL:    u,
		Page:       1,
		PageSize:   10,
		Count:      10,
		TotalCount: 50,
	}

	links := res.ToPagingLinks()

	assert.Nil(t, links.Previous, "If the current page is the first page, we should not have a previous page")
}

func TestPagingResponse_ToPagingLinksMissingNext(t *testing.T) {
	u, err := url.Parse("https://gazebosim.org/")
	require.NoError(t, err)

	res := PagingResponse{
		BaseURL:    u,
		Page:       5,
		PageSize:   10,
		Count:      10,
		TotalCount: 50,
	}

	links := res.ToPagingLinks()

	assert.Nil(t, links.Next, "If the current page is the last page, we should not have a next page")
}

func TestPagingLinks_ToLinkHeaders(t *testing.T) {
	u, err := url.Parse("https://gazebosim.org/")
	require.NoError(t, err)

	res := PagingResponse{
		BaseURL:    u,
		Page:       2,
		PageSize:   10,
		Count:      10,
		TotalCount: 50,
	}

	links := res.ToPagingLinks()

	values := links.ToLinkHeaders()

	assert.Equal(t, "<https://gazebosim.org/?page=1&page_size=10>; rel=\"first\"", values[0])
	assert.Equal(t, "<https://gazebosim.org/?page=1&page_size=10>; rel=\"prev\"", values[1])
	assert.Equal(t, "<https://gazebosim.org/?page=3&page_size=10>; rel=\"next\"", values[2])
	assert.Equal(t, "<https://gazebosim.org/?page=5&page_size=10>; rel=\"last\"", values[3])
}

func TestPaginationResponseHTTP(t *testing.T) {
	rr := httptest.NewRecorder()

	u, err := url.Parse("https://gazebosim.org/")
	require.NoError(t, err)
	res := PagingResponse{
		BaseURL:    u,
		Page:       2,
		PageSize:   10,
		Count:      10,
		TotalCount: 50,
	}

	assert.NoError(t, WriteResponse(rr, res))

	count, err := strconv.Atoi(rr.Header().Get("X-Total-Count"))
	require.NoError(t, err)
	assert.Equal(t, res.TotalCount, uint64(count))
	assert.Equal(t, strings.Join(res.ToPagingLinks().ToLinkHeaders(), ", "), rr.Header().Get("Link"))
}

func TestLastPage(t *testing.T) {
	u, err := url.Parse("https://gazebosim.org/")
	require.NoError(t, err)
	res := PagingResponse{
		BaseURL:    u,
		Page:       2,
		PageSize:   10,
		Count:      10,
		TotalCount: 50,
	}

	assert.Equal(t, uint64(5), res.getLastPage())

	res.TotalCount = 20
	assert.Equal(t, uint64(2), res.getLastPage())

	res.TotalCount = 100
	assert.Equal(t, uint64(10), res.getLastPage())

	res.TotalCount = 105
	assert.Equal(t, uint64(11), res.getLastPage())

	res.TotalCount = 30
	assert.Equal(t, uint64(3), res.getLastPage())

}
