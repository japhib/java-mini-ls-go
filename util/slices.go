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

// CombineSlices takes any number of slices and combines their contents into one single slice.
func CombineSlices[T any](inputs ...[]T) []T {
	totalLength := 0
	for _, slice := range inputs {
		totalLength += len(slice)
	}

	ret := make([]T, 0, totalLength)
	for _, slice := range inputs {
		ret = append(ret, slice...)
	}

	return ret
}

// Reverse reverses a slice.
func Reverse[T any](slice []T) []T {
	ret := make([]T, 0, len(slice))

	for i := len(slice) - 1; i >= 0; i-- {
		ret = append(ret, slice[i])
	}

	return ret
}

func Keys[K comparable, V any](m map[K]V) []K {
	ret := []K{}
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}

func Values[K comparable, V any](m map[K]V) []V {
	ret := []V{}
	for _, v := range m {
		ret = append(ret, v)
	}
	return ret
}
