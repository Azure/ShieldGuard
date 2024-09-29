package source

import "github.com/open-policy-agent/opa/ast"

// Source is the interface for a validation source.
type Source interface {
	// Name returns the name of the target.
	Name() string
	// ParsedConfigurations returns the parsed configurations of the target.
	ParsedConfigurations() ([]ast.Value, error)
}
