package engine

import (
	"fmt"
	"sync"

	"github.com/Azure/ShieldGuard/sg/internal/result"
)

func (k queryCacheKey) cacheKey() string {
	return fmt.Sprintf(
		"%s:%d:%s",
		k.compilerKey,
		k.parsedInput.Hash(),
		k.query,
	)
}

type noopQueryCacheT struct{}

func (noopQueryCacheT) set(_ queryCacheKey, _ []result.Result) {}

func (noopQueryCacheT) get(_ queryCacheKey) ([]result.Result, bool) {
	return nil, false
}

var noopQueryCache = noopQueryCacheT{}

type queryCache struct {
	mu    *sync.RWMutex
	items map[string][]result.Result
}

// NewQueryCache creates a new QueryCache.
func NewQueryCache() QueryCache {
	return &queryCache{
		mu:    &sync.RWMutex{},
		items: make(map[string][]result.Result),
	}
}

var _ QueryCache = (*queryCache)(nil)

func (qc *queryCache) set(key queryCacheKey, value []result.Result) {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	qc.items[key.cacheKey()] = value
}

func (qc *queryCache) get(key queryCacheKey) ([]result.Result, bool) {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	rv, ok := qc.items[key.cacheKey()]
	return rv, ok
}
