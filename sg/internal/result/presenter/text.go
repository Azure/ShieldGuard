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
		var totalTest int
		var totalPass int
		var totalFailures int
		var totalWarnings int
		var totalExceptions int

		for _, queryResultObj := range queryResultsObjList { // TODO: maybe we should sort & group results before iterating them
			totalPass += queryResultObj.Success

			totalFailures += len(queryResultObj.Failures)

			totalWarnings += len(queryResultObj.Warnings)

			totalExceptions += len(queryResultObj.Exceptions)

			totalTest += totalPass + totalFailures + totalWarnings + totalExceptions

			for _, failureResultObj := range queryResultObj.Failures {
				fmt.Fprintf(w, "FAIL - %s - %s - %s\n", queryResultObj.Filename, queryResultObj.Namespace, failureResultObj.Message)
			}
			for _, warningResultObj := range queryResultObj.Warnings {
				fmt.Fprintf(w, "WARN - %s - %s - %s\n", queryResultObj.Filename, queryResultObj.Namespace, warningResultObj.Message)
			}
			for _, exceptionResultObj := range queryResultObj.Exceptions {
				fmt.Fprintf(w, "Exception - %s - %s - %s\n", queryResultObj.Filename, queryResultObj.Namespace, exceptionResultObj.Message)
			}
		}

		fmt.Fprintf(w, "%d tests, %d passed, %d failures %d warnings, %d exceptions\n", totalTest, totalPass, totalFailures, totalWarnings, totalExceptions)

		return nil
	})
}

func String(totalTest int) {
	panic("unimplemented")
}
