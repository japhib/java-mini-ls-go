package util

type Stack[T any] struct {
	contents []T
}

func NewStack[T any]() Stack[T] {
	return Stack[T]{
		contents: make([]T, 0),
	}
}

func (st *Stack[T]) Push(t T) {
	st.contents = append(st.contents, t)
}

func (st *Stack[T]) Pop() T {
	if st.Size() == 0 {
		// return the zero value of type T
		var ret T
		return ret
	}

	ret := st.Top()
	st.contents = st.contents[:st.Size()-1]
	return ret
}

func (st *Stack[T]) Top() T {
	if len(st.contents) == 0 {
		// return the zero value of type T
		var ret T
		return ret
	}

	return st.contents[len(st.contents)-1]
}

func (st *Stack[T]) TopMinus(offsetFromTop int) T {
	if len(st.contents)-offsetFromTop <= 0 {
		// return the zero value of type T
		var ret T
		return ret
	}

	return st.contents[len(st.contents)-1-offsetFromTop]
}

func (st *Stack[T]) Size() int {
	return len(st.contents)
}
