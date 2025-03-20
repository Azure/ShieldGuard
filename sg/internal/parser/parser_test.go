package parser

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/stretchr/testify/assert"
)

func Test_ParseArmTemplateDefaults(t *testing.T) {
	t.Parallel()

	jsonStr := `{
		"parameters": {
			"myParam": {
				"type": "bool",
				"defaultValue": true
			}
		},
		"resources": [
			{
				"name": "MyResource",
				"properties": {
					"mustBeTrue": "[parameters('myParam')]"
				}
			}
		]
	}`

	term := jsonToTerm(t, jsonStr)

	// parse defaults
	ParseArmTemplateDefaults(term)

	// should not contain param after parsing
	assert.False(t, strings.Contains(term.Value.String(), "[parameters('myParam')]"))

}

// helper function to convert string to *ast.Term
func jsonToTerm(t *testing.T, jsonStr string) *ast.Term {
	t.Helper()

	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		t.Errorf("failed to unmarshal JSON: %s", err)
	}
	val, err := ast.InterfaceToValue(obj)
	if err != nil {
		t.Errorf("failed to convert interface to OPA value: %s", err)
	}

	return ast.NewTerm(val)
}
