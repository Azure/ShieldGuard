package policy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewRegoCompiler_Integration(t *testing.T) {
	fixtures := []struct {
		path                string
		expectedCompilerKey string
	}{
		{
			path:                "./testdata/basic",
			expectedCompilerKey: "16834659949892238337",
		},
	}

	for _, f := range fixtures {
		t.Run(fmt.Sprintf("fixture: %s", f.path), func(t *testing.T) {
			packages, err := LoadPackagesFromPaths([]string{f.path})
			assert.NoError(t, err)

			compiler, compilerKey, err := NewRegoCompiler(packages)
			assert.NoError(t, err)

			assert.NotNil(t, compiler)

			assert.Equal(t, f.expectedCompilerKey, compilerKey)
		})
	}
}
