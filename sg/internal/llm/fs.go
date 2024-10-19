package llm

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/conftest/parser"
)

type fsSource struct {
	// filePath is the full path of the read file.
	filePath string
	// content is the loaded content.
	content string
}

var _ Source = (*fsSource)(nil)

func (s *fsSource) Name() string {
	return s.filePath
}

func (s *fsSource) Content() (string, error) {
	return s.content, nil
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

	var rv []Source
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read file %q: %w", file, err)
		}
		rv = append(rv, &fsSource{
			filePath: relativeToContextRoot(file),
			content:  string(content),
		})
	}

	return rv, nil
}
