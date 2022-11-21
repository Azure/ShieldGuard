package test

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Azure/ShieldGuard/sg/internal/project"
	"github.com/spf13/pflag"
)

// cliApp is the CLI cliApplication for the test subcommand.
type cliApp struct {
	projectSpecFile string
	contextRoot     string
}

func newCliApp() *cliApp {
	return &cliApp{}
}

func (cliApp *cliApp) Run() error {
	if err := cliApp.defaults(); err != nil {
		return fmt.Errorf("defaults: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	projectSpec, err := project.ReadFromFile(cliApp.projectSpecFile)
	if err != nil {
		return fmt.Errorf("read project spec: %w", err)
	}

	for _, target := range projectSpec.Files {
		if err := cliApp.runFileTarget(ctx, cliApp.contextRoot, target); err != nil {
			return fmt.Errorf("run target (%s): %w", target.Name, err)
		}
	}

	return nil
}

func (cliApp *cliApp) BindCLIFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&cliApp.projectSpecFile, "config", "c", project.SpecFileName, "Path to the project spec file.")
}

func (cliApp *cliApp) defaults() error {
	var err error

	if cliApp.projectSpecFile == "" {
		return fmt.Errorf("project spec file is not specified")
	}
	cliApp.projectSpecFile, err = filepath.Abs(cliApp.projectSpecFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of the project spec file: %w", err)
	}

	if cliApp.contextRoot == "" {
		cliApp.contextRoot = "."
	}
	cliApp.contextRoot, err = filepath.Abs(cliApp.contextRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of the context root: %w", err)
	}

	return nil
}

func (cliApp *cliApp) runFileTarget(
	ctx context.Context,
	contextRoot string,
	target project.FileTargetSpec,
) error {
	return nil
}
