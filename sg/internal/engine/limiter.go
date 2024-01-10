package engine

import "runtime"

// channelLimiter is a channel based implementation of Limiter.
type channelLimiter chan struct{}

var _ Limiter = (*channelLimiter)(nil)

func (l channelLimiter) acquire() func() {
	l <- struct{}{}
	return func() { <-l }
}

func newLimiter(concurrency int) channelLimiter {
	return make(chan struct{}, concurrency)
}

func newCPUBoundedLimiter() channelLimiter {
	return newLimiter(runtime.GOMAXPROCS(0))
}
