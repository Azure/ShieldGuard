package policy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/OneOfOne/xxhash"
	"github.com/open-policy-agent/opa/ast"
)

func regoCompilerKey(packages []Package, _ []RegoCompilerOptions) string {
	packageIDs := make([]string, 0, len(packages))
	for _, p := range packages {
		packageIDs = append(packageIDs, p.QualifiedID())
	}
	sort.Strings(packageIDs)

	k := xxhash.Checksum64([]byte(strings.Join(packageIDs, ",")))

	return fmt.Sprint(k)
}

// RegoCompilerOptions configs the RegoCompiler.
type RegoCompilerOptions struct{}

// NewRegoCompiler creates a compiler from policy packages.
func NewRegoCompiler(
	packages []Package,
	opts ...RegoCompilerOptions,
) (*ast.Compiler, string, error) {
	modules := map[string]*ast.Module{}
	for _, p := range packages {
		for name, m := range p.ParsedModules() {
			modules[name] = m
		}
	}

	compiler := ast.NewCompiler()
	compiler.Compile(modules)
	if compiler.Failed() {
		return nil, "", fmt.Errorf("failed to create compiler: %w", compiler.Errors)
	}

	compilerKey := regoCompilerKey(packages, opts)

	return compiler, compilerKey, nil
}
