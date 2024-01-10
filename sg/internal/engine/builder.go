package engine

import (
	"fmt"

	"github.com/Azure/ShieldGuard/sg/internal/policy"
)

// QueryerBuilder constructs a Queryer.
type QueryerBuilder struct {
	packages []policy.Package
	err      error
}

// QueryWithPolicy creates a QueryerBuilder with loading packages from the given paths.
func QueryWithPolicy(policyPaths []string) *QueryerBuilder {
	qb := &QueryerBuilder{}

	qb.packages, qb.err = policy.LoadPackagesFromPaths(policyPaths)
	if qb.err != nil {
		return qb
	}

	return qb
}

// Complete constructs the Queryer.
func (qb *QueryerBuilder) Complete() (Queryer, error) {
	if qb.err != nil {
		return nil, qb.err
	}

	compiler, err := policy.NewRegoCompiler(qb.packages)
	if err != nil {
		return nil, fmt.Errorf("failed to create compiler from packages: %w", err)
	}

	rv := &RegoEngine{
		policyPackages: qb.packages,
		compiler:       compiler,
		// NOTE: we limit the actual query by CPU count as policy evaluation is CPU bounded.
		//       For input actions like reading policy files / source code, we allow them to run unbounded,
		//       as the actual limiting is done by this limiter.
		limiter: newCPUBoundedLimiter(),
	}
	return rv, nil
}
