package main

import (
	"context"
	"fmt"

	"github.com/Azure/ShieldGuard/gator-auto/internal/gatorshim/reader"
	gatortest "github.com/open-policy-agent/gatekeeper/v3/pkg/gator/test"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type GatorTestParams struct {
	Sources []string
}

func (p *GatorTestParams) BindCLIFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&p.Sources, "filename", nil, "Paths to source file")
}

func gatorTest(
	ctx context.Context,
	params GatorTestParams,
) error {
	targets, err := reader.ReadTargets(params.Sources)
	if err != nil {
		return err
	}

	responses, err := gatortest.Test(
		targets.Objects,
		gatortest.Opts{
			IncludeTrace: true,
			GatherStats:  true,
		},
	)
	if err != nil {
		return err
	}

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
			gatorTest(cmd.Context(), params)

			return nil
		},
	}

	params.BindCLIFlags(cmd.Flags())

	return cmd
}
