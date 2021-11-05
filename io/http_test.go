package io

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ignitionrobotics/web/ign-go/encoders"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type inputTest struct {
	Data string `json:"data"`
}

func (in *inputTest) ToJSON() ([]byte, error) {
	return json.Marshal(in)
}

func (in *inputTest) FromJSON(data []byte) error {
	return json.Unmarshal(data, in)
}

type outputTest struct {
	Result string `json:"result"`
}

func (out *outputTest) ToJSON() ([]byte, error) {
	return json.Marshal(out)
}

func (out *outputTest) FromJSON(data []byte) error {
	return json.Unmarshal(data, out)
}

func TestHTTPClient_CallWithIOErrors(t *testing.T) {
	u, err := url.Parse("http://localhost")
	require.NoError(t, err)

	c := NewClientHTTP(ClientOptions{
		Timeout: 10 * time.Second,
		URL:     u,
	})

	err = c.Call(context.Background(), http.MethodPost, "/test", encoders.FormatJSON, nil, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrNilValuesIO, err)

	err = c.Call(context.Background(), http.MethodPost, "/test", encoders.FormatJSON, &inputTest{}, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrNilValuesIO, err)

	err = c.Call(context.Background(), http.MethodPost, "/test", encoders.FormatJSON, nil, &outputTest{})
	assert.Error(t, err)
	assert.Equal(t, ErrNilValuesIO, err)
}

func TestHttpClient_Call(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in inputTest

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = in.FromJSON(body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		out := outputTest{Result: in.Data}

		body, err = out.ToJSON()
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

	c := NewClientHTTP(ClientOptions{
		Timeout: 10 * time.Second,
		URL:     u,
	})

	in := inputTest{Data: "test"}
	var out outputTest

	require.NoError(t, c.Call(context.Background(), http.MethodPost, "/test", encoders.FormatJSON, &in, &out))

	assert.Equal(t, "test", out.Result)
}
