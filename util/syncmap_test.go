package util

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSyncMapWithGoroutines(t *testing.T) {
	expected := make(map[int]int)

	for i := 0; i < 1000; i++ {
		expected[i] = rand.Int()
	}

	sm := NewSyncMap[int, int]()

	// for waiting on all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(500)

	// writer goroutines
	for i := 0; i < 100; i++ {
		go func(i int) {
			for j := 0; j < 10; j++ {
				index := i*10 + j
				sm.Set(index, expected[index])
				time.Sleep(1 * time.Millisecond)
			}
			wg.Done()
		}(i)
	}

	// reader goroutines
	for i := 0; i < 400; i++ {
		go func(i int) {
			for j := 0; j < 10; j++ {
				sm.Get(rand.Intn(1000))
				time.Sleep(1 * time.Millisecond)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

	for i := 0; i < 1000; i++ {
		actual, _ := sm.Get(i)
		assert.Equal(t, actual, expected[i])
	}
}
