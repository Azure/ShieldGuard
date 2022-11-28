package test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Azure/ShieldGuard/sg/internal/result/presenter"
	"github.com/stretchr/testify/assert"
)

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
		cliApp.projectSpecFile = resolveTestdataPath(t, "./testdata/basic/project.yaml")
	})

	assert.NoError(t, cliApp.Run(), "cliApp run")
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
		cliApp.projectSpecFile = resolveTestdataPath(t, "./testdata/basic/project.yaml")
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
