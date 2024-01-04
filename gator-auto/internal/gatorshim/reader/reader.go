package reader

import "context"

type LoadParams struct {
	// FileSources - list of file paths to read from
	FileSources []string
	// KustomizeSources - list of kustomize paths to read from
	KustomizeSources []string

	// HelmCommand - helm command to run
	HelmCommand string
	// HelmSources - list of helm chart paths to read from
	HelmSources []string

	// TODO: support helm values overrides
	// HelmValues - values to use for rendering helm templates
	// HelmValues map[string]interface{}
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

	if len(params.HelmSources) > 0 {
		targets, err := readHelmCharts(params.HelmCommand, params.HelmSources)
		if err != nil {
			return nil, err
		}
		rv = rv.merge(targets)
	}

	return rv, nil
}
