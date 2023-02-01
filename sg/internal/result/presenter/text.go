package presenter

import (
	"fmt"

	"io"

	"github.com/Azure/ShieldGuard/sg/internal/result"
)

// Text creates a new text presenter.
func Text(queryResultsList []result.QueryResults) WriteQueryResultTo {
	queryResultsObjList := asQueryResultsObjList(queryResultsList)

	return writeQueryResultToFunc(func(w io.Writer) error {
		var totalTests int
		var totalPasses int
		var totalFailures int
		var totalWarnings int
		var totalExceptions int

		for _, queryResultObj := range queryResultsObjList { // TODO: maybe we should sort & group results before iterating them
			totalPasses += queryResultObj.Success

			totalFailures += len(queryResultObj.Failures)

			totalWarnings += len(queryResultObj.Warnings)

			totalExceptions += len(queryResultObj.Exceptions)

			for _, failureResultObj := range queryResultObj.Failures {
				fmt.Fprintf(w, "FAIL - %s - %s - %s\n", queryResultObj.Filename, queryResultObj.Namespace, failureResultObj.Message)
			}
			for _, warningResultObj := range queryResultObj.Warnings {
				fmt.Fprintf(w, "WARN - %s - %s - %s\n", queryResultObj.Filename, queryResultObj.Namespace, warningResultObj.Message)
			}
			for _, exceptionResultObj := range queryResultObj.Exceptions {
				fmt.Fprintf(w, "EXCEPTION - %s - %s - %s\n", queryResultObj.Filename, queryResultObj.Namespace, exceptionResultObj.Message)
			}
		}

		totalTests = totalPasses + totalFailures + totalWarnings + totalExceptions
		fmt.Fprintf(w, "%d tests, %d passed, %d failures %d warnings, %d exceptions\n", totalTests, totalPasses, totalFailures, totalWarnings, totalExceptions)

		return nil
	})
}
