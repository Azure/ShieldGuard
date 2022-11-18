package source

// SourceBuilder constructs a collection of source readers.
type SourceBuilder struct {
	paths []string
	err   error
}

// FromPath creates a SourceBuilder with loading sources from the given paths.
func FromPath(paths []string) *SourceBuilder {
	return &SourceBuilder{
		paths: paths,
	}
}

func (sb *SourceBuilder) Complete() ([]Source, error) {
	if sb.err != nil {
		return nil, sb.err
	}

	var rv []Source

	// load from paths
	{
		sources, err := loadSourceFromPaths(sb.paths)
		if err != nil {
			return nil, err
		}
		rv = append(rv, sources...)
	}

	return rv, nil
}
