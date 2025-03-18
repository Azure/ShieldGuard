package test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/Azure/ShieldGuard/sg/internal/engine"
	"github.com/Azure/ShieldGuard/sg/internal/project"
	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/Azure/ShieldGuard/sg/internal/result/presenter"
	"github.com/Azure/ShieldGuard/sg/internal/source"
	"github.com/Azure/ShieldGuard/sg/internal/utils"
	"github.com/sourcegraph/conc/iter"
	"github.com/spf13/pflag"
)

var errTestFailure = errors.New("test failed")

type failSettings struct {
	noFail         bool
	failOnWarnings bool
}

func (s *failSettings) BindCLIFlags(fs *pflag.FlagSet) {
	fs.BoolVar(
		&s.failOnWarnings, "fail-on-warn", false,
		"Fail the command if any query returns warnings.",
	)
	fs.BoolVar(
		&s.noFail, "no-fail", false,
		"Do not fail the command if any query fails. When specified to true, it suppress the --fail-on-warn flag.",
	)
}

func (s *failSettings) CheckQueryResults(results []result.QueryResults) error {
	if s.noFail {
		return nil
	}

	countFailures := 0
	countWarnings := 0
	for _, result := range results {
		countFailures += len(result.Failures)
		countWarnings += len(result.Warnings)
	}
	if countFailures < 1 && countWarnings < 1 {
		return nil
	}

	err := fmt.Errorf(
		"%w: found %d failure(s), %d warning(s)",
		errTestFailure, countFailures, countWarnings,
	)
	if countFailures > 0 {
		return err
	}
	if s.failOnWarnings && countWarnings > 0 {
		return err
	}

	return nil
}

// cliApp is the CLI cliApplication for the test subcommand.
type cliApp struct {
	projectSpecFile          string
	contextRoot              string
	outputFormat             string
	failSettings             *failSettings
	enableQueryCache         bool
	parseArmTemplateDefaults bool

	stdout io.Writer
}

func newCliApp(ms ...func(*cliApp)) *cliApp {
	rv := &cliApp{
		outputFormat: presenter.FormatJSON,
		failSettings: new(failSettings),
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

	queryCache := engine.NewQueryCache()

	var queryResultsList []result.QueryResults
	for _, target := range projectSpec.Files {
		queryResult, err := cliApp.queryFileTarget(ctx, cliApp.contextRoot, target, queryCache)
		if err != nil {
			return fmt.Errorf("run target (%s): %w", target.Name, err)
		}
		queryResultsList = append(queryResultsList, queryResult...)
	}

	if err := presenter.QueryResultsList(cliApp.outputFormat, queryResultsList).
		WriteQueryResultTo(cliApp.stdout); err != nil {
		return fmt.Errorf("write query results: %w", err)
	}

	if err := cliApp.failSettings.CheckQueryResults(queryResultsList); err != nil {
		return err
	}

	return nil
}

func (cliApp *cliApp) BindCLIFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&cliApp.projectSpecFile, "config", "c", project.SpecFileName, "Path to the project spec file.")
	fs.StringVarP(
		&cliApp.outputFormat, "output", "o", cliApp.outputFormat,
		fmt.Sprintf("Output format. Available formats: %s", presenter.AvailableFormatsHelp()),
	)
	fs.BoolVarP(&cliApp.enableQueryCache, "enable-query-cache", "", false, "Enable query cache (experimental).")
	fs.BoolVarP(&cliApp.parseArmTemplateDefaults, "parse-defaults", "p", false, "Parse default values from arm templates (experimental).")
	cliApp.failSettings.BindCLIFlags(fs)
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
	queryCache engine.QueryCache,
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

	qb := engine.QueryWithPolicy(policyPaths)
	if cliApp.enableQueryCache {
		qb.WithQueueCache(queryCache)
	}
	qb.QueryWithParsingArmTemplateDefaults(cliApp.parseArmTemplateDefaults)

	queryer, err := qb.Complete()
	if err != nil {
		return nil, fmt.Errorf("create queryer failed: %w", err)
	}

	queryMapper := iter.Mapper[source.Source, result.QueryResults]{
		MaxGoroutines: len(sources),
	}

	return queryMapper.MapErr(sources, func(s *source.Source) (result.QueryResults, error) {
		return queryer.Query(ctx, *s,  &engine.QueryOptions{})
	})
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
