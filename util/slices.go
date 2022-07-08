package util

import (
	"fmt"
	"sync"
)

// Map maps all elements of a slice of any type into the given function
// and returns a new slice with the result
func Map[T any, U any](input []T, f func(T) U) []U {
	ret := make([]U, 0, len(input))

	for _, i := range input {
		ret = append(ret, f(i))
	}

	return ret
}

// MapAsync is like Map but it spins up a separate goroutine for each
// element of the input, executes the function in parallel, and waits for
// the result.
func MapAsync[T any, U any](input []T, f func(T) U) []U {
	result := make([]U, len(input))

	var wg sync.WaitGroup
	wg.Add(len(input))
	for idx, value := range input {
		go func(i int, val T) {
			result[i] = f(val)
			wg.Done()
		}(idx, value)
	}
	wg.Wait()

	return result
}

// EachAsync is like MapAsync but it doesn't return a value
func EachAsync[T any](input []T, f func(T)) {
	var wg sync.WaitGroup
	wg.Add(len(input))
	for idx, value := range input {
		go func(i int, val T) {
			f(val)
			wg.Done()
		}(idx, value)
	}
	wg.Wait()
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
