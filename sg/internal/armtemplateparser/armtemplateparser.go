package armtemplateparser

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

type visitor struct {
	parentKey string
	defaults  map[ast.Value]ast.Value
}

func (v *visitor) Visit(x interface{}) ast.Visitor {
	n, ok := x.(*ast.Term)
	if ok {
		// add key:val to defaults mapping
		hasDefault := n.Get(ast.StringTerm("defaultValue"))
		if hasDefault != nil {
			k := fmt.Sprintf("[parameters(%s)]", v.parentKey)
			key := ast.StringTerm(k)
			v.defaults[key.Value] = hasDefault.Value
		} else {
			v.parentKey = strings.ReplaceAll(n.String(), "\"", "'")
		}

		// query defaults mapping
		if val, exists := v.defaults[n.Value]; exists {
			n.Value = val
		}
	}
	return v
}

// Replaces parameters with their default values in ARM templates
//
// It renders expressions like:
//
//	"[parameters('paramName')]" -> "defaultParamName"
//
// by substituting the parameter values from the provided default value.
func ParseArmTemplateDefaults(t *ast.Term) {
	ast.Walk(&visitor{defaults: map[ast.Value]ast.Value{}}, t)
}
