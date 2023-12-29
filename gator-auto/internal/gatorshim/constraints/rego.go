package constraints

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/loader"
	"github.com/stoewer/go-strcase"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ConstraintTargets struct {
	Constraints         []*unstructured.Unstructured
	ConstraintTemplates []*unstructured.Unstructured
}

var dns1035LabelPrefixRegex = regexp.MustCompile("^[a-z]")

// package -> Constraint
func regoFileToConstraint(
	regoFile *loader.RegoFile,
) (constraintTemplate *unstructured.Unstructured, constraint *unstructured.Unstructured) {
	enforcementAction := "deny"
	for _, annotation := range regoFile.Parsed.Annotations {
		if annotation.Scope != "package" {
			continue
		}

		for k, v := range annotation.Custom {
			if k == "enforcementAction" {
				vv := fmt.Sprint(v)
				if vv == "dryrun" || vv == "deny" || vv == "warn" {
					enforcementAction = vv
				}
			}
		}
	}

	// TODO: refine logic
	constraintName := string(regoFile.Parsed.Package.Path[1].Value.(ast.String))
	if !dns1035LabelPrefixRegex.MatchString(constraintName) {
		constraintName = "gator-auto-" + constraintName
	}
	constraintTemplateName := constraintName
	constraintTemplateKind := strcase.UpperCamelCase(constraintName)

	content := string(regoFile.Raw)

	constraintTemplate = &unstructured.Unstructured{}
	constraintTemplate.SetAPIVersion("templates.gatekeeper.sh/v1")
	constraintTemplate.SetKind("ConstraintTemplate")
	constraintTemplate.SetName(constraintTemplateName)
	constraintTemplate.Object["spec"] = map[string]interface{}{
		"crd": map[string]interface{}{
			"spec": map[string]interface{}{
				"names": map[string]interface{}{
					"kind": constraintTemplateKind,
				},
				"validation": map[string]interface{}{
					"openAPIV3Schema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		},
		"targets": []interface{}{
			map[string]interface{}{
				"target": "admission.k8s.gatekeeper.sh",
				"rego":   content,
				// TODO: support libs
			},
		},
	}

	constraint = &unstructured.Unstructured{}
	constraint.SetAPIVersion("constraints.gatekeeper.sh/v1beta1")
	constraint.SetKind(constraintTemplateKind)
	constraint.SetName(constraintName)
	constraint.Object["spec"] = map[string]interface{}{
		"match": map[string]interface{}{
			"kinds": []interface{}{
				map[string]interface{}{
					"apiGroups": []interface{}{"*"},
					"kinds":     []interface{}{"*"},
				},
			},
		},
		"parameters":        map[string]interface{}{},
		"enforcementAction": enforcementAction,
	}

	return constraintTemplate, constraint
}

type LoadParams struct {
	// RegoPaths is a list of paths to rego files.
	RegoPaths []string
}

func Load(
	ctx context.Context,
	params LoadParams,
) (*ConstraintTargets, error) {
	policies, err := loader.NewFileLoader().WithProcessAnnotation(true).
		Filtered(params.RegoPaths, func(abspath string, info fs.FileInfo, depth int) bool {
			return !info.IsDir() && !strings.HasSuffix(info.Name(), bundle.RegoExt)
		})
	if err != nil {
		return nil, err
	}

	rv := &ConstraintTargets{}
	for _, regoFile := range policies.Modules {
		constraintTemplate, constraint := regoFileToConstraint(regoFile)
		rv.ConstraintTemplates = append(rv.ConstraintTemplates, constraintTemplate)
		rv.Constraints = append(rv.Constraints, constraint)
	}

	return rv, nil
}
