package validate

import (
	"testing"
)

func TestIsZero(t *testing.T) {
	type Test struct {
		V int
	}

	var zeroStruct Test
	var zeroStructPtr *Test
	var zeroMap map[string]any
	var zeroSlice []string

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "string_zero", value: "", want: true},
		{name: "string_not_zero", value: "test", want: false},

		{name: "int_zero", value: 0, want: true},
		{name: "int_not_zero", value: 4, want: false},

		{name: "int8_zero", value: int8(0), want: true},
		{name: "int8_not_zero", value: int8(4), want: false},

		{name: "int16_zero", value: int16(0), want: true},
		{name: "int16_not_zero", value: int16(4), want: false},

		{name: "int32_zero", value: int32(0), want: true},
		{name: "int32_not_zero", value: int32(4), want: false},

		{name: "int64_zero", value: int64(0), want: true},
		{name: "int64_not_zero", value: int64(4), want: false},

		{name: "bool_zero", value: false, want: true},
		{name: "bool_not_zero", value: true, want: false},

		{name: "uint_zero", value: uint(0), want: true},
		{name: "uint_not_zero", value: uint(4), want: false},

		{name: "uint8_zero", value: uint8(0), want: true},
		{name: "uint8_not_zero", value: uint8(4), want: false},

		{name: "uint16_zero", value: uint16(0), want: true},
		{name: "uint16_not_zero", value: uint16(4), want: false},

		{name: "uint32_zero", value: uint32(0), want: true},
		{name: "uint32_not_zero", value: uint32(4), want: false},

		{name: "uint64_zero", value: uint64(0), want: true},
		{name: "uint64_not_zero", value: uint64(4), want: false},

		{name: "float32_zero", value: float32(0), want: true},
		{name: "float32_not_zero", value: float32(4), want: false},

		{name: "float64_zero", value: float64(0), want: true},
		{name: "float64_not_zero", value: float64(4), want: false},

		{name: "complex64_zero", value: complex64(0), want: true},
		{name: "complex64_not_zero", value: complex64(4), want: false},

		{name: "complex128_zero", value: complex128(0), want: true},
		{name: "complex128_not_zero", value: complex128(4), want: false},

		{name: "slice_zero", value: zeroSlice, want: true},
		{name: "slice_not_zero", value: []string{}, want: false},

		{name: "map_zero", value: zeroMap, want: true},
		{name: "map_not_zero", value: map[string]any{}, want: false},

		{name: "struct_zero", value: zeroStruct, want: true},
		{name: "struct_not_zero", value: Test{V: 123}, want: false},

		{name: "structptr_zero", value: zeroStructPtr, want: true},
		{name: "structptr_not_zero", value: &Test{}, want: false},

		{name: "bool_zero", value: false, want: true},
		{name: "bool_not_zero", value: true, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsZero(tt.value); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasZero(t *testing.T) {
	tests := []struct {
		name   string
		values []any
		want   bool
	}{
		{
			name:   "no_zeros",
			values: []any{1, "OK", true},
			want:   false,
		},
		{
			name:   "some_zeros",
			values: []any{1, "OK", true, 0, "", false},
			want:   true,
		},
		{
			name:   "all_zeros",
			values: []any{0, "", false},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasZero(tt.values...); got != tt.want {
				t.Errorf("HasZero() = %v, want %v", got, tt.want)
			}
		})
	}
}
