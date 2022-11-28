package test

import "strings"

var testdataBasicJSONOutputGolden = strings.TrimSpace(`
[
  {
    "filename": "configurations/data.yaml",
    "namespace": "main",
    "success": 2,
    "failures": [
      {
        "query": "data.main.deny_foo",
        "rule": {
          "name": "foo"
        },
        "message": "name cannot be foo"
      }
    ],
    "warnings": [
      {
        "query": "data.main.warn_foo",
        "rule": {
          "name": "foo"
        },
        "message": "name is foo"
      }
    ],
    "exceptions": [
      {
        "query": "data.main.exception[_][_] == \"foo\"",
        "rule": {
          "name": "foo"
        },
        "message": ""
      },
      {
        "query": "data.main.exception[_][_] == \"foo\"",
        "rule": {
          "name": "foo"
        },
        "message": ""
      }
    ]
  }
]
`)
