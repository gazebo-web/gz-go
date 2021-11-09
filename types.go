package ign

// Int returns a pointer to the int value passed in.
func Int(v int) *int {
	return &v
}

// Int64 returns a pointer to the int64 value passed in.
func Int64(v int64) *int64 {
	return &v
}

// Float64 returns a pointer to the float64 value passed in.
func Float64(v float64) *float64 {
	return &v
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

// StringSlice converts a slice of string values into a slice of
// string pointers
func StringSlice(src []string) []*string {
	dst := make([]*string, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = &(src[i])
	}
	return dst
}

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}
