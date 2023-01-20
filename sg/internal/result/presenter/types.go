package presenter

import "io"

// WriteQueryResultTo is an interface for writing results.
type WriteQueryResultTo interface {
	// WriteQueryResultTo writes the query result to a writer.
	WriteQueryResultTo(w io.Writer) error
}

type writeQueryResultToFunc func(w io.Writer) error

var _ WriteQueryResultTo = writeQueryResultToFunc(nil)

func (f writeQueryResultToFunc) WriteQueryResultTo(w io.Writer) error {
	return f(w)
}
