package engine

import (
	"sync"
	"testing"

	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/open-policy-agent/opa/ast"
	"github.com/stretchr/testify/assert"
)

func Test_noopQueryCacheT(t *testing.T) {
	t.Parallel()

	_, ok := noopQueryCache.get(queryCacheKey{})
	assert.False(t, ok)
	noopQueryCache.set(queryCacheKey{}, nil)
	_, ok = noopQueryCache.get(queryCacheKey{})
	assert.False(t, ok)
}

func Test_QueryCache(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		cacheKey1 := queryCacheKey{
			compilerKey: "compilerKey1",
			parsedInput: ast.String("input1"),
			query:       "query1",
		}
		cacheKey2 := queryCacheKey{
			compilerKey: "compilerKey2",
			parsedInput: ast.String("input2"),
			query:       "query2",
		}
		queryResult := []result.Result{
			{
				Query: "query",
			},
		}

		qc := NewQueryCache()

		{
			_, ok := qc.get(cacheKey1)
			assert.False(t, ok)
			_, ok = qc.get(cacheKey2)
			assert.False(t, ok)
		}

		{
			qc.set(cacheKey1, queryResult)
			cached, ok := qc.get(cacheKey1)
			assert.True(t, ok)
			assert.Equal(t, queryResult, cached)

			_, ok = qc.get(cacheKey2)
			assert.False(t, ok)
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		t.Parallel()

		cacheKey := queryCacheKey{
			compilerKey: "compilerKey",
			parsedInput: ast.String("input"),
			query:       "query",
		}
		queryResult := []result.Result{
			{
				Query: "query",
			},
		}

		qc := NewQueryCache()

		{
			_, ok := qc.get(cacheKey)
			assert.False(t, ok)
		}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				qc.set(cacheKey, queryResult)
			}()
		}
		wg.Wait()

		{
			cached, ok := qc.get(cacheKey)
			assert.True(t, ok)
			assert.Equal(t, queryResult, cached)
		}
	})
}
