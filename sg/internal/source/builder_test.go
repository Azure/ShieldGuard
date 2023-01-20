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

func Test_SourceBuilder_ContextRoot(t *testing.T) {
	t.Run("empty context root", func(t *testing.T) {
		sources, err := FromPath([]string{"./testdata/sample"}).ContextRoot("").Complete()
		assert.NoError(t, err)
		assert.Len(t, sources, 2)
	})

	t.Run("relative context root", func(t *testing.T) {
		sources, err := FromPath([]string{"./testdata/sample"}).ContextRoot("./testdata").Complete()
		assert.NoError(t, err)
		assert.Len(t, sources, 2)
		assert.Equal(t, "sample/deployment+service.yaml", sources[0].Name())
		assert.Equal(t, "sample/service.yaml", sources[1].Name())
	})

	t.Run("relative context root nested", func(t *testing.T) {
		sources, err := FromPath([]string{"./testdata/sample"}).ContextRoot("./testdata/sample").Complete()
		assert.NoError(t, err)
		assert.Len(t, sources, 2)
		assert.Equal(t, "deployment+service.yaml", sources[0].Name())
		assert.Equal(t, "service.yaml", sources[1].Name())
	})

	t.Run("non-relative context root", func(t *testing.T) {
		_, err := FromPath([]string{"./testdata/sample"}).ContextRoot("foobar").Complete()
		assert.Error(t, err)
	})
}
