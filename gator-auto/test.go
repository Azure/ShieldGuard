package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Azure/ShieldGuard/gator-auto/internal/gatorshim/constraints"
	"github.com/Azure/ShieldGuard/gator-auto/internal/gatorshim/reader"
	gatortest "github.com/open-policy-agent/gatekeeper/v3/pkg/gator/test"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type GatorTestParams struct {
	Sources          []string
	KustomizeSources []string
	Policies         []string
}

func (p *GatorTestParams) BindCLIFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&p.Sources, "filename", nil, "Paths to source file")
	fs.StringSliceVar(&p.KustomizeSources, "kustomize", nil, "Paths to kustomize sources")
	fs.StringSliceVar(&p.Policies, "policy", nil, "Paths to rego policy file")
}

type TestResult struct {
	// Message is the message that was returned by the rule.
	Message string
	// Metadata is the extra metadata that was returned by the rule.
	Metadata map[string]interface{}
}

type TestResults struct {
	Source   reader.ObjectSource
	Failures []TestResult
	Warnings []TestResult
}

func toTestResult(gatorTestResult *gatortest.GatorResult) TestResult {
	rv := TestResult{
		Message: gatorTestResult.Msg,
	}

	if gatorTestResult.Metadata != nil {
		rv.Metadata = make(map[string]interface{})
		for k, v := range gatorTestResult.Metadata {
			rv.Metadata[k] = v
		}
	}

	return rv
}

func toTestResults(
	testTargets *reader.TestTargets,
	gatorTestResults []*gatortest.GatorResult,
) []*TestResults {
	resultsBySource := make(map[reader.ObjectSource]*TestResults)

	for _, gatorTestResult := range gatorTestResults {
		testResult := toTestResult(gatorTestResult)
		source := testTargets.ObjectSources[gatorTestResult.ViolatingObject]
		results, ok := resultsBySource[source]
		if !ok {
			results = &TestResults{
				Source: source,
			}
			resultsBySource[source] = results
		}

		switch gatorTestResult.EnforcementAction {
		case constraints.EnforcementActionDeny:
			results.Failures = append(results.Failures, testResult)
		case constraints.EnforcementActionWarn:
			results.Warnings = append(results.Warnings, testResult)
		case constraints.EnforcementActionDryRun:
			results.Warnings = append(results.Warnings, testResult)
		default:
			// TODO: log unknown enforcement action
		}
	}

	var rv []*TestResults
	for _, results := range resultsBySource {
		rv = append(rv, results)
	}
	sort.Slice(rv, func(i, j int) bool {
		ik := fmt.Sprintf("%s:%s", rv[i].Source.SourceType, rv[i].Source.FilePath)
		jk := fmt.Sprintf("%s:%s", rv[j].Source.SourceType, rv[j].Source.FilePath)

		return strings.Compare(ik, jk) < 0
	})

	return rv
}

func gatorTest(
	ctx context.Context,
	params GatorTestParams,
) error {
	constraintTargets, err := constraints.Load(ctx, constraints.LoadParams{
		RegoPaths: params.Policies,
	})
	if err != nil {
		return fmt.Errorf("load constraints: %w", err)
	}

	testTargets, err := reader.Load(ctx, reader.LoadParams{
		FileSources:      params.Sources,
		KustomizeSources: params.KustomizeSources,
	})
	if err != nil {
		return fmt.Errorf("load test targets: %w", err)
	}

	var objects []*unstructured.Unstructured
	objects = append(objects, testTargets.Objects...)
	objects = append(objects, constraintTargets.Constraints...)
	objects = append(objects, constraintTargets.ConstraintTemplates...)

	responses, err := gatortest.Test(
		objects,
		gatortest.Opts{
			IncludeTrace: true,
			GatherStats:  true,
		},
	)
	if err != nil {
		return fmt.Errorf("gator test: %w", err)
	}

	testResults := toTestResults(testTargets, responses.Results())
	for _, testResult := range testResults {
		for _, o := range testResult.Failures {
			fmt.Printf("%s FAIL %s\n", testResult.Source.FilePath, o.Message)
		}
		for _, o := range testResult.Warnings {
			fmt.Printf("%s WARN %s\n", testResult.Source.FilePath, o.Message)
		}
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
