package policy

import (
	"github.com/open-policy-agent/opa/ast"
)

// QueryKind specifies the kind of the query.
type QueryKind string

const (
	QueryKindUnknown   QueryKind = "unknown"
	QueryKindWarn      QueryKind = "warn"
	QueryKindDeny      QueryKind = "deny"
	QueryKindViolation QueryKind = "violation"
	QueryKindException QueryKind = "exception"
)

// Rule specifies a policy rule.
// For example:
//   - "data.main.deny_x" => Kind: deny, Name: "x"
//   - "data.main.violation_y" => Kind: violation, Name: "y"
//   - "data.main.warn_z" => Kind: warn, Name: "z"
//
// For naming conventions, see: https://www.conftest.dev/exceptions/
type Rule struct {
	// Kind specifies the kind of the query.
	Kind QueryKind
	// Name provides the name of the rule.
	Name string
	// Namespace specifies the namespace of the rule.
	Namespace string
	// SourceLocation is the source definition of the rule.
	SourceLocation *ast.Location
}

// Package defines the access methods to a policy package.
type Package interface {
	// Spec returns the package spec.
	Spec() PackageSpec

	// Rules lists all the rules in the package.
	// NOTE: <Kind> + <Name> is the primary key to a rule query. Therefore, a rule (by name)
	//       can be returned more than once.
	Rules() []Rule

	// ParsedModules returns the parsed rego modules.
	ParsedModules() map[string]*ast.Module
}
