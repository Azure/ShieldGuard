package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func main() {
	cmd := createMainCmd()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func createMainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "gator-auto",
	}

	cmd.AddCommand(
		createGatorTestCommand(),
	)

	return cmd
}
