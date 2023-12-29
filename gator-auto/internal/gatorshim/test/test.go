package test

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/frameworks/constraint/pkg/apis"
	templatesv1 "github.com/open-policy-agent/frameworks/constraint/pkg/apis/templates/v1"
	constraintclient "github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/k8scel"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers/rego"
	gatortypes "github.com/open-policy-agent/frameworks/constraint/pkg/types"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/expansion"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/gator/expand"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/gator/reader"
	gatortest "github.com/open-policy-agent/gatekeeper/v3/pkg/gator/test"
	mutationtypes "github.com/open-policy-agent/gatekeeper/v3/pkg/mutation/types"
	"github.com/open-policy-agent/gatekeeper/v3/pkg/target"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/Azure/ShieldGuard/gator-auto/internal/gatorshim/types"
)

type Opts = gatortest.Opts

var scheme *runtime.Scheme

func init() {
	scheme = runtime.NewScheme()
	err := apis.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
}

func Test(
	ctx context.Context,
	targets *types.TestTargets,
	tOpts Opts,
) (*gatortest.GatorResponses, error) {
	args := []constraintclient.Opt{constraintclient.Targets(&target.K8sValidationTarget{})}
	k8sDriver, err := k8scel.New()
	if err != nil {
		return nil, fmt.Errorf("creating K8s native driver: %w", err)
	}
	args = append(args, constraintclient.Driver(k8sDriver))

	driver, err := makeRegoDriver(tOpts)
	if err != nil {
		return nil, fmt.Errorf("creating Rego driver: %w", err)
	}
	args = append(args, constraintclient.Driver(driver))

	client, err := constraintclient.NewClient(args...)
	if err != nil {
		return nil, fmt.Errorf("creating OPA client: %w", err)
	}

	// search for templates, add them if they exist
	for _, obj := range targets.Objects {
		if !isTemplate(obj) {
			continue
		}

		templ, err := reader.ToTemplate(scheme, obj)
		if err != nil {
			return nil, fmt.Errorf("converting unstructured %q to template: %w", obj.GetName(), err)
		}

		_, err = client.AddTemplate(ctx, templ)
		if err != nil {
			return nil, fmt.Errorf("adding template %q: %w", templ.GetName(), err)
		}
	}

	// add all constraints.  A constraint must be added after its associated
	// template or OPA will return an error
	for _, obj := range targets.Objects {
		if !isConstraint(obj) {
			continue
		}

		_, err := client.AddConstraint(ctx, obj)
		if err != nil {
			return nil, fmt.Errorf("adding constraint %q: %w", obj.GetName(), err)
		}
	}

	// finally, add all the data.
	for _, obj := range targets.Objects {
		_, err := client.AddData(ctx, obj)
		if err != nil {
			return nil, fmt.Errorf("adding data of GVK %q: %w", obj.GroupVersionKind().String(), err)
		}
	}

	// create the expander
	er, err := expand.NewExpander(targets.Objects)
	if err != nil {
		return nil, fmt.Errorf("error creating expander: %w", err)
	}

	// now audit all objects
	responses := &gatortest.GatorResponses{
		ByTarget: make(map[string]*gatortest.GatorResponse),
	}
	for _, obj := range targets.Objects {
		// Try to attach the namespace if it was supplied (ns will be nil otherwise)
		ns, _ := er.NamespaceForResource(obj)
		au := target.AugmentedUnstructured{
			Object:    *obj,
			Namespace: ns,
			Source:    mutationtypes.SourceTypeOriginal,
		}

		review, err := client.Review(ctx, au)
		if err != nil {
			return nil, fmt.Errorf("reviewing %v %s/%s: %w",
				obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName(), err)
		}

		// Attempt to expand the obj and review resultant resources (if any)
		resultants, err := er.Expand(obj)
		if err != nil {
			return nil, fmt.Errorf("expanding resource %s: %w", obj.GetName(), err)
		}
		for _, resultant := range resultants {
			au := target.AugmentedUnstructured{
				Object:    *resultant.Obj,
				Namespace: ns,
				Source:    mutationtypes.SourceTypeGenerated,
			}
			resultantReview, err := client.Review(ctx, au)
			if err != nil {
				return nil, fmt.Errorf("reviewing expanded resource %v %s/%s: %w",
					resultant.Obj.GroupVersionKind(), resultant.Obj.GetNamespace(), resultant.Obj.GetName(), err)
			}
			expansion.OverrideEnforcementAction(resultant.EnforcementAction, resultantReview)
			expansion.AggregateResponses(resultant.TemplateName, review, resultantReview)
			expansion.AggregateStats(resultant.TemplateName, review, resultantReview)
		}

		for targetName, r := range review.ByTarget {
			targetResponse := responses.ByTarget[targetName]
			if targetResponse == nil {
				targetResponse = &gatortest.GatorResponse{}
				targetResponse.Target = targetName
			}

			// convert framework results to gator results, which contain a
			// reference to the violating resource
			gResults := make([]*gatortest.GatorResult, len(r.Results))
			for i, r := range r.Results {
				gResults[i] = fromFrameworkResult(r, obj)
			}
			targetResponse.Results = append(targetResponse.Results, gResults...)

			if r.Trace != nil {
				var trace string
				if targetResponse.Trace != nil {
					trace = *targetResponse.Trace
				}

				trace = trace + "\n\n" + *r.Trace
				targetResponse.Trace = &trace
			}
			responses.ByTarget[targetName] = targetResponse
		}

		responses.StatsEntries = append(responses.StatsEntries, review.StatsEntries...)
	}

	return responses, nil
}

func isTemplate(u *unstructured.Unstructured) bool {
	gvk := u.GroupVersionKind()
	return gvk.Group == templatesv1.SchemeGroupVersion.Group && gvk.Kind == "ConstraintTemplate"
}

func isConstraint(u *unstructured.Unstructured) bool {
	gvk := u.GroupVersionKind()
	return gvk.Group == "constraints.gatekeeper.sh"
}

func makeRegoDriver(tOpts Opts) (*rego.Driver, error) {
	var args []rego.Arg
	if tOpts.GatherStats {
		args = append(args, rego.GatherStats())
	}
	if tOpts.IncludeTrace {
		args = append(args, rego.Tracing(tOpts.IncludeTrace))
	}

	return rego.New(args...)
}

func fromFrameworkResult(frameResult *gatortypes.Result, violatingObject *unstructured.Unstructured) *gatortest.GatorResult {
	gResult := &gatortest.GatorResult{Result: *frameResult}

	// do a deep copy to detach us from the Constraint Framework's references
	gResult.Constraint = frameResult.Constraint.DeepCopy()

	// set the violating object, which is no longer part of framework results
	gResult.ViolatingObject = violatingObject

	return gResult
}