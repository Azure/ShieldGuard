package main

import (
	"context"
	"fmt"

	"github.com/Azure/ShieldGuard/gator-auto/internal/gatorshim/constraints"
	"github.com/Azure/ShieldGuard/gator-auto/internal/gatorshim/reader"
	gatortest "github.com/open-policy-agent/gatekeeper/v3/pkg/gator/test"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type GatorTestParams struct {
	Sources  []string
	Policies []string
}

func (p *GatorTestParams) BindCLIFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&p.Sources, "filename", nil, "Paths to source file")
	fs.StringSliceVar(&p.Policies, "policy", nil, "Paths to rego policy file")
}

func gatorTest(
	ctx context.Context,
	params GatorTestParams,
) error {
	constraintTargets, err := constraints.LoadGatorConstraints(ctx, params.Policies)
	if err != nil {
		return err
	}

	targets, err := reader.ReadTargets(params.Sources)
	if err != nil {
		return err
	}

	var objects []*unstructured.Unstructured
	objects = append(objects, targets.Objects...)
	objects = append(objects, constraintTargets.Constraints...)
	objects = append(objects, constraintTargets.ConstraintTemplates...)

	for _, obj := range objects {
		fmt.Println("loaded object", obj.GroupVersionKind(), obj.GetName())
	}

	responses, err := gatortest.Test(
		objects,
		gatortest.Opts{
			IncludeTrace: true,
			GatherStats:  true,
		},
	)
	if err != nil {
		return err
	}

	fmt.Println(responses.ByTarget)
	fmt.Println(responses.StatsEntries)
	for _, result := range responses.Results() {
		fmt.Println(result)
		path := targets.ObjectSources[result.ViolatingObject]
		fmt.Println(path)
	}
	return nil
}

func createGatorTestCommand() *cobra.Command {
	params := GatorTestParams{}

	cmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := gatorTest(cmd.Context(), params); err != nil {
				return err
			}

			return nil
		},
	}

	params.BindCLIFlags(cmd.Flags())

	return cmd
}
