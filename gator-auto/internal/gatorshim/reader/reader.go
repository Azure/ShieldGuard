package reader

import "context"

type LoadParams struct {
	// FileSources - list of file paths to read from
	FileSources []string
	// KustomizeSources - list of kustomize paths to read from
	KustomizeSources []string
}

func Load(ctx context.Context, params LoadParams) (*TestTargets, error) {
	var rv *TestTargets

	if len(params.FileSources) > 0 {
		targets, err := readFiles(params.FileSources)
		if err != nil {
			return nil, err
		}
		rv = rv.merge(targets)
	}

	if len(params.KustomizeSources) > 0 {
		targets, err := readKustomizes(params.KustomizeSources)
		if err != nil {
			return nil, err
		}
		rv = rv.merge(targets)
	}

	return rv, nil
}
