package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ReadHelpers(t *testing.T) {
	validConfigContent := `
files:
  - name: foo
    paths:
      - ./foo
    policies:
      - ./foo
    data:
      - ./foo
`

	expectIsValidConfigSpec := func(t *testing.T, spec Spec) {
		assert.Len(t, spec.Files, 1)
		fileTarget := spec.Files[0]
		assert.Equal(t, "foo", fileTarget.Name)
		assert.Equal(t, []string{"./foo"}, fileTarget.Paths)
		assert.Equal(t, []string{"./foo"}, fileTarget.Policies)
		assert.Equal(t, []string{"./foo"}, fileTarget.Data)
	}

	t.Run("ReadFromYAML", func(t *testing.T) {
		spec, err := ReadFromYAML(strings.NewReader(validConfigContent))
		assert.NoError(t, err)

		expectIsValidConfigSpec(t, spec)
	})

	t.Run("ReadFromFile", func(t *testing.T) {
		tempDir := t.TempDir()
		specFile := filepath.Join(tempDir, SpecFileName)
		err := os.WriteFile(specFile, []byte(validConfigContent), 0644)
		assert.NoError(t, err, "write config file")

		spec, err := ReadFromFile(specFile)
		assert.NoError(t, err)

		expectIsValidConfigSpec(t, spec)
	})
}
