package io

import (
	"bytes"
	"context"
	"errors"
	"gitlab.com/ignitionrobotics/web/ign-go/encoders"
	"io"
	"net/http"
	"net/url"
	"reflect"
)

// httpClient is an implementation of Client using an HTTP transport layer.
type httpClient struct {
	// client is the underlying HTTP client.
	client *http.Client
	// baseURL is the base URL where all the requests should be routed to.
	baseURL *url.URL
}

// Call calls a method to a given endpoint using an HTTP request.
// It parses the request and the response to different formats defined by the format argument.
func (c *httpClient) Call(ctx context.Context, method, endpoint string, format encoders.Format, in, out encoders.Serializer) error {
	if in == nil || out == nil {
		return ErrNilValuesIO
	}
	if reflect.ValueOf(out).Kind() != reflect.Ptr {
		return ErrOutputMustBePointer
	}

	var err error
	switch format {
	case encoders.FormatJSON:
		err = c.callJSON(ctx, method, endpoint, in, out)
	}
	if err != nil {
		return err
	}

	return nil
}

// callJSON is a helper function of Call used for JSON-specific HTTP requests.
func (c *httpClient) callJSON(ctx context.Context, method, endpoint string, in, out encoders.JSON) error {
	body, err := in.ToJSON()
	if err != nil {
		return err
	}
	buff := bytes.NewBuffer(body)

	u, err := c.baseURL.Parse(endpoint)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), buff)
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return errors.New(string(body))
	}

	if err = out.FromJSON(body); err != nil {
		return err
	}
	return nil
}

// NewClientHTTP initializes a new Client implementation using HTTP as transport layer.
func NewClientHTTP(opts ClientOptions) Client {
	return &httpClient{
		client: &http.Client{
			Timeout: opts.Timeout,
		},
		baseURL: opts.URL,
	}
}
