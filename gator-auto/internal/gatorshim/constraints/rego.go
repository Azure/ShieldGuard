package constraints

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/open-policy-agent/opa/loader"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type ConstraintTargets struct {
	Constraints         []*unstructured.Unstructured
	ConstraintTemplates []*unstructured.Unstructured
}

var dns1035LabelPrefixRegex = regexp.MustCompile("^[a-z]")

func regoFileToConstraint(
	regoFile *loader.RegoFile,
) (constraintTemplate *unstructured.Unstructured, constraint *unstructured.Unstructured) {
	content := string(regoFile.Raw)
	regoFileName := filepath.Base(regoFile.Name)
	regoFileExt := filepath.Ext(regoFileName)
	regoFileNameNoExt := regoFileName[:len(regoFileName)-len(regoFileExt)]
	regoFileNameNormalized := strings.ToLower(regoFileNameNoExt)
	// TODO: update logic
	if !dns1035LabelPrefixRegex.MatchString(regoFileNameNormalized) {
		regoFileNameNormalized = "gator-auto-" + regoFileNameNormalized
	}

	constraintTemplate = &unstructured.Unstructured{}
	constraintTemplate.SetAPIVersion("templates.gatekeeper.sh/v1")
	constraintTemplate.SetKind("ConstraintTemplate")
	constraintTemplate.SetName(regoFileNameNormalized)
	constraintTemplate.Object["spec"] = map[string]interface{}{
		"crd": map[string]interface{}{
			"spec": map[string]interface{}{
				"names": map[string]interface{}{
					"kind": regoFileNameNormalized,
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
			},
		},
	}

	constraint = &unstructured.Unstructured{}
	constraint.SetAPIVersion("constraints.gatekeeper.sh/v1beta1")
	constraint.SetKind(regoFileNameNormalized)
	constraint.SetName(regoFileNameNormalized)
	constraint.Object["spec"] = map[string]interface{}{
		"match": map[string]interface{}{
			"kinds": []interface{}{
				map[string]interface{}{
					"apiGroups": []interface{}{"*"},
					"kinds":     []interface{}{"*"},
				},
			},
		},
		"parameters": map[string]interface{}{},
	}

	return constraintTemplate, constraint
}

func LoadGatorConstraints(
	ctx context.Context,
	paths []string,
) (*ConstraintTargets, error) {
	policies, err := loader.AllRegos(paths)
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
