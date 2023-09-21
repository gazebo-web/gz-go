package encoders

import (
	"context"
	"github.com/gazebo-web/gz-go/v8/telemetry"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/otel/codes"
	"io"
)

// WriterEncoder encodes and writes data to a Writer.
type WriterEncoder interface {
	// Write encodes v in a certain format and writes it to w.
	Write(w io.Writer, v interface{}) error
}

// Marshaller marshals and unmarshals data to/from specific formats.
type Marshaller interface {
	// Marshal marshals the given data structure to a certain format and returns the representation in bytes.
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal parses a byte representation of a given data in a certain format and loads v with matching parsed values.
	Unmarshal(data []byte, v interface{}) error
}

// Unmarshal allows unmarshalling body into a value of T using the given Marshaller.
func Unmarshal[T any](ctx context.Context, m Marshaller, body []byte) (T, error) {
	_, span := telemetry.NewChildSpan(ctx, "Unmarshal")
	defer span.End()
	var value T
	span.AddEvent("Unmarshalling event")
	if err := m.Unmarshal(body, &value); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to unmarshal")
		var zero T
		return zero, err
	}
	return value, nil
}

// newEncoderFunc returns a new runtime.EncoderFunc that uses the given Marshaller to marshal and write bytes to the
// given io.Writer.
func newEncoderFunc(w io.Writer, m Marshaller) runtime.EncoderFunc {
	return func(v interface{}) error {
		b, err := m.Marshal(v)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			return err
		}
		return nil
	}
}

// newDecoderFunc returns a new runtime.DecoderFunc that uses the given Marshaller to unmarshal the bytes read from
// the given io.Reader.
func newDecoderFunc(r io.Reader, m Marshaller) runtime.DecoderFunc {
	return func(v interface{}) error {
		raw, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		if c, ok := r.(io.Closer); ok {
			defer c.Close()
		}
		return m.Unmarshal(raw, v)
	}
}
