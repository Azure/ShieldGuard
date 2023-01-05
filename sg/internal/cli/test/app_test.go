package test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/Azure/ShieldGuard/sg/internal/result/presenter"
	"github.com/stretchr/testify/assert"
)

func Test_failSettings_CheckQueryResults(t *testing.T) {
	queryResultFailed := func() result.QueryResults {
		return result.QueryResults{
			Failures: []result.Result{
				{Query: "failure"},
			},
		}
	}
	queryResultWarn := func() result.QueryResults {
		return result.QueryResults{
			Warnings: []result.Result{
				{Query: "warning"},
			},
		}
	}

	cases := []struct {
		failSettings *failSettings
		results      []result.QueryResults
		expectErr    bool
	}{
		{
			failSettings: &failSettings{},
			results:      []result.QueryResults{},
			expectErr:    false,
		},
		// 1 failure
		{
			failSettings: &failSettings{},
			results: []result.QueryResults{
				queryResultFailed(),
			},
			expectErr: true,
		},
		// 0 failures, 1 warning
		{
			failSettings: &failSettings{},
			results: []result.QueryResults{
				queryResultWarn(),
			},
			expectErr: false,
		},
		{
			failSettings: &failSettings{
				noFail:         true,
				failOnWarnings: true,
			},
			results:   []result.QueryResults{},
			expectErr: false,
		},
		// noFail = true
		{
			failSettings: &failSettings{
				noFail:         true,
				failOnWarnings: false,
			},
			results: []result.QueryResults{
				queryResultFailed(),
				queryResultWarn(),
			},
			expectErr: false,
		},
		// failOnWarnings = true, 0 failures, 1 warning
		{
			failSettings: &failSettings{
				noFail:         false,
				failOnWarnings: true,
			},
			results: []result.QueryResults{
				queryResultWarn(),
			},
			expectErr: true,
		},
	}

	for idx := range cases {
		t.Run(fmt.Sprintf("case #%d", idx), func(t *testing.T) {
			c := cases[idx]
			err := c.failSettings.CheckQueryResults(c.results)
			if c.expectErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errTestFailure)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func resolveTestdataPath(t *testing.T, path string) string {
	t.Helper()
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("resolve testdata path: %v", err)
	}
	return absPath
}

func withDebugOutput(w io.Writer) io.Writer {
	return io.MultiWriter(w, os.Stderr)
}

func Test_cliApp_basic(t *testing.T) {
	var output bytes.Buffer

	cliApp := newCliApp(func(cliApp *cliApp) {
		cliApp.contextRoot = resolveTestdataPath(t, "./testdata/basic")
		cliApp.outputFormat = presenter.FormatJSON
		cliApp.stdout = withDebugOutput(&output)
		cliApp.projectSpecFile = resolveTestdataPath(t, "./testdata/basic/sg-project.yaml")
	})

	runErr := cliApp.Run()
	assert.Error(t, runErr, "cliApp run should return error")
	assert.ErrorIs(t, runErr, errTestFailure)
	assert.ErrorContains(t, runErr, "found 1 failure(s), 1 warning(s)")
	assert.Equal(
		t,
		testdataBasicJSONOutputGolden, strings.TrimSpace(output.String()),
		"should generate expected output",
	)
}

func Test_cliApp_defaults(t *testing.T) {
	validCliApp := func(cliApp *cliApp) {
		cliApp.contextRoot = resolveTestdataPath(t, "./testdata/basic")
		cliApp.outputFormat = presenter.FormatJSON
		cliApp.stdout = os.Stderr
		cliApp.projectSpecFile = resolveTestdataPath(t, "./testdata/basic/sg-project.yaml")
	}

	cases := []*cliApp{
		newCliApp(
			validCliApp,
			func(cliApp *cliApp) {
				cliApp.projectSpecFile = ""
			},
		),
		newCliApp(
			validCliApp,
			func(cliApp *cliApp) {
				cliApp.outputFormat = "foobar"
			},
		),
	}

	for idx := range cases {
		t.Run(fmt.Sprintf("case #%d", idx), func(t *testing.T) {
			assert.Error(t, cases[idx].defaults())
		})
	}
}
