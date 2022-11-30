package policy

import (
	"fmt"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/stretchr/testify/assert"
)

func Test_ResolveRuleDocLink(t *testing.T) {
	makeRule := func(ms ...func(*Rule)) Rule {
		rv := Rule{
			Kind:      QueryKindDeny,
			Name:      "foo",
			Namespace: "main",
		}

		for _, m := range ms {
			m(&rv)
		}

		return rv
	}

	makePackageSpec := func(ms ...func(*PackageSpec)) PackageSpec {
		rv := defaultPackageSpec()

		for _, m := range ms {
			m(&rv)
		}

		return rv
	}

	cases := []struct {
		spec      PackageSpec
		rule      Rule
		expected  string
		expectErr bool
	}{
		// no rule doc link
		{
			spec:     makePackageSpec(),
			rule:     makeRule(),
			expected: "",
		},
		// invalid rule doc link format
		{
			spec: makePackageSpec(func(ps *PackageSpec) {
				ps.Rule = &RuleSpec{
					DocLink: "https://example.com/{{",
				}
			}),
			rule:      makeRule(),
			expectErr: true,
		},
		// no source location
		{
			spec: makePackageSpec(func(ps *PackageSpec) {
				ps.Rule = &RuleSpec{
					DocLink: "https://example.com/{{.Name}}/{{.Kind}}/{{.SourceFileName}}",
				}
			}),
			rule:     makeRule(),
			expected: "https://example.com/foo/deny/",
		},
		// with source location
		{
			spec: makePackageSpec(func(ps *PackageSpec) {
				ps.Rule = &RuleSpec{
					DocLink: "https://example.com/{{.Name}}/{{.Kind}}/{{.SourceFileName}}",
				}
			}),
			rule: makeRule(func(r *Rule) {
				r.SourceLocation = &ast.Location{
					File: "path/to/the/foo-rego-file.rego",
				}
			}),
			expected: "https://example.com/foo/deny/foo-rego-file",
		},
	}

	for idx := range cases {
		c := cases[idx]
		t.Run(fmt.Sprintf("case #%d", idx), func(t *testing.T) {
			actual, err := ResolveRuleDocLink(c.spec, c.rule)
			if c.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, actual)
			}
		})
	}
}
