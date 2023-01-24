package testsource

import "github.com/Azure/ShieldGuard/sg/internal/source"

type TestSource struct {
	NameFunc func() string

	ParsedConfigurationsFunc func() ([]interface{}, error)
}

var _ source.Source = (*TestSource)(nil)

func (s *TestSource) Name() string {
	return s.NameFunc()
}

func (s *TestSource) ParsedConfigurations() ([]interface{}, error) {
	return s.ParsedConfigurationsFunc()
}
