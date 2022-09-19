package repository

// Filter is used by a Repository to filter data.
type Filter struct {
	// Template contains a filter template.
	// SQL:
	//   Contains a template string with placeholders (?).
	//   All values in a filter that are filled in using user-defined values should use a placeholder (?) to prevent
	//   SQL injection.
	// Example: `name = ? AND age = ?`
	Template string
	// Values contains a sequence of values for the placeholders defined in the template.
	// The values are replaced in the order they are defined in the template.
	// Example: `["Test", 33]`
	Values []interface{}
}
