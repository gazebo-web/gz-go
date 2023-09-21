package encoders

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jszwec/csvutil"
	"io"
)

var _ runtime.Marshaler = (*csvEncoder)(nil)

// csvEncoder implements Marshaller for the CSV format.
type csvEncoder struct {
}

// NewDecoder returns a Decoder which reads byte sequence from "r".
func (e csvEncoder) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(v interface{}) error {
		raw, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		return e.Unmarshal(raw, v)
	})
}

// NewEncoder returns an Encoder which writes bytes sequence into "w".
func (e csvEncoder) NewEncoder(w io.Writer) runtime.Encoder {
	return runtime.EncoderFunc(func(v interface{}) error {
		b, err := e.Marshal(v)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			return err
		}
		return nil
	})
}

// ContentType returns the Content-Type which this marshaler is responsible for.
// The parameter describes the type which is being marshalled, which can sometimes
// affect the content type returned.
func (e csvEncoder) ContentType(v interface{}) string {
	return "text/csv"
}

// Marshal returns the CSV encoding of v.
func (csvEncoder) Marshal(v interface{}) ([]byte, error) {
	return csvutil.Marshal(v)
}

// Unmarshal parses the CSV-encoded data and stores the results in the value pointed to by v.
// NOTE: Given the nature of CSV files, v must be a non-nil pointer to a slice.
func (csvEncoder) Unmarshal(data []byte, v interface{}) error {
	return csvutil.Unmarshal(data, v)
}

// CSV holds a csv encoder instance implementing Marshaller.
var CSV = &csvEncoder{}
