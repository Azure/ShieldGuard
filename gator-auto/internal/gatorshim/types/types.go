package types

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type TestTargets struct {
	Objects       []*unstructured.Unstructured
	ObjectSources map[*unstructured.Unstructured]string // TODO: expand to struct (e.g. stdin / file / etc.)
}
