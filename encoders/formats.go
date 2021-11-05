package encoders

// Format describes to what format a certain payload should be deserialized and serialized to.
type Format string

const (
	// FormatJSON is used when a payload is in JSON format.
	FormatJSON Format = "json"

	// FormatXML is used when a payload is in XML format.
	FormatXML Format = "xml"

	// FormatText is used when a payload is in plain text format.
	FormatText Format = "text"

	// FormatProtobuf is used when a payload is in protobuf format.
	FormatProtobuf Format = "protobuf"
)

// JSON holds methods to serialize and deserialize data structures to and from json format.
type JSON interface {
	// ToJSON converts the data structure to JSON format.
	ToJSON() ([]byte, error)

	// FromJSON fills the data structure with the JSON values given by data.
	FromJSON(data []byte) error
}

// Serializer serializes and deserializes data to/from specific formats.
type Serializer interface {
	JSON
}
