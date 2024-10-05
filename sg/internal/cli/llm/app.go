package llm

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/Azure/ShieldGuard/sg/internal/project"
	"github.com/spf13/pflag"
)

type cliApp struct {
	projectSpecFile string
	contextRoot     string

	stdout io.Writer
}

func newCliApp(ms ...func(*cliApp)) *cliApp {
	rv := &cliApp{}

	for _, m := range ms {
		m(rv)
	}

	return rv
}

func (cliApp *cliApp) Run() error {
	if err := cliApp.defaults(); err != nil {
		return fmt.Errorf("defaults: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println(ctx)

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
