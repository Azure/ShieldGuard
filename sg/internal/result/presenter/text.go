package presenter

import (
	"encoding"
	"fmt"
	//"encoding/base64"

	"io"

	"github.com/Azure/ShieldGuard/sg/internal/result"
)

// Text creates a new text presenter.
func Text(queryResultsList []result.QueryResults) WriteQueryResultTo {
	return writeQueryResultToFunc(func(w io.Writer) error {
		var failuresList []Result
		var warningsList []Result
		var exceptionsList []Result
		var totalTest int
		var totalPass int
		var totalFailures int
		var totalWarnings int 
		var totalExceptions int
		for _, queryResult := range queryResultsList { // TODO: maybe we should sort & group results before iterating them
			totalPass += queryResult.Successes
			
			failuresList = append(failuresList,queryResult.Failures...)
			totalFailures += len(queryResult.Failures)
			
			warningsList = append(warningsList,queryResult.Warnings...)
			totalWarnings += len(queryResult.Warnings)
			
			exceptionsList = append(warningsList,queryResult.Exceptions...)
			totalExceptions += len(queryResult.Exceptions)
			
			totalTest += totalPass + totalFailures + totalWarnings + totalExceptions	
		}
		for _, failureResult := range failuresList {
			fmt.Fprintln(w, "FAIL - %s - %s - %s", <file-path>, <namespace>, failureResult.Message)
		}
		for _, warningResult := range warningsList {
			fmt.Fprintln(w, "WARN - %s - %s - %s", <file-path>, <namespace>, warningResult.Message)
		}
		for _, exceptionResult := range exceptionsList {
			fmt.Fprintln(w, "Exception - %s - %s - %s", <file-path>, <namespace>, exceptionResult.Message)
		}
		fmt.Fprintf(w, "%s tests, %s passed, %s failures %s warnings, %s exceptions", totalTest, totalPass, totalFailures, totalWarnings, totalExceptions)
	})
}
