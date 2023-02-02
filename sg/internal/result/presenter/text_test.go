package presenter

import (
	"bytes"
	"testing"

	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/Azure/ShieldGuard/sg/internal/source/testsource"
	"github.com/stretchr/testify/assert"
)

func Test_Text(t *testing.T) {
	presenter := Text([]result.QueryResults{
		{
			Source: &testsource.TestSource{NameFunc: func() string {
				return "file name"
			}},
			Successes:  2,
			Failures:   []result.Result{{Message: "fail message1"}, {Message: "fail message2"}},
			Warnings:   []result.Result{{Message: "warn message1"}, {Message: "warn message2"}},
			Exceptions: []result.Result{{Message: "exception message1"}},
		},
		{
			Source: &testsource.TestSource{NameFunc: func() string {
				return ""
			}},
			Successes:  0,
			Failures:   []result.Result{},
			Warnings:   []result.Result{},
			Exceptions: []result.Result{},
		},
	})
	output := new(bytes.Buffer)
	err := presenter.WriteQueryResultTo(output)
	assert.NoError(t, err)
	assert.Equal(t, output.String(), "FAIL - file name - main - fail message1\nFAIL - file name - main - fail message2\nWARN - file name - main - warn message1\nWARN - file name - main - warn message2\nEXCEPTION - file name - main - exception message1\n7 tests, 2 passed, 2 failures 2 warnings, 1 exceptions\n")
}
