package test

import (
	"errors"

	"github.com/spf13/cobra"
)

// CreateCLI creates the CLI for the test subcommand.
func CreateCLI() *cobra.Command {
	app := newCliApp()

	cmd := &cobra.Command{
		Use:   "test [PROJECT-PATH]",
		Short: "Test targets under the project.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app.contextRoot = args[0]
			app.stdout = cmd.OutOrStdout()

			appRunErr := app.Run()
			if errors.Is(appRunErr, errTestFailure) {
				// the test has ran and failed, but we don't want to show help message
				cmd.SilenceUsage = true
			}
			return appRunErr
		},
	}

	app.BindCLIFlags(cmd.Flags())

	return cmd
}
