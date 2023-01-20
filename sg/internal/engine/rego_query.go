package engine

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"

	"github.com/Azure/ShieldGuard/sg/internal/policy"
	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/Azure/ShieldGuard/sg/internal/source"
	"github.com/Azure/ShieldGuard/sg/internal/utils"
)

type loadedConfiguration struct {
	Name          string
	Configuration interface{}
}

func loadSource(source source.Source) ([]loadedConfiguration, error) {
	var rv []loadedConfiguration

	configurations, err := source.ParsedConfigurations()
	if err != nil {
		return nil, err
	}

	for _, configuration := range configurations {
		rv = append(rv, loadedConfiguration{
			Name:          source.Name(),
			Configuration: configuration,
		})
	}

	return rv, nil
}

// PackageMain is the name of the main package.
// To ease the usage, we will only use rules from main package.
const PackageMain = "main"

// RegoEngine is the OPA based query engine implementation.
type RegoEngine struct {
	policyPackages []policy.Package
	compiler       *ast.Compiler
}

var _ Queryer = (*RegoEngine)(nil)

func (engine *RegoEngine) Query(
	ctx context.Context,
	source source.Source,
	opts ...*QueryOptions,
) (result.QueryResults, error) {
	loadedConfigurations, err := loadSource(source)
	if err != nil {
		return result.QueryResults{}, fmt.Errorf("failed to load source: %w", err)
	}

	rv := result.QueryResults{
		Source: source,
	}
	for _, loadedConfiguration := range loadedConfigurations {
		for _, policyPackage := range engine.policyPackages {
			if err := engine.queryPackage(ctx, policyPackage, loadedConfiguration, &rv); err != nil {
				return result.QueryResults{}, err
			}
		}
	}

	return rv, nil
}

func (engine *RegoEngine) queryPackage(
	ctx context.Context,
	policyPackage policy.Package,
	loadedConfiguration loadedConfiguration,
	queryResult *result.QueryResults,
) error {
	for _, rule := range policyPackage.Rules() {
		if rule.Namespace != PackageMain {
			// we only care about rules in the main package
			continue
		}

		if !rule.IsKind(policy.QueryKindWarn, policy.QueryKindDeny, policy.QueryKindViolation) {
			// not a query rule
			continue
		}

		if err := engine.queryRule(ctx, rule, loadedConfiguration, queryResult); err != nil {
			return fmt.Errorf("failed to query rule: %w", err)
		}
	}

	return nil
}

func (engine *RegoEngine) queryRule(
	ctx context.Context,
	policyRule policy.Rule,
	loadedConfiguration loadedConfiguration,
	queryResult *result.QueryResults,
) error {
	// execute exception query
	exceptionQuery := fmt.Sprintf("data.%s.exception[_][_] == %q", PackageMain, policyRule.Name)
	exceptions, err := engine.executeOneQuery(ctx, loadedConfiguration.Configuration, exceptionQuery)
	if err != nil {
		return fmt.Errorf("failed to execute exception query (%q): %w", exceptionQuery, err)
	}
	exceptions = utils.Filter(exceptions, func(x result.Result) bool { return x.Passed() })

	// execute query
	// NOTE: even if the exception query returns true, we still execute the query
	query := fmt.Sprintf("data.%s.%s", PackageMain, policyRule.Query())
	results, err := engine.executeOneQuery(ctx, loadedConfiguration.Configuration, query)
	if err != nil {
		return fmt.Errorf("failed to execute query (%q): %w", query, err)
	}

	// excluded by at least one exception
	if len(exceptions) > 0 {
		for idx := range exceptions {
			exceptions[idx].Rule = policyRule
		}
		queryResult.Exceptions = append(queryResult.Exceptions, exceptions...)
		return nil
	}

	for _, result := range results {
		if result.Passed() {
			queryResult.Successes += 1
			continue
		}

		result.Rule = policyRule

		switch {
		case policyRule.IsKind(policy.QueryKindWarn):
			queryResult.Warnings = append(queryResult.Warnings, result)
		case policyRule.IsKind(policy.QueryKindViolation, policy.QueryKindDeny):
			queryResult.Failures = append(queryResult.Failures, result)
		}
	}

	return nil
}

func (engine *RegoEngine) createRegoInstance(
	input interface{},
	query string,
) *rego.Rego {
	opts := []func(*rego.Rego){
		rego.Input(input),
		rego.Query(query),
		rego.Compiler(engine.compiler),
	}

	return rego.New(opts...)
}

func (engine *RegoEngine) executeOneQuery(
	ctx context.Context,
	input interface{},
	query string,
) ([]result.Result, error) {
	regoInstance := engine.createRegoInstance(input, query)
	resultSet, err := regoInstance.Eval(ctx)
	if err != nil {
		return nil, err
	}

	var rv []result.Result
	for _, evalResult := range resultSet {
		for _, expression := range evalResult.Expressions {
			loadedResults, err := result.FromRegoExpression(query, expression)
			if err != nil {
				return nil, fmt.Errorf("failed to load result: %w", err)
			}
			rv = append(rv, loadedResults...)
		}
	}

	return rv, nil
}
