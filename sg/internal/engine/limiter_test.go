package engine

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	t.Run("newLimiter with 1 concurrency", func(t *testing.T) {
		limiter := newLimiter(1)

		wg := &sync.WaitGroup{}
		p := make(chan int, 10)
		for i := 0; i < len(p); i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer limiter.acquire()()

				p <- i
			}()
		}
		wg.Wait()

		for i := 0; i < len(p); i++ {
			assert.Equal(t, i, <-p)
		}
	})
}
