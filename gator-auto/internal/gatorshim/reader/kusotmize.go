package reader

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type wrappedFS struct {
	filesys.FileSystem
}

func (w *wrappedFS) ReadFile(path string) ([]byte, error) {
	isLoadingKustomizeFile := false
	filename := filepath.Base(path)
	for _, n := range konfig.RecognizedKustomizationFileNames() {
		if filename == n {
			isLoadingKustomizeFile = true
			break
		}
	}

	b, err := w.FileSystem.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if isLoadingKustomizeFile {
		// TODO: patch the file
		var v types.Kustomization
		if err := v.Unmarshal(b); err != nil {
			// TODO: log the error
			fmt.Println("failed to unmarshal", err)
			return b, nil
		}

		hasOriginalAnnotations := false
		hasTransformerAnnotations := false
		for _, vv := range v.BuildMetadata {
			if vv == types.OriginAnnotations {
				hasOriginalAnnotations = true
			}
			if vv == types.TransformerAnnotations {
				hasTransformerAnnotations = true
			}
		}
		if !hasOriginalAnnotations {
			v.BuildMetadata = append(v.BuildMetadata, types.OriginAnnotations)
		}
		if !hasTransformerAnnotations {
			v.BuildMetadata = append(v.BuildMetadata, types.TransformerAnnotations)
		}
		if hasOriginalAnnotations && hasTransformerAnnotations {
			return b, nil
		}

		updatedContent, err := json.Marshal(v)
		if err != nil {
			// TODO: log the error
			fmt.Println("failed to marshal", err)
			return b, nil
		}

		fmt.Println("updated", string(updatedContent))
		b = updatedContent
	}

	return b, nil
}

func readKustomize(path string) error {
	opts := krusty.MakeDefaultOptions()
	k := krusty.MakeKustomizer(opts)

	fs := &wrappedFS{filesys.MakeFsOnDisk()}

	res, err := k.Run(fs, path)
	if err != nil {
		return err
	}

	res.Debug("test")

	return nil
}

func readKustomizes(paths []string) error {
	for _, p := range paths {
		readKustomize(p)
	}

	return nil
}
