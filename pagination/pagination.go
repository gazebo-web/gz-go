package pagination

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	page, err := strconv.ParseUint(r.URL.Query().Get("page"), 10, 64)
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.ParseUint(r.URL.Query().Get("page_size"), 10, 64)
	if err != nil {
		pageSize = 30
	}

	return PagingRequest{
		Page:     page,
		PageSize: pageSize,
	}
}

// PagingResponse contains the values that should be added to an HTTP response.
type PagingResponse struct {
	// BaseURL contains the base url that should be used to redirect users when accessing .
	BaseURL *url.URL
	// Page contains the page that was requested.
	Page uint
	// PageSize contains the amount of elements that were requested for this Page.
	PageSize uint
	// Count contains the amount of elements that were retrieved for this Page.
	Count uint
	// TotalCount contains the elements available, usually used to calculate the amount of pages available.
	TotalCount uint
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
func (r PagingResponse) getPreviousPage() uint {
	return r.Page - 1
}

// getNextPage returns the next page index.
func (r PagingResponse) getNextPage() uint {
	return r.Page + 1
}

// getFirstPage returns the first page index.
func (r PagingResponse) getFirstPage() uint {
	return 1
}

// getLastPage returns the last page index.
func (r PagingResponse) getLastPage() uint {
	lastPage := r.TotalCount / r.PageSize
	if r.TotalCount%r.PageSize > 0 {
		lastPage++
	}
	return lastPage
}

// generateURL adds pagination query parameters to an URL.
func generateURL(u url.URL, page uint, size uint) *url.URL {
	q := u.Query()

	q.Set("page", strconv.Itoa(int(page)))
	q.Set("page_size", strconv.Itoa(int(size)))

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
		h = append(h, generateLinkHeader(p.First, "first"))
	}
	if p.Previous != nil {
		h = append(h, generateLinkHeader(p.Previous, "prev"))
	}
	if p.Next != nil {
		h = append(h, generateLinkHeader(p.Next, "next"))
	}
	if p.Last != nil {
		h = append(h, generateLinkHeader(p.Last, "last"))
	}
	return h
}

// generateLinkHeader generates a valid "Link" HTTP header value by using the given URL and rel values.
func generateLinkHeader(u *url.URL, rel string) string {
	return fmt.Sprintf("<%s>; rel=\"%s\"", u.String(), rel)
}

// WriteResponse writes to http.ResponseWriter the pagination values determined by PagingResponse.
func WriteResponse(w http.ResponseWriter, res PagingResponse) error {
	w.Header().Set("X-Total-Count", strconv.Itoa(int(res.TotalCount)))
	headers := res.ToPagingLinks().ToLinkHeaders()
	for _, h := range headers {
		w.Header().Add("Link", h)
	}
	return nil
}
