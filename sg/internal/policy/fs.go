package policy

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
)

// FSPackage is a policy package loaded from the file system.
type FSPackage struct {
	packageSpec   PackageSpec
	rules         []Rule
	parsedModules map[string]*ast.Module
}

func loadPackageFromPath(path string) (Package, error) {
	rv := &FSPackage{}

	// load rules
	{
		policies, err := loader.AllRegos([]string{path})
		if err != nil {
			return nil, fmt.Errorf("failed to load policies: %w", err)
		}
		if len(policies.Modules) == 0 {
			return nil, fmt.Errorf("no policies found from path: %s", path)
		}

		rv.parsedModules = policies.ParsedModules()
		for _, module := range rv.parsedModules {
			rv.rules = append(rv.rules, loadRulesFromModule(module)...)
		}
	}

	// load package spec
	{
		projectSpec, err := loadPackageSpecFromDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load package spec: %w", err)
		}
		rv.packageSpec = projectSpec
	}

	return rv, nil
}

var _ Package = (*FSPackage)(nil)

func (p *FSPackage) Spec() PackageSpec {
	return p.packageSpec
}

func (p *FSPackage) Rules() []Rule {
	return p.rules
}

func (p *FSPackage) ParsedModules() map[string]*ast.Module {
	return p.parsedModules
}

// LoadPackagesFromPaths loads policy packages from the given paths.
func LoadPackagesFromPaths(paths []string) ([]Package, error) {
	var rv []Package

	for _, path := range paths {
		p, err := loadPackageFromPath(path)
		if err != nil {
			return nil, err
		}
		rv = append(rv, p)
	}

	return rv, nil
}
