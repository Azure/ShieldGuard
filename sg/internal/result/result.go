package result

import (
	"fmt"

	"github.com/open-policy-agent/opa/rego"
)

func empty(query string) Result {
	return Result{Query: query}
}

func fromString(query string, message string) Result {
	return Result{
		Query:   query,
		Message: message,
	}
}

func fromMetadata(query string, metadata map[string]interface{}) (Result, error) {
	const msgField = "msg"

	if _, ok := metadata[msgField]; !ok {
		return Result{}, fmt.Errorf("rule missing %s field: %v", msgField, metadata)
	}
	if _, ok := metadata[msgField].(string); !ok {
		return Result{}, fmt.Errorf("%s field must be string: %v", msgField, metadata)
	}

	rv := fromString(query, metadata[msgField].(string))
	rv.Metadata = make(map[string]interface{})

	for k, v := range metadata {
		if k == msgField {
			continue
		}
		rv.Metadata[k] = v
	}

	return rv, nil
}

// FromRegoExpression resolves result from rego expression.
// Based on: https://github.com/open-policy-agent/conftest/blob/f18b7bbde2fdbd766c8348dff3a0a24792eb98c7/policy/engine.go#L439-L443
func FromRegoExpression(
	query string,
	expression *rego.ExpressionValue,
) ([]Result, error) {
	// Rego rules that are intended for evaluation should return a slice of values.
	// For example, deny[msg] or violation[{"msg": msg}].
	//
	// When an expression does not have a slice of values, the expression did not
	// evaluate to true, and no message was returned.
	var expressionValues []interface{}
	if v, ok := expression.Value.([]interface{}); ok {
		expressionValues = v
	}
	if len(expressionValues) == 0 {
		return []Result{empty(query)}, nil
	}

	var rv []Result
	for _, v := range expressionValues {
		switch val := v.(type) {
		// Policies that only return a single string (e.g. deny[msg])
		case string:
			rv = append(rv, fromString(query, val))
			// Policies that return metadata (e.g. deny[{"msg": msg}])
		case map[string]interface{}:
			v, err := fromMetadata(query, val)
			if err != nil {
				return nil, fmt.Errorf("failed to load from metadata: %w", err)
			}
			rv = append(rv, v)
		}
	}

	return rv, nil
}

// Passed tells if the result is passed.
func (r Result) Passed() bool {
	return r.Message == ""
}

// Merge merges two results into a new one.
// The new result uses Source from the first result.
func (qr QueryResults) Merge(other QueryResults) QueryResults {
	return QueryResults{
		Source:     qr.Source,
		Successes:  qr.Successes + other.Successes,
		Failures:   append(qr.Failures, other.Failures...),
		Warnings:   append(qr.Warnings, other.Warnings...),
		Exceptions: append(qr.Exceptions, other.Exceptions...),
	}
}
