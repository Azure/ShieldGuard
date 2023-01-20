package policy

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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
	//                        If the rule is not defined in a source file, this will be empty.
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

// ResolveRuleDocLink resolves the rule document link.
func ResolveRuleDocLink(spec PackageSpec, rule Rule) (string, error) {
	if spec.Rule == nil || spec.Rule.DocLink == "" {
		// not set
		return "", nil
	}

	tmpl, err := template.New("doc-link").Parse(spec.Rule.DocLink)
	if err != nil {
		return "", fmt.Errorf("parse %q as template: %w", spec.Rule.DocLink, err)
	}

	var b bytes.Buffer

	tmplPayload := map[string]interface{}{
		"Name":           rule.Name,
		"Kind":           rule.Kind,
		"SourceFileName": "",
	}
	if rule.SourceLocation != nil {
		f := filepath.Base(rule.SourceLocation.File)
		f = strings.TrimSuffix(f, filepath.Ext(f))
		tmplPayload["SourceFileName"] = f
	}

	if err := tmpl.Execute(&b, tmplPayload); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return b.String(), nil
}
