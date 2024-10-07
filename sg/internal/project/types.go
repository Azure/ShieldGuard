package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// Spec defines the project specification.
type Spec struct {
	Files []FileTargetSpec `json:"files"`
}

// FileTargetSpec defines the specification of a file target.
// Without further specification, paths are relative to the context root which is defined during execution.
//
// NOTE: we should always use json tag here, please see the comment in ReadFromYAML for context.
type FileTargetSpec struct {
	// Name - name of the target.
	Name string `json:"name"`
	// Paths - paths to the targets to check.
	Paths []string `json:"paths"`
	// Policies - paths to the policy to load.
	Policies strListOrMap `json:"policies"`
	// Data - paths to the (extra) data to load.
	Data         []string `json:"data"`
	SystemPrompt string   `json:"systemPrompt"`
}

// strListOrMap is a helper type to support specifying string value using list or map (keys).
// When marshaling back to JSON/YAML, it will always be marshaled as an ordered list.
//
// It is useful when we want to support YAML anchors. For example, in targets list, users can specify either:
//
// 1. a list of paths to policy;
// 2. a map with policy paths as key.
//
// The map usage is for defining and reusing via YAML anchors, which is useful for large mono-repo.
//
// ```yaml
//
//	shared-policies: &shared-policies
//	 ? policies/policy-a
//	 ? policies/policy-b
//
//	shared-policies-2: &shared-policies-2
//	 ? policies/policy-c
//	 ? policies/policy-d
//
//	files:
//	 - name: project-a
//	   policies:
//	    <<: *shared-policies
//	    ? policies/policy-for-project-a
//	 - name: project-b
//	   policies:
//	    <<: [*shared-policies, *shared-policies-2]
//	    ? policies/policy-for-project-b
//
// ```
type strListOrMap []string

var _ json.Unmarshaler = (*strListOrMap)(nil)

func (p *strListOrMap) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil {
		return err
	}
	v, ok := t.(json.Delim)
	if !ok {
		return fmt.Errorf("unsupported value, only list and map are supported")
	}

	switch v {
	case '[':
		var values []string
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
		for _, value := range values {
			*p = append(*p, value)
		}
	case '{':
		var values map[string]interface{}
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
		for k := range values {
			*p = append(*p, k)
		}
	}

	sort.Strings(*p)

	return nil
}

// Values returns the sorted values.
func (p strListOrMap) Values() []string {
	return p
}
