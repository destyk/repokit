package config

import "fmt"

// ToolingConfig describes the shared tooling repository.
type ToolingConfig struct {
	Repository string        `yaml:"repository"`
	Ref        string        `yaml:"ref"`
	RefType    string        `yaml:"ref_type"`
	Files      []FileMapping `yaml:"files"`
	Make       []string      `yaml:"make"`
}

// FileMapping describes a single file (or glob) copied from the tooling
// repository into the target repository.
type FileMapping struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	Required    bool   `yaml:"required"`
}

// Validate validates tooling configuration.
// Defaults: RefType=tag.
func (t *ToolingConfig) Validate() error {
	if t.Repository == "" {
		return fmt.Errorf("repository is required")
	}

	if t.Ref == "" {
		return fmt.Errorf("ref is required")
	}

	if t.RefType == "" {
		t.RefType = "tag"
	}

	if t.RefType != "tag" && t.RefType != "branch" {
		return fmt.Errorf("ref_type must be tag or branch, got %q", t.RefType)
	}

	for i := range t.Files {
		if err := t.Files[i].Validate(); err != nil {
			return fmt.Errorf("files[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate validates a single file mapping.
func (m *FileMapping) Validate() error {
	if m.Source == "" {
		return fmt.Errorf("source is required")
	}

	if m.Destination == "" {
		return fmt.Errorf("destination is required")
	}

	if err := validateRelativePath(m.Destination, true); err != nil {
		return fmt.Errorf("destination: %w", err)
	}

	return nil
}
