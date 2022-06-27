package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStack(t *testing.T) {
	stack := NewStack[int]()

	// stack is empty
	assert.Equal(t, 0, stack.Size())
	assert.Equal(t, 0, stack.Pop())
	assert.Equal(t, 0, stack.Top())
	assert.Equal(t, 0, stack.TopMinus(1))

	// add some values
	stack.Push(1)
	stack.Push(2)
	stack.Push(4)

	// stack is full, verify all expected values are there
	assert.Equal(t, 3, stack.Size())
	assert.Equal(t, 4, stack.Top())
	assert.Equal(t, 2, stack.TopMinus(1))
	assert.Equal(t, 1, stack.TopMinus(2))
	assert.Equal(t, 0, stack.TopMinus(3))

	// make the stack empty again
	assert.Equal(t, 4, stack.Pop())
	assert.Equal(t, 2, stack.Pop())
	assert.Equal(t, 1, stack.Pop())
	assert.Equal(t, 0, stack.Pop())
	assert.Equal(t, 0, stack.Pop())

	// test empty state again
	assert.Equal(t, 0, stack.Size())
	assert.Equal(t, 0, stack.Pop())
	assert.Equal(t, 0, stack.Top())
	assert.Equal(t, 0, stack.TopMinus(1))
}
