package engine

import (
	"context"
	"testing"

	"github.com/Azure/ShieldGuard/sg/internal/policy"
	"github.com/Azure/ShieldGuard/sg/internal/source"
	"github.com/stretchr/testify/assert"
)

func Test_Integration_BrokenPolicy(t *testing.T) {
	t.Parallel()

	_, err := QueryWithPolicy([]string{
		"./testdata/broken-policy",
	}).Complete()
	assert.Error(t, err)
}

func Test_Integration_Basic(t *testing.T) {
	t.Parallel()

	queryer, err := QueryWithPolicy([]string{
		"./testdata/basic/policy",
	}).Complete()
	assert.NoError(t, err)
	assert.NotNil(t, queryer)

	sources, err := source.FromPath([]string{
		"./testdata/basic/configurations",
	}).Complete()
	assert.NoError(t, err)
	assert.Len(t, sources, 1)
	dataYAMLSource := sources[0]

	ctx := context.Background()

	queryResult, err := queryer.Query(ctx, dataYAMLSource)
	assert.NoError(t, err)
	assert.NotNil(t, queryResult)
	assert.Equal(t, queryResult.Source, dataYAMLSource)
	assert.Equal(t, queryResult.Successes, 2, "one document passes the test")

	assert.Len(t, queryResult.Exceptions, 2, "one document emits two exceptions")
	{
		denyExcResult := queryResult.Exceptions[0]
		assert.Equal(t, denyExcResult.Query, `data.main.exception[_][_] == "foo"`)
		assert.Equal(t, denyExcResult.Rule.Kind, policy.QueryKindDeny)
		assert.Equal(t, denyExcResult.Rule.Name, "foo")
		assert.Equal(t, denyExcResult.Rule.Namespace, "main")
		assert.Equal(t, denyExcResult.RuleDocLink, "https://example.com/foo-deny-001-foo")

		warnExcResult := queryResult.Exceptions[1]
		assert.Equal(t, warnExcResult.Query, `data.main.exception[_][_] == "foo"`)
		assert.Equal(t, warnExcResult.Rule.Kind, policy.QueryKindWarn)
		assert.Equal(t, warnExcResult.Rule.Name, "foo")
		assert.Equal(t, warnExcResult.Rule.Namespace, "main")
		assert.Equal(t, warnExcResult.RuleDocLink, "https://example.com/foo-warn-001-foo")
	}

	assert.Len(t, queryResult.Warnings, 1, "one document emits warning")
	{
		warnResult := queryResult.Warnings[0]
		assert.Equal(t, warnResult.Message, "name is foo")
		assert.Equal(t, warnResult.Query, "data.main.warn_foo")
		assert.Equal(t, warnResult.Rule.Kind, policy.QueryKindWarn)
		assert.Equal(t, warnResult.Rule.Name, "foo")
		assert.Equal(t, warnResult.Rule.Namespace, "main")
		assert.Equal(t, warnResult.RuleDocLink, "https://example.com/foo-warn-001-foo")
	}

	assert.Len(t, queryResult.Failures, 1, "one document fails the test")
	{
		failureResult := queryResult.Failures[0]
		assert.Equal(t, failureResult.Message, "name cannot be foo")
		assert.Equal(t, failureResult.Query, "data.main.deny_foo")
		assert.Equal(t, failureResult.Rule.Kind, policy.QueryKindDeny)
		assert.Equal(t, failureResult.Rule.Name, "foo")
		assert.Equal(t, failureResult.Rule.Namespace, "main")
		assert.Equal(t, failureResult.RuleDocLink, "https://example.com/foo-deny-001-foo")
	}
}

func Test_Integration_Basic_QueryCacheEnabled(t *testing.T) {
	t.Parallel()

	queryCache := NewQueryCache()
	queryer, err := QueryWithPolicy([]string{
		"./testdata/basic/policy",
	}).
		WithQueueCache(queryCache).
		Complete()
	assert.NoError(t, err)
	assert.NotNil(t, queryer)

	sources, err := source.FromPath([]string{
		"./testdata/basic/configurations",
	}).Complete()
	assert.NoError(t, err)
	assert.Len(t, sources, 1)
	dataYAMLSource := sources[0]

	round := func() {
		ctx := context.Background()

		queryResult, err := queryer.Query(ctx, dataYAMLSource)
		assert.NoError(t, err)
		assert.NotNil(t, queryResult)
		assert.Equal(t, queryResult.Source, dataYAMLSource)
		assert.Equal(t, queryResult.Successes, 2, "one document passes the test")

		assert.Len(t, queryResult.Exceptions, 2, "one document emits two exceptions")
		{
			denyExcResult := queryResult.Exceptions[0]
			assert.Equal(t, denyExcResult.Query, `data.main.exception[_][_] == "foo"`)
			assert.Equal(t, denyExcResult.Rule.Kind, policy.QueryKindDeny)
			assert.Equal(t, denyExcResult.Rule.Name, "foo")
			assert.Equal(t, denyExcResult.Rule.Namespace, "main")
			assert.Equal(t, denyExcResult.RuleDocLink, "https://example.com/foo-deny-001-foo")

			warnExcResult := queryResult.Exceptions[1]
			assert.Equal(t, warnExcResult.Query, `data.main.exception[_][_] == "foo"`)
			assert.Equal(t, warnExcResult.Rule.Kind, policy.QueryKindWarn)
			assert.Equal(t, warnExcResult.Rule.Name, "foo")
			assert.Equal(t, warnExcResult.Rule.Namespace, "main")
			assert.Equal(t, warnExcResult.RuleDocLink, "https://example.com/foo-warn-001-foo")
		}

		assert.Len(t, queryResult.Warnings, 1, "one document emits warning")
		{
			warnResult := queryResult.Warnings[0]
			assert.Equal(t, warnResult.Message, "name is foo")
			assert.Equal(t, warnResult.Query, "data.main.warn_foo")
			assert.Equal(t, warnResult.Rule.Kind, policy.QueryKindWarn)
			assert.Equal(t, warnResult.Rule.Name, "foo")
			assert.Equal(t, warnResult.Rule.Namespace, "main")
			assert.Equal(t, warnResult.RuleDocLink, "https://example.com/foo-warn-001-foo")
		}

		assert.Len(t, queryResult.Failures, 1, "one document fails the test")
		{
			failureResult := queryResult.Failures[0]
			assert.Equal(t, failureResult.Message, "name cannot be foo")
			assert.Equal(t, failureResult.Query, "data.main.deny_foo")
			assert.Equal(t, failureResult.Rule.Kind, policy.QueryKindDeny)
			assert.Equal(t, failureResult.Rule.Name, "foo")
			assert.Equal(t, failureResult.Rule.Namespace, "main")
			assert.Equal(t, failureResult.RuleDocLink, "https://example.com/foo-deny-001-foo")
		}
	}

	for i := 0; i < 10; i++ {
		round()
	}
}
