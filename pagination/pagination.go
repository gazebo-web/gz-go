package pagination

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	headerTotalCount = "X-Total-Count"

	headerLink         = "Link"
	headerLinkFirst    = "first"
	headerLinkPrevious = "prev"
	headerLinkNext     = "next"
	headerLinkLast     = "last"

	queryParamPage     = "page"
	queryParamPageSize = "page_size"

	defaultPage     = 1
	defaultPageSize = 50
	maxPageSize     = 1000
)

// PagingRequest is used to describe pagination values from incoming requests.
type PagingRequest struct {
	// Page is the page that should be shown. When using ReadRequest, this value defaults to 1.
	Page uint64
	// PageSize contains the amount of elements that a Page will include. When using ReadRequest, this value defaults to 30.
	PageSize uint64
}

// ReadRequest reads the PagingRequest from a http.Request.
func ReadRequest(r *http.Request) PagingRequest {
	page, err := strconv.ParseUint(r.URL.Query().Get(queryParamPage), 10, 64)
	if err != nil {
		page = defaultPage
	}
	pageSize, err := strconv.ParseUint(r.URL.Query().Get(queryParamPageSize), 10, 64)
	if err != nil {
		pageSize = defaultPageSize
	}

	return PagingRequest{
		Page:     page,
		PageSize: pageSize,
	}
}

// PagingResponse contains the values that should be added to an HTTP response.
type PagingResponse struct {
	// BaseURL contains the base url that should be used to interact with the desired API list method that supports pagination.
	// If BaseURL contains query params for page and page_size, those values will be overwritten.
	BaseURL *url.URL
	// Page contains the page that was requested.
	Page uint64
	// PageSize contains the amount of elements that were requested for this Page.
	PageSize uint64
	// Count contains the amount of elements that were retrieved for this Page.
	Count uint64
	// TotalCount contains the elements available, usually used to calculate the amount of pages available.
	TotalCount uint64
}

// ToPagingLinks converts the current PagingResponse to a group of links described by PagingLinks.
func (r PagingResponse) ToPagingLinks() PagingLinks {
	if r.BaseURL == nil {
		return PagingLinks{}
	}

	var prev *url.URL
	if p := r.getPreviousPage(); p >= r.getFirstPage() {
		prev = generateURL(*r.BaseURL, p, r.PageSize)
	}

	var next *url.URL
	if p := r.getNextPage(); p <= r.getLastPage() {
		next = generateURL(*r.BaseURL, p, r.PageSize)
	}

	return PagingLinks{
		First:    generateURL(*r.BaseURL, r.getFirstPage(), r.PageSize),
		Previous: prev,
		Next:     next,
		Last:     generateURL(*r.BaseURL, r.getLastPage(), r.PageSize),
	}
}

// getPreviousPage returns the previous page index.
func (r PagingResponse) getPreviousPage() uint64 {
	return r.Page - 1
}

// getNextPage returns the next page index.
func (r PagingResponse) getNextPage() uint64 {
	return r.Page + 1
}

// getFirstPage returns the first page index.
func (r PagingResponse) getFirstPage() uint64 {
	return 1
}

// getLastPage returns the last page index.
func (r PagingResponse) getLastPage() uint64 {
	return uint64(math.Ceil(float64(r.TotalCount) / float64(r.PageSize)))
}

// generateURL adds pagination query parameters to an URL.
func generateURL(u url.URL, page uint64, size uint64) *url.URL {
	q := u.Query()

	q.Set(queryParamPage, strconv.Itoa(int(page)))
	q.Set(queryParamPageSize, strconv.Itoa(int(size)))

	u.RawQuery = q.Encode()
	return &u
}

// PagingLinks contains a set of URLs to help with pagination.
type PagingLinks struct {
	// First contains the URL where clients should query the first page of elements.
	First *url.URL
	// Previous contains the URL where clients should query the first page of elements.
	Previous *url.URL
	// Next contains the URL where clients should query the first page of elements.
	Next *url.URL
	// Last contains the URL where clients should query the first page of elements.
	Last *url.URL
}

// ToLinkHeaders converts the links in PagingLinks to an array of links to use in the "Link" HTTP header.
func (p PagingLinks) ToLinkHeaders() []string {
	h := make([]string, 0, 4)
	if p.First != nil {
		h = append(h, generateLinkHeader(p.First, headerLinkFirst))
	}
	if p.Previous != nil {
		h = append(h, generateLinkHeader(p.Previous, headerLinkPrevious))
	}
	if p.Next != nil {
		h = append(h, generateLinkHeader(p.Next, headerLinkNext))
	}
	if p.Last != nil {
		h = append(h, generateLinkHeader(p.Last, headerLinkLast))
	}
	return h
}

// generateLinkHeader generates a valid "Link" HTTP header value by using the given URL and rel values.
func generateLinkHeader(u *url.URL, rel string) string {
	return fmt.Sprintf("<%s>; rel=\"%s\"", u.String(), rel)
}

// WriteResponse writes to http.ResponseWriter the pagination values determined by PagingResponse.
func WriteResponse(w http.ResponseWriter, res PagingResponse) error {
	w.Header().Set(headerTotalCount, strconv.Itoa(int(res.TotalCount)))
	headers := res.ToPagingLinks().ToLinkHeaders()
	w.Header().Set(headerLink, strings.Join(headers, ", "))
	return nil
}
