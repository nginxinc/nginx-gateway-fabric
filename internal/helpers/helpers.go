package helpers

import "github.com/google/go-cmp/cmp"

// Diff prints the diff between two structs.
// It is useful in testing to compare two structs when they are large. In such a case, without Diff it will be difficult
// to pinpoint the difference between the two structs.
func Diff(x, y interface{}) string {
	r := cmp.Diff(x, y)

	if r != "" {
		return "(-want +got)\n" + r
	}
	return r
}

// GetStringPointer takes a string and returns a pointer to it. Useful in unit tests when initializing structs.
func GetStringPointer(s string) *string {
	return &s
}

// GetInt32Pointer takes an int32 and returns a pointer to it. Useful in unit tests when initializing structs.
func GetInt32Pointer(i int32) *int32 {
	return &i
}
