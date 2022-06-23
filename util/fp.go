package util

import "fmt"

// Map maps all elements of a slice of any type into the given function
// and returns a new slice with the result
func Map[T any, U any](input []T, f func(T) U) []U {
	ret := make([]U, 0, len(input))

	for _, i := range input {
		ret = append(ret, f(i))
	}

	return ret
}

// MapToString maps all elements of a slice of any type into String() and
// returns a new slice with the result
func MapToString[T fmt.Stringer](input []T) []string {
	ret := make([]string, 0, len(input))

	for _, i := range input {
		ret = append(ret, i.String())
	}

	return ret
}
