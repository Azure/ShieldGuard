package project

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// ReadFromYAML reads a project specification from YAML.
func ReadFromYAML(src io.Reader) (Spec, error) {
	// NOTE: since we want to support YAML anchors and mixing string list and string map,
	//       therefore, we firstly use yaml.Decoder to resolve the anchors and encode
	//       the spec to a untyped object. Then, we use json.Unmarshal to decode the resolved
	//       values. Finally, the mixed string list and string map will be decoded via strListOrMap.

	var untypedObj any
	yamlDecoder := yaml.NewDecoder(src)
	if err := yamlDecoder.Decode(&untypedObj); err != nil {
		return Spec{}, fmt.Errorf("decode yaml: %w", err)
	}

	resolvedJSON, err := json.Marshal(untypedObj)
	if err != nil {
		return Spec{}, fmt.Errorf("resolve to json: %w", err)
	}

	var rv Spec
	if err := json.Unmarshal(resolvedJSON, &rv); err != nil {
		return Spec{}, fmt.Errorf("decode yaml: %w", err)
	}

	return rv, nil
}

// ReadFromFile reads a project specification from a file.
func ReadFromFile(p string) (Spec, error) {
	f, err := os.Open(p)
	if err != nil {
		return Spec{}, fmt.Errorf("read file %q: %w", p, err)
	}
	defer f.Close()

	return ReadFromYAML(f)
}
