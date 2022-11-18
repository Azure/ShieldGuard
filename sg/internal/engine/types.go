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
	Query(
		ctx context.Context,
		source source.Source,
		opts ...*QueryOptions,
	) (*result.QueryResults, error)
}
