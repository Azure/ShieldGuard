package presenter

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_plainText(t *testing.T) {
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
}
