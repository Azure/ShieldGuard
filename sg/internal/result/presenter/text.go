package presenter

import (
	"fmt"
	"io"

	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/b4fun/ci"
	"github.com/b4fun/ci/cilog"
)

func plainText(queryResultsList []result.QueryResults) WriteQueryResultTo {
	queryResultsObjList := asQueryResultsObjList(queryResultsList)

	printQueryResultObj := func(
		w io.Writer,
		category string,
		filename string,
		o resultObj,
		printDocumentLink bool,
	) {
		var messageDetails string
		if o.Message == "" {
			messageDetails = fmt.Sprintf("(%s)", o.Rule.Name)
		} else {
			messageDetails = fmt.Sprintf("%s (%s)", o.Message, o.Rule.Name)
		}

		fmt.Fprintf(
			w,
			"%s - %s - %s\n",
			category, filename, messageDetails,
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
		var (
			totalTests      int
			totalPasses     int
			totalFailures   int
			totalWarnings   int
			totalExceptions int
		)

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

func ciText(logger cilog.Logger, queryResultsList []result.QueryResults) WriteQueryResultTo {
	queryResultsObjList := asQueryResultsObjList(queryResultsList)

	printQueryResultObj := func(
		println func(s string),
		category string,
		filename string,
		o resultObj,
		printDocumentLink bool,
	) {
		var messageDetails string
		if o.Message == "" {
			messageDetails = fmt.Sprintf("(%s)", o.Rule.Name)
		} else {
			messageDetails = fmt.Sprintf("%s (%s)", o.Message, o.Rule.Name)
		}

		println(fmt.Sprintf("%s - %s - %s", category, filename, messageDetails))
		if printDocumentLink && o.Rule.DocLink != "" {
			println(fmt.Sprintf("Document: %s", o.Rule.DocLink))
		}
	}

	return writeQueryResultToFunc(func(w io.Writer) error {
		var (
			totalTests      int
			totalPasses     int
			totalFailures   int
			totalWarnings   int
			totalExceptions int
		)

		for _, queryResultObj := range queryResultsObjList {
			totalPasses += queryResultObj.Success
			totalFailures += len(queryResultObj.Failures)
			totalWarnings += len(queryResultObj.Warnings)
			totalExceptions += len(queryResultObj.Exceptions)

			printFailures := logger.Log
			if cilog.Can(logger, cilog.CapabilityErrorLog) {
				printFailures = logger.ErrorLog
			}
			printWarnings := logger.Log
			if cilog.Can(logger, cilog.CapabilityWarningLog) {
				printWarnings = logger.WarningLog
			}

			for _, o := range queryResultObj.Failures {
				printQueryResultObj(printFailures, "FAIL", queryResultObj.Filename, o, true)
			}
			for _, o := range queryResultObj.Warnings {
				printQueryResultObj(printWarnings, "WARN", queryResultObj.Filename, o, true)
			}

			if len(queryResultObj.Exceptions) > 0 {
				excLogger := logger
				endGroup := func() {}
				if cilog.Can(logger, cilog.CapabilityGroupLog) {
					groupName := fmt.Sprintf("EXCEPTIONS (%d)", len(queryResultObj.Exceptions))
					excLogger, endGroup = logger.GroupLog(cilog.GroupLogParams{Name: groupName})
				}
				for _, o := range queryResultObj.Exceptions {
					printQueryResultObj(excLogger.Log, "EXCEPTION", queryResultObj.Filename, o, false)
				}
				endGroup()
			}
		}

		totalTests = totalPasses + totalFailures + totalWarnings + totalExceptions
		logger.Log(fmt.Sprintf(
			"%d test(s), %d passed, %d failure(s) %d warning(s), %d exception(s)",
			totalTests, totalPasses, totalFailures, totalWarnings, totalExceptions,
		))

		return nil
	})
}

// Text creates a new text presenter.
func Text(queryResultsList []result.QueryResults) WriteQueryResultTo {
	switch name := ci.Detect(); name {
	case ci.GithubActions, ci.AzurePipelines:
		return ciText(cilog.Get(name), queryResultsList)
	default:
		return plainText(queryResultsList)
	}
}
