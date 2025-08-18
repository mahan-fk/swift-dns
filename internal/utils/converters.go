package utils

func ToAnySlice[T any](s []T) []any {
	a := make([]any, len(s))

	for i, v := range s {
		a[i] = v
	}

	return a
}
