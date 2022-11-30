package source

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

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

func relativeToContextRootFn(contextRoot string) func(string) string {
	if contextRoot == "" {
		return func(path string) string {
			return path
		}
	}

	return func(s string) string {
		rel, err := filepath.Rel(contextRoot, s)
		if err != nil {
			// NOTE: we don't expect this as we guard this in loadSourceFromPaths.
			panic(fmt.Sprintf("path %q is not relative to context root %q", s, contextRoot))
		}
		return rel
	}
}

// ref: https://github.com/open-policy-agent/conftest/blob/f18b7bbde2fdbd766c8348dff3a0a24792eb98c7/runner/test.go#L99
func loadSourceFromPaths(contextRoot string, paths []string) ([]Source, error) {
	// when contextRoot specified, all paths must be relative to contextRoot.
	// FIXME(hbc): this implementation may not be correct in Windows (see context in `filepath.HasPrefix`)
	//             We should revisit this in later changes.
	if contextRoot != "" {
		contextRootAbs, err := filepath.Abs(contextRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path of context root %q: %w", contextRoot, err)
		}

		for _, p := range paths {
			pAbs, err := filepath.Abs(p)
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute path of %q: %w", p, err)
			}

			if !strings.HasPrefix(pAbs, contextRootAbs) {
				return nil, fmt.Errorf("path %q is not relative to context root %q", p, contextRoot)
			}
		}
	}
	relativeToContextRoot := relativeToContextRootFn(contextRoot)

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
	filePathsSorted := make([]string, 0, len(configurations))
	for filePath := range configurations {
		filePathsSorted = append(filePathsSorted, filePath)
	}
	sort.Strings(filePathsSorted)

	var rv []Source
	for _, filePath := range filePathsSorted {
		c := configurations[filePath]
		var subConfigurations []interface{}
		if cc, ok := c.([]interface{}); ok {
			subConfigurations = cc
		} else {
			subConfigurations = []interface{}{c}
		}

		rv = append(rv, &fsSource{
			filePath:       relativeToContextRoot(filePath),
			configurations: subConfigurations,
		})
	}

	return rv, nil
}
