package source

import "path/filepath"

// SourceBuilder constructs a collection of source readers.
type SourceBuilder struct {
	paths       []string
	contextRoot string
	err         error
}

// FromPath creates a SourceBuilder with loading sources from the given paths.
func FromPath(paths []string) *SourceBuilder {
	absolutePaths := make([]string, len(paths))
	for idx, path := range paths {
		p, err := filepath.Abs(path)
		if err != nil {
			return &SourceBuilder{err: err}
		}
		absolutePaths[idx] = p
	}

	return &SourceBuilder{
		paths: absolutePaths,
	}
}

// ContextRoot binds the context root to the SourceBuilder.
// When context root is specified, all paths must be relative to context root,
// and source names will be relative to context root.
func (sb *SourceBuilder) ContextRoot(contextRoot string) *SourceBuilder {
	if sb.err != nil {
		return sb
	}

	p, err := filepath.Abs(contextRoot)
	if err != nil {
		sb.err = err
		return sb
	}

	sb.contextRoot = p
	return sb
}

func (sb *SourceBuilder) Complete() ([]Source, error) {
	if sb.err != nil {
		return nil, sb.err
	}

	var rv []Source

	// load from paths
	{
		sources, err := loadSourceFromPaths(sb.contextRoot, sb.paths)
		if err != nil {
			return nil, err
		}
		rv = append(rv, sources...)
	}

	return rv, nil
}
