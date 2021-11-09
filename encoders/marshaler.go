package encoders

// Marshaller marshals and unmarshals data to/from specific formats.
type Marshaller interface {
	// Marshal marshals the given data structure to a certain format and returns the representation in bytes.
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal parses a byte representation of a given data in a certain format and loads v with matching parsed values.
	Unmarshal(data []byte, v interface{}) error
}
