package test

import "github.com/spf13/cobra"

// CreateCLI creates the CLI for the test subcommand.
func CreateCLI() *cobra.Command {
	app := newCliApp()

	cmd := &cobra.Command{
		Use:   "test [PROJECT-PATH]",
		Short: "Test targets under the project.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app.contextRoot = args[0]

			return app.Run()
		},
	}

	app.BindCLIFlags(cmd.Flags())

	return cmd
}
