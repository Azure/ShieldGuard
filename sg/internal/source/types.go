package source

// Source is the interface for a validation source.
type Source interface {
	// Name returns the name of the target.
	Name() string
	// ParsedConfigurations returns the parsed configurations of the target.
	ParsedConfigurations() ([]interface{}, error)
}
