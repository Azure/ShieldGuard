package engine

import "runtime"

// channelLimiter is a channel based implementation of Limiter.
type channelLimiter chan struct{}

var _ limiter = (*channelLimiter)(nil)

func (l channelLimiter) acquire() func() {
	p := <-l
	return func() { l <- p }
}

func newLimiter(concurrency int) channelLimiter {
	rv := make(chan struct{}, concurrency)

	for i := 0; i < concurrency; i++ {
		rv <- struct{}{}
	}

	return rv
}

func newLimiterFromMaxProcs() channelLimiter {
	return newLimiter(runtime.GOMAXPROCS(0))
}
