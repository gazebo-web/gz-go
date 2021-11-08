package io

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// httpDialer is a Dialer implementation using HTTP as transport layer.
type httpDialer struct {
	// client is the HTTP client used to create requests and receive responses from a certain web server.
	client *http.Client

	// baseURL is the base URL where all the requests should be routed to.
	baseURL *url.URL
}

// Dial establishes a connection with a certain endpoint sending the given slice of bytes as input,
// it returns the response's body as a slice of bytes.
func (h httpDialer) Dial(ctx context.Context, endpoint string, in []byte) ([]byte, error) {
	method, path := h.resolveEndpoint(endpoint)

	u, err := h.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	buff := bytes.NewBuffer(in)
	req, err := http.NewRequestWithContext(ctx, method, u.String(), buff)
	if err != nil {
		return nil, err
	}

	res, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	var out []byte
	out, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New(string(out))
	}

	return out, nil
}

// resolveEndpoint resolves if the given endpoint is a valid endpoint
func (h httpDialer) resolveEndpoint(endpoint string) (method string, path string) {
	values := strings.Split(endpoint, " ")
	if len(values) != 2 {
		return http.MethodGet, "/"
	}
	method = values[0]
	path = values[1]

	return method, path
}

// NewDialerHTTP initializes a new HTTP Dialer.
func NewDialerHTTP(baseURL *url.URL, timeout time.Duration) Dialer {
	return &httpDialer{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}
