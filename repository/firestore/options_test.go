package firestore

import (
	"github.com/gazebo-web/gz-go/v7/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetMaxResults(t *testing.T) {
	var opts []repository.Option

	opts, err := setMaxResults(opts, nil)
	assert.NoError(t, err)
	assert.Len(t, opts, 1)
}

func TestSetMaxResults_Empty(t *testing.T) {
	var opts []repository.Option

	req := TestPaginationRequest{}

	opts, err := setMaxResults(opts, &req)
	assert.NoError(t, err)
	assert.Len(t, opts, 1)
}

func TestSetMaxResults_Max(t *testing.T) {
	var opts []repository.Option

	req := TestPaginationRequest{
		size: 1001,
	}

	opts, err := setMaxResults(opts, &req)
	assert.NoError(t, err)
	assert.Len(t, opts, 1)
}

func TestSetMaxResults_Invalid(t *testing.T) {
	var opts []repository.Option

	req := TestPaginationRequest{
		size: -100,
	}

	var err error
	opts, err = setMaxResults(opts, &req)
	assert.Error(t, err)
	assert.Len(t, opts, 0)
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
