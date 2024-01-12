package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	t.Run("newLimiter with 1 concurrency", func(t *testing.T) {
		limiter := newLimiter(1)

		times := 10
		p := make(chan int, times)
		for i := 0; i < times; i++ {
			i := i
			doneLimit := limiter.acquire()
			go func() {
				defer doneLimit()

				p <- i
			}()
		}

		for i := 0; i < times; i++ {
			assert.Equal(t, i, <-p)
		}
	})
}
