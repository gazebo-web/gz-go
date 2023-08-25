package encoders

import "github.com/jszwec/csvutil"

// csvEncoder implements Marshaller for the CSV format.
type csvEncoder struct {
}

// Marshal returns the CSV encoding of v.
func (csvEncoder) Marshal(v interface{}) ([]byte, error) {
	return csvutil.Marshal(v)
}

// Unmarshal parses the CSV-encoded data and stores the results in the value pointed to by v.
// NOTE: Given the nature of CSV files, v must be a not nil pointer to a slice.
func (csvEncoder) Unmarshal(data []byte, v interface{}) error {
	return csvutil.Unmarshal(data, v)
}

// CSV holds a csv encoder instance implementing Marshaller.
var CSV = &csvEncoder{}
