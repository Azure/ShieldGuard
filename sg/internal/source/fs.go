package source

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/open-policy-agent/conftest/parser"
)

type fsSource struct {
	// filePath is the full path of the read file.
	filePath string
	// configurations is the loaded configurations.
	configurations []interface{}
}

var _ Source = (*fsSource)(nil)

func (s *fsSource) Name() string {
	return s.filePath
}

func (s *fsSource) ParsedConfigurations() ([]interface{}, error) {
	return s.configurations, nil
}

// ref: https://github.com/open-policy-agent/conftest/blob/f18b7bbde2fdbd766c8348dff3a0a24792eb98c7/runner/test.go#L99
func loadSourceFromPaths(paths []string) ([]Source, error) {
	var files []string

	walk := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if parser.FileSupported(path) {
			files = append(files, path)
		}

		return nil
	}

	for _, path := range paths {
		if err := filepath.WalkDir(path, walk); err != nil {
			return nil, fmt.Errorf("walk path %q: %w", path, err)
		}
	}

	if len(files) < 1 {
		return nil, fmt.Errorf("no files found from given paths: %v", paths)
	}

	configurations, err := parser.ParseConfigurations(files)
	if err != nil {
		return nil, fmt.Errorf("parse configurations: %w", err)
	}

	var rv []Source
	for filePath, c := range configurations {
		var subConfigurations []interface{}
		if cc, ok := c.([]interface{}); ok {
			subConfigurations = cc
		} else {
			subConfigurations = []interface{}{c}
		}

		rv = append(rv, &fsSource{
			filePath:       filePath,
			configurations: subConfigurations,
		})
	}

	return rv, nil
}
