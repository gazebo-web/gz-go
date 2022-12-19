package gz

import (
	"errors"
	"github.com/stretchr/testify/assert"
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
