package crud

// IntsToAnys converts a slice of any integer type to a slice of any.
// This is a helper function for the WithIn option when using integer keys for relations.
func IntsToAnys[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](s []T) []any {
	ans := make([]any, len(s))
	for i, v := range s {
		ans[i] = v
	}
	return ans
}
