package main

import (
	"os"

	"github.com/spf13/cobra"

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
		Short:             "ShieldGuard secures your code from day one.",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rv.AddCommand(
		test.CreateCLI(),
	)

	return rv
}
