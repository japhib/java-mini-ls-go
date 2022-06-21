package util

// Map maps all elements of a slice of any type into the given function
// and returns a new slice with the result
func Map[T any, U any](input []T, f func(T) U) []U {
	ret := make([]U, 0, len(input))

	for _, i := range input {
		ret = append(ret, f(i))
	}

	return ret
}
