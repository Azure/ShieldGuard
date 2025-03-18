package engine

import (
	"context"

	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/Azure/ShieldGuard/sg/internal/source"
	"github.com/open-policy-agent/opa/ast"
)

// QueryOptions controls the query behavior.
type QueryOptions struct {
	ParseArmTemplateDefaults bool
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

// limiter limits the query concurrency.
type limiter interface {
	// acquire acquires a resource. Caller must call release() when done.
	acquire() (release func())
}

// queryCacheKey is the key for the query cache.
type queryCacheKey struct {
	// compilerKey represents the configuration combinations of the compiler.
	// It should include the policy packages and the compiler options.
	compilerKey string
	// parsedInput is the parsed input in ast.Value representation.
	// The hash of the parsed input is used to identify the input.
	parsedInput ast.Value
	// query is the query string.
	query string
}

// QueryCache provides the internal query cache.
type QueryCache interface {
	set(key queryCacheKey, value []result.Result)
	get(key queryCacheKey) ([]result.Result, bool)
}
