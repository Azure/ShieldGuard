package reader

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type SourceType string

const (
	SourceTypeFile      SourceType = "file"
	SourceTypeImage     SourceType = "image"
	SourceTypeStdin     SourceType = "stdin"
	SourceTypeKustomize SourceType = "kustomize"
)

type ObjectSource struct {
	SourceType SourceType
	FilePath   string
}

type TestTargets struct {
	Objects       []*unstructured.Unstructured
	ObjectSources map[*unstructured.Unstructured]ObjectSource
}

func (tt *TestTargets) merge(other *TestTargets) *TestTargets {
	if other == nil {
		return tt
	}
	if tt == nil {
		return other
	}

	rv := &TestTargets{
		Objects:       append(tt.Objects, other.Objects...),
		ObjectSources: make(map[*unstructured.Unstructured]ObjectSource),
	}
	for k, v := range tt.ObjectSources {
		rv.ObjectSources[k] = v
	}
	for k, v := range other.ObjectSources {
		rv.ObjectSources[k] = v
	}
	return rv
}
