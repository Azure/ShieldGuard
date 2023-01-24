package test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Azure/ShieldGuard/sg/internal/result/presenter"
	"github.com/stretchr/testify/assert"
)

type testSuiteRunCheckFunc func(
	t *testing.T,
	suite *testdataTestSuite,
	runErr error,
	output string,
)

// testdataTestSuite describes a test suite with fixtures from testdata folder.
type testdataTestSuite struct {
	// Name - name of the test suite.
	// Required.
	Name string
	// ProjectSpecFile - path to the project spec file.
	// Defaults to <testdata>/<.Name>/sg-project.yaml
	ProjectSpecFile string
	// GoldenJSONOutput - path to the JSON output file.
	// Sets to non-empty path to enable golden file testing.
	GoldenJSONOutput string
	// Checkers - additional checkers to run after the test suite.
	Checkers []testSuiteRunCheckFunc
}

func (ts *testdataTestSuite) resolveTestdataPath(t *testing.T, paths ...string) string {
	return resolveTestdataPath(t, append([]string{"testdata", ts.Name}, paths...)...)
}

func (ts *testdataTestSuite) resolveCliApp(t *testing.T) (*cliApp, *bytes.Buffer) {
	if ts.Name == "" {
		t.Errorf("testdataTestSuite.Name is required")
	}

	output := new(bytes.Buffer)

	cliApp := newCliApp(
		func(cliApp *cliApp) {
			cliApp.contextRoot = ts.resolveTestdataPath(t)
			cliApp.outputFormat = presenter.FormatJSON
			cliApp.stdout = withDebugOutput(output)
			cliApp.projectSpecFile = defaults(
				ts.ProjectSpecFile,
				ts.resolveTestdataPath(t, "sg-project.yaml"),
			)
		},
	)

	return cliApp, output
}

// Run invokes the test suite.
func (ts *testdataTestSuite) Run(t *testing.T) {
	cliApp, output := ts.resolveCliApp(t)

	runErr := cliApp.Run()
	if ts.GoldenJSONOutput != "" {
		goldenJSONOutput := resolveTestdataPath(t, ts.GoldenJSONOutput)
		assert.Equal(
			t, goldenJSONOutput, strings.TrimSpace(output.String()),
			"should generate expected output",
		)
	}
	for _, checker := range ts.Checkers {
		checker(t, ts, runErr, output.String())
	}
}

func expectRunErrorWith(numFailures int, numWarnings int) testSuiteRunCheckFunc {
	expectedErr := fmt.Sprintf("test failed: found %d failure(s), %d warning(s)", numFailures, numWarnings)

	return func(t *testing.T, ts *testdataTestSuite, runErr error, output string) {
		assert.Error(t, runErr, "should return error")
		assert.Equal(t, runErr.Error(), expectedErr)
	}
}

func expectGoldenOutput(goldenOutputFileName string) testSuiteRunCheckFunc {
	return func(t *testing.T, ts *testdataTestSuite, runErr error, output string) {
		goldenOutputFilePath := ts.resolveTestdataPath(t, goldenOutputFileName)
		b, err := os.ReadFile(goldenOutputFilePath)
		assert.NoError(t, err, "read golden output file: %q", goldenOutputFilePath)
		expectedOutput := strings.TrimSpace(string(b))

		assert.Equal(t, expectedOutput, strings.TrimSpace(output))
	}
}

func Test_cliApp_testdataTestSuites(t *testing.T) {
	testSuites := []*testdataTestSuite{
		{
			Name: "basic",
			Checkers: []testSuiteRunCheckFunc{
				expectRunErrorWith(1, 1),
				expectGoldenOutput("golden-output.json"),
			},
		},
		{
			Name: "bug25",
			Checkers: []testSuiteRunCheckFunc{
				expectRunErrorWith(2, 1),
				expectGoldenOutput("golden-output.json"),
			},
		},
	}

	for idx := range testSuites {
		t.Run(fmt.Sprintf("test suite #%d", idx), func(t *testing.T) {
			testSuites[idx].Run(t)
		})
	}
}
