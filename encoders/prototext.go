package encoders

import (
	"errors"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

var _ Marshaller = (*protoTextEncoder)(nil)

// protoTextEncoder implements Marshaller for the ProtoText format.
type protoTextEncoder struct{}

// Marshal returns the ProtoText encoding of v.
func (protoTextEncoder) Marshal(v interface{}) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, errors.New("invalid proto message")
	}
	return prototext.Marshal(m)
}

// Unmarshal parses the ProtoText-encoded data and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
func (protoTextEncoder) Unmarshal(data []byte, v interface{}) error {
	m, ok := v.(proto.Message)
	if !ok {
		return errors.New("invalid proto message")
	}
	return prototext.Unmarshal(data, m)
}

// ProtoText holds a proto text encoder instance implementing Marshaller.
var ProtoText = &protoTextEncoder{}
