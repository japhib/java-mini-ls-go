package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStack(t *testing.T) {
	stack := NewStack[int]()

	assert.Equal(t, 0, stack.Size())
	assert.Equal(t, 0, stack.Pop())

	stack.Push(1)
	stack.Push(2)
	stack.Push(4)

	assert.Equal(t, 3, stack.Size())
	assert.Equal(t, 4, stack.Top())

	assert.Equal(t, 4, stack.Pop())
	assert.Equal(t, 2, stack.Pop())
	assert.Equal(t, 1, stack.Pop())
	assert.Equal(t, 0, stack.Pop())
	assert.Equal(t, 0, stack.Pop())

	assert.Equal(t, 0, stack.Size())
}
