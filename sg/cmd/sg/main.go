package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/Azure/ShieldGuard/sg/internal/cli/llm"
	"github.com/Azure/ShieldGuard/sg/internal/cli/test"
)

func main() {
	cmd := createMainCmd()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func createMainCmd() *cobra.Command {
	rv := &cobra.Command{
		Use:               "sg",
		Short:             "Enables best security practices for your project from day zero.",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rv.AddCommand(
		test.CreateCLI(),
		llm.CreateCLI(),
	)

	return rv
}
