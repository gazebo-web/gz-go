package io

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// httpCaller is a Caller implementation using HTTP as transport layer.
type httpCaller struct {
	// client is the HTTP client used to create requests and receive responses from a certain web server.
	client *http.Client

	// baseURL is the base URL where all the requests should be routed to.
	baseURL *url.URL

	// endpoints contains a set of HTTP endpoints that this caller can communicate with.
	endpoints map[string]EndpointHTTP
}

// Call establishes a connection with a certain endpoint sending the given slice of bytes as input,
// it returns the response's body as a slice of bytes.
func (h *httpCaller) Call(ctx context.Context, endpoint string, in []byte) ([]byte, error) {
	e := h.resolveEndpoint(endpoint)

	u, err := h.baseURL.Parse(e.Path)
	if err != nil {
		return nil, err
	}

	buff := bytes.NewBuffer(in)
	req, err := http.NewRequestWithContext(ctx, e.Method, u.String(), buff)
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
func (h *httpCaller) resolveEndpoint(endpoint string) EndpointHTTP {
	e, ok := h.endpoints[endpoint]
	if !ok {
		return defaultEndpointHTTP
	}
	return e
}

// defaultEndpointHTTP is a default endpoint returned when no HTTP endpoint has been found.
var defaultEndpointHTTP = EndpointHTTP{
	Method: http.MethodGet,
	Path:   "/",
}

// EndpointHTTP represents an HTTP endpoint.
type EndpointHTTP struct {
	// Method is the HTTP verb supported by this endpoint.
	Method string
	// Path is the relative path where this endpoint is located.
	// Example: /example/test
	Path string
}

// NewCallerHTTP initializes a new HTTP Caller.
func NewCallerHTTP(baseURL *url.URL, endpoints map[string]EndpointHTTP, timeout time.Duration) Caller {
	return &httpCaller{
		baseURL:   baseURL,
		endpoints: endpoints,
		client:    &http.Client{Timeout: timeout},
	}
}
