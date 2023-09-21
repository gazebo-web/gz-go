package encoders

import (
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"io"
	"reflect"
)

// typeOfBytes holds the type of a nil slice of bytes.
var typeOfBytes = reflect.TypeOf([]byte(nil))

var _ runtime.Marshaler = (*rawEncoder)(nil)

// rawEncoder provides methods to copy data into and from slices of bytes.
type rawEncoder struct {
}

// Marshal returns the raw encoding of v. This method expects a slice of bytes.
func (e rawEncoder) Marshal(v interface{}) ([]byte, error) {
	b, ok := v.([]byte)
	if !ok {
		return nil, errors.New("type must be []byte")
	}
	return b, nil
}

// Unmarshal writes data into v, v must be a pointer to a slice of bytes.
func (e rawEncoder) Unmarshal(data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("%T is not a pointer", v)
	}

	rv = rv.Elem()
	if rv.Type() != typeOfBytes {
		return fmt.Errorf("type must be []byte but got %T", v)
	}

	rv.Set(reflect.ValueOf(data))
	return nil
}

// NewEncoder returns an Encoder which writes bytes sequence into "w".
func (e rawEncoder) NewEncoder(w io.Writer) runtime.Encoder {
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

// NewDecoder returns a Decoder which reads byte sequence from "r".
func (e rawEncoder) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(v interface{}) error {
		raw, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		return e.Unmarshal(raw, v)
	})
}

// ContentType returns the Content-Type which this marshaler is responsible for.
// The parameter describes the type which is being marshalled, which can sometimes
// affect the content type returned.
func (e rawEncoder) ContentType(v interface{}) string {
	return "application/octet-stream"
}
