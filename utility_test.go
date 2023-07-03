package gz

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/type/money"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestParseURL(t *testing.T) {
	u, err := ParseURL("wrong")
	assert.Error(t, err)
	assert.Nil(t, u)

	u, err = ParseURL("1234")
	assert.Error(t, err)
	assert.Nil(t, u)

	u, err = ParseURL("http://localhost:3333")
	assert.NoError(t, err)
	assert.NotNil(t, u)

	u, err = ParseURL("https://gazebosim.org")
	assert.NoError(t, err)
	assert.NotNil(t, u)

	u, err = ParseURL("ws://gazebosim.org/simulations/123456")
	assert.NoError(t, err)
	assert.NotNil(t, u)

	u, err = ParseURL("wss://gazebosim.org/simulations/123456")
	assert.NoError(t, err)
	assert.NotNil(t, u)

	u, err = ParseURL("s3://gazebosim.org/simulations/123456")
	assert.NoError(t, err)
	assert.NotNil(t, u)
}

func TestNewDateTime(t *testing.T) {
	now := time.Now()
	date := NewDateTime(now)
	require.NotNil(t, date)
	assert.NotEmpty(t, date.GetTimeZone().String())
	assert.Equal(t, now.Year(), int(date.GetYear()))
	assert.Equal(t, int(now.Month()), int(date.GetMonth()))
	assert.Equal(t, now.Day(), int(date.GetDay()))
	assert.Equal(t, now.Hour(), int(date.GetHours()))
	assert.Equal(t, now.Minute(), int(date.GetMinutes()))
	assert.Equal(t, now.Second(), int(date.GetSeconds()))
	assert.Equal(t, now.Nanosecond(), int(date.GetNanos()))
	assert.NotNil(t, date.TimeOffset)

	_, offset := now.Zone()
	assert.Equal(t, time.Duration(offset)*time.Second, date.GetUtcOffset().AsDuration())
	t.Log("Offset:", offset, "Got:", date.GetUtcOffset().AsDuration().String())
}

// NewMoney converts the given cents into a money.Money value.
func NewMoney(currency string, cents int64) *money.Money {
	u := cents / 100
	n := int32(cents-(u*100)) * int32(math.Pow10(7))
	return &money.Money{
		CurrencyCode: currency,
		Units:        u,
		Nanos:        n,
	}
}

func TestNewMoney(t *testing.T) {
	m := NewMoney("usd", 0)

	// If `units` is zero, `nanos` can be positive, zero, or negative.
	assert.Equal(t, "usd", m.GetCurrencyCode())
	assert.Equal(t, int64(0), m.GetUnits())
	assert.Equal(t, int32(0), m.GetNanos())

	// $-1.75 is represented as `units`=-1 and `nanos`=-750,000,000.
	m = NewMoney("usd", -175)
	assert.Equal(t, "usd", m.GetCurrencyCode())
	assert.Equal(t, int64(-1), m.GetUnits())
	assert.Equal(t, int32(-750_000_000), m.GetNanos())

	// $99.99 is represented as `units`=99 and `nanos`=990,000,000.
	m = NewMoney("usd", 9999)
	assert.Equal(t, "usd", m.GetCurrencyCode())
	assert.Equal(t, int64(99), m.GetUnits())
	assert.Equal(t, int32(990_000_000), m.GetNanos())

	// $19.99 is represented as `units`=19 and `nanos`=990,000,000.
	m = NewMoney("usd", 1999)
	assert.Equal(t, "usd", m.GetCurrencyCode())
	assert.Equal(t, int64(19), m.GetUnits())
	assert.Equal(t, int32(990_000_000), m.GetNanos())

	// $9.99 is represented as `units`=9 and `nanos`=990,000,000.
	m = NewMoney("usd", 999)
	assert.Equal(t, "usd", m.GetCurrencyCode())
	assert.Equal(t, int64(9), m.GetUnits())
	assert.Equal(t, int32(990_000_000), m.GetNanos())
}

func FuzzNewMoney(f *testing.F) {
	f.Add("usd", int64(175))
	f.Add("usd", int64(990))
	f.Add("usd", int64(1999))
	f.Add("usd", int64(9990))
	f.Add("usd", int64(99990))
	f.Fuzz(func(t *testing.T, c string, v int64) {
		NewMoney(c, v)
	})
}
