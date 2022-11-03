package encoders

import (
	"encoding/json"
	"io"
)

var _ Marshaller = (*jsonEncoder)(nil)
var _ WriterEncoder = (*jsonEncoder)(nil)

// jsonEncoder implements Marshaller for the JSON format.
type jsonEncoder struct{}

// Write writes the JSON encoding of v to w.
func (jsonEncoder) Write(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

// Marshal returns the JSON encoding of v.
func (jsonEncoder) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
func (jsonEncoder) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// JSON holds a json encoder instance implementing Marshaller.
var JSON = &jsonEncoder{}
