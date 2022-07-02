package util

// Set is basically a hash set
type Set[T comparable] struct {
	values map[T]bool
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		values: make(map[T]bool),
	}
}

func SetFromSlice[T comparable](slice []T) *Set[T] {
	values := make(map[T]bool)
	for _, item := range slice {
		values[item] = true
	}

	return &Set[T]{
		values: values,
	}
}

func SetFromValues[T comparable](slice ...T) *Set[T] {
	return SetFromSlice(slice)
}

func (s *Set[T]) Add(item T) {
	s.values[item] = true
}

func (s *Set[T]) Remove(item T) {
	delete(s.values, item)
}

func (s *Set[T]) Contains(item T) bool {
	_, ok := s.values[item]
	return ok
}
