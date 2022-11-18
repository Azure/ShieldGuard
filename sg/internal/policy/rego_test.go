package policy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewRegoCompiler_Integration(t *testing.T) {
	fixtures := []string{
		"./testdata/basic",
	}

	for _, f := range fixtures {
		t.Run(fmt.Sprintf("fixture: %s", f), func(t *testing.T) {
			packages, err := LoadPackagesFromPaths([]string{f})
			assert.NoError(t, err)

			compiler, err := NewRegoCompiler(packages)
			assert.NoError(t, err)

			assert.NotNil(t, compiler)
		})
	}
}
