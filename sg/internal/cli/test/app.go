package test

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/Azure/ShieldGuard/sg/internal/engine"
	"github.com/Azure/ShieldGuard/sg/internal/project"
	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/Azure/ShieldGuard/sg/internal/result/presenter"
	"github.com/Azure/ShieldGuard/sg/internal/source"
	"github.com/Azure/ShieldGuard/sg/internal/utils"
	"github.com/spf13/pflag"
)

// cliApp is the CLI cliApplication for the test subcommand.
type cliApp struct {
	projectSpecFile string
	contextRoot     string
	outputFormat    string

	stdout io.Writer
}

func newCliApp(ms ...func(*cliApp)) *cliApp {
	rv := &cliApp{
		outputFormat: presenter.FormatJSON,
	}

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

	projectSpec, err := project.ReadFromFile(cliApp.projectSpecFile)
	if err != nil {
		return fmt.Errorf("read project spec: %w", err)
	}

	var queryResultsList []result.QueryResults
	for _, target := range projectSpec.Files {
		queryResult, err := cliApp.queryFileTarget(ctx, cliApp.contextRoot, target)
		if err != nil {
			return fmt.Errorf("run target (%s): %w", target.Name, err)
		}
		queryResultsList = append(queryResultsList, queryResult...)
	}

	if err := presenter.QueryResultsList(cliApp.outputFormat, queryResultsList).
		WriteQueryResultTo(cliApp.stdout); err != nil {
		return fmt.Errorf("write query results: %w", err)
	}

	return nil
}

func (cliApp *cliApp) BindCLIFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&cliApp.projectSpecFile, "config", "c", project.SpecFileName, "Path to the project spec file.")
	fs.StringVarP(
		&cliApp.outputFormat, "output", "o", cliApp.outputFormat,
		fmt.Sprintf("Output format. Available formats: %s", presenter.AvailableFormatsHelp()),
	)
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

	if _, exists := presenter.AvailableFormats[cliApp.outputFormat]; !exists {
		return fmt.Errorf(
			"output format %q is not supported. Supported formats are: %s",
			cliApp.outputFormat, presenter.AvailableFormatsHelp(),
		)
	}

	return nil
}

func (cliApp *cliApp) queryFileTarget(
	ctx context.Context,
	contextRoot string,
	target project.FileTargetSpec,
) ([]result.QueryResults, error) {
	resolveToContextRoot := resolveToContextRootFn(contextRoot)

	policyPaths := utils.Map(target.Policies, resolveToContextRoot)
	paths := utils.Map(target.Paths, resolveToContextRoot)
	// TODO: load data paths
	// dataPaths := utils.Map(target.Data, resolveToContextRoot)

	sources, err := source.FromPath(paths).ContextRoot(contextRoot).Complete()
	if err != nil {
		return nil, fmt.Errorf("load sources failed: %w", err)
	}

	queryer, err := engine.QueryWithPolicy(policyPaths).Complete()
	if err != nil {
		return nil, fmt.Errorf("create queryer failed: %w", err)
	}

	var rv []result.QueryResults
	for _, source := range sources {
		queryOpts := &engine.QueryOptions{}
		queryResult, err := queryer.Query(ctx, source, queryOpts)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}

		rv = append(rv, queryResult)
	}

	return rv, nil
}

func resolveToContextRootFn(contextRoot string) func(string) string {
	return func(path string) string {
		// FIXME(hbc): handle absolute paths input
		//             We should limit the input to be relative to the context root.

		fullPath := filepath.Join(contextRoot, path)
		fullPath = filepath.Clean(fullPath)
		return fullPath
	}
}
