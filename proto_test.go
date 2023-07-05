package gz

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

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

func TestNewMoney(t *testing.T) {
	m := NewMoney("usd", 0)

	assert.NotNil(t, m)

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
