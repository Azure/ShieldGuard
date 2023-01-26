package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ReadHelpers(t *testing.T) {
	cases := []struct {
		content      string
		validateSpec func(t *testing.T, spec Spec)
	}{
		// policies using string list
		{
			content: `
files:
  - name: foo
    paths:
      - ./foo
    policies:
      - ./foo2
      - ./foo1
    data:
      - ./foo
`,
			validateSpec: func(t *testing.T, spec Spec) {
				assert.Len(t, spec.Files, 1)
				fileTarget := spec.Files[0]
				assert.Equal(t, "foo", fileTarget.Name)
				assert.Equal(t, []string{"./foo"}, fileTarget.Paths)
				assert.Equal(t, []string{"./foo"}, fileTarget.Data)
				assert.Equal(
					t,
					[]string{"./foo1", "./foo2"},
					fileTarget.Policies.Values(),
				)
			},
		},
		// policies using string map
		{
			content: `
files:
  - name: foo
    paths:
      - ./foo
    policies:
      ? ./foo2
      ? ./foo1
    data:
      - ./foo
`,
			validateSpec: func(t *testing.T, spec Spec) {
				assert.Len(t, spec.Files, 1)
				fileTarget := spec.Files[0]
				assert.Equal(t, "foo", fileTarget.Name)
				assert.Equal(t, []string{"./foo"}, fileTarget.Paths)
				assert.Equal(t, []string{"./foo"}, fileTarget.Data)
				assert.Equal(
					t,
					[]string{"./foo1", "./foo2"},
					fileTarget.Policies.Values(),
				)
			},
		},
		// policies using string map with anchors
		{
			content: `
shared-policies: &shared-policies
  ? ./foo2
  ? ./foo3

another-shared-policies: &another-shared-policies
  ? ./foo4

files:
  - name: foo
    paths:
      - ./foo
    policies:
      <<: [*shared-policies, *another-shared-policies]
      ? ./foo1
    data:
      - ./foo
`,
			validateSpec: func(t *testing.T, spec Spec) {
				assert.Len(t, spec.Files, 1)
				fileTarget := spec.Files[0]
				assert.Equal(t, "foo", fileTarget.Name)
				assert.Equal(t, []string{"./foo"}, fileTarget.Paths)
				assert.Equal(t, []string{"./foo"}, fileTarget.Data)
				assert.Equal(
					t,
					[]string{"./foo1", "./foo2", "./foo3", "./foo4"},
					fileTarget.Policies.Values(),
				)
			},
		},
	}

	for idx := range cases {
		c := cases[idx]
		t.Run(fmt.Sprintf("[%d] MarshalYAML", idx), func(t *testing.T) {
			spec, err := ReadFromYAML(strings.NewReader(c.content))
			assert.NoError(t, err)

			c.validateSpec(t, spec)
		})

		t.Run(fmt.Sprintf("[%d] ReadFromFile", idx), func(t *testing.T) {
			tempDir := t.TempDir()
			specFile := filepath.Join(tempDir, SpecFileName)
			err := os.WriteFile(specFile, []byte(c.content), 0644)
			assert.NoError(t, err, "write config file")

			spec, err := ReadFromFile(specFile)
			assert.NoError(t, err)

			c.validateSpec(t, spec)
		})
	}

}
