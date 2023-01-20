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

	t.Run("check rules", func(t *testing.T) {
		checkers := map[string]ruleCheckFunc{
			"deny_foo": func(t *testing.T, r Rule) {
				assert.Equal(t, "foo", r.Name)
				assert.Equal(t, "main", r.Namespace)
				assert.Equal(t, QueryKindDeny, r.Kind)
				assert.NotNil(t, r.SourceLocation)
			},
			"warn_foo": func(t *testing.T, r Rule) {
				assert.Equal(t, "foo", r.Name)
				assert.Equal(t, "main", r.Namespace)
				assert.Equal(t, QueryKindWarn, r.Kind)
				assert.NotNil(t, r.SourceLocation)
			},
			"deny_no_bar": func(t *testing.T, r Rule) {
				assert.Equal(t, "no_bar", r.Name)
				assert.Equal(t, "main", r.Namespace)
				assert.Equal(t, QueryKindDeny, r.Kind)
				assert.NotNil(t, r.SourceLocation)
			},
			"violation_no_baz": func(t *testing.T, r Rule) {
				assert.Equal(t, "no_baz", r.Name)
				assert.Equal(t, "main", r.Namespace)
				assert.Equal(t, QueryKindViolation, r.Kind)
				assert.NotNil(t, r.SourceLocation)
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
	})

	t.Run("check package spec", func(t *testing.T) {
		spec := pkg.Spec()
		assert.NotNil(t, spec.Rule)
		assert.Equal(t, "https://example.com/docs/{{.SourceFileName}}.md", spec.Rule.DocLink)
	})
}

func Test_LoadPackagesFromPaths_no_package_spec(t *testing.T) {
	pkgs, err := LoadPackagesFromPaths([]string{"./testdata/no-package-spec"})
	assert.NoError(t, err)
	assert.Len(t, pkgs, 1)

	pkg := pkgs[0]

	t.Run("check rules", func(t *testing.T) {
		checkers := map[string]ruleCheckFunc{
			"deny_foo": func(t *testing.T, r Rule) {
				assert.Equal(t, "foo", r.Name)
				assert.Equal(t, "main", r.Namespace)
				assert.Equal(t, QueryKindDeny, r.Kind)
				assert.NotNil(t, r.SourceLocation)
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
	})

	t.Run("check package spec", func(t *testing.T) {
		spec := pkg.Spec()
		assert.Equal(t, defaultPackageSpec(), spec)
	})
}

func Test_LoadPackages_multiple(t *testing.T) {
	pkgs, err := LoadPackagesFromPaths([]string{"./testdata/basic", "./testdata/no-package-spec"})
	assert.NoError(t, err)
	assert.Len(t, pkgs, 2)

	assert.NotEqual(t, defaultPackageSpec(), pkgs[0].Spec())
	assert.Equal(t, defaultPackageSpec(), pkgs[1].Spec())
}
