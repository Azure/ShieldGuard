package project

// Spec defines the project specification.
type Spec struct {
	Files []FileTargetSpec `json:"files" yaml:"files"`
}

// FileTargetSpec defines the specification of a file target.
// Without further specification, paths are relative to the context root which is defined during execution.
type FileTargetSpec struct {
	// Name - name of the target.
	Name string `json:"name" yaml:"name"`
	// Paths - paths to the targets to check.
	Paths []string `json:"paths" yaml:"paths"`
	// Policies - paths to the policy to load.
	Policies []string `json:"policies" yaml:"policies"`
	// Data - paths to the (extra) data to load.
	Data []string `json:"data" yaml:"data"`
}
