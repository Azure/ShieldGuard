package reader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"

	gatorreader "github.com/open-policy-agent/gatekeeper/v3/pkg/gator/reader"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type kustomizePatchedFileSystem struct {
	filesys.FileSystem
}

func (w *kustomizePatchedFileSystem) ReadFile(path string) ([]byte, error) {
	b, err := w.FileSystem.ReadFile(path)
	if err != nil {
		return nil, err
	}

	isLoadingKustomizeFile := false
	filename := filepath.Base(path)
	for _, n := range konfig.RecognizedKustomizationFileNames() {
		if filename == n {
			isLoadingKustomizeFile = true
			break
		}
	}

	if !isLoadingKustomizeFile {
		return b, nil
	}

	var kust types.Kustomization
	if err := kust.Unmarshal(b); err != nil {
		// the kustomize config is invalid
		// TODO: log the error
		return b, nil
	}

	hasOriginalAnnotations := false
	hasTransformerAnnotations := false
	for _, v := range kust.BuildMetadata {
		switch v {
		case types.OriginAnnotations:
			hasOriginalAnnotations = true
		case types.TransformerAnnotations:
			hasTransformerAnnotations = true
		}
	}
	if hasOriginalAnnotations && hasTransformerAnnotations {
		// no need to patch
		return b, nil
	}

	if !hasOriginalAnnotations {
		kust.BuildMetadata = append(kust.BuildMetadata, types.OriginAnnotations)
	}
	if !hasTransformerAnnotations {
		kust.BuildMetadata = append(kust.BuildMetadata, types.TransformerAnnotations)
	}

	patched, err := json.Marshal(kust)
	if err != nil {
		// failed to marshal the kustomize config
		return nil, fmt.Errorf("marshal patched kustomize coonfig: %w", err)
	}

	return patched, nil
}

func objectSourceFromKustomizeAnnotations(
	kustomizeDir string,
	annotations map[string]string,
) ObjectSource {
	const (
		// ref: sigs.k8s.io/kustomize/api/internal/utils/annotations.go
		originAnnotationKey = "config.kubernetes.io/origin"
	)

	rv := ObjectSource{
		SourceType: SourceTypeKustomize,
	}

	if annotations == nil {
		return rv
	}

	originValue, ok := annotations[originAnnotationKey]
	if !ok {
		return rv
	}

	var origin resource.Origin

	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(originValue)), 1024)
	if err := dec.Decode(&origin); err != nil {
		// TODO: log the error
		return rv
	}

	// TODO: handle the case where the origin is remote source
	rv.FilePath = filepath.Join(kustomizeDir, origin.Path)

	return rv
}

func readKustomize(kustomizeDir string) (*TestTargets, error) {
	opts := krusty.MakeDefaultOptions()
	k := krusty.MakeKustomizer(opts)

	fs := &kustomizePatchedFileSystem{filesys.MakeFsOnDisk()}

	resMap, err := k.Run(fs, kustomizeDir)
	if err != nil {
		return nil, err
	}

	rv := &TestTargets{
		ObjectSources: make(map[*unstructured.Unstructured]ObjectSource),
	}
	for _, resource := range resMap.Resources() {
		// TODO: avoid the double marshalling
		resourceYAML, err := resource.AsYAML()
		if err != nil {
			return nil, fmt.Errorf("marshal resource to YAML: %w", err)
		}

		us, err := gatorreader.ReadK8sResources(bytes.NewReader(resourceYAML))
		if err != nil {
			return nil, fmt.Errorf("read resource from YAML: %w", err)
		}

		rv.Objects = append(rv.Objects, us...)
		for _, obj := range us {
			annotations := obj.GetAnnotations()
			rv.ObjectSources[obj] = objectSourceFromKustomizeAnnotations(kustomizeDir, annotations)
		}
	}

	return rv, nil
}

func readKustomizes(kustomizeDirs []string) (*TestTargets, error) {
	var rv *TestTargets
	for _, kustomizeDir := range kustomizeDirs {
		p, err := readKustomize(kustomizeDir)
		if err != nil {
			return nil, err
		}

		rv = rv.merge(p)
	}

	return rv, nil
}
