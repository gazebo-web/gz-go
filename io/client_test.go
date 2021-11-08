package io

import (
	"context"
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

type outputTest struct {
	Result string `json:"result"`
}

func TestHTTPClient_CallWithIOErrors(t *testing.T) {
	u, err := url.Parse("http://localhost")
	require.NoError(t, err)

	d := NewDialerHTTP(u, time.Second)

	c := NewClient(d, encoders.JSON)

	err = c.Call(context.Background(), "GET /", nil, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrNilValuesIO, err)

	err = c.Call(context.Background(), "GET /", &inputTest{}, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrNilValuesIO, err)

	err = c.Call(context.Background(), "GET /", nil, &outputTest{})
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

	d := NewDialerHTTP(u, time.Second)

	c := NewClient(d, encoders.JSON)

	in := inputTest{Data: "test"}
	var out outputTest

	require.NoError(t, c.Call(context.Background(), "POST /test", &in, &out))

	assert.Equal(t, "test", out.Result)
}
