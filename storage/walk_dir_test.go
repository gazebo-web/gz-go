package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestWalkDir(t *testing.T) {
	ctx := context.Background()
	var count int

	fn := WalkDirFunc(func(ctx context.Context, path string, body io.Reader) error {
		count++
		return nil
	})

	assert.NoError(t, WalkDir(ctx, "./testdata/example", fn))
	assert.Equal(t, 4, count)
}
