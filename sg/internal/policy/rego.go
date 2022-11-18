package policy

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"
)

// RegoCompilerOptions configs the RegoCompiler.
type RegoCompilerOptions struct{}

// NewRegoCompiler creates a compiler from policy packages.
func NewRegoCompiler(
	packages []Package,
	_ ...RegoCompilerOptions,
) (*ast.Compiler, error) {
	modules := map[string]*ast.Module{}
	for _, p := range packages {
		for name, m := range p.ParsedModules() {
			modules[name] = m
		}
	}

	compiler := ast.NewCompiler()
	compiler.Compile(modules)
	if compiler.Failed() {
		return nil, fmt.Errorf("failed to create compiler: %w", compiler.Errors)
	}

	return compiler, nil
}
