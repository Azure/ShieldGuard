package policy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

var queryRegex *regexp.Regexp

func init() {
	re := fmt.Sprintf("^(%s)(_[a-zA-Z0-9_]+)*$", strings.Join([]string{
		string(QueryKindWarn),
		string(QueryKindDeny),
		string(QueryKindViolation),
	}, "|"))

	queryRegex = regexp.MustCompile(re)
}

// Query creates the query string.
func (r Rule) Query() string {
	return fmt.Sprintf("%s_%s", r.Kind, r.Name)
}

// IsKind checks if the query string is of the specified kind.
func (r Rule) IsKind(kind QueryKind, others ...QueryKind) bool {
	for _, k := range append([]QueryKind{kind}, others...) {
		if r.Kind == k {
			return true
		}
	}
	return false
}

func loadRulesFromModule(module *ast.Module) []Rule {
	var rv []Rule

	moduleNamespace := strings.Replace(module.Package.Path.String(), "data.", "", 1)

	for _, regoRule := range module.Rules {
		ruleString := regoRule.Head.Name.String()
		ps := queryRegex.FindAllStringSubmatch(ruleString, -1)
		if len(ps) != 1 || len(ps[0]) != 3 {
			continue
		}
		parsed := ps[0]

		rule := Rule{
			Kind:           QueryKind(parsed[1]),
			Name:           strings.TrimPrefix(ruleString, parsed[1]+"_"),
			Namespace:      moduleNamespace,
			SourceLocation: regoRule.Location,
		}

		rv = append(rv, rule)
	}

	return rv
}
