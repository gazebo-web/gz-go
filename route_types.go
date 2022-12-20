package gz

import (
	"net/http"
)

// Route is a definition of a route
type Route struct {

	// Name of the route
	Name string `json:"name"`

	// Description of the route
	Description string `json:"description"`

	// URI pattern
	URI string `json:"uri"`

	// Headers required by the route
	Headers []Header `json:"headers"`

	// HTTP methods supported by the route
	Methods Methods `json:"methods"`

	// Secure HTTP methods supported by the route
	SecureMethods SecureMethods `json:"secure_methods"`
}

// Routes is an array of Route
type Routes []Route

// Header stores the information about headers included in a request.
type Header struct {
	Name          string `json:"name"`
	HeaderDetails Detail `json:"details"`
}

// Detail stores information about a paramter.
type Detail struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// Method associates an HTTP method (GET, POST, PUT, DELETE) with a list of
// handlers.
type Method struct {
	// GET, POST, PUT, DELETE
	// \todo: Make this an enum
	Type string `json:"type"`

	// Description of the method
	Description string `json:"description"`

	// A slice of hanlders used to process this method.
	Handlers FormatHandlers `json:"handler"`
}

// Methods is a slice of Method.
type Methods []Method

// SecureMethods is a slice of Method that require authentication.
type SecureMethods []Method

// FormatHandler represents a format type string, and handler function pair. Handlers are called in response to a route request.
type FormatHandler struct {
	// Format (eg: .json, .proto, .html)
	Extension string `json:"extension"`

	// Processor for the url pattern
	Handler http.Handler `json:"-"`
}

// FormatHandlers is a slice of FormatHandler values.
type FormatHandlers []FormatHandler

// AuthHeadersRequired is an array of Headers needed when authentication is
// required.
var AuthHeadersRequired = []Header{
	{
		Name: "authorization: Bearer <YOUR_JWT_TOKEN>",
		HeaderDetails: Detail{
			Required: true,
		},
	},
}

// AuthHeadersOptional is an array of Headers needed when authentication is
// optional.
var AuthHeadersOptional = []Header{
	{
		Name: "authorization: Bearer <YOUR_JWT_TOKEN>",
		HeaderDetails: Detail{
			Required: false,
		},
	},
}
