package engine

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	t.Run("newLimiter with 1 concurrency", func(t *testing.T) {
		limiter := newLimiter(1)

		times := 10
		wg := &sync.WaitGroup{}
		p := make(chan int, times)
		for i := 0; i < times; i++ {
			i := i
			wg.Add(1)
			doneLimit := limiter.acquire()
			go func() {
				defer wg.Done()
				defer doneLimit()

				p <- i
			}()
		}
		wg.Wait()

		for i := 0; i < times; i++ {
			assert.Equal(t, i, <-p)
		}
	})
}
