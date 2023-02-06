package presenter

import (
	"fmt"
	"io"

	"github.com/Azure/ShieldGuard/sg/internal/result"
)

// Text creates a new text presenter.
func Text(queryResultsList []result.QueryResults) WriteQueryResultTo {
	queryResultsObjList := asQueryResultsObjList(queryResultsList)

	printQueryResultObj := func(
		w io.Writer,
		category string,
		filename string,
		o resultObj,
		printDocumentLink bool,
	) {
		fmt.Fprintf(
			w,
			"%s - %s - %s\n",
			category, filename, o.Message,
		)
		if printDocumentLink && o.Rule.DocLink != "" {
			fmt.Fprintf(
				w,
				"Document: %s\n",
				o.Rule.DocLink,
			)
		}
	}

	return writeQueryResultToFunc(func(w io.Writer) error {
		var totalTests int
		var totalPasses int
		var totalFailures int
		var totalWarnings int
		var totalExceptions int

		for _, queryResultObj := range queryResultsObjList {
			totalPasses += queryResultObj.Success

			totalFailures += len(queryResultObj.Failures)

			totalWarnings += len(queryResultObj.Warnings)

			totalExceptions += len(queryResultObj.Exceptions)

			for _, o := range queryResultObj.Failures {
				printQueryResultObj(w, "FAIL", queryResultObj.Filename, o, true)
			}
			for _, o := range queryResultObj.Warnings {
				printQueryResultObj(w, "WARN", queryResultObj.Filename, o, true)
			}
			for _, o := range queryResultObj.Exceptions {
				printQueryResultObj(w, "EXCEPTION", queryResultObj.Filename, o, false)
			}
		}

		totalTests = totalPasses + totalFailures + totalWarnings + totalExceptions
		fmt.Fprintf(
			w,
			"%d test(s), %d passed, %d failure(s) %d warning(s), %d exception(s)\n",
			totalTests, totalPasses, totalFailures, totalWarnings, totalExceptions,
		)

		return nil
	})
}
