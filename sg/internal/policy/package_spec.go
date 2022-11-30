package policy

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// RuleSpec specifies the policy rule settings.
type RuleSpec struct {
	// DocLink specifies the policy rule document link format.
	//
	// The value will be formatted using text.Template. Following variables are available:
	//
	// - {{.Name}}: the name of the rule.
	// - {{.Kind}}: the kind of the rule. See `QueryKind` for available options.
	// - {{.SourceFileName}}: the source file name (without the .rego extension) of the rule.
	DocLink string `json:"doc_link,omitempty" yaml:"doc_link,omitempty"`
}

// PackageSpec specifies the package settings.
type PackageSpec struct {
	// Rule specifies the policy rule settings.
	Rule *RuleSpec `json:"rule,omitempty" yaml:"rule,omitempty"`
}

// rule:
//   doc_link: https://example.com/docs/{{.Kind}}/{{.SourceFileName}}.md

func defaultPackageSpec() PackageSpec {
	return PackageSpec{
		Rule: &RuleSpec{},
	}
}

// PackageSpecFileName is the default name of the package specification file.
const PackageSpecFileName = "sg-package.yaml"

func loadPackageSpecFromDir(dir string) (PackageSpec, error) {
	specFile := filepath.Join(dir, PackageSpecFileName)
	if b, err := os.ReadFile(specFile); err != nil {
		if os.IsNotExist(err) {
			return defaultPackageSpec(), nil
		}
		return PackageSpec{}, fmt.Errorf("failed to read package spec file: %w", err)
	} else {
		var spec PackageSpec
		if err := yaml.Unmarshal(b, &spec); err != nil {
			return PackageSpec{}, fmt.Errorf("failed to unmarshal package spec: %w", err)
		}
		return spec, nil
	}
}
