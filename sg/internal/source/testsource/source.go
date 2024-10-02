package testsource

import (
	"github.com/Azure/ShieldGuard/sg/internal/source"
	"github.com/open-policy-agent/opa/ast"
)

type TestSource struct {
	NameFunc func() string

	ParsedConfigurationsFunc func() ([]ast.Value, error)
}

var _ source.Source = (*TestSource)(nil)

func (s *TestSource) Name() string {
	return s.NameFunc()
}

func (s *TestSource) ParsedConfigurations() ([]ast.Value, error) {
	return s.ParsedConfigurationsFunc()
}
