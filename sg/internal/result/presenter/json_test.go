package presenter

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_JSON(t *testing.T) {
	presenter := JSON(testQueryResults())
	output := new(bytes.Buffer)
	err := presenter.WriteQueryResultTo(output)
	assert.NoError(t, err)
	t.Log("\n" + output.String())
	assert.JSONEq(
		t,
		`[
			{
			  "filename": "file name",
			  "namespace": "main",
			  "success": 2,
			  "failures": [
				{
				  "query": "",
				  "rule": {
					"name": "001-rule",
					"doc_link": "https://github.com/Azure/ShieldGuard/docs/001-rego.md"
				  },
				  "message": "fail message1"
				},
				{
				  "query": "",
				  "rule": {
					"name": "002-rule",
					"doc_link": "https://github.com/Azure/ShieldGuard/docs/002-rego.md"
				  },
				  "message": "fail message2"
				}
			  ],
			  "warnings": [
				{
				  "query": "",
				  "rule": {
					"name": "001-rule",
					"doc_link": "https://github.com/Azure/ShieldGuard/docs/001-rego.md"
				  },
				  "message": "warn message1"
				},
				{
				  "query": "",
				  "rule": {
					"name": "002-rule",
					"doc_link": "https://github.com/Azure/ShieldGuard/docs/002-rego.md"
				  },
				  "message": "warn message2"
				}
			  ],
			  "exceptions": [
				{
				  "query": "",
				  "rule": {
					"name": "003-rule",
					"doc_link": "https://github.com/Azure/ShieldGuard/docs/003-rego.md"
				  },
				  "message": ""
				}
			  ]
			},
			{
			  "filename": "",
			  "namespace": "main",
			  "success": 0,
			  "failures": [],
			  "warnings": [],
			  "exceptions": []
			}
		  ]`,
		output.String(),
	)
}
