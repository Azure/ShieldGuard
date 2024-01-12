package test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Azure/ShieldGuard/sg/internal/project"
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

func resolveTestdataPath(t *testing.T, paths ...string) string {
	t.Helper()
	absPath, err := filepath.Abs(filepath.Join(paths...))
	if err != nil {
		t.Fatalf("resolve testdata path: %v", err)
	}
	return absPath
}

func defaults[T comparable](v T, defaultValue T) T {
	var zeroValue T

	if v == zeroValue {
		return defaultValue
	}
	return v
}

func withDebugOutput(w io.Writer) io.Writer {
	return io.MultiWriter(w, os.Stderr)
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

func Test_cliApp_perf(t *testing.T) {
	if testing.Short() {
		// this test is slow, skip it in short mode
		t.Skip("skipping test in short mode")
		return
	}

	t.Parallel()

	// generates filesCount * rulesCount targets
	const filesCount = 200
	const rulesCount = 100

	tempDir := t.TempDir()
	sgProjectConfigFile := filepath.Join(tempDir, "sg-project.yaml")
	configurationsDir := "configurations"
	policyDir := "policy"

	for _, dir := range []string{configurationsDir, policyDir} {
		assert.NoError(t, os.Mkdir(filepath.Join(tempDir, dir), 0755))
	}

	// construct configurations folder
	spec := project.Spec{}
	for i := 0; i < filesCount; i++ {
		subConfigurationsDir := fmt.Sprintf("test-%d", i)
		targetPath := filepath.Join(configurationsDir, subConfigurationsDir)

		filePath := filepath.Join(tempDir, targetPath, "test.json")
		assert.NoError(t, os.MkdirAll(filepath.Dir(filePath), 0755))
		assert.NoError(t, os.WriteFile(filePath, []byte("{}"), 0644))

		spec.Files = append(spec.Files, project.FileTargetSpec{
			Name:  fmt.Sprintf("test-%d", i),
			Paths: []string{targetPath},
			Policies: []string{
				policyDir,
			},
		})
	}

	spProjectContent, err := json.Marshal(spec)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(sgProjectConfigFile, spProjectContent, 0644))

	// construct rules
	for i := 0; i < rulesCount; i++ {
		fileName := fmt.Sprintf("%03d-test.rego", i)
		filePath := filepath.Join(tempDir, policyDir, fileName)
		policyContent := fmt.Sprintf(`
package main

deny_deny_rule_%d[msg] {
	msg := "deny_rule_%d"
}

exception[rules] {
	rules = ["deny_rule_%d"]
} 
`, i, i, i)
		assert.NoError(t, os.WriteFile(filePath, []byte(policyContent), 0644))
	}

	cliApp := newCliApp(
		func(cliApp *cliApp) {
			cliApp.contextRoot = tempDir
			cliApp.projectSpecFile = sgProjectConfigFile
			cliApp.stdout = io.Discard
		},
	)

	t.Log("starting perf test")

	start := time.Now()
	runErr := cliApp.Run()
	duration := time.Since(start)
	t.Logf(
		"perf test duration: %s (CPU=%d files=%d, rules=%d)",
		duration, runtime.GOMAXPROCS(0), filesCount, rulesCount,
	)

	assert.NoError(t, runErr)
}
