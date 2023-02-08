package presenter

import (
	"bytes"
	"testing"

	"github.com/b4fun/ci"
	"github.com/stretchr/testify/assert"
)

func Test_Text(t *testing.T) {
	t.Run("plain text", func(t *testing.T) {
		t.Setenv("CI_NAME", "CUSTOM")

		presenter := Text(testQueryResults())
		output := new(bytes.Buffer)
		err := presenter.WriteQueryResultTo(output)
		assert.NoError(t, err)
		t.Log("\n" + output.String())
		assert.Equal(
			t,
			`FAIL -  - (002-rule) fail message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
FAIL -  - (002-rule) fail message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
WARNING -  - (002-rule) warn message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
WARNING -  - (002-rule) warn message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
EXCEPTION -  - (003-rule)
7 test(s), 2 passed, 2 failure(s) 2 warning(s), 1 exception(s)
`,
			output.String(),
		)
	})

	t.Run("azure pipelines", func(t *testing.T) {
		t.Setenv("CI_NAME", ci.AzurePipelines)

		presenter := Text(testQueryResults())
		output := new(bytes.Buffer)
		err := presenter.WriteQueryResultTo(output)
		assert.NoError(t, err)
		t.Log("\n" + output.String())
		assert.Equal(
			t,
			`##vso[task.logissue type=error]FAIL -  - (002-rule) fail message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
##vso[task.logissue type=error]FAIL -  - (002-rule) fail message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
##vso[task.logissue type=warning]WARNING -  - (002-rule) warn message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
##vso[task.logissue type=warning]WARNING -  - (002-rule) warn message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
##[group]EXCEPTIONS (1)
EXCEPTION -  - (003-rule)
##[endgroup]
7 test(s), 2 passed, 2 failure(s) 2 warning(s), 1 exception(s)
`,
			output.String(),
		)
	})

	t.Run("github actions", func(t *testing.T) {
		t.Setenv("CI_NAME", ci.GithubActions)

		presenter := Text(testQueryResults())
		output := new(bytes.Buffer)
		err := presenter.WriteQueryResultTo(output)
		assert.NoError(t, err)
		t.Log("\n" + output.String())
		assert.Equal(
			t,
			`::error::FAIL -  - (002-rule) fail message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
::error::FAIL -  - (002-rule) fail message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
::warning::WARNING -  - (002-rule) warn message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
::warning::WARNING -  - (002-rule) warn message2
Document: https://github.com/Azure/ShieldGuard/docs/002-rego.md
::group::EXCEPTIONS (1)
EXCEPTION -  - (003-rule)
::endgroup::
7 test(s), 2 passed, 2 failure(s) 2 warning(s), 1 exception(s)
`,
			output.String(),
		)
	})
}
