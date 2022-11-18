package source

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_loadSourceFromPaths(t *testing.T) {
	t.Run("sample", func(t *testing.T) {
		sources, err := loadSourceFromPaths([]string{"./testdata/sample"})
		assert.NoError(t, err)

		checkers := map[string]func(source Source){
			"deployment+service.yaml": func(source Source) {
				configs, err := source.ParsedConfigurations()
				assert.NoError(t, err)
				assert.Len(t, configs, 2)
			},
			"service.yaml": func(source Source) {
				configs, err := source.ParsedConfigurations()
				assert.NoError(t, err)
				assert.Len(t, configs, 1)
			},
		}
		assert.Len(t, sources, len(checkers))

		for _, source := range sources {
			baseName := filepath.Base(source.Name())
			assert.Contains(t, checkers, baseName)
			checkers[baseName](source)
		}
	})
}
