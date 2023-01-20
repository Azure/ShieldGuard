package presenter

import (
	"encoding/json"
	"io"

	"github.com/Azure/ShieldGuard/sg/internal/result"
)

// JSON creates a new JSON presenter.
func JSON(queryResultsList []result.QueryResults) WriteQueryResultTo {
	queryResultsObjList := asQueryResultsObjList(queryResultsList)

	return writeQueryResultToFunc(func(w io.Writer) error {
		marshaler := json.NewEncoder(w)
		marshaler.SetIndent("", "  ")
		return marshaler.Encode(queryResultsObjList)
	})
}
