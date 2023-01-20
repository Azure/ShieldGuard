package project

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// ReadFromYAML reads a project specification from YAML.
func ReadFromYAML(src io.Reader) (Spec, error) {
	rv := Spec{}

	dec := yaml.NewDecoder(src)
	if err := dec.Decode(&rv); err != nil {
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
