package presenter

import (
	"fmt"
	"io"

	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/b4fun/ci"
	"github.com/b4fun/ci/cilog"
)

// Text creates a new text presenter.
func Text(queryResultsList []result.QueryResults) WriteQueryResultTo {
	const (
		categoryFAIL      = "FAIL"
		categoryWARNING   = "WARNING"
		categoryEXCEPTION = "EXCEPTION"
	)

	queryResultsObjList := asQueryResultsObjList(queryResultsList)

	var logger cilog.Logger
	switch name := ci.Detect(); name {
	case ci.AzurePipelines:
		// we use LogIssue in Azure pipelines to prompt the error in build details.
		logger = cilog.AzurePipeline(cilog.AzurePipelineUseLogIssue(true))
	default:
		logger = cilog.Get(name)
	}

	printResultObj := func(
		logger cilog.Logger,
		category string,
		filename string,
		o resultObj,
	) {
		var messageDetails string
		if o.Message == "" {
			messageDetails = fmt.Sprintf("(%s)", o.Rule.Name)
		} else {
			messageDetails = fmt.Sprintf("(%s) %s", o.Rule.Name, o.Message)
		}

		println := logger.Log
		switch category {
		case categoryFAIL:
			println = func(s string) { cilog.Error(logger, s) }
		case categoryWARNING:
			println = func(s string) { cilog.Warning(logger, s) }
		}

		println(fmt.Sprintf("%s - %s - %s", category, filename, messageDetails))
	}

	printDocumentLink := func(logger cilog.Logger, docLink string) {
		if docLink == "" {
			return
		}
		logger.Log(fmt.Sprintf("Document: %s", docLink))
	}

	return writeQueryResultToFunc(func(w io.Writer) error {
		logger.SetOutput(w)

		var (
			totalTests  int
			totalPasses int

			failures   []func(cilog.Logger)
			warnings   []func(cilog.Logger)
			exceptions []func(cilog.Logger)
		)

		for _, queryResultObj := range queryResultsObjList {
			totalPasses += queryResultObj.Success

			for _, o := range queryResultObj.Failures {
				o := o
				fileName := queryResultObj.Filename
				failures = append(failures, func(l cilog.Logger) {
					printResultObj(logger, categoryFAIL, fileName, o)
					printDocumentLink(logger, o.Rule.DocLink)
				})
			}
			for _, o := range queryResultObj.Warnings {
				o := o
				fileName := queryResultObj.Filename
				warnings = append(warnings, func(l cilog.Logger) {
					printResultObj(logger, categoryWARNING, fileName, o)
					printDocumentLink(logger, o.Rule.DocLink)
				})
			}
			for _, o := range queryResultObj.Exceptions {
				o := o
				fileName := queryResultObj.Filename
				exceptions = append(exceptions, func(l cilog.Logger) {
					printResultObj(logger, categoryEXCEPTION, fileName, o)
				})
			}
		}

		for _, cb := range failures {
			cb(logger)
		}
		for _, cb := range warnings {
			cb(logger)
		}
		if c := len(exceptions); c > 0 {
			groupName := fmt.Sprintf("EXCEPTIONS (%d)", c)
			excLogger, endExc := cilog.Group(logger, cilog.GroupLogParams{Name: groupName})
			for _, cb := range exceptions {
				cb(excLogger)
			}

			endExc()
		}

		totalTests = totalPasses + len(failures) + len(warnings) + len(exceptions)
		logger.Log(fmt.Sprintf(
			"%d test(s), %d passed, %d failure(s) %d warning(s), %d exception(s)",
			totalTests, totalPasses, len(failures), len(warnings), len(exceptions),
		))

		return nil
	})
}
