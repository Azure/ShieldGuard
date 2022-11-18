package source

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SourceBuilder_FromPath(t *testing.T) {
	sources, err := FromPath([]string{"./testdata/sample"}).Complete()
	assert.NoError(t, err)
	assert.Len(t, sources, 2)
}
