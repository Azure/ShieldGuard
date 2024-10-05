package llm

import "github.com/spf13/cobra"

func CreateCLI() *cobra.Command {
	app := newCliApp()

	cmd := &cobra.Command{
		Use:   "llm [PROJECT-PATH]",
		Short: "INspect targets under the project with LLM.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app.contextRoot = args[0]
			app.stdout = cmd.OutOrStdout()

			return app.Run()
		},
	}

	app.BindCLIFlags(cmd.Flags())

	return cmd
}