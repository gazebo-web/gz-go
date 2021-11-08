package encoders

// Serializer serializes and deserializes data to/from specific formats.
type Serializer interface {
	// Marshal serializes the given data structure to a certain format and returns the representation in bytes.
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal tries to parse the given data with a certain format and fill the v with those values.
	Unmarshal(data []byte, v interface{}) error
}
