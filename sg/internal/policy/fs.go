package policy

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
)

// FSPackage is a policy package loaded from the file system.
type FSPackage struct {
	rules         []Rule
	parsedModules map[string]*ast.Module
}

func loadPackageFromPaths(paths []string) (Package, error) {
	policies, err := loader.AllRegos(paths)
	if err != nil {
		return nil, fmt.Errorf("failed to load policies: %w", err)
	}
	if len(policies.Modules) == 0 {
		return nil, fmt.Errorf("no policies found from path: %s", paths)
	}

	rv := &FSPackage{
		parsedModules: policies.ParsedModules(),
	}
	for _, module := range rv.parsedModules {
		rv.rules = append(rv.rules, loadRulesFromModule(module)...)
	}

	return rv, nil
}

var _ Package = (*FSPackage)(nil)

func (p *FSPackage) Rules() []Rule {
	return p.rules
}

func (p *FSPackage) ParsedModules() map[string]*ast.Module {
	return p.parsedModules
}

// LoadPackagesFromPaths loads policy packages from the given paths.
func LoadPackagesFromPaths(paths []string) ([]Package, error) {
	p, err := loadPackageFromPaths(paths)
	if err != nil {
		return nil, err
	}

	return []Package{p}, nil
}
