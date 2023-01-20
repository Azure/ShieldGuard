package result

import (
	"github.com/Azure/ShieldGuard/sg/internal/policy"
	"github.com/Azure/ShieldGuard/sg/internal/source"
)

// Result specifies the result of a policy rule query.
type Result struct {
	// Query is the OPA query that was executed.
	Query string
	// Rule is the rule executed by the query.
	Rule policy.Rule
	// RuleDocLink is the link to the documentation of the rule.
	RuleDocLink string
	// Message is the message that was returned by the rule.
	Message string
	// Metadata is the extra metadata that was returned by the rule.
	Metadata map[string]interface{}
}

// QueryResults specifies the results against a target.
type QueryResults struct {
	// Source specifies the target that was tested.
	Source source.Source
	// Successes is the number of successes queries.
	Successes int
	// Failures is the list of failed queries.
	Failures []Result
	// Warnings is the list of warning queries.
	Warnings []Result
	// Exceptions is the list of exception queries.
	Exceptions []Result
}
