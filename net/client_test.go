package net

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gazebo-web/gz-go/v9/encoders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type inputTest struct {
	Data string `json:"data"`
}

type outputTest struct {
	Result string `json:"result"`
}

func TestHTTPClient_CallWithIOErrors(t *testing.T) {
	u, err := url.Parse("http://localhost")
	require.NoError(t, err)

	d := NewCallerHTTP(u, map[string]EndpointHTTP{
		"TestEndpoint": {
			Method: "GET",
			Path:   "/test",
		},
	}, time.Second)

	c := NewClient(d, encoders.JSON)

	err = c.Call(context.Background(), "TestEndpoint", nil, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrNilValuesIO, err)

	err = c.Call(context.Background(), "TestEndpoint", &inputTest{}, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrNilValuesIO, err)

	err = c.Call(context.Background(), "TestEndpoint", nil, &outputTest{})
	assert.Error(t, err)
	assert.Equal(t, ErrNilValuesIO, err)
}

func TestHttpClient_Call(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in inputTest

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/test", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = encoders.JSON.Unmarshal(body, &in); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		out := outputTest{Result: in.Data}

		body, err = encoders.JSON.Marshal(out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	u, err := url.ParseRequestURI(server.URL)
	require.NoError(t, err)

	d := NewCallerHTTP(u, map[string]EndpointHTTP{
		"CreateTest": {
			Method: "POST",
			Path:   "/test",
		},
	}, time.Second)

	c := NewClient(d, encoders.JSON)

	in := inputTest{Data: "test"}
	var out outputTest

	require.NoError(t, c.Call(context.Background(), "CreateTest", &in, &out))
	assert.Equal(t, "test", out.Result)
}
