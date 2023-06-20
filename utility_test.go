package gz

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests for utility file

// TestStrToSlice tests the StrToSlice func
func TestStrToSlice(t *testing.T) {

	type exp struct {
		input string
		exp   []string
	}
	var inputs = []exp{
		{" tag middle space,  test_tag2 ,   , test_tag_1,  ",
			[]string{"tag middle space", "test_tag_1", "test_tag2"},
		},
	}

	for _, i := range inputs {
		got := StrToSlice(i.input)
		for _, s := range got {
			t.Log("got:", strings.Replace(s, " ", "%s", -1))
		}
		assert.True(t, SameElements(i.exp, got), "Didn't get expected string slice exp:[%s] got:[%s]", i.exp, got)
	}
}

func TestIsError(t *testing.T) {
	target := errors.New("test")
	err := errors.New("this is a test error")
	assert.True(t, IsError(err, target))

	err = errors.New("another error")
	assert.False(t, IsError(err, target))
}

func TestRemoveIfFound(t *testing.T) {
	path, err := os.MkdirTemp("", "test")
	require.NoError(t, err)

	fp := filepath.Join(path, "test.txt")
	f, err := os.Create(fp)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	assert.NoError(t, RemoveIfFound(fp))
	_, err = os.Stat(fp)
	assert.Error(t, err) // File should not exist

	require.NoError(t, os.RemoveAll(path))
}

func TestValidateURL(t *testing.T) {
	assert.Error(t, ValidateURL("wrong"))
	assert.Error(t, ValidateURL("1234"))
	assert.NoError(t, ValidateURL("http://localhost:3333"))
	assert.NoError(t, ValidateURL("https://gazebosim.org"))
}
