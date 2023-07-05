package gz

import (
	"google.golang.org/genproto/googleapis/type/datetime"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/durationpb"
	"math"
	"time"
)

// NewDateTime initializes a new datetime.DateTime from the current time.Time.
// It infers the datetime.DateTime's timezone from the current location.
// NOTE: datetime.DateTime is a protobuf message.
func NewDateTime(t time.Time) *datetime.DateTime {
	_, offset := t.Zone()
	return &datetime.DateTime{
		Year:       int32(t.Year()),
		Month:      int32(t.Month()),
		Day:        int32(t.Day()),
		Hours:      int32(t.Hour()),
		Minutes:    int32(t.Minute()),
		Seconds:    int32(t.Second()),
		Nanos:      int32(t.Nanosecond()),
		TimeOffset: &datetime.DateTime_UtcOffset{UtcOffset: durationpb.New(time.Duration(offset) * time.Second)},
	}
}

// NewMoney converts the given cents into a money.Money value.
// NOTE: money.Money is a protobuf message.
func NewMoney(currency string, cents int64) *money.Money {
	u := cents / 100
	n := int32(cents-(u*100)) * int32(math.Pow10(7))
	return &money.Money{
		CurrencyCode: currency,
		Units:        u,
		Nanos:        n,
	}
}
