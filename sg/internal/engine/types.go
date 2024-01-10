package engine

import (
	"context"

	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/Azure/ShieldGuard/sg/internal/source"
)

// QueryOptions controls the query behavior.
type QueryOptions struct {
}

// Queryer performs queries against a target.
type Queryer interface {
	// Query executes the query.
	// The query call is expected to concurrent safe.
	Query(
		ctx context.Context,
		source source.Source,
		opts ...*QueryOptions,
	) (result.QueryResults, error)
}

// Limiter limits the query concurrency.
type Limiter interface {
	// acquire acquires a resource. Caller must call release() when done.
	acquire() (release func())
}
