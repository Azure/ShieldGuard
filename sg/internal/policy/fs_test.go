package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type ruleCheckFunc func(t *testing.T, r Rule)

func Test_LoadPackagesFromPaths_empty(t *testing.T) {
	_, err := LoadPackagesFromPaths([]string{"./testdata/empty"})
	assert.Error(t, err, "should return error when no policies found")
}

func Test_LoadPackagesFromPaths_basic(t *testing.T) {
	pkgs, err := LoadPackagesFromPaths([]string{"./testdata/basic"})
	assert.NoError(t, err)
	assert.Len(t, pkgs, 1)

	pkg := pkgs[0]

	checkers := map[string]ruleCheckFunc{
		"deny_foo": func(t *testing.T, r Rule) {
			assert.Equal(t, "foo", r.Name)
			assert.Equal(t, "main", r.Namespace)
			assert.Equal(t, QueryKindDeny, r.Kind)
			assert.NotNil(t, r.SourceLocation)
			assert.NotNil(t, r.DocLink)
		},
		"warn_foo": func(t *testing.T, r Rule) {
			assert.Equal(t, "foo", r.Name)
			assert.Equal(t, "main", r.Namespace)
			assert.Equal(t, QueryKindWarn, r.Kind)
			assert.NotNil(t, r.SourceLocation)
			assert.NotNil(t, r.DocLink)
		},
		"deny_no_bar": func(t *testing.T, r Rule) {
			assert.Equal(t, "no_bar", r.Name)
			assert.Equal(t, "main", r.Namespace)
			assert.Equal(t, QueryKindDeny, r.Kind)
			assert.NotNil(t, r.SourceLocation)
			assert.NotNil(t, r.DocLink)
		},
		"violation_no_baz": func(t *testing.T, r Rule) {
			assert.Equal(t, "no_baz", r.Name)
			assert.Equal(t, "main", r.Namespace)
			assert.Equal(t, QueryKindViolation, r.Kind)
			assert.NotNil(t, r.SourceLocation)
			assert.Nil(t, r.DocLink, "no doc for no_baz rule")
		},
	}

	parsedModules := pkg.ParsedModules()
	assert.Len(t, parsedModules, len(checkers))

	rules := pkg.Rules()
	assert.Len(t, rules, len(checkers))

	for _, rule := range rules {
		assert.Contains(t, checkers, rule.Query())
		assert.Contains(t, parsedModules, rule.SourceLocation.File)

		checkers[rule.Query()](t, rule)
	}
}
