package presenter

import (
	"strings"

	"github.com/Azure/ShieldGuard/sg/internal/result"
)

const (
	FormatJSON = "json"
	FormatText = "text"
)

// AvailableFormats book-keeps the available formats.
var AvailableFormats = map[string]struct{}{
	FormatJSON: {},
	FormatText: {}
}

// AvailableFormatsHelp returns help message for available formats.
func AvailableFormatsHelp() string {
	rv := make([]string, 0, len(AvailableFormats))
	for format := range AvailableFormats {
		rv = append(rv, string(format))
	}
	return strings.Join(rv, ", ")
}

// QueryResultsList presents a list of query results.
func QueryResultsList(
	format string,
	queryResultsList []result.QueryResults,
) WriteQueryResultTo {
	switch strings.ToLower(format) {
	case FormatJSON:
		return JSON(queryResultsList)
	case FormatText:
		return Text(queryResultsList)	
	default:
		// defaults to JSON
		return JSON(queryResultsList)
	}
}
