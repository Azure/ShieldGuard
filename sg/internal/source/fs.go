package source

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/util"
)

type fsSource struct {
	// filePath is the full path of the read file.
	filePath string
	// configurations is the loaded configurations.
	configurations []ast.Value
}

var _ Source = (*fsSource)(nil)

func (s *fsSource) Name() string {
	return s.filePath
}

func (s *fsSource) ParsedConfigurations() ([]ast.Value, error) {
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

// ref: https://github.com/open-policy-agent/opa/blob/af8f915846fa325fe009dd6c226122bb205cff0a/rego/rego.go#L1937-L1956
func parseRawConfiguration(s any) (ast.Value, error) {
	rawPtr := util.Reference(s)
	if err := util.RoundTrip(rawPtr); err != nil {
		return nil, fmt.Errorf("convert raw configuration to JSON: %w", err)
	}
	rv, err := ast.InterfaceToValue(rawPtr)
	if err != nil {
		return nil, fmt.Errorf("convert raw configuration to OPA value: %w", err)
	}
	return rv, nil
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
		var subConfigurations []any
		if cc, ok := c.([]any); ok {
			subConfigurations = cc
		} else {
			subConfigurations = []any{c}
		}

		var parsedConfigurations []ast.Value
		for _, rawConfiguration := range subConfigurations {
			parsedConfiguration, err := parseRawConfiguration(rawConfiguration)
			if err != nil {
				return nil, fmt.Errorf("parse raw configuration: %w", err)
			}
			parsedConfigurations = append(parsedConfigurations, parsedConfiguration)
		}

		rv = append(rv, &fsSource{
			filePath:       relativeToContextRoot(filePath),
			configurations: parsedConfigurations,
		})
	}

	return rv, nil
}
